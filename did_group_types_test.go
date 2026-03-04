package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDIDGroupTypesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_group_types": {status: http.StatusOK, fixture: "did_group_types/index.json"},
	})

	types, err := client.DIDGroupTypes().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, types, 6)
}

func TestDIDGroupTypesFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_group_types/d6530a8c-924c-469a-98c0-9525602e6192": {status: http.StatusOK, fixture: "did_group_types/show.json"},
	})

	dgt, err := client.DIDGroupTypes().Find(context.Background(), "d6530a8c-924c-469a-98c0-9525602e6192")
	require.NoError(t, err)

	assert.Equal(t, "d6530a8c-924c-469a-98c0-9525602e6192", dgt.ID)
	assert.Equal(t, "Global", dgt.Name)
}
