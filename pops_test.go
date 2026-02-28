package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestPopsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/pops": {status: http.StatusOK, fixture: "pops/index.json"},
	})

	pops, err := client.Pops().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pops) != 2 {
		t.Fatalf("expected 2 pops, got %d", len(pops))
	}

	if pops[0].ID == "" {
		t.Error("expected non-empty ID")
	}
	if pops[0].Name == "" {
		t.Error("expected non-empty Name")
	}
}
