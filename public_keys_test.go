package didww

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestPublicKeysList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/public_keys": {status: http.StatusOK, fixture: "public_keys/index.json"},
	})

	keys, err := client.PublicKeys().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(keys) != 2 {
		t.Fatalf("expected 2 public keys, got %d", len(keys))
	}
	if keys[0].ID != "dcf2bfcb-a1d0-3b58-bbf0-3ec22a510ba8" {
		t.Errorf("expected ID 'dcf2bfcb-a1d0-3b58-bbf0-3ec22a510ba8', got %q", keys[0].ID)
	}
	if !strings.HasPrefix(keys[0].Key, "-----BEGIN PUBLIC KEY-----") {
		t.Errorf("expected key to start with '-----BEGIN PUBLIC KEY-----', got %q", keys[0].Key[:30])
	}
	if keys[1].ID != "f40e1176-a4ff-36e6-b2ed-c2c2d18097a3" {
		t.Errorf("expected ID 'f40e1176-a4ff-36e6-b2ed-c2c2d18097a3', got %q", keys[1].ID)
	}
}

func TestPublicKeysNoAuthHeader(t *testing.T) {
	var capturedAuth string
	var capturedAPIVersion string
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/public_keys": {status: http.StatusOK, fixture: "public_keys/index.json"},
	}, func(r *http.Request) {
		capturedAuth = r.Header.Get("Api-Key")
		capturedAPIVersion = r.Header.Get("X-DIDWW-API-Version")
	})

	_, err := server.client.PublicKeys().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAuth != "" {
		t.Errorf("expected no Api-Key header for public_keys endpoint, got %q", capturedAuth)
	}
	if capturedAPIVersion != apiVersion {
		t.Errorf("expected X-DIDWW-API-Version %q, got %q", apiVersion, capturedAPIVersion)
	}
}

func TestNonPublicEndpointHasAuthHeader(t *testing.T) {
	var capturedAuth string
	var capturedAPIVersion string
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/countries": {status: http.StatusOK, fixture: "countries/index.json"},
	}, func(r *http.Request) {
		capturedAuth = r.Header.Get("Api-Key")
		capturedAPIVersion = r.Header.Get("X-DIDWW-API-Version")
	})

	_, err := server.client.Countries().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAuth != "test-api-key" {
		t.Errorf("expected Api-Key 'test-api-key', got %q", capturedAuth)
	}
	if capturedAPIVersion != apiVersion {
		t.Errorf("expected X-DIDWW-API-Version %q, got %q", apiVersion, capturedAPIVersion)
	}
}
