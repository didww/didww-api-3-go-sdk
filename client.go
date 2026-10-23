package didww

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/didww/didww-api-3-go-sdk/v4/resource"
)

// Client is the DIDWW API client. It is safe for concurrent use.
//
// Every request is authenticated with a short-lived bearer token minted from
// the API key by the token service; the API key itself is sent to the token
// service only.
type Client struct {
	apiKey            string
	env               Environment
	baseURL           string
	authURL           string
	voiceInPricesURL  string
	voiceOutPricesURL string
	smsPricesURL      string
	apiVersion        string
	scope             Scope
	httpClient        *http.Client
	noRedirectClient  *http.Client
	tokenStore        TokenStore
	logger            *slog.Logger
	tokens            *tokenProvider
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithEnvironment sets the API environment, Sandbox or Production. The JSON:API
// base URL, the token service and the price hosts follow it unless overridden.
func WithEnvironment(env Environment) ClientOption {
	return func(c *Client) {
		c.env = env
		c.baseURL = string(env)
	}
}

// WithBaseURL sets a custom JSON:API base URL. The token service and the price
// hosts are not affected.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client. It is used as given: configure
// proxies, TLS and timeouts on it. Note that a total Timeout also bounds the
// download of large files.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithAPIVersion sets the API version requested for access tokens. It must be
// a YYYY-MM-DD date not earlier than DefaultAPIVersion; NewClient returns an
// error otherwise.
func WithAPIVersion(v string) ClientOption {
	return func(c *Client) {
		c.apiVersion = v
	}
}

// WithScope sets the scope requested for access tokens (ScopeRead or
// ScopeFull). Without it no scope is requested and the token service grants
// full access.
func WithScope(s Scope) ClientOption {
	return func(c *Client) {
		c.scope = s
	}
}

// WithTokenStore sets the store that caches access tokens, for example one
// shared by several processes. The store holds live bearer tokens and must be
// protected accordingly. The default is NewMemoryTokenStore.
func WithTokenStore(s TokenStore) ClientOption {
	return func(c *Client) {
		c.tokenStore = s
	}
}

// WithLogger sets the logger for token warnings and debug events. The default
// is slog.Default(); nil disables logging. Tokens and API keys are never logged.
func WithLogger(l *slog.Logger) ClientOption {
	return func(c *Client) {
		if l == nil {
			l = slog.New(discardHandler{})
		}
		c.logger = l
	}
}

// WithAuthURL sets a custom token service base URL.
func WithAuthURL(url string) ClientOption {
	return func(c *Client) {
		c.authURL = url
	}
}

// WithVoiceInPricesURL sets a custom Voice IN prices base URL (with or without /v3).
func WithVoiceInPricesURL(url string) ClientOption {
	return func(c *Client) {
		c.voiceInPricesURL = url
	}
}

// WithVoiceOutPricesURL sets a custom Voice OUT prices base URL (with or without /v3).
func WithVoiceOutPricesURL(url string) ClientOption {
	return func(c *Client) {
		c.voiceOutPricesURL = url
	}
}

// WithSMSPricesURL sets a custom SMS prices base URL (with or without /v3).
func WithSMSPricesURL(url string) ClientOption {
	return func(c *Client) {
		c.smsPricesURL = url
	}
}

// NewClient creates a new DIDWW API client.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if err := validateAPIKey(apiKey); err != nil {
		return nil, err
	}

	c := &Client{
		apiKey:     apiKey,
		env:        Sandbox,
		baseURL:    string(Sandbox),
		apiVersion: DefaultAPIVersion,
	}

	for _, opt := range opts {
		opt(c)
	}

	if err := c.validateOptions(); err != nil {
		return nil, err
	}
	if err := c.resolveEndpoints(); err != nil {
		return nil, err
	}
	if c.httpClient == nil {
		c.httpClient = defaultHTTPClient()
	}
	noRedirect := *c.httpClient
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c.noRedirectClient = &noRedirect
	if c.tokenStore == nil {
		c.tokenStore = NewMemoryTokenStore()
	}
	c.tokens = newTokenProvider(c)

	return c, nil
}

func validateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	if strings.ContainsFunc(apiKey, func(r rune) bool { return (r < 0x20 && r != '\t') || r == 0x7f }) {
		return fmt.Errorf("API key contains characters not allowed in an HTTP header")
	}
	return nil
}

// validateOptions checks the API version and the scope set by the options.
func (c *Client) validateOptions() error {
	if !apiVersionPattern.MatchString(c.apiVersion) {
		return fmt.Errorf("API version must be a YYYY-MM-DD date, got %q", c.apiVersion)
	}
	if c.apiVersion < DefaultAPIVersion {
		return fmt.Errorf("API version %s is not supported: this SDK requires %s or later; "+
			"use module github.com/didww/didww-api-3-go-sdk/v3 (branch release-3) for 2026-04-16", c.apiVersion, DefaultAPIVersion)
	}
	switch c.scope {
	case "", ScopeRead, ScopeFull:
		return nil
	default:
		return fmt.Errorf("scope must be %q or %q, got %q", ScopeRead, ScopeFull, c.scope)
	}
}

// resolveEndpoints fills the service URLs not set by an option from the
// environment and checks that each one is an http or https URL.
func (c *Client) resolveEndpoints() error {
	e, ok := environmentEndpoints[c.env]
	if !ok {
		return fmt.Errorf("environment must be didww.Sandbox or didww.Production, got %q; "+
			"use WithBaseURL, WithAuthURL and the prices URL options for other hosts", c.env)
	}
	if c.authURL == "" {
		c.authURL = e.auth
	}
	if c.voiceInPricesURL == "" {
		c.voiceInPricesURL = e.voiceIn
	}
	if c.voiceOutPricesURL == "" {
		c.voiceOutPricesURL = e.voiceOut
	}
	if c.smsPricesURL == "" {
		c.smsPricesURL = e.sms
	}
	for name, u := range map[string]string{
		"token service": c.authURL, "Voice IN prices": c.voiceInPricesURL,
		"Voice OUT prices": c.voiceOutPricesURL, "SMS prices": c.smsPricesURL,
	} {
		if !isHTTPURL(u) {
			return fmt.Errorf("%s URL must be an http or https URL, got %q", name, u)
		}
	}
	return nil
}

func isHTTPURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// defaultHTTPClient returns a client without a total timeout, so that large
// files can be downloaded, but with connect, TLS handshake and response header
// timeouts.
func defaultHTTPClient() *http.Client {
	return &http.Client{Transport: defaultTransport()}
}

// defaultTransport is built once, so that clients created without
// WithHTTPClient share one connection pool.
var defaultTransport = sync.OnceValue(func() http.RoundTripper {
	return newDefaultTransport(http.DefaultTransport)
})

// newDefaultTransport clones base with the SDK timeouts. When base is not an
// *http.Transport, for example a mock that replaced http.DefaultTransport, it
// returns nil, so that every request reads http.DefaultTransport as it is then.
func newDefaultTransport(base http.RoundTripper) http.RoundTripper {
	t, ok := base.(*http.Transport)
	if !ok {
		return nil
	}
	transport := t.Clone()
	transport.DialContext = (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.ResponseHeaderTimeout = 60 * time.Second
	return transport
}

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// APIKey returns the API key of the client.
func (c *Client) APIKey() string {
	return c.apiKey
}

// DownloadExport downloads an export file from the given URL and writes it to
// the destination writer unchanged (gzip-compressed CSV).
//
// Only a 200 answer is written; redirects are not followed. Any other status,
// including 3xx and 204, returns an *APIError. An error from dest is returned
// as is.
func (c *Client) DownloadExport(ctx context.Context, downloadURL string, dest io.Writer) error {
	resp, err := c.send(ctx, c.noRedirectClient, audienceAPI, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
		if err != nil {
			return nil, &ClientError{Message: fmt.Sprintf("failed to create download request: %v", err)}
		}
		return req, nil
	})
	if err != nil {
		return wrapSendError("download request failed", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return unexpectedStatusError(resp.StatusCode, body)
	}

	w := &writeRecorder{w: dest}
	if _, err := io.Copy(w, resp.Body); err != nil {
		if w.err != nil || errors.Is(err, io.ErrShortWrite) {
			return err
		}
		return &ClientError{Message: fmt.Sprintf("failed to read export data: %v", err)}
	}
	return nil
}

// writeRecorder remembers the error of the underlying writer, so that write
// failures can be told apart from read failures after io.Copy.
type writeRecorder struct {
	w   io.Writer
	err error
}

func (r *writeRecorder) Write(p []byte) (int, error) {
	n, err := r.w.Write(p)
	if err != nil {
		r.err = err
	}
	return n, err
}

// DownloadAndDecompressExport downloads a gzip-compressed export file (.csv.gz) and writes the decompressed CSV to dest.
func (c *Client) DownloadAndDecompressExport(ctx context.Context, downloadURL string, dest io.Writer) error {
	var buf bytes.Buffer
	if err := c.DownloadExport(ctx, downloadURL, &buf); err != nil {
		return err
	}

	gz, err := gzip.NewReader(&buf)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to create gzip reader: %v", err)}
	}
	defer gz.Close() //nolint:errcheck // best-effort close

	const maxDecompressedSize = 1 << 30 // 1 GB
	if _, err := io.Copy(dest, io.LimitReader(gz, maxDecompressedSize)); err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to decompress export data: %v", err)}
	}
	return nil
}

// --- Repository Accessors ---

func (c *Client) Balance() *BalanceRepository {
	return &BalanceRepository{client: c}
}

func (c *Client) Countries() *Repository[resource.Country] { return NewRepository[resource.Country](c) }
func (c *Client) Regions() *Repository[resource.Region]    { return NewRepository[resource.Region](c) }
func (c *Client) Cities() *Repository[resource.City]       { return NewRepository[resource.City](c) }
func (c *Client) Areas() *Repository[resource.Area]        { return NewRepository[resource.Area](c) }
func (c *Client) VoiceInTrunks() *Repository[resource.VoiceInTrunk] {
	return NewRepository[resource.VoiceInTrunk](c)
}
func (c *Client) VoiceInTrunkGroups() *Repository[resource.VoiceInTrunkGroup] {
	return NewRepository[resource.VoiceInTrunkGroup](c)
}
func (c *Client) VoiceOutTrunks() *Repository[resource.VoiceOutTrunk] {
	return NewRepository[resource.VoiceOutTrunk](c)
}
func (c *Client) DIDs() *Repository[resource.DID] { return NewRepository[resource.DID](c) }
func (c *Client) DIDGroups() *Repository[resource.DIDGroup] {
	return NewRepository[resource.DIDGroup](c)
}
func (c *Client) DIDReservations() *Repository[resource.DIDReservation] {
	return NewRepository[resource.DIDReservation](c)
}
func (c *Client) AvailableDIDs() *Repository[resource.AvailableDID] {
	return NewRepository[resource.AvailableDID](c)
}
func (c *Client) Orders() *Repository[resource.Order] { return NewRepository[resource.Order](c) }
func (c *Client) Identities() *Repository[resource.Identity] {
	return NewRepository[resource.Identity](c)
}
func (c *Client) Addresses() *Repository[resource.Address] { return NewRepository[resource.Address](c) }
func (c *Client) AddressVerifications() *Repository[resource.AddressVerification] {
	return NewRepository[resource.AddressVerification](c)
}
func (c *Client) Proofs() *Repository[resource.Proof] { return NewRepository[resource.Proof](c) }
func (c *Client) ProofTypes() *Repository[resource.ProofType] {
	return NewRepository[resource.ProofType](c)
}
func (c *Client) AddressRequirements() *Repository[resource.AddressRequirement] {
	return NewRepository[resource.AddressRequirement](c)
}
func (c *Client) AddressRequirementValidations() *Repository[resource.AddressRequirementValidation] {
	return NewRepository[resource.AddressRequirementValidation](c)
}
func (c *Client) Exports() *Repository[resource.Export] { return NewRepository[resource.Export](c) }
func (c *Client) CapacityPools() *Repository[resource.CapacityPool] {
	return NewRepository[resource.CapacityPool](c)
}
func (c *Client) SharedCapacityGroups() *Repository[resource.SharedCapacityGroup] {
	return NewRepository[resource.SharedCapacityGroup](c)
}
func (c *Client) PublicKeys() *Repository[resource.PublicKey] {
	return NewRepository[resource.PublicKey](c)
}
func (c *Client) EncryptedFiles() *Repository[resource.EncryptedFile] {
	return NewRepository[resource.EncryptedFile](c)
}
func (c *Client) SupportingDocumentTemplates() *Repository[resource.SupportingDocumentTemplate] {
	return NewRepository[resource.SupportingDocumentTemplate](c)
}
func (c *Client) PermanentSupportingDocuments() *Repository[resource.PermanentSupportingDocument] {
	return NewRepository[resource.PermanentSupportingDocument](c)
}
func (c *Client) NanpaPrefixes() *Repository[resource.NanpaPrefix] {
	return NewRepository[resource.NanpaPrefix](c)
}
func (c *Client) EmergencyCallingServices() *Repository[resource.EmergencyCallingService] {
	return NewRepository[resource.EmergencyCallingService](c)
}
func (c *Client) EmergencyVerifications() *Repository[resource.EmergencyVerification] {
	return NewRepository[resource.EmergencyVerification](c)
}
func (c *Client) EmergencyRequirementValidations() *Repository[resource.EmergencyRequirementValidation] {
	return NewRepository[resource.EmergencyRequirementValidation](c)
}
func (c *Client) EmergencyRequirements() *Repository[resource.EmergencyRequirement] {
	return NewRepository[resource.EmergencyRequirement](c)
}
func (c *Client) DIDHistory() *Repository[resource.DIDHistory] {
	return NewRepository[resource.DIDHistory](c)
}
func (c *Client) VoiceOutTrunkRegenerateCredentials() *Repository[resource.VoiceOutTrunkRegenerateCredential] {
	return NewRepository[resource.VoiceOutTrunkRegenerateCredential](c)
}
func (c *Client) SmsTrunks() *Repository[resource.SmsTrunk] {
	return NewRepository[resource.SmsTrunk](c)
}
func (c *Client) SmsTrunkGroups() *Repository[resource.SmsTrunkGroup] {
	return NewRepository[resource.SmsTrunkGroup](c)
}
func (c *Client) SmsTrunkRegenerateCredentials() *Repository[resource.SmsTrunkRegenerateCredential] {
	return NewRepository[resource.SmsTrunkRegenerateCredential](c)
}

// SenderIDVerifications supports List, Find and Update only.
func (c *Client) SenderIDVerifications() *Repository[resource.SenderIDVerification] {
	return NewRepository[resource.SenderIDVerification](c)
}
