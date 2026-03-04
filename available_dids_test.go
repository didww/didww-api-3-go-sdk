package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvailableDIDsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/available_dids": {status: http.StatusOK, fixture: "available_dids/index.json"},
	})

	dids, err := client.AvailableDIDs().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, dids)
}

func TestAvailableDIDsFindWithIncludedDIDGroup(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/available_dids/0b76223b-9625-412f-b0f3-330551473e7e": {status: http.StatusOK, fixture: "available_dids/show.json"},
	})

	params := NewQueryParams().Include("did_group")
	did, err := client.AvailableDIDs().Find(context.Background(), "0b76223b-9625-412f-b0f3-330551473e7e", params)
	require.NoError(t, err)

	assert.Equal(t, "0b76223b-9625-412f-b0f3-330551473e7e", did.ID)
	assert.Equal(t, "16169886810", did.Number)
	require.NotNil(t, did.DIDGroup)
	assert.Equal(t, "a9e3d346-d7bc-4a85-adb0-8ef1119cf237", did.DIDGroup.ID)
	assert.Equal(t, "Grand Rapids", did.DIDGroup.AreaName)
}

func TestAvailableDIDsListWithNanpaPrefix(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/available_dids": {status: http.StatusOK, fixture: "available_dids/index_with_nanpa.json"},
	})

	dids, err := client.AvailableDIDs().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, dids, 1)
	assert.Equal(t, "aa13b01c-36c8-405c-b5a8-1427aa7966ea", dids[0].ID)
	assert.Equal(t, "18649204444", dids[0].Number)
}

func TestAvailableDIDsFindWithNanpaPrefix(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/available_dids/ID": {status: http.StatusOK, fixture: "available_dids/show_with_nanpa_prefix.json"},
	})

	params := NewQueryParams().Include("nanpa_prefix")
	did, err := client.AvailableDIDs().Find(context.Background(), "ID", params)
	require.NoError(t, err)

	require.NotNil(t, did.NanpaPrefix)
	assert.Equal(t, "201", did.NanpaPrefix.NPA)
	assert.Equal(t, "221", did.NanpaPrefix.NXX)
}
