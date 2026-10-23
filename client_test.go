package didww

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v4/resource"
)

func TestClientSendsCorrectHeaders(t *testing.T) {
	var (
		receivedContentType   string
		receivedAccept        string
		receivedAuthorization string
		receivedUserAgent     string
		receivedHeader        http.Header
	)

	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		receivedAccept = r.Header.Get("Accept")
		receivedAuthorization = r.Header.Get("Authorization")
		receivedUserAgent = r.Header.Get("User-Agent")
		receivedHeader = r.Header.Clone()

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		w.Write(loadFixture(t, "balance/index.json"))
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)

	_, err := client.Balance().Find(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "application/vnd.api+json", receivedContentType)
	assert.Equal(t, "application/vnd.api+json", receivedAccept)
	assert.Equal(t, "Bearer "+testAccessToken, receivedAuthorization)
	assert.Equal(t, "didww-go-sdk/"+sdkVersion, receivedUserAgent)
	assert.NotContains(t, receivedHeader, "Api-Key")
	assert.NotContains(t, receivedHeader, "X-Didww-Api-Version")
}

func TestClientHandlesHTTPErrors(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		makeRequest  func(*Client) error
		checkErrors  func(*testing.T, *APIError)
	}{
		{
			name:         "404 Not Found",
			statusCode:   http.StatusNotFound,
			responseBody: `{"errors":[{"title":"not found","detail":"Resource not found","status":"404"}]}`,
			makeRequest: func(c *Client) error {
				_, err := c.Countries().Find(context.Background(), "nonexistent-id")
				return err
			},
		},
		{
			name:         "500 Internal Server Error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"errors":[{"title":"server error","detail":"Internal server error","status":"500"}]}`,
			makeRequest: func(c *Client) error {
				_, err := c.Balance().Find(context.Background())
				return err
			},
		},
		{
			name:         "422 Unprocessable Entity",
			statusCode:   http.StatusUnprocessableEntity,
			responseBody: `{"errors":[{"title":"is invalid","detail":"name - is invalid","code":"100","source":{"pointer":"/data/attributes/name"},"status":"422"}]}`,
			makeRequest: func(c *Client) error {
				_, err := c.VoiceInTrunks().Create(context.Background(), &resource.VoiceInTrunk{Name: "test"})
				return err
			},
			checkErrors: func(t *testing.T, apiErr *APIError) {
				require.Len(t, apiErr.Errors, 1)
				assert.Equal(t, "100", apiErr.Errors[0].Code)
				assert.Equal(t, "/data/attributes/name", apiErr.Errors[0].Source.Pointer)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.responseBody))
			})))
			defer server.Close()

			client := newTestClient(t, server.URL)

			err := tc.makeRequest(client)
			require.Error(t, err)

			apiErr, ok := err.(*APIError)
			require.True(t, ok, "expected *APIError")
			assert.Equal(t, tc.statusCode, apiErr.HTTPStatus)
			if tc.checkErrors != nil {
				tc.checkErrors(t, apiErr)
			}
		})
	}
}

func TestClientWithQueryParamsAppendedToURL(t *testing.T) {
	var requestURL string
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURL = r.URL.String()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		w.Write(loadFixture(t, "countries/index.json"))
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)

	params := NewQueryParams().
		Filter("prefix", "44").
		Sort("name").
		Include("regions").
		Page(1, 25)

	_, err := client.Countries().List(context.Background(), params)
	require.NoError(t, err)

	// Verify query params are appended to URL
	assertContains(t, requestURL, "filter[prefix]=44")
	assertContains(t, requestURL, "sort=name")
	assertContains(t, requestURL, "include=regions")
	assertContains(t, requestURL, "page[number]=1")
	assertContains(t, requestURL, "page[size]=25")
}

func TestClientWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	client, err := NewClient("test-api-key", WithHTTPClient(custom))
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Same(t, custom, client.httpClient)
	assert.Nil(t, custom.CheckRedirect, "the user's client must not be modified")

	// Every call goes through the custom client's transport.
	var mu sync.Mutex
	var seen []string
	custom.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		seen = append(seen, r.Method+" "+r.URL.Path)
		mu.Unlock()
		return http.DefaultTransport.RoundTrip(r)
	})
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v3/balance":
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.Write(loadFixture(t, "balance/index.json"))
		case "/v3/exports/export.csv.gz":
			w.Write(loadFixture(t, "exports/download.csv.gz"))
		case "/v3/encrypted_files":
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusCreated)
			w.Write(loadFixture(t, "encrypted_files/create.json"))
		default:
			w.WriteHeader(http.StatusOK)
		}
	})))
	defer server.Close()

	client = newTestClient(t, server.URL, WithHTTPClient(custom), WithVoiceInPricesURL(server.URL+"/vin"))
	ctx := context.Background()
	_, err = client.Balance().Find(ctx)
	require.NoError(t, err)
	require.NoError(t, client.DownloadExport(ctx, server.URL+"/v3/exports/export.csv.gz", io.Discard))
	_, err = client.UploadEncryptedFile(ctx, []byte("encrypted-content"), "sample.pdf.enc", "fingerprint-123", "sample.pdf")
	require.NoError(t, err)
	_, err = client.VoiceInPrices().Head(ctx, time.Time{})
	require.NoError(t, err)

	assert.Equal(t, []string{
		"POST /token", "GET /v3/balance", "GET /v3/exports/export.csv.gz",
		"POST /v3/encrypted_files", "POST /token", "HEAD /vin/v3/prices.csv.gz",
	}, seen)
	assert.NotSame(t, custom, client.noRedirectClient)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClientDefaultHTTPClient(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	assert.Zero(t, client.httpClient.Timeout)
	transport, ok := client.httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, 10*time.Second, transport.TLSHandshakeTimeout)
	assert.Equal(t, 60*time.Second, transport.ResponseHeaderTimeout)
	assert.Nil(t, client.httpClient.CheckRedirect)

	require.NotNil(t, client.noRedirectClient.CheckRedirect)
	assert.Equal(t, http.ErrUseLastResponse, client.noRedirectClient.CheckRedirect(nil, nil))
	assert.Same(t, transport, client.noRedirectClient.Transport)
}

func TestNewClientEndpoints(t *testing.T) {
	t.Run("sandbox by default", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		require.NoError(t, err)
		assert.Equal(t, "https://auth-sandbox.didww.com", client.authURL)
		assert.Equal(t, "https://api-vin-sandbox.didww.com/v3", client.voiceInPricesURL)
		assert.Equal(t, "https://api-vout-sandbox.didww.com/v3", client.voiceOutPricesURL)
		assert.Equal(t, "https://api-sms-sandbox.didww.com/v3", client.smsPricesURL)
	})

	t.Run("production", func(t *testing.T) {
		client, err := NewClient("test-api-key", WithEnvironment(Production))
		require.NoError(t, err)
		assert.Equal(t, string(Production), client.BaseURL())
		assert.Equal(t, "https://auth.didww.com", client.authURL)
		assert.Equal(t, "https://api-vin.didww.com/v3", client.voiceInPricesURL)
		assert.Equal(t, "https://api-vout.didww.com/v3", client.voiceOutPricesURL)
		assert.Equal(t, "https://api-sms.didww.com/v3", client.smsPricesURL)
	})

	t.Run("overrides", func(t *testing.T) {
		client, err := NewClient("test-api-key",
			WithEnvironment(Production),
			WithBaseURL("http://localhost:3000"),
			WithAuthURL("http://localhost:3001"),
			WithVoiceInPricesURL("http://localhost:3002/v3/"),
			WithVoiceOutPricesURL("http://localhost:3003"),
			WithSMSPricesURL("http://localhost:3004/v3"),
		)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:3000", client.BaseURL())
		assert.Equal(t, "http://localhost:3001", client.authURL)
		assert.Equal(t, "http://localhost:3002/v3/", client.voiceInPricesURL)
		assert.Equal(t, "http://localhost:3003", client.voiceOutPricesURL)
		assert.Equal(t, "http://localhost:3004/v3", client.smsPricesURL)
	})

	t.Run("API key with a newline", func(t *testing.T) {
		_, err := NewClient("test-api-key\n")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API key contains characters not allowed")
	})

	for name, opt := range map[string]func(string) ClientOption{
		"Voice IN prices":  WithVoiceInPricesURL,
		"Voice OUT prices": WithVoiceOutPricesURL,
		"SMS prices":       WithSMSPricesURL,
	} {
		t.Run(name+" URL without a scheme", func(t *testing.T) {
			_, err := NewClient("test-api-key", opt("api-prices.example.com"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), name+" URL must be an http or https URL")
		})
	}

	for name, opt := range map[string]func(string) ClientOption{
		"API base":      WithBaseURL,
		"token service": WithAuthURL,
		"SMS prices":    WithSMSPricesURL,
	} {
		for _, u := range []string{"https://auth.example.com#custom", "https://api-prices.example.com/v3?region=eu", "https://api-prices.example.com/v3?"} {
			t.Run(name+" URL with a query or fragment "+u, func(t *testing.T) {
				_, err := NewClient("test-api-key", opt(u))
				require.Error(t, err)
				assert.Contains(t, err.Error(), name+" URL must be an http or https URL")
			})
		}
	}

	t.Run("token service URL without a scheme", func(t *testing.T) {
		_, err := NewClient("test-api-key", WithAuthURL("auth.didww.com"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token service URL must be an http or https URL")
	})

	t.Run("unknown environment", func(t *testing.T) {
		_, err := NewClient("test-api-key", WithEnvironment(Environment("https://api.didww.com/v3/")))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "environment must be didww.Sandbox or didww.Production")
	})

	t.Run("custom base URL keeps the environment's token service", func(t *testing.T) {
		client, err := NewClient("test-api-key", WithBaseURL("http://localhost:3000/v3"))
		require.NoError(t, err)
		assert.Equal(t, "https://auth-sandbox.didww.com", client.authURL)
	})
}

func TestJoinV3(t *testing.T) {
	assert.Equal(t, "http://h/v3/countries", joinV3("http://h", "countries"))
	assert.Equal(t, "http://h/v3/countries", joinV3("http://h/v3", "countries"))
	assert.Equal(t, "http://h/v3/countries", joinV3("http://h/v3/", "countries"))
	assert.Equal(t, "http://h/v3/sms_in/prices.csv.gz", joinV3("http://h/", "/sms_in/prices.csv.gz"))
}

func TestNewClientAPIVersion(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)
	assert.Equal(t, DefaultAPIVersion, client.apiVersion)
	assert.Equal(t, "2026-10-22", DefaultAPIVersion)

	client, err = NewClient("test-api-key", WithAPIVersion("2099-01-01"))
	require.NoError(t, err)
	assert.Equal(t, "2099-01-01", client.apiVersion)

	for _, v := range []string{"", "latest", "2026-10-22T00:00:00Z", "26-10-22"} {
		_, err = NewClient("test-api-key", WithAPIVersion(v))
		require.Error(t, err, v)
		assert.Contains(t, err.Error(), "YYYY-MM-DD", v)
	}

	_, err = NewClient("test-api-key", WithAPIVersion("2026-04-16"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "release-3")
	assert.Contains(t, err.Error(), "github.com/didww/didww-api-3-go-sdk/v3 ")
}

func TestNewClientScope(t *testing.T) {
	for _, s := range []Scope{ScopeRead, ScopeFull} {
		client, err := NewClient("test-api-key", WithScope(s))
		require.NoError(t, err)
		assert.Equal(t, s, client.scope)
	}

	_, err := NewClient("test-api-key", WithScope("write"))
	require.Error(t, err)
	_, err = NewClient("test-api-key", WithScope("read:prices"))
	require.Error(t, err)
}

// unauthorizedBody is a JSON:API 401 answer.
const unauthorizedBody = `{"errors":[{"title":"Unauthorized","detail":"Authorization failed","code":"401","status":"401"}]}`

type recordedRequest struct {
	authorization string
	body          string
}

// requestLog records the API requests received by a test server.
type requestLog struct {
	mu       sync.Mutex
	requests []recordedRequest
}

func (l *requestLog) record(r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.requests = append(l.requests, recordedRequest{authorization: r.Header.Get("Authorization"), body: string(body)})
}

func (l *requestLog) all() []recordedRequest {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]recordedRequest(nil), l.requests...)
}

func TestClientRetriesOnceWithNewTokenAfter401(t *testing.T) {
	var log requestLog
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Header.Get("Authorization") == "Bearer token-1" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(unauthorizedBody))
			return
		}
		w.Write(loadFixture(t, "countries/index.json"))
	})
	client := newTestClient(t, srv.URL)

	countries, err := client.Countries().List(context.Background(), nil)
	require.NoError(t, err)

	assert.NotEmpty(t, countries)
	assert.Equal(t, 2, srv.mints())
	assert.JSONEq(t, `{"aud":["api.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(0)))
	assert.JSONEq(t, `{"aud":["api.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(1)))
	requests := log.all()
	require.Len(t, requests, 2)
	assert.Equal(t, "Bearer token-1", requests[0].authorization)
	assert.Equal(t, "Bearer token-2", requests[1].authorization)
}

func TestClientReplaysPatchBodyAfter401(t *testing.T) {
	var log requestLog
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Header.Get("Authorization") == "Bearer token-1" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(unauthorizedBody))
			return
		}
		w.Write(loadFixture(t, "exports/update.json"))
	})
	client := newTestClient(t, srv.URL)

	extRef := "export-ext-ref"
	export, err := client.Exports().Update(context.Background(), &resource.Export{
		ID:                  "da15f006-5da4-45ca-b0df-735baeadf423",
		ExternalReferenceID: &extRef,
	})
	require.NoError(t, err)

	assert.Equal(t, "da15f006-5da4-45ca-b0df-735baeadf423", export.ID)
	requests := log.all()
	require.Len(t, requests, 2)
	assertRequestJSON(t, []byte(requests[0].body), "exports/update_request.json")
	assert.Equal(t, requests[0].body, requests[1].body)
	assert.Equal(t, "Bearer token-1", requests[0].authorization)
	assert.Equal(t, "Bearer token-2", requests[1].authorization)
}

func TestClientSecond401DropsReplacementToken(t *testing.T) {
	var log requestLog
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Header.Get("Authorization") != "Bearer token-3" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(unauthorizedBody))
			return
		}
		w.Write(loadFixture(t, "countries/index.json"))
	})
	client := newTestClient(t, srv.URL)

	_, err := client.Countries().List(context.Background(), nil)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnauthorized, apiErr.HTTPStatus)
	assert.Len(t, log.all(), 2)
	assert.Equal(t, 2, srv.mints())

	_, err = client.Countries().List(context.Background(), nil)
	require.NoError(t, err)

	requests := log.all()
	require.Len(t, requests, 3)
	assert.Equal(t, "Bearer token-1", requests[0].authorization)
	assert.Equal(t, "Bearer token-2", requests[1].authorization)
	assert.Equal(t, "Bearer token-3", requests[2].authorization)
	assert.Equal(t, 3, srv.mints())
}

func TestClientForbiddenNotRetried(t *testing.T) {
	var log requestLog
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope"`)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"errors":[{"title":"Insufficient scope","code":"403","status":"403"}]}`))
	})
	client := newTestClient(t, srv.URL, WithScope(ScopeRead))

	_, err := client.VoiceInTrunks().Create(context.Background(), &resource.VoiceInTrunk{Name: "test"})

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.HTTPStatus)
	assert.Equal(t, "Insufficient scope", apiErr.Errors[0].Title)
	assert.Len(t, log.all(), 1)
	assert.Equal(t, 1, srv.mints())
}

func TestClientKeepsTokenMintedByAnotherCaller(t *testing.T) {
	var client *Client
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		if r.Header.Get("Authorization") == "Bearer token-1" {
			// Another caller replaces the token before this answer arrives.
			client.tokens.invalidate(context.Background(), audienceAPI, "token-1")
			_, err := client.tokens.token(context.Background(), audienceAPI)
			assert.NoError(t, err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(unauthorizedBody))
			return
		}
		w.Write(loadFixture(t, "countries/index.json"))
	})
	client = newTestClient(t, srv.URL)

	_, err := client.Countries().List(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, 2, srv.mints())
}

func TestClientReturnsAuthErrors(t *testing.T) {
	var log requestLog
	srv := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeTokenError(t, w, http.StatusUnauthorized, "token/invalid_client.json")
	}, func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		w.WriteHeader(http.StatusOK)
	})
	client := newTestClient(t, srv.URL)

	_, err := client.Countries().List(context.Background(), nil)

	require.ErrorIs(t, err, ErrInvalidClient)
	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	assert.Empty(t, log.all())
	assert.Equal(t, 1, srv.mints())
}

func TestClientWrapsTransportErrors(t *testing.T) {
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		require.True(t, ok)
		conn, _, err := hj.Hijack()
		require.NoError(t, err)
		conn.Close()
	})
	client := newTestClient(t, srv.URL)

	_, err := client.Countries().List(context.Background(), nil)

	var clientErr *ClientError
	require.ErrorAs(t, err, &clientErr)
	assert.Contains(t, clientErr.Message, "request failed")
	var authErr *AuthError
	assert.False(t, errors.As(err, &authErr))
}

func TestSendStripsIfNoneMatch(t *testing.T) {
	var header http.Header
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Clone()
		w.WriteHeader(http.StatusNotModified)
	})
	client := newTestClient(t, srv.URL)

	resp, err := client.send(context.Background(), client.httpClient, audienceVoiceIn, func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/v3/prices.csv.gz", http.NoBody)
		if err != nil {
			return nil, err
		}
		req.Header.Set("If-None-Match", `"abc"`)
		req.Header.Set("If-Modified-Since", "Wed, 14 Oct 2026 06:00:00 GMT")
		return req, nil
	})
	require.NoError(t, err)
	resp.Body.Close()

	assert.NotContains(t, header, "If-None-Match")
	assert.Equal(t, "Wed, 14 Oct 2026 06:00:00 GMT", header.Get("If-Modified-Since"))
	assert.Equal(t, "Bearer token-1", header.Get("Authorization"))
	assert.JSONEq(t, `{"aud":["api-vin.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(0)))
}

func TestDefaultClientsShareTransport(t *testing.T) {
	first, err := NewClient("key-1")
	require.NoError(t, err)
	second, err := NewClient("key-2")
	require.NoError(t, err)

	assert.Same(t, first.httpClient.Transport, second.httpClient.Transport)
}

func TestNewDefaultTransport(t *testing.T) {
	base := &http.Transport{}

	transport, ok := newDefaultTransport(base).(*http.Transport)

	require.True(t, ok)
	assert.NotSame(t, base, transport)
	assert.Equal(t, 10*time.Second, transport.TLSHandshakeTimeout)
	assert.Equal(t, 60*time.Second, transport.ResponseHeaderTimeout)
	assert.Nil(t, newDefaultTransport(roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, nil })))
}
