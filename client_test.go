package didww

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSendsCorrectHeaders(t *testing.T) {
	var (
		receivedContentType string
		receivedAccept      string
		receivedAPIKey      string
		receivedAPIVersion  string
		receivedUserAgent   string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		receivedAccept = r.Header.Get("Accept")
		receivedAPIKey = r.Header.Get("Api-Key")
		receivedAPIVersion = r.Header.Get("X-DIDWW-API-Version")
		receivedUserAgent = r.Header.Get("User-Agent")

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		w.Write(loadFixture(t, "balance/index.json"))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Balance().Find(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedContentType != "application/vnd.api+json" {
		t.Errorf("expected Content-Type 'application/vnd.api+json', got %q", receivedContentType)
	}
	if receivedAccept != "application/vnd.api+json" {
		t.Errorf("expected Accept 'application/vnd.api+json', got %q", receivedAccept)
	}
	if receivedAPIKey != "test-api-key" {
		t.Errorf("expected Api-Key 'test-api-key', got %q", receivedAPIKey)
	}
	if receivedAPIVersion != apiVersion {
		t.Errorf("expected X-DIDWW-API-Version %q, got %q", apiVersion, receivedAPIVersion)
	}
	if receivedUserAgent == "" {
		t.Error("expected non-empty User-Agent header")
	}
}

func TestClientHandles404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"title":"not found","detail":"Resource not found","status":"404"}]}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Countries().Find(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.HTTPStatus != http.StatusNotFound {
		t.Errorf("expected HTTP status 404, got %d", apiErr.HTTPStatus)
	}
}

func TestClientHandles500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"title":"server error","detail":"Internal server error","status":"500"}]}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.Balance().Find(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("expected HTTP status 500, got %d", apiErr.HTTPStatus)
	}
}

func TestClientHandles422(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"errors":[{"title":"is invalid","detail":"name - is invalid","code":"100","source":{"pointer":"/data/attributes/name"},"status":"422"}]}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.VoiceInTrunks().Create(context.Background(), &VoiceInTrunk{Name: "test"})
	if err == nil {
		t.Fatal("expected error for 422 response")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.HTTPStatus != http.StatusUnprocessableEntity {
		t.Errorf("expected HTTP status 422, got %d", apiErr.HTTPStatus)
	}
	if len(apiErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(apiErr.Errors))
	}
	if apiErr.Errors[0].Code != "100" {
		t.Errorf("expected code '100', got %q", apiErr.Errors[0].Code)
	}
	if apiErr.Errors[0].Source.Pointer != "/data/attributes/name" {
		t.Errorf("expected source pointer '/data/attributes/name', got %q", apiErr.Errors[0].Source.Pointer)
	}
}

func TestClientWithQueryParamsAppendedToURL(t *testing.T) {
	var requestURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURL = r.URL.String()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		w.Write(loadFixture(t, "countries/index.json"))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params := NewQueryParams().
		Filter("prefix", "44").
		Sort("name").
		Include("regions").
		Page(1, 25)

	_, err = client.Countries().List(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify query params are appended to URL
	assertContains(t, requestURL, "filter[prefix]=44")
	assertContains(t, requestURL, "sort=name")
	assertContains(t, requestURL, "include=regions")
	assertContains(t, requestURL, "page[number]=1")
	assertContains(t, requestURL, "page[size]=25")
}
