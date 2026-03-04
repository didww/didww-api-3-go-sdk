package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCitiesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/cities": {status: http.StatusOK, fixture: "cities/index.json"},
	})

	cities, err := client.Cities().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, cities)
}

func TestCitiesFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/cities/368bf92f-c36e-473f-96fc-d53ed1b4028b": {status: http.StatusOK, fixture: "cities/show.json"},
	})

	params := NewQueryParams().Include("country,region")
	city, err := client.Cities().Find(context.Background(), "368bf92f-c36e-473f-96fc-d53ed1b4028b", params)
	require.NoError(t, err)

	assert.Equal(t, "368bf92f-c36e-473f-96fc-d53ed1b4028b", city.ID)
	assert.Equal(t, "New York", city.Name)
	require.NotNil(t, city.Country)
	assert.Equal(t, "United States", city.Country.Name)
	require.NotNil(t, city.Region)
	assert.Equal(t, "New York", city.Region.Name)
}
