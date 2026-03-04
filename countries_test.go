package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountriesFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/countries/7eda11bb-0e66-4146-98e7-57a5281f56c8": {status: http.StatusOK, fixture: "countries/show.json"},
	})

	country, err := client.Countries().Find(context.Background(), "7eda11bb-0e66-4146-98e7-57a5281f56c8")
	require.NoError(t, err)

	assert.Equal(t, "7eda11bb-0e66-4146-98e7-57a5281f56c8", country.ID)
	assert.Equal(t, "United Kingdom", country.Name)
	assert.Equal(t, "44", country.Prefix)
	assert.Equal(t, "GB", country.ISO)
}

func TestCountriesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/countries": {status: http.StatusOK, fixture: "countries/index.json"},
	})

	countries, err := client.Countries().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, countries)

	first := countries[0]
	assert.NotEmpty(t, first.ID)
	assert.NotEmpty(t, first.Name)
	assert.NotEmpty(t, first.Prefix)
	assert.NotEmpty(t, first.ISO)
}

func TestCountriesFindWithIncludedRegions(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/countries/661d8448-8897-4765-acda-00cc1740148d": {status: http.StatusOK, fixture: "countries/show_with_regions.json"},
	})

	params := NewQueryParams().Include("regions")
	country, err := client.Countries().Find(context.Background(), "661d8448-8897-4765-acda-00cc1740148d", params)
	require.NoError(t, err)

	assert.Equal(t, "661d8448-8897-4765-acda-00cc1740148d", country.ID)
	assert.Equal(t, "Lithuania", country.Name)
	assert.Equal(t, "370", country.Prefix)
	assert.Equal(t, "LT", country.ISO)
	require.Len(t, country.Regions, 10)
	assert.Equal(t, "Alytaus Apskritis", country.Regions[0].Name)
}
