package didww

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/exports": {status: http.StatusOK, fixture: "exports/index.json"},
	})

	exports, err := client.Exports().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, exports, 1)

	export := exports[0]
	assert.Equal(t, "da15f006-5da4-45ca-b0df-735baeadf423", export.ID)
	assert.Equal(t, enums.ExportStatusCompleted, export.Status)
	assert.Equal(t, enums.ExportTypeCdrIn, export.ExportType)
	require.NotNil(t, export.URL)
	assert.NotEmpty(t, *export.URL)
}

func TestExportsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/exports": {status: http.StatusCreated, fixture: "exports/create.json"},
	})

	export, err := server.client.Exports().Create(context.Background(), &resource.Export{
		ExportType: enums.ExportTypeCdrIn,
		Filters: map[string]interface{}{
			"did_number": "1234556789",
			"from":       "2026-04-01 00:00:00",
			"to":         "2026-04-15 23:59:59",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "da15f006-5da4-45ca-b0df-735baeadf423", export.ID)
	assert.Equal(t, enums.ExportStatusPending, export.Status)
	assert.Equal(t, enums.ExportTypeCdrIn, export.ExportType)

	assertRequestJSON(t, *capturedBodyPtr, "exports/create_request.json")
}

func TestExportsCreateCdrOut(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/exports": {status: http.StatusCreated, fixture: "exports/create_cdr_out.json"},
	})

	export, err := client.Exports().Create(context.Background(), &resource.Export{
		ExportType: enums.ExportTypeCdrOut,
		Filters: map[string]interface{}{
			"from": "2026-04-01 00:00:00",
			"to":   "2026-04-30 23:59:59",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "da15f006-5da4-45ca-b0df-735baeadf423", export.ID)
	assert.Equal(t, enums.ExportTypeCdrOut, export.ExportType)
	assert.Equal(t, enums.ExportStatusPending, export.Status)
}

func TestExportsCreateUnauthorized(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/exports": {status: http.StatusUnauthorized, fixture: "exports/create_error_unauthorized.json"},
	})

	_, err := client.Exports().Create(context.Background(), &resource.Export{
		ExportType: enums.ExportTypeCdrIn,
	})
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError")
	assert.Equal(t, http.StatusUnauthorized, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "Unauthorized", apiErr.Errors[0].Title)
}

func TestExportsFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/exports/da15f006-5da4-45ca-b0df-735baeadf423": {status: http.StatusOK, fixture: "exports/show.json"},
	})

	export, err := client.Exports().Find(context.Background(), "da15f006-5da4-45ca-b0df-735baeadf423")
	require.NoError(t, err)

	assert.Equal(t, "da15f006-5da4-45ca-b0df-735baeadf423", export.ID)
	assert.Equal(t, enums.ExportStatusCompleted, export.Status)
	assert.Equal(t, enums.ExportTypeCdrIn, export.ExportType)
	require.NotNil(t, export.URL)
	assert.NotEmpty(t, *export.URL)
}

func TestExportsUpdateExternalReferenceID(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/exports/da15f006-5da4-45ca-b0df-735baeadf423": {status: http.StatusOK, fixture: "exports/update.json"},
	})

	extRef := "export-ext-ref"
	export, err := server.client.Exports().Update(context.Background(), &resource.Export{
		ID:                  "da15f006-5da4-45ca-b0df-735baeadf423",
		ExternalReferenceID: &extRef,
	})
	require.NoError(t, err)

	assert.Equal(t, "da15f006-5da4-45ca-b0df-735baeadf423", export.ID)
	require.NotNil(t, export.ExternalReferenceID)
	assert.Equal(t, "export-ext-ref", *export.ExternalReferenceID)

	assertRequestJSON(t, *capturedBodyPtr, "exports/update_request.json")
}

func TestExportsUpdateExternalReferenceIDFromLoaded(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"GET /v3/exports/da15f006-5da4-45ca-b0df-735baeadf423":   {status: http.StatusOK, fixture: "exports/show.json"},
		"PATCH /v3/exports/da15f006-5da4-45ca-b0df-735baeadf423": {status: http.StatusOK, fixture: "exports/update.json"},
	})

	export, err := server.client.Exports().Find(context.Background(), "da15f006-5da4-45ca-b0df-735baeadf423")
	require.NoError(t, err)

	extRef := "export-ext-ref"
	export.ExternalReferenceID = &extRef

	export, err = server.client.Exports().Update(context.Background(), export)
	require.NoError(t, err)

	assert.Equal(t, "da15f006-5da4-45ca-b0df-735baeadf423", export.ID)
	require.NotNil(t, export.ExternalReferenceID)
	assert.Equal(t, "export-ext-ref", *export.ExternalReferenceID)

	assertRequestJSON(t, *capturedBodyPtr, "exports/update_request.json")
}

func TestDownloadExport(t *testing.T) {
	gzData := loadFixture(t, "exports/download.csv.gz")

	var capturedHeader http.Header
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Clone()
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(gzData)
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)

	var buf bytes.Buffer
	err := client.DownloadExport(context.Background(), server.URL+"/v3/exports/02bf6df4.csv.gz", &buf)
	require.NoError(t, err)

	assert.Equal(t, gzData, buf.Bytes())
	assert.Equal(t, "Bearer "+testAccessToken, capturedHeader.Get("Authorization"))
	assert.NotContains(t, capturedHeader, "Api-Key")
	assert.NotContains(t, capturedHeader, "X-Didww-Api-Version")
	assert.Equal(t, "didww-go-sdk/"+sdkVersion, capturedHeader.Get("User-Agent"))
}

func TestDownloadAndDecompressExport(t *testing.T) {
	gzData := loadFixture(t, "exports/download.csv.gz")

	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(gzData)
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)

	var buf bytes.Buffer
	err := client.DownloadAndDecompressExport(context.Background(), server.URL+"/v3/exports/02bf6df4.csv.gz", &buf)
	require.NoError(t, err)

	content := buf.String()
	assert.Contains(t, content, "Date/Time Start (UTC)")
	assert.Contains(t, content, "972397239159652")
}

func TestDownloadExportError(t *testing.T) {
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)

	var buf bytes.Buffer
	err := client.DownloadExport(context.Background(), server.URL+"/v3/exports/missing.csv.gz", &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "not found", apiErr.Errors[0].Detail)
	assert.Empty(t, buf.Bytes())
}

func TestDownloadExportJSONAPIError(t *testing.T) {
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"title":"Not found","detail":"Export file not found","status":"404"}]}`))
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)

	err := client.DownloadExport(context.Background(), server.URL+"/v3/exports/missing.csv.gz", io.Discard)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "Export file not found", apiErr.Errors[0].Detail)
}

func TestDownloadExportDoesNotFollowRedirects(t *testing.T) {
	var redirectedHits int
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/elsewhere/export.csv.gz" {
			redirectedHits++
			w.Write(loadFixture(t, "exports/download.csv.gz"))
			return
		}
		http.Redirect(w, r, "/elsewhere/export.csv.gz", http.StatusFound)
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)

	var buf bytes.Buffer
	err := client.DownloadExport(context.Background(), server.URL+"/v3/exports/02bf6df4.csv.gz", &buf)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusFound, apiErr.HTTPStatus)
	assert.Equal(t, 0, redirectedHits)
	assert.Empty(t, buf.Bytes())
}

func TestDownloadExportNoContent(t *testing.T) {
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)

	err := client.DownloadExport(context.Background(), server.URL+"/v3/exports/02bf6df4.csv.gz", io.Discard)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNoContent, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "Unexpected status", apiErr.Errors[0].Title)
}

func TestDownloadExportRetriedAfter401(t *testing.T) {
	gzData := loadFixture(t, "exports/download.csv.gz")
	var log requestLog
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		if r.Header.Get("Authorization") == "Bearer token-1" {
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(unauthorizedBody))
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(gzData)
	})
	client := newTestClient(t, srv.URL)

	var buf bytes.Buffer
	err := client.DownloadExport(context.Background(), srv.URL+"/v3/exports/02bf6df4.csv.gz", &buf)
	require.NoError(t, err)

	assert.Equal(t, gzData, buf.Bytes())
	assert.Len(t, log.all(), 2)
	assert.Equal(t, 2, srv.mints())
	assert.JSONEq(t, `{"aud":["api.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(0)))
	assert.JSONEq(t, `{"aud":["api.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(1)))
}

func TestDownloadExportRefusesURLOutsideBaseURL(t *testing.T) {
	var otherHits int
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otherHits++
		w.Write(loadFixture(t, "exports/download.csv.gz"))
	}))
	defer other.Close()

	var apiHits int
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		apiHits++
		w.Write(loadFixture(t, "exports/download.csv.gz"))
	})
	httpsBase := "https://" + strings.TrimPrefix(srv.URL, "http://")

	tests := []struct {
		name    string
		baseURL string
		url     string
		call    func(c *Client, u string, w io.Writer) error
	}{
		{"other host", srv.URL, other.URL + "/v3/exports/02bf6df4.csv.gz", downloadExport},
		{"other host decompressed", srv.URL, other.URL + "/v3/exports/02bf6df4.csv.gz", downloadAndDecompressExport},
		{"http on an https base URL", httpsBase, srv.URL + "/v3/exports/02bf6df4.csv.gz", downloadExport},
		{"relative URL", srv.URL, "/v3/exports/02bf6df4.csv.gz", downloadExport},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, srv.URL, WithBaseURL(tt.baseURL))
			var buf bytes.Buffer

			err := tt.call(client, tt.url, &buf)

			var clientErr *ClientError
			require.ErrorAs(t, err, &clientErr)
			assert.Contains(t, clientErr.Message, "does not match the API base URL")
			assert.Empty(t, buf.Bytes())
		})
	}
	assert.Equal(t, 0, otherHits)
	assert.Equal(t, 0, apiHits)
	assert.Equal(t, 0, srv.mints())
}

func TestCheckExportURLOrigin(t *testing.T) {
	const proxy = "https://proxy.example.com/v3"
	tests := []struct {
		name    string
		baseURL string
		url     string
		allowed bool
	}{
		{"base URL host", proxy, "https://proxy.example.com/v3/exports/1.csv.gz", true},
		{"environment host through a proxy base URL", proxy, "https://api.didww.com/v3/exports/1.csv.gz", true},
		{"environment host with the default port", proxy, "https://api.didww.com:443/v3/exports/1.csv.gz", true},
		{"environment host in upper case with a trailing dot", proxy, "https://API.didww.com./v3/exports/1.csv.gz", true},
		{"base URL with the default port", "https://proxy.example.com:443/v3", "https://proxy.example.com/v3/exports/1.csv.gz", true},
		{"base URL with a trailing dot", "https://proxy.example.com./v3", "https://proxy.example.com/v3/exports/1.csv.gz", true},
		{"http base URL with the default port", "http://proxy.example.com/v3", "http://proxy.example.com:80/v3/exports/1.csv.gz", true},
		{"third host", proxy, "https://other.example.com/v3/exports/1.csv.gz", false},
		{"other environment host", proxy, "https://sandbox-api.didww.com/v3/exports/1.csv.gz", false},
		{"environment host over http", proxy, "http://api.didww.com/v3/exports/1.csv.gz", false},
		{"environment host on another port", proxy, "https://api.didww.com:8443/v3/exports/1.csv.gz", false},
		{"base URL host on another port", proxy, "https://proxy.example.com:8443/v3/exports/1.csv.gz", false},
		{"default https port over http", proxy, "http://proxy.example.com:443/v3/exports/1.csv.gz", false},
		{"relative URL", proxy, "/v3/exports/1.csv.gz", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient("test-api-key", WithEnvironment(Production), WithBaseURL(tt.baseURL), WithLogger(nil))
			require.NoError(t, err)

			err = client.checkExportURL(tt.url)

			if tt.allowed {
				assert.NoError(t, err)
				return
			}
			var clientErr *ClientError
			require.ErrorAs(t, err, &clientErr)
			assert.Contains(t, clientErr.Message, "does not match the API base URL")
		})
	}
}

func downloadExport(c *Client, u string, w io.Writer) error {
	return c.DownloadExport(context.Background(), u, w)
}

func downloadAndDecompressExport(c *Client, u string, w io.Writer) error {
	return c.DownloadAndDecompressExport(context.Background(), u, w)
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestDownloadExportReturnsWriteError(t *testing.T) {
	server := httptest.NewServer(withTokenEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(loadFixture(t, "exports/download.csv.gz"))
	})))
	defer server.Close()

	client := newTestClient(t, server.URL)
	writeErr := errors.New("disk full")

	err := client.DownloadExport(context.Background(), server.URL+"/v3/exports/02bf6df4.csv.gz", failingWriter{err: writeErr})

	assert.Same(t, writeErr, err)
}
