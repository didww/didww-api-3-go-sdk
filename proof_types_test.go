package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProofTypesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/proof_types": {status: http.StatusOK, fixture: "proof_types/index.json"},
	})

	proofTypes, err := client.ProofTypes().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, proofTypes, 5)

	first := proofTypes[0]
	assert.NotEmpty(t, first.ID)
	assert.NotEmpty(t, first.Name)
	assert.NotEmpty(t, first.EntityType)
}
