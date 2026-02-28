package didww

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestApiErrorParsing(t *testing.T) {
	body := `{
		"errors": [
			{
				"title": "is invalid",
				"detail": "voice_in_trunk_group - is invalid",
				"code": "100",
				"source": {
					"pointer": "/data/attributes/voice_in_trunk_group_id"
				},
				"status": "422"
			}
		]
	}`

	apiErr, err := ParseAPIErrors([]byte(body), http.StatusUnprocessableEntity)
	if err != nil {
		t.Fatalf("unexpected error parsing API errors: %v", err)
	}

	if apiErr.HTTPStatus != http.StatusUnprocessableEntity {
		t.Errorf("expected HTTP status 422, got %d", apiErr.HTTPStatus)
	}

	if len(apiErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(apiErr.Errors))
	}

	e := apiErr.Errors[0]
	if e.Title != "is invalid" {
		t.Errorf("expected title 'is invalid', got %q", e.Title)
	}
	if e.Detail != "voice_in_trunk_group - is invalid" {
		t.Errorf("expected detail 'voice_in_trunk_group - is invalid', got %q", e.Detail)
	}
	if e.Code != "100" {
		t.Errorf("expected code '100', got %q", e.Code)
	}
	if e.Source.Pointer != "/data/attributes/voice_in_trunk_group_id" {
		t.Errorf("expected source pointer '/data/attributes/voice_in_trunk_group_id', got %q", e.Source.Pointer)
	}
	if e.Status != "422" {
		t.Errorf("expected status '422', got %q", e.Status)
	}
}

func TestApiErrorMultipleErrors(t *testing.T) {
	body := `{
		"errors": [
			{
				"title": "can't be blank",
				"detail": "name - can't be blank",
				"source": {"pointer": "/data/attributes/name"},
				"status": "422"
			},
			{
				"title": "is invalid",
				"detail": "configuration - is invalid",
				"source": {"pointer": "/data/attributes/configuration"},
				"status": "422"
			}
		]
	}`

	apiErr, err := ParseAPIErrors([]byte(body), http.StatusUnprocessableEntity)
	if err != nil {
		t.Fatalf("unexpected error parsing API errors: %v", err)
	}

	if len(apiErr.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(apiErr.Errors))
	}

	if apiErr.Errors[0].Detail != "name - can't be blank" {
		t.Errorf("expected first error detail 'name - can't be blank', got %q", apiErr.Errors[0].Detail)
	}
	if apiErr.Errors[1].Detail != "configuration - is invalid" {
		t.Errorf("expected second error detail 'configuration - is invalid', got %q", apiErr.Errors[1].Detail)
	}
}

func TestApiErrorWithoutCode(t *testing.T) {
	body := `{
		"errors": [
			{
				"title": "not found",
				"detail": "Resource not found",
				"status": "404"
			}
		]
	}`

	apiErr, err := ParseAPIErrors([]byte(body), http.StatusNotFound)
	if err != nil {
		t.Fatalf("unexpected error parsing API errors: %v", err)
	}

	if apiErr.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected HTTP status 404, got %d", apiErr.HTTPStatus)
	}

	if apiErr.Errors[0].Code != "" {
		t.Errorf("expected empty code, got %q", apiErr.Errors[0].Code)
	}
}

func TestApiErrorWithoutSourcePointer(t *testing.T) {
	body := `{
		"errors": [
			{
				"title": "server error",
				"detail": "Internal server error",
				"status": "500"
			}
		]
	}`

	apiErr, err := ParseAPIErrors([]byte(body), http.StatusInternalServerError)
	if err != nil {
		t.Fatalf("unexpected error parsing API errors: %v", err)
	}

	if apiErr.Errors[0].Source.Pointer != "" {
		t.Errorf("expected empty source pointer, got %q", apiErr.Errors[0].Source.Pointer)
	}
}

func TestApiErrorEmptyBody(t *testing.T) {
	_, err := ParseAPIErrors([]byte(""), http.StatusInternalServerError)
	if err == nil {
		t.Fatal("expected error when parsing empty body")
	}
}

func TestApiErrorInvalidJSON(t *testing.T) {
	_, err := ParseAPIErrors([]byte("not json"), http.StatusInternalServerError)
	if err == nil {
		t.Fatal("expected error when parsing invalid JSON")
	}
}

func TestApiErrorImplementsError(t *testing.T) {
	apiErr := &APIError{
		HTTPStatus: 422,
		Errors: []ErrorDetail{
			{Title: "is invalid", Detail: "name - is invalid"},
		},
	}

	errMsg := apiErr.Error()
	if errMsg == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestClientError(t *testing.T) {
	err := &ClientError{Message: "connection timeout"}
	if err.Error() != "connection timeout" {
		t.Errorf("expected 'connection timeout', got %q", err.Error())
	}
}

func TestErrorDetailJSONRoundTrip(t *testing.T) {
	original := ErrorDetail{
		Title:  "is invalid",
		Detail: "name - is invalid",
		Code:   "100",
		Status: "422",
		Source: ErrorSource{Pointer: "/data/attributes/name"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("unexpected error marshalling: %v", err)
	}

	var decoded ErrorDetail
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("unexpected error unmarshalling: %v", err)
	}

	if decoded.Title != original.Title {
		t.Errorf("expected title %q, got %q", original.Title, decoded.Title)
	}
	if decoded.Detail != original.Detail {
		t.Errorf("expected detail %q, got %q", original.Detail, decoded.Detail)
	}
	if decoded.Code != original.Code {
		t.Errorf("expected code %q, got %q", original.Code, decoded.Code)
	}
	if decoded.Status != original.Status {
		t.Errorf("expected status %q, got %q", original.Status, decoded.Status)
	}
	if decoded.Source.Pointer != original.Source.Pointer {
		t.Errorf("expected source pointer %q, got %q", original.Source.Pointer, decoded.Source.Pointer)
	}
}
