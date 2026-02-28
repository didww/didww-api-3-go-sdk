package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestNanpaPrefixesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/nanpa_prefixes": {status: http.StatusOK, fixture: "nanpa_prefixes/index.json"},
	})

	prefixes, err := client.NanpaPrefixes().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prefixes) == 0 {
		t.Fatal("expected non-empty nanpa prefixes list")
	}
}

func TestNanpaPrefixesFindWithIncludedCountry(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/nanpa_prefixes/6c16d51d-d376-4395-91c4-012321317e48": {status: http.StatusOK, fixture: "nanpa_prefixes/show.json"},
	})

	params := NewQueryParams().Include("country")
	prefix, err := client.NanpaPrefixes().Find(context.Background(), "6c16d51d-d376-4395-91c4-012321317e48", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if prefix.ID != "6c16d51d-d376-4395-91c4-012321317e48" {
		t.Errorf("expected ID '6c16d51d-d376-4395-91c4-012321317e48', got %q", prefix.ID)
	}
	if prefix.NPA != "864" {
		t.Errorf("expected NPA '864', got %q", prefix.NPA)
	}
	if prefix.NXX != "920" {
		t.Errorf("expected NXX '920', got %q", prefix.NXX)
	}
	if prefix.Country == nil {
		t.Fatal("expected non-nil Country")
	}
	if prefix.Country.Name != "United States" {
		t.Errorf("expected country name 'United States', got %q", prefix.Country.Name)
	}
}
