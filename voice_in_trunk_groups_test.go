package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
)

func intPtr(v int) *int {
	return &v
}

func TestVoiceInTrunkGroupsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_in_trunk_groups": {status: http.StatusOK, fixture: "voice_in_trunk_groups/index.json"},
	})

	groups, err := client.VoiceInTrunkGroups().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, groups)
}

func TestVoiceInTrunkGroupsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/voice_in_trunk_groups": {status: http.StatusCreated, fixture: "voice_in_trunk_groups/create.json"},
	})

	group, err := server.client.VoiceInTrunkGroups().Create(context.Background(), &resource.VoiceInTrunkGroup{
		Name:            "trunk group sample with 2 trunks",
		CapacityLimit:   intPtr(1000),
		VoiceInTrunkIDs: []string{"7c15bca2-7f17-46fb-9486-7e2a17158c7e", "b07a4cab-48c6-4b3a-9670-11b90b81bdef"},
	})
	require.NoError(t, err)

	assert.Equal(t, "b2319703-ce6c-480d-bb53-614e7abcfc96", group.ID)
	assert.Equal(t, "trunk group sample with 2 trunks", group.Name)

	// Verify included voice_in_trunks
	require.Len(t, group.VoiceInTrunks, 2)
	assert.Equal(t, "test custom11", group.VoiceInTrunks[0].Name)

	assertRequestJSON(t, *capturedBodyPtr, "voice_in_trunk_groups/create_request.json")
}

func TestVoiceInTrunkGroupsUpdate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/voice_in_trunk_groups/b2319703-ce6c-480d-bb53-614e7abcfc96": {status: http.StatusOK, fixture: "voice_in_trunk_groups/update.json"},
	})

	group, err := client.VoiceInTrunkGroups().Update(context.Background(), &resource.VoiceInTrunkGroup{
		ID:            "b2319703-ce6c-480d-bb53-614e7abcfc96",
		Name:          "trunk group sample updated with 2 trunks",
		CapacityLimit: intPtr(500),
	})
	require.NoError(t, err)

	assert.Equal(t, "b2319703-ce6c-480d-bb53-614e7abcfc96", group.ID)
	assert.Equal(t, "trunk group sample updated with 2 trunks", group.Name)
	require.NotNil(t, group.CapacityLimit)
	assert.Equal(t, 500, *group.CapacityLimit)
}

func TestVoiceInTrunkGroupsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/voice_in_trunk_groups/b2319703-ce6c-480d-bb53-614e7abcfc96": {status: http.StatusNoContent},
	})

	err := client.VoiceInTrunkGroups().Delete(context.Background(), "b2319703-ce6c-480d-bb53-614e7abcfc96")
	require.NoError(t, err)
}
