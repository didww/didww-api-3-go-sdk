package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

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
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/exports": {status: http.StatusCreated, fixture: "exports/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	export, err := server.client.Exports().Create(context.Background(), &Export{
		ExportType: enums.ExportTypeCdrIn,
		Filters: map[string]interface{}{
			"did_number": "1234556789",
			"year":       "2019",
			"month":      "01",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "da15f006-5da4-45ca-b0df-735baeadf423", export.ID)
	assert.Equal(t, enums.ExportStatusPending, export.Status)
	assert.Equal(t, enums.ExportTypeCdrIn, export.ExportType)

	assertRequestJSON(t, capturedBody, "exports/create_request.json")
}

func TestExportsCreateCdrOut(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/exports": {status: http.StatusCreated, fixture: "exports/create_cdr_out.json"},
	})

	export, err := client.Exports().Create(context.Background(), &Export{
		ExportType: enums.ExportTypeCdrOut,
		Filters: map[string]interface{}{
			"year":  "2019",
			"month": "01",
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

	_, err := client.Exports().Create(context.Background(), &Export{
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
