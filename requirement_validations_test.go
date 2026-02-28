package didww

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestRequirementValidationsCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/requirement_validations": {status: http.StatusCreated, fixture: "requirement_validations/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	rv, err := server.client.RequirementValidations().Create(context.Background(), &RequirementValidation{
		AddressID:     "d3414687-40f4-4346-a267-c2c65117d28c",
		RequirementID: "aea92b24-a044-4864-9740-89d3e15b65c7",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rv.ID == "" {
		t.Error("expected non-empty ID")
	}

	assertRequestJSON(t, capturedBody, "requirement_validations/create_request.json")
}

func TestRequirementValidationsCreateError(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/requirement_validations": {status: http.StatusUnprocessableEntity, fixture: "requirement_validations/create_error_validation.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.RequirementValidations().Create(context.Background(), &RequirementValidation{
		IdentityID:    "5e9df058-50d2-4e34-b0d4-d1746b86f41a",
		AddressID:     "d3414687-40f4-4346-a267-c2c65117d28c",
		RequirementID: "2efc3427-8ba6-4d50-875d-f2de4a068de8",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if len(apiErr.Errors) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(apiErr.Errors))
	}

	assertRequestJSON(t, capturedBody, "requirement_validations/create_request_failed.json")
}
