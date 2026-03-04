package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAreasList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/areas": {status: http.StatusOK, fixture: "areas/index.json"},
	})

	areas, err := client.Areas().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, areas)
}

func TestAreasFindWithIncludedCountry(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/areas/ab2adc18-7c94-42d9-bdde-b28dfc373a22": {status: http.StatusOK, fixture: "areas/show.json"},
	})

	params := NewQueryParams().Include("country")
	area, err := client.Areas().Find(context.Background(), "ab2adc18-7c94-42d9-bdde-b28dfc373a22", params)
	require.NoError(t, err)

	assert.Equal(t, "ab2adc18-7c94-42d9-bdde-b28dfc373a22", area.ID)
	assert.Equal(t, "Tuscany", area.Name)
	require.NotNil(t, area.Country)
	assert.Equal(t, "Italy", area.Country.Name)
	assert.Equal(t, "39", area.Country.Prefix)
	assert.Equal(t, "IT", area.Country.ISO)
}
