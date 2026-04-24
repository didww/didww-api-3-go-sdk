package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
)

func TestCapacityPoolsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/capacity_pools": {status: http.StatusOK, fixture: "capacity_pools/index.json"},
	})

	pools, err := client.CapacityPools().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, pools, 2)
}

func TestCapacityPoolsFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/capacity_pools/f288d07c-e2fc-4ae6-9837-b18fb469c324": {status: http.StatusOK, fixture: "capacity_pools/show.json"},
	})

	params := NewQueryParams().Include("countries,shared_capacity_groups,qty_based_pricings")
	pool, err := client.CapacityPools().Find(context.Background(), "f288d07c-e2fc-4ae6-9837-b18fb469c324", params)
	require.NoError(t, err)

	assert.Equal(t, "f288d07c-e2fc-4ae6-9837-b18fb469c324", pool.ID)
	assert.Equal(t, "Standard", pool.Name)
	assert.Equal(t, 34, pool.TotalChannelsCount)
	assert.Equal(t, "0.0", pool.SetupPrice)
	assert.Equal(t, "15.0", pool.MonthlyPrice)

	// Verify countries are resolved (fixture has many)
	assert.NotEmpty(t, pool.Countries)
}

func TestCapacityPoolsUpdate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/capacity_pools/f288d07c-e2fc-4ae6-9837-b18fb469c324": {status: http.StatusOK, fixture: "capacity_pools/update.json"},
	})

	pool, err := server.client.CapacityPools().Update(context.Background(), &resource.CapacityPool{
		ID:                 "f288d07c-e2fc-4ae6-9837-b18fb469c324",
		TotalChannelsCount: 25,
	})
	require.NoError(t, err)

	assert.Equal(t, "f288d07c-e2fc-4ae6-9837-b18fb469c324", pool.ID)

	assertRequestJSON(t, *capturedBodyPtr, "capacity_pools/update_request.json")
}
