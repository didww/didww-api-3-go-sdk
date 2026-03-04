package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedCapacityGroupsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/shared_capacity_groups": {status: http.StatusOK, fixture: "shared_capacity_groups/index.json"},
	})

	groups, err := client.SharedCapacityGroups().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, groups, 4)
}

func TestSharedCapacityGroupsFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/shared_capacity_groups/89f987e2-0862-4bf4-a3f4-cdc89af0d875": {status: http.StatusOK, fixture: "shared_capacity_groups/show.json"},
	})

	params := NewQueryParams().Include("capacity_pool,dids")
	group, err := client.SharedCapacityGroups().Find(context.Background(), "89f987e2-0862-4bf4-a3f4-cdc89af0d875", params)
	require.NoError(t, err)

	assert.Equal(t, "89f987e2-0862-4bf4-a3f4-cdc89af0d875", group.ID)
	assert.Equal(t, "didww", group.Name)
	assert.Equal(t, 19, group.SharedChannelsCount)

	// Verify capacity pool is resolved
	require.NotNil(t, group.CapacityPool)
	assert.Equal(t, "f288d07c-e2fc-4ae6-9837-b18fb469c324", group.CapacityPool.ID)
}

func TestSharedCapacityGroupsCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/shared_capacity_groups": {status: http.StatusCreated, fixture: "shared_capacity_groups/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	group, err := server.client.SharedCapacityGroups().Create(context.Background(), &SharedCapacityGroup{
		Name:                 "java-sdk",
		SharedChannelsCount:  5,
		MeteredChannelsCount: 0,
		CapacityPoolID:       "f288d07c-e2fc-4ae6-9837-b18fb469c324",
	})
	require.NoError(t, err)

	assert.NotEmpty(t, group.ID)

	assertRequestJSON(t, capturedBody, "shared_capacity_groups/create_request.json")
}

func TestSharedCapacityGroupsCreateWithChannels(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/shared_capacity_groups": {status: http.StatusCreated, fixture: "shared_capacity_groups/create_with_channels.json"},
	})

	group, err := client.SharedCapacityGroups().Create(context.Background(), &SharedCapacityGroup{
		Name:                 "java-sdk",
		SharedChannelsCount:  5,
		MeteredChannelsCount: 0,
		CapacityPoolID:       "f288d07c-e2fc-4ae6-9837-b18fb469c324",
	})
	require.NoError(t, err)

	assert.Equal(t, "3688a9c3-354f-4e16-b458-1d2df9f02547", group.ID)
	assert.Equal(t, 5, group.SharedChannelsCount)
	assert.Equal(t, 0, group.MeteredChannelsCount)
}

func TestSharedCapacityGroupsCreateMissingPool(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/shared_capacity_groups": {status: http.StatusUnprocessableEntity, fixture: "shared_capacity_groups/create_error_missing_pool.json"},
	})

	_, err := client.SharedCapacityGroups().Create(context.Background(), &SharedCapacityGroup{
		Name: "missing pool",
	})
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError")
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "capacity_pool - can't be blank", apiErr.Errors[0].Detail)
}

func TestSharedCapacityGroupsUpdate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/shared_capacity_groups/89f987e2-0862-4bf4-a3f4-cdc89af0d875": {status: http.StatusOK, fixture: "shared_capacity_groups/update.json"},
	})

	group, err := client.SharedCapacityGroups().Update(context.Background(), &SharedCapacityGroup{
		ID:   "89f987e2-0862-4bf4-a3f4-cdc89af0d875",
		Name: "didww1",
	})
	require.NoError(t, err)

	assert.Equal(t, "89f987e2-0862-4bf4-a3f4-cdc89af0d875", group.ID)
	assert.Equal(t, "didww1", group.Name)
	assert.Equal(t, 10, group.SharedChannelsCount)
	assert.Equal(t, 2, group.MeteredChannelsCount)
}

func TestSharedCapacityGroupsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/shared_capacity_groups/89f987e2-0862-4bf4-a3f4-cdc89af0d875": {status: http.StatusNoContent},
	})

	err := client.SharedCapacityGroups().Delete(context.Background(), "89f987e2-0862-4bf4-a3f4-cdc89af0d875")
	require.NoError(t, err)
}
