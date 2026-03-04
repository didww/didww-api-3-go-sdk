package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegionsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/regions": {status: http.StatusOK, fixture: "regions/index.json"},
	})

	regions, err := client.Regions().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, regions)
}

func TestRegionsFindWithIncludedCountry(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/regions/c11b1f34-16cf-4ba6-8497-f305b53d5b01": {status: http.StatusOK, fixture: "regions/show.json"},
	})

	params := NewQueryParams().Include("country")
	region, err := client.Regions().Find(context.Background(), "c11b1f34-16cf-4ba6-8497-f305b53d5b01", params)
	require.NoError(t, err)

	assert.Equal(t, "c11b1f34-16cf-4ba6-8497-f305b53d5b01", region.ID)
	assert.Equal(t, "California", region.Name)
	require.NotNil(t, region.Country)
	assert.Equal(t, "United States", region.Country.Name)
	assert.Equal(t, "1", region.Country.Prefix)
	assert.Equal(t, "US", region.Country.ISO)
}
