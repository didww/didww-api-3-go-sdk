package didww

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixturesDir returns the absolute path to the testdata/fixtures directory.
func fixturesDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata", "fixtures")
}

// loadFixture reads a fixture file from the testdata/fixtures directory.
func loadFixture(t *testing.T, path string) []byte {
	t.Helper()
	fullPath := filepath.Join(fixturesDir(), path)
	data, err := os.ReadFile(fullPath)
	require.NoError(t, err, "failed to load fixture %s", path)
	return data
}

// testAccessToken is the access token granted by withTokenEndpoint.
const testAccessToken = "test-access-token"

// withTokenEndpoint serves POST /token with a token granted for the requested
// audience and API version, and passes every other request to h.
func withTokenEndpoint(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/token" {
			writeTestToken(w, r, testAccessToken)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// writeTestToken answers a token request with accessToken, echoing the
// requested audience and API version, scope "full" and expires_in 900.
func writeTestToken(w http.ResponseWriter, r *http.Request, accessToken string) {
	var req struct {
		Audience   []string `json:"aud"`
		APIVersion string   `json:"api_version"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   900,
		"aud":          req.Audience,
		"scope":        "full",
		"api_version":  req.APIVersion,
	})
}

// newTestClient builds a client whose JSON:API and token service are both
// served at url, with logging disabled. opts are applied last.
func newTestClient(t *testing.T, url string, opts ...ClientOption) *Client {
	t.Helper()
	opts = append([]ClientOption{WithBaseURL(url), WithAuthURL(url), WithLogger(nil)}, opts...)
	client, err := NewClient("test-api-key", opts...)
	require.NoError(t, err)
	return client
}

// newTestServer creates an httptest.Server that serves fixture files based on the request path.
// The handler maps URL paths to fixture file paths; POST /token is served by withTokenEndpoint.
func newTestServer(t *testing.T, routes map[string]testRoute) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			// Try without method prefix for simple GET routes
			route, ok = routes[r.URL.Path]
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors":[{"title":"not found","status":"404"}]}`))
			return
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(route.status)
		if route.fixture != "" {
			data := loadFixture(t, route.fixture)
			w.Write(data)
		} else if route.body != "" {
			w.Write([]byte(route.body))
		}
	})))

	client := newTestClient(t, server.URL)

	t.Cleanup(func() {
		server.Close()
	})

	return server, client
}

// testRoute defines a route for the test server.
type testRoute struct {
	status  int
	fixture string // path to fixture file relative to testdata/fixtures/
	body    string // raw body to return (used if fixture is empty)
}

// testServerWithClient wraps a test server with an associated client.
type testServerWithClient struct {
	server *http.Server
	client *Client
}

// newTestServerWithInspector creates a test server that calls an inspector function for each request.
func newTestServerWithInspector(t *testing.T, routes map[string]testRoute, inspector func(r *http.Request)) *testServerWithClient {
	t.Helper()
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if inspector != nil {
			inspector(r)
		}

		route, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			route, ok = routes[r.URL.Path]
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors":[{"title":"not found","status":"404"}]}`))
			return
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(route.status)
		if route.fixture != "" {
			data := loadFixture(t, route.fixture)
			w.Write(data)
		} else if route.body != "" {
			w.Write([]byte(route.body))
		}
	})))

	client := newTestClient(t, server.URL)

	t.Cleanup(func() {
		server.Close()
	})

	return &testServerWithClient{client: client}
}

// newTestServerWithDynamicPatch creates a test server where PATCH requests are handled
// dynamically based on call count. Other methods use static routes.
func newTestServerWithDynamicPatch(t *testing.T, routes map[string]testRoute, inspector func(r *http.Request), patchRouter func(call int) testRoute) *testServerWithClient {
	t.Helper()
	var patchCount int
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if inspector != nil {
			inspector(r)
		}

		var route testRoute
		var ok bool
		if r.Method == http.MethodPatch {
			patchCount++
			route = patchRouter(patchCount)
			ok = true
		} else {
			route, ok = routes[r.Method+" "+r.URL.Path]
			if !ok {
				route, ok = routes[r.URL.Path]
			}
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors":[{"title":"not found","status":"404"}]}`))
			return
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(route.status)
		if route.fixture != "" {
			data := loadFixture(t, route.fixture)
			w.Write(data)
		} else if route.body != "" {
			w.Write([]byte(route.body))
		}
	})))

	client := newTestClient(t, server.URL)

	t.Cleanup(func() {
		server.Close()
	})

	return &testServerWithClient{client: client}
}

// assertRequestJSON compares the actual request body against a fixture file using semantic JSON comparison.
// Key order is ignored; both are normalized to map[string]any before comparison.
func assertRequestJSON(t *testing.T, actual []byte, fixturePath string) {
	t.Helper()
	expected := loadFixture(t, fixturePath)

	var actualObj any
	require.NoError(t, json.Unmarshal(actual, &actualObj), "failed to parse actual request body")

	var expectedObj any
	require.NoError(t, json.Unmarshal(expected, &expectedObj), "failed to parse expected fixture %s", fixturePath)

	assert.Equal(t, expectedObj, actualObj, "request body mismatch for fixture %s", fixturePath)
}

// captureRequestBody creates a test server with an inspector that captures the request body.
// Returns the server wrapper and a pointer to the captured body bytes.
func captureRequestBody(t *testing.T, routes map[string]testRoute) (*testServerWithClient, *[]byte) {
	t.Helper()
	var body []byte
	server := newTestServerWithInspector(t, routes, func(r *http.Request) {
		if r.Method == http.MethodPatch || r.Method == http.MethodPost {
			body, _ = io.ReadAll(r.Body)
		}
	})
	return server, &body
}

// authTestServer serves POST /token through mint and every other request
// through api, recording the token requests.
type authTestServer struct {
	*httptest.Server
	mu          sync.Mutex
	mintBodies  [][]byte
	mintHeaders []http.Header
}

// newAuthTestServer starts an authTestServer. mint receives the 1-based number
// of the token request; nil grants "token-1", "token-2", ... in turn. api may be
// nil when only the token service is called.
func newAuthTestServer(t *testing.T, mint func(n int, w http.ResponseWriter, r *http.Request), api http.HandlerFunc) *authTestServer {
	t.Helper()
	if mint == nil {
		mint = func(n int, w http.ResponseWriter, r *http.Request) {
			writeTestToken(w, r, fmt.Sprintf("token-%d", n))
		}
	}
	s := &authTestServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/token" {
			body, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
			s.mu.Lock()
			s.mintBodies = append(s.mintBodies, body)
			s.mintHeaders = append(s.mintHeaders, r.Header.Clone())
			n := len(s.mintBodies)
			s.mu.Unlock()
			mint(n, w, r)
			return
		}
		if api == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		api(w, r)
	}))
	t.Cleanup(s.Close)
	return s
}

// mints returns the number of token requests received.
func (s *authTestServer) mints() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.mintBodies)
}

// mintBody returns the body of the i-th (0-based) token request.
func (s *authTestServer) mintBody(i int) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mintBodies[i]
}

// mintHeader returns the headers of the i-th (0-based) token request.
func (s *authTestServer) mintHeader(i int) http.Header {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mintHeaders[i]
}

// writeTokenError answers a token request with status and a fixture body.
func writeTokenError(t *testing.T, w http.ResponseWriter, status int, fixture string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(loadFixture(t, fixture))
}

// stubBackoff replaces the provider's sleep and jitter, returning the recorded delays.
func stubBackoff(c *Client, jitter float64) *[]time.Duration {
	var mu sync.Mutex
	sleeps := []time.Duration{}
	c.tokens.random = func() float64 { return jitter }
	c.tokens.sleep = func(_ context.Context, d time.Duration) error {
		mu.Lock()
		defer mu.Unlock()
		sleeps = append(sleeps, d)
		return nil
	}
	return &sleeps
}

// forgetRateLimit removes the 429 window of c when the test ends, so that a
// later test served on the same address is not blocked.
func forgetRateLimit(t *testing.T, c *Client) {
	t.Cleanup(func() {
		rateLimits.Lock()
		defer rateLimits.Unlock()
		delete(rateLimits.until, c.tokens.rateLimitKey())
	})
}

// fakeClock is a manually advanced clock.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// syncBuffer is a bytes.Buffer safe for concurrent use, for capturing logs.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *syncBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}
