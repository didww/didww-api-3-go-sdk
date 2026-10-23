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
	var capturedHeader http.Header
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Clone()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write(loadFixture(t, "public_keys/index.json"))
	})
	client := newTestClient(t, srv.URL)

	_, err := client.PublicKeys().List(context.Background(), nil)
	require.NoError(t, err)

	assert.NotContains(t, capturedHeader, "Authorization")
	assert.NotContains(t, capturedHeader, "Api-Key")
	assert.NotContains(t, capturedHeader, "X-Didww-Api-Version")
	assert.Equal(t, "didww-go-sdk/"+sdkVersion, capturedHeader.Get("User-Agent"))
	assert.Equal(t, 0, srv.mints())
}

func TestNonPublicEndpointHasAuthHeader(t *testing.T) {
	var capturedHeader http.Header
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/countries": {status: http.StatusOK, fixture: "countries/index.json"},
	}, func(r *http.Request) {
		capturedHeader = r.Header.Clone()
	})

	_, err := server.client.Countries().List(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "Bearer "+testAccessToken, capturedHeader.Get("Authorization"))
	assert.NotContains(t, capturedHeader, "Api-Key")
	assert.NotContains(t, capturedHeader, "X-Didww-Api-Version")
}
