package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDIDGroupsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_groups": {status: http.StatusOK, fixture: "did_groups/index.json"},
	})

	groups, err := client.DIDGroups().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, groups)
}

func TestDIDGroupsFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_groups/2187c36d-28fb-436f-8861-5a0f5b5a3ee1": {status: http.StatusOK, fixture: "did_groups/show.json"},
	})

	params := NewQueryParams().Include("country,city,did_group_type,stock_keeping_units")
	group, err := client.DIDGroups().Find(context.Background(), "2187c36d-28fb-436f-8861-5a0f5b5a3ee1", params)
	require.NoError(t, err)

	assert.Equal(t, "2187c36d-28fb-436f-8861-5a0f5b5a3ee1", group.ID)
	assert.Equal(t, "241", group.Prefix)
	assert.Equal(t, "Aachen", group.AreaName)
	assert.True(t, group.AllowAdditionalChannels)

	// Verify included country
	require.NotNil(t, group.Country)
	assert.Equal(t, "Germany", group.Country.Name)
	assert.Equal(t, "DE", group.Country.ISO)

	// Verify included city
	require.NotNil(t, group.City)
	assert.Equal(t, "Aachen", group.City.Name)

	// Verify included DID group type
	require.NotNil(t, group.DIDGroupType)
	assert.Equal(t, "Local", group.DIDGroupType.Name)

	// Verify included stock keeping units
	require.Len(t, group.StockKeepingUnits, 2)
	assert.Equal(t, "0.4", group.StockKeepingUnits[0].SetupPrice)
	assert.Equal(t, "0.8", group.StockKeepingUnits[0].MonthlyPrice)
	assert.Equal(t, 0, group.StockKeepingUnits[0].ChannelsIncludedCount)
	assert.Equal(t, 2, group.StockKeepingUnits[1].ChannelsIncludedCount)
}

func TestDIDGroupsFindWithIncludedRequirement(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_groups/2187c36d-28fb-436f-8861-5a0f5b5a3ee1": {status: http.StatusOK, fixture: "did_groups/show_with_requirement.json"},
	})

	params := NewQueryParams().Include("requirement")
	group, err := client.DIDGroups().Find(context.Background(), "2187c36d-28fb-436f-8861-5a0f5b5a3ee1", params)
	require.NoError(t, err)

	assert.Equal(t, "2187c36d-28fb-436f-8861-5a0f5b5a3ee1", group.ID)
	assert.Equal(t, "241", group.Prefix)
	assert.Equal(t, "Aachen", group.AreaName)
	assert.False(t, group.IsMetered)
	assert.True(t, group.AllowAdditionalChannels)

	// Verify included requirement
	require.NotNil(t, group.Requirement)
	assert.Equal(t, "8da1e0b2-047c-4baf-9c57-57143f09b9ce", group.Requirement.ID)
	assert.Equal(t, "Any", group.Requirement.IdentityType)
	assert.Equal(t, "WorldWide", group.Requirement.PersonalAreaLevel)
	assert.Equal(t, "Country", group.Requirement.BusinessAreaLevel)
	assert.Equal(t, "City", group.Requirement.AddressAreaLevel)
	assert.Equal(t, 1, group.Requirement.PersonalProofQty)
	assert.Equal(t, 1, group.Requirement.BusinessProofQty)
	assert.Equal(t, 1, group.Requirement.AddressProofQty)
	assert.False(t, group.Requirement.ServiceDescriptionRequired)
}
