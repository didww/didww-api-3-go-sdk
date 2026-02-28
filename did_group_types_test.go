package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestDIDGroupTypesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_group_types": {status: http.StatusOK, fixture: "did_group_types/index.json"},
	})

	types, err := client.DIDGroupTypes().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(types) != 6 {
		t.Fatalf("expected 6 did group types, got %d", len(types))
	}
}

func TestDIDGroupTypesFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_group_types/d6530a8c-924c-469a-98c0-9525602e6192": {status: http.StatusOK, fixture: "did_group_types/show.json"},
	})

	dgt, err := client.DIDGroupTypes().Find(context.Background(), "d6530a8c-924c-469a-98c0-9525602e6192")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dgt.ID != "d6530a8c-924c-469a-98c0-9525602e6192" {
		t.Errorf("expected ID 'd6530a8c-924c-469a-98c0-9525602e6192', got %q", dgt.ID)
	}
	if dgt.Name != "Global" {
		t.Errorf("expected Name 'Global', got %q", dgt.Name)
	}
}
