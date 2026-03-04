package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicKeysList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/public_keys": {status: http.StatusOK, fixture: "public_keys/index.json"},
	})

	keys, err := client.PublicKeys().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, keys, 2)
	assert.Equal(t, "dcf2bfcb-a1d0-3b58-bbf0-3ec22a510ba8", keys[0].ID)
	assert.Contains(t, keys[0].Key, "-----BEGIN PUBLIC KEY-----")
	assert.Equal(t, "f40e1176-a4ff-36e6-b2ed-c2c2d18097a3", keys[1].ID)
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
	require.NoError(t, err)

	assert.Equal(t, "", capturedAuth)
	assert.Equal(t, apiVersion, capturedAPIVersion)
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
	require.NoError(t, err)

	assert.Equal(t, "test-api-key", capturedAuth)
	assert.Equal(t, apiVersion, capturedAPIVersion)
}
