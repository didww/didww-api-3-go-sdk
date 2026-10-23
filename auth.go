package didww

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	tokenCacheKeyPrefix = "didww:token:v1:" //nolint:gosec // cache key prefix, not a credential
	maxMintAttempts     = 4
	mintTimeout         = 10 * time.Second
	storeTimeout        = 10 * time.Second
	backoffBase         = 500 * time.Millisecond
	backoffCap          = 8 * time.Second
)

// send authenticates the request built by build for audience and sends it.
// audience "" sends no Authorization (public_keys).
//
// A 401 answer evicts the refused token, and the request is built again and
// sent once more with a newly minted token. Token errors (*AuthError) are
// returned as is; transport errors are returned unwrapped.
func (c *Client) send(ctx context.Context, hc *http.Client, audience string, build func() (*http.Request, error)) (*http.Response, error) {
	req, err := build()
	if err != nil {
		return nil, err
	}
	prepareRequest(req)
	if audience == "" {
		return hc.Do(req)
	}

	tok, err := c.tokens.token(ctx, audience)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	resp, err := hc.Do(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}

	drainAndClose(resp.Body)
	c.tokens.invalidate(ctx, audience, tok.AccessToken)
	if req, err = build(); err != nil {
		return nil, err
	}
	prepareRequest(req)
	if tok, err = c.tokens.token(ctx, audience); err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	resp, err = hc.Do(req)
	if err == nil && resp.StatusCode == http.StatusUnauthorized {
		c.tokens.invalidate(ctx, audience, tok.AccessToken)
	}
	return resp, err
}

// prepareRequest sets the headers common to every request. If-None-Match is
// never sent: the services would ignore If-Modified-Since next to it.
func prepareRequest(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Del("If-None-Match")
}

func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, maxErrorBodySize))
	_ = body.Close()
}

// tokenProvider mints access tokens per audience, caches them in the token
// store and lets concurrent callers share one mint.
type tokenProvider struct {
	authURL      string
	apiKey       string
	apiKeyDigest string
	apiVersion   string
	scope        Scope
	store        TokenStore
	hc           *http.Client
	logger       *slog.Logger // nil means slog.Default()

	sleep        func(ctx context.Context, d time.Duration) error
	random       func() float64
	now          func() time.Time
	mintTimeout  time.Duration
	storeTimeout time.Duration

	mu       sync.Mutex // guards inflight and storeLocks only, never held during store I/O
	inflight map[string]*flight
	// storeLocks keep the write of a minted token and the compare-and-delete
	// of invalidate atomic against each other per cache key within this
	// Client; clients or processes sharing a store are not serialized.
	storeLocks map[string]chan struct{}
}

// flight is a mint in progress; its result is shared by every waiter.
type flight struct {
	done chan struct{}
	tok  Token
	err  error
}

// rateLimits holds the 429 windows of the token service per auth URL and API
// key, shared by every client of the process.
var rateLimits = struct {
	sync.Mutex
	until map[string]time.Time
}{until: map[string]time.Time{}}

func newTokenProvider(c *Client) *tokenProvider {
	return &tokenProvider{
		authURL:      strings.TrimSuffix(c.authURL, "/"),
		apiKey:       c.apiKey,
		apiKeyDigest: sha256Hex(c.apiKey),
		apiVersion:   c.apiVersion,
		scope:        c.scope,
		store:        c.tokenStore,
		hc:           c.noRedirectClient,
		logger:       c.logger,
		sleep:        sleepContext,
		random:       rand.Float64,
		now:          time.Now,
		mintTimeout:  mintTimeout,
		storeTimeout: storeTimeout,
		inflight:     map[string]*flight{},
		storeLocks:   map[string]chan struct{}{},
	}
}

// token returns the cached token for audience or mints one. Concurrent callers
// for the same audience wait for a single mint, each within its own ctx.
func (p *tokenProvider) token(ctx context.Context, audience string) (Token, error) {
	if err := ctx.Err(); err != nil {
		return Token{}, err
	}
	key := p.cacheKey(audience)
	if t, ok := p.cached(ctx, key); ok {
		return t, nil
	}
	if err := ctx.Err(); err != nil {
		return Token{}, err
	}

	p.mu.Lock()
	f, ok := p.inflight[key]
	if !ok {
		f = &flight{done: make(chan struct{})}
		p.inflight[key] = f
		go p.fly(context.WithoutCancel(ctx), key, audience, f)
	}
	p.mu.Unlock()

	select {
	case <-f.done:
		return f.tok, f.err
	case <-ctx.Done():
		return Token{}, ctx.Err()
	}
}

// fly mints the token of a flight on its own goroutine. A panic on it, for
// example in the token store's Set, the logger or a custom transport, becomes
// the flight's error instead of crashing the process.
func (p *tokenProvider) fly(ctx context.Context, key, audience string, f *flight) {
	defer func() {
		if r := recover(); r != nil {
			f.tok, f.err = Token{}, &AuthError{Err: fmt.Errorf("panic while minting a token: %v", r)}
		}
		p.mu.Lock()
		delete(p.inflight, key)
		p.mu.Unlock()
		close(f.done)
	}()

	if t, ok := p.cachedWithin(ctx, key); ok {
		f.tok = t
		return
	}
	t, err := p.mint(ctx, audience)
	if err != nil {
		f.err = err
		return
	}
	p.storeToken(ctx, key, t)
	f.tok = t
}

// cachedWithin reads the store on a detached ctx, bounded by storeTimeout.
func (p *tokenProvider) cachedWithin(ctx context.Context, key string) (Token, bool) {
	ctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
	defer cancel()
	t, ok, err := p.store.Get(ctx, key)
	if err != nil {
		p.log().Warn("DIDWW token store read failed", "error", err)
		return Token{}, false
	}
	return t, ok && t.AccessToken != ""
}

func (p *tokenProvider) storeToken(ctx context.Context, key string, t Token) { //nolint:gocritic // Token is passed by value like TokenStore.Set
	ctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
	defer cancel()
	unlock, err := p.lockKey(ctx, key)
	if err != nil {
		p.log().Warn("DIDWW token store write skipped", "error", err)
		return
	}
	defer unlock()
	if err := p.store.Set(ctx, key, t, time.Duration(t.ExpiresIn)*time.Second); err != nil {
		p.log().Warn("DIDWW token store write failed", "error", err)
	}
}

// lockKey takes the store lock of key, giving up when ctx ends.
func (p *tokenProvider) lockKey(ctx context.Context, key string) (func(), error) {
	p.mu.Lock()
	lock, ok := p.storeLocks[key]
	if !ok {
		lock = make(chan struct{}, 1)
		p.storeLocks[key] = lock
	}
	p.mu.Unlock()
	select {
	case lock <- struct{}{}:
		return func() { <-lock }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// invalidate evicts the cached token for audience only while it is still the
// refused one, so a token another caller has just minted survives. It waits
// for the store lock within ctx and bounds the store calls by storeTimeout.
func (p *tokenProvider) invalidate(ctx context.Context, audience, accessToken string) {
	key := p.cacheKey(audience)
	unlock, err := p.lockKey(ctx, key)
	if err != nil {
		return
	}
	defer unlock()
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), p.storeTimeout)
	defer cancel()
	t, ok, err := p.store.Get(ctx, key)
	if err != nil {
		p.log().Warn("DIDWW token store read failed", "error", err)
		return
	}
	if !ok || t.AccessToken != accessToken {
		return
	}
	if err := p.store.Delete(ctx, key); err != nil {
		p.log().Warn("DIDWW token store delete failed", "error", err)
	}
}

func (p *tokenProvider) cached(ctx context.Context, key string) (Token, bool) {
	t, ok, err := p.store.Get(ctx, key)
	if err != nil {
		if ctx.Err() == nil {
			p.log().Warn("DIDWW token store read failed", "error", err)
		}
		return Token{}, false
	}
	return t, ok && t.AccessToken != ""
}

func (p *tokenProvider) cacheKey(audience string) string {
	scope := string(p.scope)
	if scope == "" {
		scope = "-"
	}
	return tokenCacheKeyPrefix + sha256Hex(strings.Join([]string{p.authURL, p.apiKeyDigest, audience, p.apiVersion, scope}, "\n"))
}

func (p *tokenProvider) rateLimitKey() string {
	return p.authURL + "\n" + p.apiKeyDigest
}

func (p *tokenProvider) log() *slog.Logger {
	if p.logger != nil {
		return p.logger
	}
	return slog.Default()
}

type tokenRequest struct {
	Audience   []string `json:"aud"`
	APIVersion string   `json:"api_version"`
	Scope      string   `json:"scope,omitempty"`
}

type tokenResponse struct {
	status int
	header http.Header
	body   []byte
}

func (p *tokenProvider) mint(ctx context.Context, audience string) (Token, error) {
	if err := p.rateLimited(); err != nil {
		return Token{}, err
	}

	body, err := json.Marshal(tokenRequest{Audience: []string{audience}, APIVersion: p.apiVersion, Scope: string(p.scope)})
	if err != nil {
		return Token{}, &AuthError{Err: err}
	}
	start := p.now()
	resp, err := p.post(ctx, body)
	if err != nil {
		return Token{}, err
	}
	t, err := p.tokenFrom(resp, audience)
	if err != nil {
		return Token{}, err
	}
	p.log().Debug("DIDWW token minted",
		"aud", t.Audience, "scope", t.Scope, "api_version", t.APIVersion,
		"expires_in", t.ExpiresIn, "duration", p.now().Sub(start))
	return t, nil
}

// post sends the token request, retrying 5xx answers and transport failures
// other than TLS errors with a full-jitter backoff.
func (p *tokenProvider) post(ctx context.Context, body []byte) (*tokenResponse, error) {
	if _, err := http.NewRequest(http.MethodPost, p.authURL+"/token", http.NoBody); err != nil {
		return nil, &AuthError{Err: err, kind: ErrAuthConnection}
	}
	for attempt := 1; ; attempt++ {
		resp, err := p.attempt(ctx, body)
		switch {
		case err == nil && (resp.status < 500 || attempt == maxMintAttempts):
			return resp, nil
		case err != nil && (isTLSError(err) || attempt == maxMintAttempts):
			return nil, &AuthError{Err: err, kind: ErrAuthConnection}
		}
		if sleepErr := p.sleep(ctx, p.backoff(attempt)); sleepErr != nil {
			return nil, &AuthError{Err: sleepErr, kind: ErrAuthConnection}
		}
		if err := p.rateLimited(); err != nil {
			return nil, err
		}
	}
}

// rateLimited returns ErrRateLimited while a 429 window of the token service
// is open, so that no request is sent inside it.
func (p *tokenProvider) rateLimited() *AuthError {
	remaining := p.blockedFor()
	if remaining <= 0 {
		return nil
	}
	return &AuthError{
		StatusCode: http.StatusTooManyRequests,
		Code:       "rate_limit_exceeded",
		RetryAfter: time.Duration(math.Ceil(remaining.Seconds())) * time.Second,
		kind:       ErrRateLimited,
	}
}

func (p *tokenProvider) attempt(ctx context.Context, body []byte) (*tokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, p.mintTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.authURL+"/token", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Api-Key", p.apiKey)
	req.Header.Set("User-Agent", userAgent)

	resp, err := p.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
	if err != nil {
		return nil, err
	}
	return &tokenResponse{status: resp.StatusCode, header: resp.Header, body: b}, nil
}

func (p *tokenProvider) backoff(attempt int) time.Duration {
	limit := backoffBase << attempt
	if limit > backoffCap {
		limit = backoffCap
	}
	return time.Duration(p.random() * float64(limit))
}

func (p *tokenProvider) tokenFrom(r *tokenResponse, audience string) (Token, error) {
	var fields map[string]any
	if err := json.Unmarshal(r.body, &fields); err != nil {
		fields = nil
	}
	code, _ := fields["error"].(string)
	description, _ := fields["error_description"].(string)
	authErr := &AuthError{StatusCode: r.status, Code: code, Description: description}

	switch {
	case r.status == http.StatusOK:
		var t Token
		if err := json.Unmarshal(r.body, &t); err != nil {
			t = Token{}
		}
		t.MintedAt = p.now()
		if err := p.verifyGrant(&t, audience); err != nil {
			return Token{}, err
		}
		return t, nil
	case r.status == http.StatusBadRequest:
		authErr.kind = ErrInvalidRequest
	case r.status == http.StatusUnauthorized:
		authErr.kind = ErrInvalidClient
	case r.status == http.StatusTooManyRequests:
		authErr.kind = ErrRateLimited
		authErr.RetryAfter = parseTokenRetryAfter(r.header.Get("Retry-After"))
		rateLimits.Lock()
		rateLimits.until[p.rateLimitKey()] = p.now().Add(authErr.RetryAfter)
		rateLimits.Unlock()
	case r.status >= 500 && r.status <= 599:
		authErr.kind = ErrAuthServer
	}
	return Token{}, authErr
}

func (p *tokenProvider) verifyGrant(t *Token, audience string) error {
	if t.AccessToken == "" || !slices.Contains(t.Audience, audience) {
		return &AuthError{
			StatusCode:  http.StatusOK,
			Description: fmt.Sprintf("granted audience %v, requested %s", t.Audience, audience),
			kind:        ErrInvalidGrant,
		}
	}
	if p.scope != "" && t.Scope != string(p.scope) {
		p.log().Warn("DIDWW token scope mismatch", "requested", string(p.scope), "granted", t.Scope)
	}
	if t.APIVersion != p.apiVersion {
		p.log().Warn("DIDWW token api_version mismatch", "requested", p.apiVersion, "granted", t.APIVersion)
	}
	return nil
}

func (p *tokenProvider) blockedFor() time.Duration {
	rateLimits.Lock()
	until, ok := rateLimits.until[p.rateLimitKey()]
	rateLimits.Unlock()
	if !ok {
		return 0
	}
	return until.Sub(p.now())
}

// parseTokenRetryAfter returns the Retry-After seconds, at least one second.
func parseTokenRetryAfter(h string) time.Duration {
	n, err := strconv.Atoi(strings.TrimSpace(h))
	if err != nil || n < 1 {
		n = 1
	}
	return time.Duration(n) * time.Second
}

// isTLSError reports whether err is a certificate or TLS protocol failure,
// including a TLS alert sent by the peer, which retrying cannot fix.
func isTLSError(err error) bool {
	var unknownAuthority x509.UnknownAuthorityError
	var hostname x509.HostnameError
	var invalid x509.CertificateInvalidError
	var verification *tls.CertificateVerificationError
	var recordHeader tls.RecordHeaderError
	var alert tls.AlertError
	return errors.As(err, &unknownAuthority) ||
		errors.As(err, &hostname) ||
		errors.As(err, &invalid) ||
		errors.As(err, &verification) ||
		errors.As(err, &recordHeader) ||
		errors.As(err, &alert) ||
		isRemoteAlert(err)
}

// isRemoteAlert reports whether any *net.OpError in err's chain carries a TLS
// alert from the peer; a proxy wraps it in an outer "proxyconnect" OpError.
func isRemoteAlert(err error) bool {
	for err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Op == "remote error" {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// discardHandler is a slog.Handler that drops every record.
type discardHandler struct{}

func (discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (h discardHandler) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h discardHandler) WithGroup(string) slog.Handler           { return h }
