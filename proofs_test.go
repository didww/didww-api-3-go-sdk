package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProofsCreateWithProofTypeAndFiles(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/proofs": {status: http.StatusCreated, fixture: "proofs/create.json"},
	})

	proof, err := server.client.Proofs().Create(context.Background(), &resource.Proof{
		ProofTypeID: "19cd7b22-559b-41d4-99c9-7ad7ad63d5d1",
		FileIDs:     []string{"254b3c2d-c40c-4ff7-93b1-a677aee7fa10"},
	})
	require.NoError(t, err)

	assert.NotEmpty(t, proof.ID)

	assertRequestJSON(t, *capturedBodyPtr, "proofs/create_request.json")
}

func TestProofsCreateWithIdentityEntity(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/proofs": {status: http.StatusCreated, fixture: "proofs/create_with_identity.json"},
	})

	_, err := server.client.Proofs().Create(context.Background(), &resource.Proof{
		ProofTypeID: "d2c1b3fb-29f7-46ca-ba82-b617f4630b78",
		EntityID:    "54c92d8e-f135-4b55-ac48-748d44437509",
		EntityType:  "identities",
		FileIDs:     []string{"cc52b6b3-0627-47d3-a1c9-b54d3de42813"},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "proofs/create_with_identity_request.json")
}

func TestProofsCreateWithAddressEntity(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/proofs": {status: http.StatusCreated, fixture: "proofs/create_with_address.json"},
	})

	_, err := server.client.Proofs().Create(context.Background(), &resource.Proof{
		ProofTypeID: "d2c1b3fb-29f7-46ca-ba82-b617f4630b78",
		EntityID:    "54c92d8e-f135-4b55-ac48-748d44437509",
		EntityType:  "addresses",
		FileIDs:     []string{"cc52b6b3-0627-47d3-a1c9-b54d3de42813"},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "proofs/create_with_address_request.json")
}

func TestProofsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/proofs": {status: http.StatusOK, fixture: "proofs/index.json"},
	})

	proofs, err := client.Proofs().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, proofs)

	// Verify entity relationship is parsed from response
	proof := proofs[0]
	assert.Equal(t, "identities", proof.EntityType)
	assert.Equal(t, "54c92d8e-f135-4b55-ac48-748d44437509", proof.EntityID)
	assert.Equal(t, "19cd7b22-559b-41d4-99c9-7ad7ad63d5d1", proof.ProofTypeID)
}

func TestProofsCreateResponseParsesProofTypeFromRelationships(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/proofs": {status: http.StatusCreated, fixture: "proofs/create.json"},
	})

	proof, err := client.Proofs().Create(context.Background(), &resource.Proof{
		ProofTypeID: "19cd7b22-559b-41d4-99c9-7ad7ad63d5d1",
		EntityID:    "some-entity-id",
		EntityType:  "identities",
	})
	require.NoError(t, err)

	// create.json has proof_type with data but entity without data
	assert.Equal(t, "19cd7b22-559b-41d4-99c9-7ad7ad63d5d1", proof.ProofTypeID)
	// Entity has only links (no data) in create.json, so should be empty
	assert.Equal(t, "", proof.EntityID)
}
