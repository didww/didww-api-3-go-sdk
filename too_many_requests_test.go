package didww

import (
	"context"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tooManyRequestsBody = `{"errors":[{"title":"Too many requests","detail":"The request budget is spent","code":"too_many_requests","status":"429"}]}`

// apiErrorCalls are the calls that return an error status as an *APIError.
var apiErrorCalls = []struct {
	name string
	call func(c *Client, baseURL string) error
}{
	{"JSON:API call", func(c *Client, _ string) error {
		_, err := c.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b")
		return err
	}},
	{"export download", func(c *Client, baseURL string) error {
		return c.DownloadExport(context.Background(), baseURL+"/v3/exports/02bf6df4.csv.gz", io.Discard)
	}},
	{"encrypted file upload", func(c *Client, _ string) error {
		_, err := c.UploadEncryptedFile(context.Background(), []byte("encrypted-content"), "sample.pdf.enc", "fingerprint-123", "sample.pdf")
		return err
	}},
}

func TestTooManyRequestsReturnsRetryAfter(t *testing.T) {
	headers := []struct {
		name       string
		retryAfter string
		want       time.Duration
	}{
		{"with Retry-After", "30", 30 * time.Second},
		{"without Retry-After", "", 0},
		{"with a fractional Retry-After", "1.5", 0},
		{"with an HTTP-date Retry-After", "Wed, 21 Oct 2026 07:28:00 GMT", 0},
	}
	for _, c := range apiErrorCalls {
		for _, h := range headers {
			t.Run(c.name+" "+h.name, func(t *testing.T) {
				var requests atomic.Int32
				srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
					requests.Add(1)
					if h.retryAfter != "" {
						w.Header().Set("Retry-After", h.retryAfter)
					}
					w.Header().Set("Content-Type", jsonapiMediaType)
					w.WriteHeader(http.StatusTooManyRequests)
					_, _ = w.Write([]byte(tooManyRequestsBody))
				})
				client := newTestClient(t, srv.URL)

				err := c.call(client, srv.URL)

				var apiErr *APIError
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, http.StatusTooManyRequests, apiErr.HTTPStatus)
				assert.Equal(t, h.want, apiErr.RetryAfter)
				require.Len(t, apiErr.Errors, 1)
				assert.Equal(t, "too_many_requests", apiErr.Errors[0].Code)
				assert.Equal(t, int32(1), requests.Load(), "a 429 is not retried")
				assert.Equal(t, 1, srv.mints())
			})
		}
	}
}

func TestTooManyRequestsWithoutErrorsDocument(t *testing.T) {
	for _, c := range apiErrorCalls {
		t.Run(c.name, func(t *testing.T) {
			srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Retry-After", "7")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte("Too Many Requests"))
			})
			client := newTestClient(t, srv.URL)

			err := c.call(client, srv.URL)

			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, http.StatusTooManyRequests, apiErr.HTTPStatus)
			assert.Equal(t, 7*time.Second, apiErr.RetryAfter)
			require.Len(t, apiErr.Errors, 1)
			assert.Equal(t, "Too Many Requests", apiErr.Errors[0].Detail)
		})
	}
}

func TestRetryAfterOnOtherErrorStatus(t *testing.T) {
	for _, c := range apiErrorCalls {
		t.Run(c.name, func(t *testing.T) {
			srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Retry-After", "30")
				w.Header().Set("Content-Type", jsonapiMediaType)
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"errors":[{"title":"Service unavailable","status":"503"}]}`))
			})
			client := newTestClient(t, srv.URL)

			err := c.call(client, srv.URL)

			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, http.StatusServiceUnavailable, apiErr.HTTPStatus)
			assert.Equal(t, 30*time.Second, apiErr.RetryAfter)
		})
	}
}
