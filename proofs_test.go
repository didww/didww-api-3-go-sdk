package didww

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestProofsCreateWithProofTypeAndFiles(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/proofs": {status: http.StatusCreated, fixture: "proofs/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	proof, err := server.client.Proofs().Create(context.Background(), &Proof{
		ProofTypeID: "19cd7b22-559b-41d4-99c9-7ad7ad63d5d1",
		FileIDs:     []string{"254b3c2d-c40c-4ff7-93b1-a677aee7fa10"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proof.ID == "" {
		t.Error("expected non-empty ID after creation")
	}

	assertRequestJSON(t, capturedBody, "proofs/create_request.json")
}

func TestProofsCreateWithIdentityEntity(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/proofs": {status: http.StatusCreated, fixture: "proofs/create_with_identity.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.Proofs().Create(context.Background(), &Proof{
		ProofTypeID: "d2c1b3fb-29f7-46ca-ba82-b617f4630b78",
		EntityID:    "54c92d8e-f135-4b55-ac48-748d44437509",
		EntityType:  "identities",
		FileIDs:     []string{"cc52b6b3-0627-47d3-a1c9-b54d3de42813"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "proofs/create_with_identity_request.json")
}

func TestProofsCreateWithAddressEntity(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/proofs": {status: http.StatusCreated, fixture: "proofs/create_with_address.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.Proofs().Create(context.Background(), &Proof{
		ProofTypeID: "d2c1b3fb-29f7-46ca-ba82-b617f4630b78",
		EntityID:    "54c92d8e-f135-4b55-ac48-748d44437509",
		EntityType:  "addresses",
		FileIDs:     []string{"cc52b6b3-0627-47d3-a1c9-b54d3de42813"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "proofs/create_with_address_request.json")
}

func TestProofsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/proofs": {status: http.StatusOK, fixture: "proofs/index.json"},
	})

	proofs, err := client.Proofs().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(proofs) == 0 {
		t.Fatal("expected non-empty proofs list")
	}

	// Verify entity relationship is parsed from response
	proof := proofs[0]
	if proof.EntityType != "identities" {
		t.Errorf("expected entity type 'identities', got %q", proof.EntityType)
	}
	if proof.EntityID != "54c92d8e-f135-4b55-ac48-748d44437509" {
		t.Errorf("expected entity ID '54c92d8e-f135-4b55-ac48-748d44437509', got %q", proof.EntityID)
	}
	if proof.ProofTypeID != "19cd7b22-559b-41d4-99c9-7ad7ad63d5d1" {
		t.Errorf("expected proof_type ID '19cd7b22-559b-41d4-99c9-7ad7ad63d5d1', got %q", proof.ProofTypeID)
	}
}

func TestProofsCreateResponseParsesProofTypeFromRelationships(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/proofs": {status: http.StatusCreated, fixture: "proofs/create.json"},
	})

	proof, err := client.Proofs().Create(context.Background(), &Proof{
		ProofTypeID: "19cd7b22-559b-41d4-99c9-7ad7ad63d5d1",
		EntityID:    "some-entity-id",
		EntityType:  "identities",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// create.json has proof_type with data but entity without data
	if proof.ProofTypeID != "19cd7b22-559b-41d4-99c9-7ad7ad63d5d1" {
		t.Errorf("expected proof_type ID '19cd7b22-559b-41d4-99c9-7ad7ad63d5d1', got %q", proof.ProofTypeID)
	}
	// Entity has only links (no data) in create.json, so should be empty
	if proof.EntityID != "" {
		t.Errorf("expected empty entity ID (no data in response), got %q", proof.EntityID)
	}
}
