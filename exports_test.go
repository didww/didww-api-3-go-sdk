package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func TestExportsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/exports": {status: http.StatusOK, fixture: "exports/index.json"},
	})

	exports, err := client.Exports().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(exports))
	}

	export := exports[0]
	if export.ID != "da15f006-5da4-45ca-b0df-735baeadf423" {
		t.Errorf("expected ID 'da15f006-5da4-45ca-b0df-735baeadf423', got %q", export.ID)
	}
	if export.Status != enums.ExportStatusCompleted {
		t.Errorf("expected Status 'Completed', got %q", export.Status)
	}
	if export.ExportType != enums.ExportTypeCdrIn {
		t.Errorf("expected ExportType 'cdr_in', got %q", export.ExportType)
	}
	if export.URL == nil || *export.URL == "" {
		t.Error("expected non-nil non-empty URL")
	}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if export.ID != "da15f006-5da4-45ca-b0df-735baeadf423" {
		t.Errorf("expected ID 'da15f006-5da4-45ca-b0df-735baeadf423', got %q", export.ID)
	}
	if export.Status != enums.ExportStatusPending {
		t.Errorf("expected Status 'Pending', got %q", export.Status)
	}
	if export.ExportType != enums.ExportTypeCdrIn {
		t.Errorf("expected ExportType 'cdr_in', got %q", export.ExportType)
	}

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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if export.ID != "da15f006-5da4-45ca-b0df-735baeadf423" {
		t.Errorf("expected ID 'da15f006-5da4-45ca-b0df-735baeadf423', got %q", export.ID)
	}
	if export.ExportType != enums.ExportTypeCdrOut {
		t.Errorf("expected ExportType 'cdr_out', got %q", export.ExportType)
	}
	if export.Status != enums.ExportStatusPending {
		t.Errorf("expected Status 'Pending', got %q", export.Status)
	}
}

func TestExportsCreateUnauthorized(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/exports": {status: http.StatusUnauthorized, fixture: "exports/create_error_unauthorized.json"},
	})

	_, err := client.Exports().Create(context.Background(), &Export{
		ExportType: enums.ExportTypeCdrIn,
	})
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("expected HTTP status 401, got %d", apiErr.HTTPStatus)
	}
	if len(apiErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(apiErr.Errors))
	}
	if apiErr.Errors[0].Title != "Unauthorized" {
		t.Errorf("expected title 'Unauthorized', got %q", apiErr.Errors[0].Title)
	}
}

func TestExportsFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/exports/da15f006-5da4-45ca-b0df-735baeadf423": {status: http.StatusOK, fixture: "exports/show.json"},
	})

	export, err := client.Exports().Find(context.Background(), "da15f006-5da4-45ca-b0df-735baeadf423")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if export.ID != "da15f006-5da4-45ca-b0df-735baeadf423" {
		t.Errorf("expected ID 'da15f006-5da4-45ca-b0df-735baeadf423', got %q", export.ID)
	}
	if export.Status != enums.ExportStatusCompleted {
		t.Errorf("expected Status 'Completed', got %q", export.Status)
	}
	if export.ExportType != enums.ExportTypeCdrIn {
		t.Errorf("expected ExportType 'cdr_in', got %q", export.ExportType)
	}
	if export.URL == nil || *export.URL == "" {
		t.Error("expected non-nil non-empty URL")
	}
}
