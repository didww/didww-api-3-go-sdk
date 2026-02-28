package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestProofTypesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/proof_types": {status: http.StatusOK, fixture: "proof_types/index.json"},
	})

	proofTypes, err := client.ProofTypes().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(proofTypes) != 5 {
		t.Fatalf("expected 5 proof types, got %d", len(proofTypes))
	}

	first := proofTypes[0]
	if first.ID == "" {
		t.Error("expected non-empty ID")
	}
	if first.Name == "" {
		t.Error("expected non-empty Name")
	}
	if first.EntityType == "" {
		t.Error("expected non-empty EntityType")
	}
}
