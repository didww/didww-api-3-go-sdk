package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/pops": {status: http.StatusOK, fixture: "pops/index.json"},
	})

	pops, err := client.Pops().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, pops, 2)

	assert.NotEmpty(t, pops[0].ID)
	assert.NotEmpty(t, pops[0].Name)
}
