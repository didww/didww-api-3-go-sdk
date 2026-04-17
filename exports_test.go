package didww

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"

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

func TestDownloadExport(t *testing.T) {
	gzData := loadFixture(t, "exports/download.csv.gz")

	var capturedAuth string
	var capturedAPIVersion string
	var capturedUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Api-Key")
		capturedAPIVersion = r.Header.Get("X-DIDWW-API-Version")
		capturedUserAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(gzData)
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	var buf bytes.Buffer
	err = client.DownloadExport(context.Background(), server.URL+"/v3/exports/02bf6df4.csv.gz", &buf)
	require.NoError(t, err)

	assert.Equal(t, gzData, buf.Bytes())
	assert.Equal(t, "test-api-key", capturedAuth)
	assert.Equal(t, apiVersion, capturedAPIVersion)
	assert.Equal(t, "didww-go-sdk/"+sdkVersion, capturedUserAgent)
}

func TestDownloadAndDecompressExport(t *testing.T) {
	gzData := loadFixture(t, "exports/download.csv.gz")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(gzData)
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	var buf bytes.Buffer
	err = client.DownloadAndDecompressExport(context.Background(), server.URL+"/v3/exports/02bf6df4.csv.gz", &buf)
	require.NoError(t, err)

	content := buf.String()
	assert.Contains(t, content, "Date/Time Start (UTC)")
	assert.Contains(t, content, "972397239159652")
}

func TestDownloadExportError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	var buf bytes.Buffer
	err = client.DownloadExport(context.Background(), server.URL+"/v3/exports/missing.csv.gz", &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
}
