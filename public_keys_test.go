package didww

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestPublicKeysFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/public_keys": {status: http.StatusOK, fixture: "public_keys/index.json"},
	})

	key, err := client.PublicKeys().Find(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key.ID != "dcf2bfcb-a1d0-3b58-bbf0-3ec22a510ba8" {
		t.Errorf("expected ID 'dcf2bfcb-a1d0-3b58-bbf0-3ec22a510ba8', got %q", key.ID)
	}
	if !strings.HasPrefix(key.Key, "-----BEGIN PUBLIC KEY-----") {
		t.Errorf("expected key to start with '-----BEGIN PUBLIC KEY-----', got %q", key.Key[:30])
	}
}

func TestPublicKeysNoAuthHeader(t *testing.T) {
	var capturedAuth string
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/public_keys": {status: http.StatusOK, fixture: "public_keys/index.json"},
	}, func(r *http.Request) {
		capturedAuth = r.Header.Get("Api-Key")
	})

	_, err := server.client.PublicKeys().Find(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAuth != "" {
		t.Errorf("expected no Api-Key header for public_keys endpoint, got %q", capturedAuth)
	}
}

func TestNonPublicEndpointHasAuthHeader(t *testing.T) {
	var capturedAuth string
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/countries": {status: http.StatusOK, fixture: "countries/index.json"},
	}, func(r *http.Request) {
		capturedAuth = r.Header.Get("Api-Key")
	})

	_, err := server.client.Countries().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAuth != "test-api-key" {
		t.Errorf("expected Api-Key 'test-api-key', got %q", capturedAuth)
	}
}
