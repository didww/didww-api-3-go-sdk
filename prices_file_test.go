package didww

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	lastModifiedHeader      = "Wed, 14 Oct 2026 06:00:00 GMT"
	otherLastModifiedHeader = "Thu, 15 Oct 2026 06:00:00 GMT"
)

var (
	lastModifiedTime      = time.Date(2026, 10, 14, 6, 0, 0, 0, time.UTC)
	otherLastModifiedTime = time.Date(2026, 10, 15, 6, 0, 0, 0, time.UTC)
)

// writePriceFile answers with a price file; HEAD gets the headers only.
func writePriceFile(w http.ResponseWriter, r *http.Request, body []byte, lastModified string) {
	h := w.Header()
	h.Set("Content-Type", "application/gzip")
	h.Set("Content-Disposition", `attachment; filename="voice_in_prices.csv.gz"`)
	h.Set("Content-Length", strconv.Itoa(len(body)))
	h.Set("Last-Modified", lastModified)
	h.Set("Cache-Control", "private")
	h.Set("Accept-Ranges", "none")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		w.Write(body)
	}
}

// writeNotModified answers 304 with the headers of a price file.
func writeNotModified(w http.ResponseWriter, lastModified string) {
	w.Header().Set("Content-Disposition", `attachment; filename="voice_in_prices.csv.gz"`)
	w.Header().Set("Last-Modified", lastModified)
	w.WriteHeader(http.StatusNotModified)
}

// writeServingFailure answers 503 with retryAfter, if any.
func writeServingFailure(t *testing.T, w http.ResponseWriter, retryAfter string) {
	if retryAfter != "" {
		w.Header().Set("Retry-After", retryAfter)
	}
	writePricesJSON(w, http.StatusServiceUnavailable, loadFixture(t, "prices/errors/serving_failure.json"))
}

// stubPricesSleep replaces the 503 sleep of s, returning the recorded delays.
func stubPricesSleep[T any](s *PricesService[T]) *[]time.Duration {
	sleeps := []time.Duration{}
	s.sleep = func(_ context.Context, d time.Duration) error {
		sleeps = append(sleeps, d)
		return nil
	}
	return &sleeps
}

// dirEntries returns the names in dir.
func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

func modTime(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err)
	return info.ModTime()
}

type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) { return 0, w.err }

func TestPriceFileEndpoints(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	tests := []struct {
		name     string
		download func(c *Client, w io.Writer) (*PriceFile, error)
		path     string
		audience string
	}{
		{"VoiceIn", func(c *Client, w io.Writer) (*PriceFile, error) {
			return c.VoiceInPrices().Download(context.Background(), w, time.Time{})
		}, "/v3/prices.csv.gz", "api-vin.didww.com"},
		{"VoiceOut", func(c *Client, w io.Writer) (*PriceFile, error) {
			return c.VoiceOutPrices().Download(context.Background(), w, time.Time{})
		}, "/v3/prices.csv.gz", "api-vout.didww.com"},
		{"SMSOutA2P", func(c *Client, w io.Writer) (*PriceFile, error) {
			return c.SMSOutA2PPrices().Download(context.Background(), w, time.Time{})
		}, "/v3/sms_out/a2p/prices.csv.gz", "api-sms.didww.com"},
		{"SMSOutP2P", func(c *Client, w io.Writer) (*PriceFile, error) {
			return c.SMSOutP2PPrices().Download(context.Background(), w, time.Time{})
		}, "/v3/sms_out/p2p/prices.csv.gz", "api-sms.didww.com"},
		{"SMSOutOTP", func(c *Client, w io.Writer) (*PriceFile, error) {
			return c.SMSOutOTPPrices().Download(context.Background(), w, time.Time{})
		}, "/v3/sms_out/otp/prices.csv.gz", "api-sms.didww.com"},
		{"SMSIn", func(c *Client, w io.Writer) (*PriceFile, error) {
			return c.SMSInPrices().Download(context.Background(), w, time.Time{})
		}, "/v3/sms_in/prices.csv.gz", "api-sms.didww.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
				writePriceFile(w, r, gzipFixture, lastModifiedHeader)
			})
			var buf bytes.Buffer

			_, err := tt.download(client, &buf)
			require.NoError(t, err)

			require.Equal(t, 1, s.requestCount())
			req := s.request(0)
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, tt.path, req.URL.Path)
			assert.Empty(t, req.URL.RawQuery)
			assert.Equal(t, "Bearer token-1", req.Header.Get("Authorization"))
			assert.Equal(t, []string{tt.audience}, s.mintAudiences(t))
			assert.Equal(t, gzipFixture, buf.Bytes())
		})
	}
}

func TestPriceFileDownload(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	var buf bytes.Buffer

	file, err := client.VoiceInPrices().Download(context.Background(), &buf, time.Time{})
	require.NoError(t, err)

	assert.Equal(t, gzipFixture, buf.Bytes(), "the gzipped bytes are written unchanged")
	assert.Equal(t, http.StatusOK, file.StatusCode)
	assert.False(t, file.NotModified())
	assert.True(t, file.LastModified.Equal(lastModifiedTime))
	assert.Equal(t, int64(len(gzipFixture)), file.ContentLength)
	assert.Equal(t, "voice_in_prices.csv.gz", file.Filename)
	assert.Equal(t, int64(len(gzipFixture)), file.BytesWritten)

	req := s.request(0)
	assert.Empty(t, req.Header.Get("If-Modified-Since"))
	assert.Empty(t, req.Header.Get("If-None-Match"))
	assert.Empty(t, req.Header.Get("Api-Key"))
	assert.Equal(t, userAgent, req.Header.Get("User-Agent"))
}

func TestPriceFileDownloadSendsIfModifiedSince(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	local := time.Date(2026, 10, 14, 8, 0, 0, 0, time.FixedZone("CEST", 2*60*60))

	_, err := client.VoiceInPrices().Download(context.Background(), io.Discard, local)
	require.NoError(t, err)

	assert.Equal(t, lastModifiedHeader, s.request(0).Header.Get("If-Modified-Since"))
}

func TestPriceFileDownloadNotModified(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeNotModified(w, lastModifiedHeader)
	})
	var buf bytes.Buffer

	file, err := client.VoiceInPrices().Download(context.Background(), &buf, lastModifiedTime)
	require.NoError(t, err)

	assert.True(t, file.NotModified())
	assert.Equal(t, http.StatusNotModified, file.StatusCode)
	assert.True(t, file.LastModified.Equal(lastModifiedTime))
	assert.Equal(t, int64(-1), file.ContentLength)
	assert.Equal(t, int64(0), file.BytesWritten)
	assert.Empty(t, buf.Bytes())
	assert.Equal(t, lastModifiedHeader, s.request(0).Header.Get("If-Modified-Since"))
}

func TestPriceFileDownloadRequiresWriter(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := client.VoiceInPrices().Download(context.Background(), nil, time.Time{})

	var clientErr *ClientError
	require.ErrorAs(t, err, &clientErr)
	assert.Equal(t, 0, s.requestCount())
}

func TestPriceFileHead(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})

	file, err := client.VoiceInPrices().Head(context.Background(), time.Time{})
	require.NoError(t, err)

	assert.Equal(t, http.MethodHead, s.request(0).Method)
	assert.Equal(t, "/v3/prices.csv.gz", s.request(0).URL.Path)
	assert.Empty(t, s.request(0).Header.Get("If-Modified-Since"))
	assert.Equal(t, http.StatusOK, file.StatusCode)
	assert.True(t, file.LastModified.Equal(lastModifiedTime))
	assert.Equal(t, int64(len(gzipFixture)), file.ContentLength)
	assert.Equal(t, "voice_in_prices.csv.gz", file.Filename)
	assert.Equal(t, int64(0), file.BytesWritten)
}

func TestPriceFileHeadNotModified(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeNotModified(w, lastModifiedHeader)
	})

	file, err := client.VoiceInPrices().Head(context.Background(), lastModifiedTime)
	require.NoError(t, err)

	assert.True(t, file.NotModified())
	assert.True(t, file.LastModified.Equal(lastModifiedTime))
	assert.Equal(t, lastModifiedHeader, s.request(0).Header.Get("If-Modified-Since"))
}

func TestPriceFileHeadNotFoundWithoutBody(t *testing.T) {
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.VoiceInPrices().Head(context.Background(), time.Time{})

	require.ErrorIs(t, err, ErrPricePlanNotAssigned)
	var pricesErr *PricesError
	require.ErrorAs(t, err, &pricesErr)
	assert.Equal(t, http.StatusNotFound, pricesErr.StatusCode)
	assert.Equal(t, "price_plan_not_assigned", pricesErr.Code)
	assert.Empty(t, pricesErr.Errors)
}

func TestPriceFileHeadRetriesServiceUnavailable(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	s, client := newPricesTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	prices := client.VoiceInPrices()
	sleeps := stubPricesSleep(prices)

	file, err := prices.Head(context.Background(), time.Time{})
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, file.StatusCode)
	assert.Equal(t, []time.Duration{2 * time.Second}, *sleeps)
	require.Equal(t, 2, s.requestCount())
	assert.Equal(t, http.MethodHead, s.request(1).Method)
}

func TestPriceFileDownloadRetriesServiceUnavailable(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	s, client := newPricesTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n == 1 {
			writeServingFailure(t, w, "1")
			return
		}
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	prices := client.VoiceInPrices()
	sleeps := stubPricesSleep(prices)
	var buf bytes.Buffer

	_, err := prices.Download(context.Background(), &buf, time.Time{})
	require.NoError(t, err)

	assert.Equal(t, gzipFixture, buf.Bytes(), "the 503 body is not written")
	assert.Equal(t, []time.Duration{time.Second}, *sleeps)
	assert.Equal(t, 2, s.requestCount())
}

func TestPriceFileDownloadServiceUnavailableAfterTwoRetries(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeServingFailure(t, w, "1")
	})
	prices := client.VoiceInPrices()
	sleeps := stubPricesSleep(prices)
	var buf bytes.Buffer

	_, err := prices.Download(context.Background(), &buf, time.Time{})

	require.ErrorIs(t, err, ErrPricesServiceUnavailable)
	var pricesErr *PricesError
	require.ErrorAs(t, err, &pricesErr)
	assert.Equal(t, http.StatusServiceUnavailable, pricesErr.StatusCode)
	assert.Equal(t, "serving_failure", pricesErr.Code)
	assert.Equal(t, time.Second, pricesErr.RetryAfter)
	assert.EqualError(t, err, "didww: prices service: The price file is not available yet.")
	assert.Equal(t, 3, s.requestCount())
	assert.Equal(t, []time.Duration{time.Second, time.Second}, *sleeps)
	assert.Empty(t, buf.Bytes())
}

func TestPriceFileDownloadServiceUnavailableNotRetried(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter string
		wait       time.Duration
	}{
		{"Retry-After over a minute", "120", 120 * time.Second},
		{"Retry-After as an HTTP date", "Wed, 21 Oct 2026 07:28:00 GMT", 0},
		{"without Retry-After", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
				writeServingFailure(t, w, tt.retryAfter)
			})
			prices := client.VoiceInPrices()
			sleeps := stubPricesSleep(prices)

			_, err := prices.Download(context.Background(), io.Discard, time.Time{})

			require.ErrorIs(t, err, ErrPricesServiceUnavailable)
			var pricesErr *PricesError
			require.ErrorAs(t, err, &pricesErr)
			assert.Equal(t, tt.wait, pricesErr.RetryAfter)
			assert.Equal(t, 1, s.requestCount())
			assert.Empty(t, *sleeps)
		})
	}
}

func TestPriceFileDownloadSleepInterrupted(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeServingFailure(t, w, "1")
	})
	prices := client.VoiceInPrices()
	prices.sleep = func(context.Context, time.Duration) error { return context.Canceled }

	_, err := prices.Download(context.Background(), io.Discard, time.Time{})

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, s.requestCount())
}

func TestPriceFileDownloadTooManyRequestsNotRetried(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	prices := client.VoiceInPrices()
	sleeps := stubPricesSleep(prices)

	_, err := prices.Download(context.Background(), io.Discard, time.Time{})

	require.ErrorIs(t, err, ErrPricesTooManyRequests)
	var pricesErr *PricesError
	require.ErrorAs(t, err, &pricesErr)
	assert.Equal(t, http.StatusTooManyRequests, pricesErr.StatusCode)
	assert.Equal(t, 30*time.Second, pricesErr.RetryAfter)
	assert.Equal(t, 1, s.requestCount())
	assert.Empty(t, *sleeps)
}

func TestPriceFileDownloadRetriesRefusedTokenOnce(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	s, client := newPricesTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n == 1 {
			writePricesJSON(w, http.StatusUnauthorized, []byte(`{"errors":[{"title":"Unauthorized","code":"401","status":"401"}]}`))
			return
		}
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	var buf bytes.Buffer

	_, err := client.VoiceInPrices().Download(context.Background(), &buf, time.Time{})
	require.NoError(t, err)

	assert.Equal(t, gzipFixture, buf.Bytes(), "only the file bytes are written")
	assert.Equal(t, 2, s.requestCount())
	assert.Equal(t, []string{"api-vin.didww.com", "api-vin.didww.com"}, s.mintAudiences(t))
}

func TestPriceFileDownloadToFileNotAuthorized(t *testing.T) {
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusUnauthorized,
			[]byte(`{"errors":[{"title":"Unauthorized","detail":"Authorization failed","code":"401","status":"401"}]}`))
	})
	dir := t.TempDir()

	_, err := client.VoiceInPrices().DownloadToFile(context.Background(), filepath.Join(dir, "voice_in_prices.csv.gz"))

	require.ErrorIs(t, err, ErrPricesNotAuthorized)
	var pricesErr *PricesError
	require.ErrorAs(t, err, &pricesErr)
	assert.Equal(t, http.StatusUnauthorized, pricesErr.StatusCode)
	assert.Equal(t, "401", pricesErr.Code)
	assert.EqualError(t, err, "didww: prices service: Authorization failed")
	assert.Empty(t, dirEntries(t, dir))
}

func TestPriceFileDownloadPricePlanNotAssigned(t *testing.T) {
	fixture := loadFixture(t, "prices/errors/price_plan_not_assigned.json")
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusNotFound, fixture)
	})
	var buf bytes.Buffer

	_, err := client.VoiceInPrices().Download(context.Background(), &buf, time.Time{})

	require.ErrorIs(t, err, ErrPricePlanNotAssigned)
	assert.EqualError(t, err, "didww: prices service: No price plan is assigned to this account.")
	assert.Empty(t, buf.Bytes())
}

func TestPriceFileDownloadAccessDenied(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusForbidden, []byte(`{"errors":[{"title":"Insufficient scope","code":"403","status":"403"}]}`))
	})

	_, err := client.VoiceInPrices().Download(context.Background(), io.Discard, time.Time{})

	require.ErrorIs(t, err, ErrPricesAccessDenied)
	assert.EqualError(t, err, "didww: prices service: Insufficient scope")
	assert.Equal(t, 1, s.requestCount())
	assert.Equal(t, 1, s.mints())
}

func TestPriceFileDownloadUnexpectedStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"202", http.StatusAccepted},
		{"204", http.StatusNoContent},
		{"302", http.StatusFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Location", "/v3/elsewhere.csv.gz")
				w.WriteHeader(tt.status)
			})
			dir := t.TempDir()

			_, err := client.VoiceInPrices().DownloadToFile(context.Background(), filepath.Join(dir, "voice_in_prices.csv.gz"))

			require.ErrorIs(t, err, ErrPricesUnexpectedStatus)
			var pricesErr *PricesError
			require.ErrorAs(t, err, &pricesErr)
			assert.Equal(t, tt.status, pricesErr.StatusCode)
			assert.EqualError(t, err, "didww: prices service responded with status "+tt.name)
			assert.Equal(t, 1, s.requestCount(), "redirects are not followed")
			assert.Empty(t, dirEntries(t, dir))
		})
	}
}

func TestPriceFileDownloadWriterError(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	writeErr := errors.New("disk full")

	_, err := client.VoiceInPrices().Download(context.Background(), errWriter{err: writeErr}, time.Time{})

	assert.Equal(t, writeErr, err, "the writer's error is returned as is")
}

func TestPriceFileDownloadTruncatedBody(t *testing.T) {
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("0123456789"))
	})
	dir := t.TempDir()
	path := filepath.Join(dir, "voice_in_prices.csv.gz")
	require.NoError(t, os.WriteFile(path, []byte("old copy"), 0o600))

	_, err := client.VoiceInPrices().DownloadToFile(context.Background(), path)

	require.ErrorIs(t, err, ErrPricesConnection)
	assert.Equal(t, []byte("old copy"), readFile(t, path))
	assert.Equal(t, []string{"voice_in_prices.csv.gz"}, dirEntries(t, dir))
}

func TestPriceFileDownloadConnectionError(t *testing.T) {
	auth := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, auth.URL, WithVoiceInPricesURL(closedServerURL()))
	dir := t.TempDir()
	path := filepath.Join(dir, "voice_in_prices.csv.gz")
	require.NoError(t, os.WriteFile(path, []byte("old copy"), 0o600))

	_, err := client.VoiceInPrices().DownloadToFile(context.Background(), path)

	require.ErrorIs(t, err, ErrPricesConnection)
	var pricesErr *PricesError
	require.ErrorAs(t, err, &pricesErr)
	assert.Equal(t, 0, pricesErr.StatusCode)
	assert.Equal(t, []byte("old copy"), readFile(t, path))
	assert.Equal(t, []string{"voice_in_prices.csv.gz"}, dirEntries(t, dir))
}

func TestPriceFileDownloadToFile(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	dir := t.TempDir()
	path := filepath.Join(dir, "voice_in_prices.csv.gz")
	require.NoError(t, os.WriteFile(path, []byte("old copy"), 0o600))

	file, err := client.VoiceInPrices().DownloadToFile(context.Background(), path)
	require.NoError(t, err)

	assert.Equal(t, gzipFixture, readFile(t, path))
	assert.True(t, modTime(t, path).Equal(lastModifiedTime), "modification time is Last-Modified")
	assert.Equal(t, []string{"voice_in_prices.csv.gz"}, dirEntries(t, dir))
	assert.Equal(t, int64(len(gzipFixture)), file.BytesWritten)
	assert.Equal(t, "voice_in_prices.csv.gz", file.Filename)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
	}
}

func TestPriceFileDownloadToFileLeavesOtherTempFiles(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	dir := t.TempDir()
	path := filepath.Join(dir, "voice_in_prices.csv.gz")
	require.NoError(t, os.WriteFile(path+".tmp", []byte("concurrent download"), 0o600))

	_, err := client.VoiceInPrices().DownloadToFile(context.Background(), path)
	require.NoError(t, err)

	assert.Equal(t, gzipFixture, readFile(t, path))
	assert.Equal(t, []byte("concurrent download"), readFile(t, path+".tmp"))
	assert.ElementsMatch(t, []string{"voice_in_prices.csv.gz", "voice_in_prices.csv.gz.tmp"}, dirEntries(t, dir))
}

func TestPriceFileDownloadToFileErrorKeepsExistingFile(t *testing.T) {
	fixture := loadFixture(t, "prices/errors/price_plan_not_assigned.json")
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusNotFound, fixture)
	})
	dir := t.TempDir()
	path := filepath.Join(dir, "voice_in_prices.csv.gz")
	require.NoError(t, os.WriteFile(path, []byte("old copy"), 0o600))

	_, err := client.VoiceInPrices().DownloadToFile(context.Background(), path)

	require.ErrorIs(t, err, ErrPricePlanNotAssigned)
	assert.Equal(t, []byte("old copy"), readFile(t, path))
	assert.Equal(t, []string{"voice_in_prices.csv.gz"}, dirEntries(t, dir))
}

func TestPriceFileDownloadToFileMissingDirectory(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := client.VoiceInPrices().DownloadToFile(context.Background(), filepath.Join(t.TempDir(), "missing", "prices.csv.gz"))

	require.Error(t, err)
	assert.Equal(t, 0, s.requestCount())
}

func TestPriceFileFetchWithoutCopy(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	path := filepath.Join(t.TempDir(), "voice_in_prices.csv.gz")

	file, err := client.VoiceInPrices().Fetch(context.Background(), path)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, file.StatusCode)
	assert.Equal(t, gzipFixture, readFile(t, path))
	_, sent := s.request(0).Header["If-Modified-Since"]
	assert.False(t, sent)
}

func TestPriceFileFetchSendsModificationTime(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeNotModified(w, lastModifiedHeader)
	})
	dir := t.TempDir()
	path := filepath.Join(dir, "voice_in_prices.csv.gz")
	require.NoError(t, os.WriteFile(path, []byte("previous copy"), 0o600))
	require.NoError(t, os.Chtimes(path, lastModifiedTime, lastModifiedTime))

	file, err := client.VoiceInPrices().Fetch(context.Background(), path)
	require.NoError(t, err)

	assert.True(t, file.NotModified())
	assert.Equal(t, lastModifiedHeader, s.request(0).Header.Get("If-Modified-Since"))
	assert.Equal(t, []byte("previous copy"), readFile(t, path))
	assert.Equal(t, []string{"voice_in_prices.csv.gz"}, dirEntries(t, dir))
}

func TestPriceFileFetchReplacesChangedCopy(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	s, client := newPricesTestServer(t, func(n int, w http.ResponseWriter, r *http.Request) {
		switch n {
		case 1:
			writePriceFile(w, r, []byte("previous copy"), lastModifiedHeader)
		case 2:
			writePriceFile(w, r, gzipFixture, otherLastModifiedHeader)
		default:
			writeNotModified(w, otherLastModifiedHeader)
		}
	})
	prices := client.VoiceInPrices()
	dir := t.TempDir()
	path := filepath.Join(dir, "voice_in_prices.csv.gz")

	for i := 0; i < 3; i++ {
		_, err := prices.Fetch(context.Background(), path)
		require.NoError(t, err)
	}

	require.Equal(t, 3, s.requestCount())
	assert.Empty(t, s.request(0).Header.Get("If-Modified-Since"))
	assert.Equal(t, lastModifiedHeader, s.request(1).Header.Get("If-Modified-Since"))
	assert.Equal(t, otherLastModifiedHeader, s.request(2).Header.Get("If-Modified-Since"))
	for i := 0; i < 3; i++ {
		assert.Empty(t, s.request(i).Header.Get("If-None-Match"))
	}
	assert.Equal(t, gzipFixture, readFile(t, path))
	assert.True(t, modTime(t, path).Equal(otherLastModifiedTime))
	assert.Equal(t, []string{"voice_in_prices.csv.gz"}, dirEntries(t, dir))
}

func TestEachPriceRowVoiceIn(t *testing.T) {
	var rows []map[string]string

	err := EachPriceRow(bytes.NewReader(loadFixture(t, "prices/voice_in_prices.csv.gz")), func(row map[string]string) error {
		rows = append(rows, row)
		return nil
	})
	require.NoError(t, err)

	require.Len(t, rows, 4)
	assert.Equal(t, map[string]string{
		"Country":                         "United Kingdom",
		"Src country name":                "",
		"Source type":                     "cellular",
		"Destination":                     "44800",
		"Rate (USD, excl. VAT)":           "",
		"Connection fee (USD, excl. VAT)": "",
		"Initial interval":                "60",
		"Next interval":                   "60",
		"Reject calls":                    "true",
	}, rows[0])
	destinations := []string{}
	for _, row := range rows {
		destinations = append(destinations, row["Destination"])
	}
	assert.Equal(t, []string{"44800", "4480012", "1800", "1800"}, destinations)
}

func TestEachPriceRowVoiceOut(t *testing.T) {
	var rows []map[string]string

	err := EachPriceRow(bytes.NewReader(loadFixture(t, "prices/voice_out_prices.csv.gz")), func(row map[string]string) error {
		rows = append(rows, row)
		return nil
	})
	require.NoError(t, err)

	require.Len(t, rows, 4)
	assert.Equal(t, map[string]string{
		"Category":                 "International",
		"Country":                  "Germany",
		"Network":                  "Fixed",
		"Prefix":                   "4930",
		"Rate":                     "0.0045",
		"Connection fee":           "0.0",
		"Initial Increment":        "1",
		"Recurring Increment":      "1",
		"Reject calls":             "false",
		"Next rate":                "0.005",
		"Next connection fee":      "0.0",
		"Next initial increment":   "30",
		"Next recurring increment": "6",
		"Effective date":           "2026-11-01T00:00:00Z",
		"Comment":                  "Increase and Billing increment change",
	}, rows[0])
	assert.Equal(t, "Origin based", rows[1]["Category"])
	assert.Equal(t, "99950,99951,99952", rows[2]["Prefix"])
	assert.Equal(t, "New Breakout", rows[3]["Network"])
	assert.Equal(t, "", rows[3]["Rate"])
	assert.Equal(t, "true", rows[3]["Reject calls"])
}

func TestEachPriceRowReturnsCallbackError(t *testing.T) {
	stop := errors.New("stop")
	calls := 0

	err := EachPriceRow(bytes.NewReader(loadFixture(t, "prices/voice_in_prices.csv.gz")), func(map[string]string) error {
		calls++
		return stop
	})

	assert.Equal(t, stop, err)
	assert.Equal(t, 1, calls)
}

func TestEachPriceRowRejectsUncompressedInput(t *testing.T) {
	err := EachPriceRow(strings.NewReader("Country,Destination\nGermany,49\n"), func(map[string]string) error {
		t.Fatal("no row expected")
		return nil
	})

	assert.Error(t, err)
}

func TestEachPriceRowSkipsByteOrderMark(t *testing.T) {
	for _, csvText := range []string{"\uFEFFCountry,Rate\nUS,0.1\n", "\uFEFF\"Country\",Rate\nUS,0.1\n"} {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, err := gz.Write([]byte(csvText))
		require.NoError(t, err)
		require.NoError(t, gz.Close())

		var rows []map[string]string
		err = EachPriceRow(&buf, func(row map[string]string) error {
			rows = append(rows, row)
			return nil
		})

		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "US", rows[0]["Country"])
	}
}
