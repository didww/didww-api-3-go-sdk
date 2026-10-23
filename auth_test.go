package didww

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenMintRequest(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, srv.URL)

	tok, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assert.Equal(t, "token-1", tok.AccessToken)
	assert.Equal(t, "Bearer", tok.TokenType)
	assert.Equal(t, 900, tok.ExpiresIn)
	assert.Equal(t, []string{audienceAPI}, tok.Audience)
	assert.Equal(t, "full", tok.Scope)
	assert.Equal(t, DefaultAPIVersion, tok.APIVersion)
	assert.False(t, tok.MintedAt.IsZero())

	require.Equal(t, 1, srv.mints())
	assertRequestJSON(t, srv.mintBody(0), "token/create_request.json")
	header := srv.mintHeader(0)
	assert.Equal(t, "test-api-key", header.Get("Api-Key"))
	assert.Equal(t, "application/json", header.Get("Content-Type"))
	assert.Equal(t, "application/json", header.Get("Accept"))
	assert.Equal(t, "didww-go-sdk/"+sdkVersion, header.Get("User-Agent"))
	assert.Empty(t, header.Get("Authorization"))
}

func TestTokenMintRequestWithScope(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, srv.URL, WithScope(ScopeRead))

	_, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assertRequestJSON(t, srv.mintBody(0), "token/create_request_read_scope.json")
}

func TestTokenMintRequestWithAPIVersion(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, srv.URL, WithAPIVersion("2099-01-01"))

	tok, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assert.Equal(t, "2099-01-01", tok.APIVersion)
	assert.JSONEq(t, `{"aud":["api.didww.com"],"api_version":"2099-01-01"}`, string(srv.mintBody(0)))
}

func TestTokenReused(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, srv.URL)

	first, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	second, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assert.Equal(t, first.AccessToken, second.AccessToken)
	assert.Equal(t, 1, srv.mints())
}

func TestTokenMintedPerAudience(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, srv.URL)

	api, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	vin, err := client.tokens.token(context.Background(), audienceVoiceIn)
	require.NoError(t, err)

	assert.NotEqual(t, api.AccessToken, vin.AccessToken)
	assert.Equal(t, []string{audienceVoiceIn}, vin.Audience)
	require.Equal(t, 2, srv.mints())
	assert.JSONEq(t, `{"aud":["api.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(0)))
	assert.JSONEq(t, `{"aud":["api-vin.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(1)))
}

func TestTokenSharedThroughStore(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	store := NewMemoryTokenStore()
	first := newTestClient(t, srv.URL, WithTokenStore(store))
	second := newTestClient(t, srv.URL, WithTokenStore(store))

	_, err := first.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	tok, err := second.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assert.Equal(t, "token-1", tok.AccessToken)
	assert.Equal(t, 1, srv.mints())
}

type recordingTokenStore struct {
	TokenStore
	mu   sync.Mutex
	keys []string
	ttls []time.Duration
}

func (s *recordingTokenStore) Set(ctx context.Context, key string, t Token, ttl time.Duration) error {
	s.mu.Lock()
	s.keys = append(s.keys, key)
	s.ttls = append(s.ttls, ttl)
	s.mu.Unlock()
	return s.TokenStore.Set(ctx, key, t, ttl)
}

func TestTokenCachedForGrantedLifetime(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"t","token_type":"Bearer","expires_in":300,"aud":["api.didww.com"],"scope":"full","api_version":"2026-10-22"}`))
	}, nil)
	store := &recordingTokenStore{TokenStore: NewMemoryTokenStore()}
	client := newTestClient(t, srv.URL, WithTokenStore(store))

	_, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	require.Len(t, store.keys, 1)
	assert.True(t, strings.HasPrefix(store.keys[0], "didww:token:v1:"))
	assert.Equal(t, 300*time.Second, store.ttls[0])
}

func TestTokenCacheKey(t *testing.T) {
	newProvider := func(t *testing.T, opts ...ClientOption) *tokenProvider {
		t.Helper()
		opts = append([]ClientOption{WithAuthURL("https://auth.example.com")}, opts...)
		c, err := NewClient("key-1", opts...)
		require.NoError(t, err)
		return c.tokens
	}
	base := newProvider(t).cacheKey(audienceAPI)

	assert.Regexp(t, `^didww:token:v1:[0-9a-f]{64}$`, base)
	assert.NotContains(t, base, "key-1")
	assert.Equal(t, base, newProvider(t).cacheKey(audienceAPI))
	assert.NotEqual(t, base, newProvider(t).cacheKey(audienceSMS))
	assert.NotEqual(t, base, newProvider(t, WithAPIVersion("2099-01-01")).cacheKey(audienceAPI))
	assert.NotEqual(t, base, newProvider(t, WithScope(ScopeRead)).cacheKey(audienceAPI))
	assert.NotEqual(t, base, newProvider(t, WithAuthURL("https://auth-sandbox.example.com")).cacheKey(audienceAPI))

	other, err := NewClient("key-2", WithAuthURL("https://auth.example.com"))
	require.NoError(t, err)
	assert.NotEqual(t, base, other.tokens.cacheKey(audienceAPI))
}

func TestTokenSingleFlight(t *testing.T) {
	srv := newAuthTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		writeTestToken(w, r, fmt.Sprintf("token-%d", n))
	}, nil)
	client := newTestClient(t, srv.URL)

	const callers = 10
	tokens := make([]string, callers)
	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tok, err := client.tokens.token(context.Background(), audienceAPI)
			tokens[i], errs[i] = tok.AccessToken, err
		}(i)
	}
	wg.Wait()

	for i := 0; i < callers; i++ {
		require.NoError(t, errs[i])
		assert.Equal(t, "token-1", tokens[i])
	}
	assert.Equal(t, 1, srv.mints())
}

func TestTokenWaiterContextDoesNotCancelMint(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	srv := newAuthTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		<-release
		writeTestToken(w, r, fmt.Sprintf("token-%d", n))
	}, nil)
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	client := newTestClient(t, srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := client.tokens.token(ctx, audienceAPI)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	releaseOnce.Do(func() { close(release) })
	tok, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	assert.Equal(t, "token-1", tok.AccessToken)
	assert.Equal(t, 1, srv.mints())
}

// blockingSetTokenStore blocks every Set until release is closed and closes
// entered when the first Set starts.
type blockingSetTokenStore struct {
	TokenStore
	entered chan struct{}
	once    sync.Once
	release chan struct{}
}

func (s *blockingSetTokenStore) Set(ctx context.Context, key string, t Token, ttl time.Duration) error {
	s.once.Do(func() { close(s.entered) })
	<-s.release
	return s.TokenStore.Set(ctx, key, t, ttl)
}

func TestTokenWaiterNotBlockedByStoreWrite(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	store := &blockingSetTokenStore{TokenStore: NewMemoryTokenStore(), entered: make(chan struct{}), release: make(chan struct{})}
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(store.release) }) })
	client := newTestClient(t, srv.URL, WithTokenStore(store))

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, _ = client.tokens.token(ctx, audienceAPI)
	}()
	select {
	case <-store.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("the token store write did not start")
	}

	result := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_, err := client.tokens.token(ctx, audienceVoiceIn)
		result <- err
	}()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(5 * time.Second):
		t.Fatal("token() did not return within its context while a store write was blocked")
	}
}

// stallingTokenStore blocks every Get until its ctx ends while stalled is set.
type stallingTokenStore struct {
	TokenStore
	stalled atomic.Bool
}

func (s *stallingTokenStore) Get(ctx context.Context, key string) (Token, bool, error) {
	if s.stalled.Load() {
		<-ctx.Done()
		return Token{}, false, ctx.Err()
	}
	return s.TokenStore.Get(ctx, key)
}

func TestTokenInvalidateBoundsStalledStore(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	store := &stallingTokenStore{TokenStore: NewMemoryTokenStore()}
	client := newTestClient(t, srv.URL, WithTokenStore(store))
	client.tokens.storeTimeout = 50 * time.Millisecond
	tok, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	store.stalled.Store(true)
	done := make(chan struct{})
	go func() {
		client.tokens.invalidate(context.Background(), audienceAPI, tok.AccessToken)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("invalidate did not give up on a stalled store")
	}

	store.stalled.Store(false)
	client.tokens.invalidate(context.Background(), audienceAPI, tok.AccessToken)
	_, err = client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	assert.Equal(t, 2, srv.mints())
}

// stallingCallsTokenStore blocks the Get calls listed in stallGets (1-based)
// and every Set until their ctx ends.
type stallingCallsTokenStore struct {
	TokenStore
	gets      atomic.Int32
	stallGets map[int32]bool
	stallSet  bool
}

func (s *stallingCallsTokenStore) Get(ctx context.Context, key string) (Token, bool, error) {
	if s.stallGets[s.gets.Add(1)] {
		<-ctx.Done()
		return Token{}, false, ctx.Err()
	}
	return s.TokenStore.Get(ctx, key)
}

func (s *stallingCallsTokenStore) Set(ctx context.Context, key string, t Token, ttl time.Duration) error { //nolint:gocritic // matches TokenStore
	if s.stallSet {
		<-ctx.Done()
		return ctx.Err()
	}
	return s.TokenStore.Set(ctx, key, t, ttl)
}

func TestTokenFlightBoundsStalledStore(t *testing.T) {
	for name, store := range map[string]*stallingCallsTokenStore{
		"re-check read": {TokenStore: NewMemoryTokenStore(), stallGets: map[int32]bool{2: true}},
		"write":         {TokenStore: NewMemoryTokenStore(), stallSet: true},
	} {
		t.Run(name, func(t *testing.T) {
			srv := newAuthTestServer(t, nil, nil)
			client := newTestClient(t, srv.URL, WithTokenStore(store))
			client.tokens.storeTimeout = 50 * time.Millisecond

			result := make(chan error, 1)
			go func() {
				_, err := client.tokens.token(context.Background(), audienceAPI)
				result <- err
			}()
			select {
			case err := <-result:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("the mint did not give up on a stalled store")
			}
			assert.Equal(t, 1, srv.mints())
		})
	}
}

func TestTokenInvalidateWaitsForStoreLockWithinContext(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, srv.URL)
	unlock, err := client.tokens.lockKey(context.Background(), client.tokens.cacheKey(audienceAPI))
	require.NoError(t, err)
	t.Cleanup(unlock)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		client.tokens.invalidate(ctx, audienceAPI, "refused")
		client.tokens.invalidate(context.Background(), audienceVoiceIn, "refused")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("invalidate ignored its context or another audience's lock")
	}
}

type panickingTokenStore struct {
	TokenStore
}

func (panickingTokenStore) Set(context.Context, string, Token, time.Duration) error {
	panic("store is broken")
}

type ctxAwareTokenStore struct {
	TokenStore
	gets  *atomic.Int32
	onGet func()
}

func (s ctxAwareTokenStore) Get(ctx context.Context, _ string) (Token, bool, error) {
	s.gets.Add(1)
	if s.onGet != nil {
		s.onGet()
	}
	if err := ctx.Err(); err != nil {
		return Token{}, false, err
	}
	return Token{}, false, nil
}

func TestTokenCanceledContextIsNotLoggedAsStoreFailure(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	logger, logs := newCapturingLogger(slog.LevelWarn)
	store := ctxAwareTokenStore{TokenStore: NewMemoryTokenStore(), gets: &atomic.Int32{}}
	client := newTestClient(t, srv.URL, WithTokenStore(store), WithLogger(logger))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.tokens.token(ctx, audienceAPI)

	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, logs.String())
	assert.Equal(t, int32(0), store.gets.Load())
	assert.Equal(t, 0, srv.mints())
}

func TestTokenContextEndingDuringStoreReadIsNotLogged(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	logger, logs := newCapturingLogger(slog.LevelWarn)
	ctx, cancel := context.WithCancel(context.Background())
	store := ctxAwareTokenStore{TokenStore: NewMemoryTokenStore(), gets: &atomic.Int32{}, onGet: cancel}
	client := newTestClient(t, srv.URL, WithTokenStore(store), WithLogger(logger))

	_, err := client.tokens.token(ctx, audienceAPI)

	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, logs.String())
	assert.Equal(t, int32(1), store.gets.Load())
	assert.Equal(t, 0, srv.mints())
}

func TestTokenStorePanicReturnedAsError(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, srv.URL, WithTokenStore(panickingTokenStore{TokenStore: NewMemoryTokenStore()}))

	for i := 0; i < 2; i++ {
		tok, err := client.tokens.token(context.Background(), audienceAPI)

		var authErr *AuthError
		require.ErrorAs(t, err, &authErr)
		assert.Contains(t, err.Error(), "store is broken")
		assert.Empty(t, tok.AccessToken)
	}
	assert.Equal(t, 2, srv.mints())
}

func TestTokenInvalidClient(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeTokenError(t, w, http.StatusUnauthorized, "token/invalid_client.json")
	}, nil)
	client := newTestClient(t, srv.URL)
	sleeps := stubBackoff(client, 1)

	_, err := client.tokens.token(context.Background(), audienceAPI)

	require.ErrorIs(t, err, ErrInvalidClient)
	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, http.StatusUnauthorized, authErr.StatusCode)
	assert.Equal(t, "invalid_client", authErr.Code)
	assert.Equal(t, "Client authentication failed.", authErr.Description)
	assert.Equal(t, "didww: token service: invalid_client: Client authentication failed.", err.Error())
	assert.NotErrorIs(t, err, ErrInvalidRequest)
	assert.Equal(t, 1, srv.mints())
	assert.Empty(t, *sleeps)
}

func TestTokenInvalidRequest(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeTokenError(t, w, http.StatusBadRequest, "token/invalid_scope.json")
	}, nil)
	client := newTestClient(t, srv.URL)

	_, err := client.tokens.token(context.Background(), audienceAPI)

	require.ErrorIs(t, err, ErrInvalidRequest)
	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, http.StatusBadRequest, authErr.StatusCode)
	assert.Equal(t, "invalid_scope", authErr.Code)
	assert.Equal(t, 1, srv.mints())
}

func TestTokenRateLimited(t *testing.T) {
	srv := newAuthTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n == 1 {
			w.Header().Set("Retry-After", "30")
			writeTokenError(t, w, http.StatusTooManyRequests, "token/rate_limit_exceeded.json")
			return
		}
		writeTestToken(w, r, fmt.Sprintf("token-%d", n))
	}, nil)
	client := newTestClient(t, srv.URL)
	forgetRateLimit(t, client)
	clock := newFakeClock()
	client.tokens.now = clock.Now
	sleeps := stubBackoff(client, 1)

	_, err := client.tokens.token(context.Background(), audienceAPI)
	require.ErrorIs(t, err, ErrRateLimited)
	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, http.StatusTooManyRequests, authErr.StatusCode)
	assert.Equal(t, 30*time.Second, authErr.RetryAfter)

	clock.Advance(29*time.Second + 500*time.Millisecond)
	_, err = client.tokens.token(context.Background(), audienceAPI)
	require.ErrorIs(t, err, ErrRateLimited)
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, time.Second, authErr.RetryAfter)

	_, err = client.tokens.token(context.Background(), audienceVoiceIn)
	require.ErrorIs(t, err, ErrRateLimited)
	assert.Equal(t, 1, srv.mints())

	clock.Advance(500 * time.Millisecond)
	tok, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	assert.Equal(t, "token-2", tok.AccessToken)
	assert.Equal(t, 2, srv.mints())
	assert.Empty(t, *sleeps)
}

func TestTokenRateLimitedWithoutRetryAfter(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "Wed, 21 Oct 2026 07:28:00 GMT")
		writeTokenError(t, w, http.StatusTooManyRequests, "token/rate_limit_exceeded.json")
	}, nil)
	client := newTestClient(t, srv.URL)
	forgetRateLimit(t, client)

	_, err := client.tokens.token(context.Background(), audienceAPI)

	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.ErrorIs(t, err, ErrRateLimited)
	assert.Equal(t, time.Second, authErr.RetryAfter)
}

func TestTokenServerErrorRetriedWithBackoff(t *testing.T) {
	for _, tc := range []struct {
		jitter float64
		sleeps []time.Duration
	}{
		{jitter: 1, sleeps: []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}},
		{jitter: 0.5, sleeps: []time.Duration{500 * time.Millisecond, time.Second, 2 * time.Second}},
	} {
		t.Run(fmt.Sprintf("jitter %v", tc.jitter), func(t *testing.T) {
			srv := newAuthTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
				if n < 4 {
					writeTokenError(t, w, http.StatusServiceUnavailable, "token/server_error.json")
					return
				}
				writeTestToken(w, r, "token-ok")
			}, nil)
			client := newTestClient(t, srv.URL)
			sleeps := stubBackoff(client, tc.jitter)

			tok, err := client.tokens.token(context.Background(), audienceAPI)
			require.NoError(t, err)

			assert.Equal(t, "token-ok", tok.AccessToken)
			assert.Equal(t, 4, srv.mints())
			assert.Equal(t, tc.sleeps, *sleeps)
		})
	}
}

func TestTokenBackoffCap(t *testing.T) {
	client := newTestClient(t, "http://127.0.0.1")
	client.tokens.random = func() float64 { return 1 }

	assert.Equal(t, time.Second, client.tokens.backoff(1))
	assert.Equal(t, 4*time.Second, client.tokens.backoff(3))
	assert.Equal(t, 8*time.Second, client.tokens.backoff(4))
	assert.Equal(t, 8*time.Second, client.tokens.backoff(5))
}

func TestTokenServerErrorExhausted(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeTokenError(t, w, http.StatusInternalServerError, "token/server_error.json")
	}, nil)
	client := newTestClient(t, srv.URL)
	sleeps := stubBackoff(client, 1)

	_, err := client.tokens.token(context.Background(), audienceAPI)

	require.ErrorIs(t, err, ErrAuthServer)
	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, http.StatusInternalServerError, authErr.StatusCode)
	assert.Equal(t, "server_error", authErr.Code)
	assert.Equal(t, maxMintAttempts, srv.mints())
	assert.Len(t, *sleeps, maxMintAttempts-1)
}

func TestTokenServerErrorWithoutJSONBody(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("<html>Bad Gateway</html>"))
	}, nil)
	client := newTestClient(t, srv.URL)
	stubBackoff(client, 1)

	_, err := client.tokens.token(context.Background(), audienceAPI)

	require.ErrorIs(t, err, ErrAuthServer)
	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, http.StatusBadGateway, authErr.StatusCode)
	assert.Empty(t, authErr.Code)
	assert.Equal(t, "didww: token service responded with status 502", err.Error())
}

func TestTokenUnexpectedStatus(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}, nil)
	client := newTestClient(t, srv.URL)
	sleeps := stubBackoff(client, 1)

	_, err := client.tokens.token(context.Background(), audienceAPI)

	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, http.StatusForbidden, authErr.StatusCode)
	for _, sentinel := range []error{ErrInvalidClient, ErrInvalidRequest, ErrRateLimited, ErrAuthServer, ErrAuthConnection, ErrInvalidGrant} {
		assert.NotErrorIs(t, err, sentinel)
	}
	assert.Equal(t, 1, srv.mints())
	assert.Empty(t, *sleeps)
}

func TestTokenLaterCallMintsAgainAfterFailure(t *testing.T) {
	srv := newAuthTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n <= maxMintAttempts {
			writeTokenError(t, w, http.StatusInternalServerError, "token/server_error.json")
			return
		}
		writeTestToken(w, r, "token-ok")
	}, nil)
	client := newTestClient(t, srv.URL)
	stubBackoff(client, 1)

	_, err := client.tokens.token(context.Background(), audienceAPI)
	require.ErrorIs(t, err, ErrAuthServer)

	tok, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	assert.Equal(t, "token-ok", tok.AccessToken)
	assert.Equal(t, maxMintAttempts+1, srv.mints())
}

func TestTokenConnectionError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	url := "http://" + listener.Addr().String()
	require.NoError(t, listener.Close())

	client := newTestClient(t, url)
	sleeps := stubBackoff(client, 1)

	_, err = client.tokens.token(context.Background(), audienceAPI)

	require.ErrorIs(t, err, ErrAuthConnection)
	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, 0, authErr.StatusCode)
	require.Error(t, authErr.Unwrap())
	assert.Len(t, *sleeps, maxMintAttempts-1)
}

func TestTokenTLSErrorNotRetried(t *testing.T) {
	var connections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeTestToken(w, r, "unreachable")
	}))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	client := newTestClient(t, server.URL)
	sleeps := stubBackoff(client, 1)

	_, err := client.tokens.token(context.Background(), audienceAPI)

	require.ErrorIs(t, err, ErrAuthConnection)
	assert.True(t, isTLSError(err))
	assert.Empty(t, *sleeps)
	assert.Equal(t, int32(1), connections.Load())
}

func TestTokenTLSAlertNotRetried(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() })
	var connections atomic.Int32
	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}
			connections.Add(1)
			go func() {
				defer conn.Close()
				// A fatal handshake_failure alert, then wait for the client to
				// close so that the alert is not lost to a reset.
				conn.Write([]byte{0x15, 0x03, 0x03, 0x00, 0x02, 0x02, 0x28})
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				io.Copy(io.Discard, conn)
			}()
		}
	}()

	client := newTestClient(t, "https://"+listener.Addr().String())
	sleeps := stubBackoff(client, 1)

	_, err = client.tokens.token(context.Background(), audienceAPI)

	require.ErrorIs(t, err, ErrAuthConnection)
	assert.Contains(t, err.Error(), "remote error")
	assert.True(t, isTLSError(err))
	assert.Empty(t, *sleeps)
	assert.Equal(t, int32(1), connections.Load())
}

func TestTokenInvalidGrant(t *testing.T) {
	for name, body := range map[string]string{
		"audience mismatch": `{"access_token":"t","token_type":"Bearer","expires_in":900,"aud":["api-sms.didww.com"],"scope":"full","api_version":"2026-10-22"}`,
		"empty token":       `{"access_token":"","token_type":"Bearer","expires_in":900,"aud":["api.didww.com"],"scope":"full","api_version":"2026-10-22"}`,
		"not JSON":          `ok`,
	} {
		t.Run(name, func(t *testing.T) {
			srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(body))
			}, nil)
			client := newTestClient(t, srv.URL)

			_, err := client.tokens.token(context.Background(), audienceAPI)
			require.ErrorIs(t, err, ErrInvalidGrant)
			var authErr *AuthError
			require.ErrorAs(t, err, &authErr)
			assert.Equal(t, http.StatusOK, authErr.StatusCode)

			_, err = client.tokens.token(context.Background(), audienceAPI)
			require.ErrorIs(t, err, ErrInvalidGrant)
			assert.Equal(t, 2, srv.mints())
		})
	}
}

func newCapturingLogger(level slog.Level) (*slog.Logger, *syncBuffer) {
	buf := &syncBuffer{}
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: level})), buf
}

func TestTokenScopeMismatchWarning(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	logger, logs := newCapturingLogger(slog.LevelWarn)
	client := newTestClient(t, srv.URL, WithScope(ScopeRead), WithLogger(logger))

	tok, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assert.Equal(t, "full", tok.Scope)
	assert.Contains(t, logs.String(), `msg="DIDWW token scope mismatch" requested=read granted=full`)
	assert.NotContains(t, logs.String(), "token-1")
}

func TestTokenAPIVersionMismatchWarning(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"t","token_type":"Bearer","expires_in":900,"aud":["api.didww.com"],"scope":"full","api_version":"2099-01-01"}`))
	}, nil)
	logger, logs := newCapturingLogger(slog.LevelWarn)
	client := newTestClient(t, srv.URL, WithLogger(logger))

	tok, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assert.Equal(t, "t", tok.AccessToken)
	assert.Contains(t, logs.String(), `msg="DIDWW token api_version mismatch" requested=2026-10-22 granted=2099-01-01`)
}

func TestTokenNoWarningWhenGrantMatches(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	logger, logs := newCapturingLogger(slog.LevelWarn)
	client := newTestClient(t, srv.URL, WithScope(ScopeFull), WithLogger(logger))

	_, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assert.Empty(t, logs.String())
}

func TestTokenMintedDebugLog(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	logger, logs := newCapturingLogger(slog.LevelDebug)
	client := newTestClient(t, srv.URL, WithLogger(logger))

	_, err := client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)

	assert.Contains(t, logs.String(), `msg="DIDWW token minted" aud=[api.didww.com] scope=full api_version=2026-10-22 expires_in=900`)
	assert.NotContains(t, logs.String(), "token-1")
	assert.NotContains(t, logs.String(), "test-api-key")
}

func TestTokenInvalidate(t *testing.T) {
	t.Run("evicts the refused token", func(t *testing.T) {
		srv := newAuthTestServer(t, nil, nil)
		client := newTestClient(t, srv.URL)
		_, err := client.tokens.token(context.Background(), audienceAPI)
		require.NoError(t, err)

		client.tokens.invalidate(context.Background(), audienceAPI, "token-1")

		tok, err := client.tokens.token(context.Background(), audienceAPI)
		require.NoError(t, err)
		assert.Equal(t, "token-2", tok.AccessToken)
	})

	t.Run("keeps a token minted after the refused one", func(t *testing.T) {
		srv := newAuthTestServer(t, nil, nil)
		client := newTestClient(t, srv.URL)
		_, err := client.tokens.token(context.Background(), audienceAPI)
		require.NoError(t, err)

		client.tokens.invalidate(context.Background(), audienceAPI, "stale-token")

		tok, err := client.tokens.token(context.Background(), audienceAPI)
		require.NoError(t, err)
		assert.Equal(t, "token-1", tok.AccessToken)
		assert.Equal(t, 1, srv.mints())
	})
}

func TestAuthErrorMessage(t *testing.T) {
	transport := errors.New("connection refused")
	tests := []struct {
		err  *AuthError
		want string
	}{
		{&AuthError{StatusCode: 400, Code: "invalid_request", Description: "aud is missing"}, "didww: token service: invalid_request: aud is missing"},
		{&AuthError{StatusCode: 429, Code: "rate_limit_exceeded"}, "didww: token service: rate_limit_exceeded"},
		{&AuthError{StatusCode: 200, Description: "granted audience [], requested api.didww.com"}, "didww: token service: granted audience [], requested api.didww.com"},
		{&AuthError{Err: transport}, "didww: token service: connection refused"},
		{&AuthError{StatusCode: 503}, "didww: token service responded with status 503"},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, tc.err.Error())
	}
	assert.ErrorIs(t, &AuthError{Err: transport, kind: ErrAuthConnection}, transport)
}

func TestIsTLSErrorFindsRemoteAlertBehindProxy(t *testing.T) {
	alert := &net.OpError{Op: "remote error", Err: errors.New("tls: handshake failure")}

	assert.True(t, isTLSError(&net.OpError{Op: "proxyconnect", Err: alert}))
	assert.True(t, isTLSError(fmt.Errorf("post: %w", alert)))
	assert.False(t, isTLSError(&net.OpError{Op: "dial", Err: errors.New("connection refused")}))
}

func TestTokenMintDoesNotFollowRedirects(t *testing.T) {
	var redirected atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		redirected.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(other.Close)
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", other.URL+"/token")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}, nil)
	client := newTestClient(t, srv.URL)

	_, err := client.tokens.token(context.Background(), audienceAPI)

	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Equal(t, http.StatusTemporaryRedirect, authErr.StatusCode)
	assert.Equal(t, 1, srv.mints())
	assert.Equal(t, int32(0), redirected.Load())
}

func TestTokenMintTimeoutRetried(t *testing.T) {
	hang := make(chan struct{})
	srv := newAuthTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n == 1 {
			select {
			case <-r.Context().Done():
			case <-hang:
			}
			return
		}
		writeTestToken(w, r, "token-2")
	}, nil)
	// Registered after the server, so it runs first and frees a hung handler.
	t.Cleanup(func() { close(hang) })
	client := newTestClient(t, srv.URL)
	client.tokens.mintTimeout = 500 * time.Millisecond
	sleeps := stubBackoff(client, 1)

	tok, err := tokenWithin(t, client, 5*time.Second)

	require.NoError(t, err)
	assert.Equal(t, "token-2", tok.AccessToken)
	assert.Equal(t, 2, srv.mints())
	assert.Len(t, *sleeps, 1)
}

func TestTokenMintTimeoutExhaustsAttempts(t *testing.T) {
	hang := make(chan struct{})
	srv := newAuthTestServer(t, func(_ int, _ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-hang:
		}
	}, nil)
	// Registered after the server, so it runs first and frees a hung handler.
	t.Cleanup(func() { close(hang) })
	client := newTestClient(t, srv.URL)
	client.tokens.mintTimeout = 250 * time.Millisecond
	stubBackoff(client, 1)

	_, err := tokenWithin(t, client, 5*time.Second)

	require.ErrorIs(t, err, ErrAuthConnection)
	assert.Equal(t, maxMintAttempts, srv.mints())
}

// tokenWithin mints a token for the API audience and fails the test when
// that does not finish within limit.
func tokenWithin(t *testing.T, client *Client, limit time.Duration) (Token, error) {
	t.Helper()
	type result struct {
		tok Token
		err error
	}
	done := make(chan result, 1)
	go func() {
		tok, err := client.tokens.token(context.Background(), audienceAPI)
		done <- result{tok, err}
	}()
	select {
	case r := <-done:
		return r.tok, r.err
	case <-time.After(limit):
		t.Fatal("the mint did not finish in time")
		return Token{}, nil
	}
}

// seedingTokenStore misses on its first Get and, as a side effect, stores the
// seed token, as another process sharing the store would.
type seedingTokenStore struct {
	TokenStore
	seed  Token
	seeds atomic.Bool
}

func (s *seedingTokenStore) Get(ctx context.Context, key string) (Token, bool, error) {
	if s.seeds.CompareAndSwap(false, true) {
		_ = s.TokenStore.Set(ctx, key, s.seed, time.Minute)
		return Token{}, false, nil
	}
	return s.TokenStore.Get(ctx, key)
}

func TestTokenFlightRereadsStoreBeforeMinting(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	store := &seedingTokenStore{TokenStore: NewMemoryTokenStore(), seed: Token{AccessToken: "shared-token", ExpiresIn: 60}}
	client := newTestClient(t, srv.URL, WithTokenStore(store))

	tok, err := client.tokens.token(context.Background(), audienceAPI)

	require.NoError(t, err)
	assert.Equal(t, "shared-token", tok.AccessToken)
	assert.Equal(t, 0, srv.mints())
}

func TestTokenRetryStopsInsideRateLimitWindow(t *testing.T) {
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}, nil)
	client := newTestClient(t, srv.URL)
	forgetRateLimit(t, client)
	client.tokens.sleep = func(context.Context, time.Duration) error {
		// Another audience's mint is answered 429 while this one backs off.
		rateLimits.Lock()
		rateLimits.until[client.tokens.rateLimitKey()] = time.Now().Add(30 * time.Second)
		rateLimits.Unlock()
		return nil
	}

	_, err := client.tokens.token(context.Background(), audienceAPI)

	require.ErrorIs(t, err, ErrRateLimited)
	assert.Equal(t, 1, srv.mints())
}

func TestTokenNoScopeWarningWithoutRequestedScope(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	logger, logs := newCapturingLogger(slog.LevelWarn)
	client := newTestClient(t, srv.URL, WithLogger(logger))

	tok, err := client.tokens.token(context.Background(), audienceAPI)

	require.NoError(t, err)
	assert.Equal(t, "full", tok.Scope)
	assert.Empty(t, logs.String())
}

func TestTokenWarningsGoToDefaultLogger(t *testing.T) {
	srv := newAuthTestServer(t, nil, nil)
	logger, logs := newCapturingLogger(slog.LevelWarn)
	previous, previousOutput, previousFlags := slog.Default(), log.Writer(), log.Flags()
	slog.SetDefault(logger)
	t.Cleanup(func() {
		slog.SetDefault(previous)
		log.SetOutput(previousOutput)
		log.SetFlags(previousFlags)
	})

	client, err := NewClient("test-api-key", WithBaseURL(srv.URL), WithAuthURL(srv.URL), WithScope(ScopeRead))
	require.NoError(t, err)
	_, err = client.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	assert.Contains(t, logs.String(), "DIDWW token scope mismatch")

	logs.Reset()
	silent, err := NewClient("test-api-key", WithBaseURL(srv.URL), WithAuthURL(srv.URL), WithScope(ScopeRead), WithLogger(nil))
	require.NoError(t, err)
	_, err = silent.tokens.token(context.Background(), audienceAPI)
	require.NoError(t, err)
	assert.Empty(t, logs.String())
}
