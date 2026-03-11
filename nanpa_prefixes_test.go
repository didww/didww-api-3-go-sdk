package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNanpaPrefixesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/nanpa_prefixes": {status: http.StatusOK, fixture: "nanpa_prefixes/index.json"},
	})

	prefixes, err := client.NanpaPrefixes().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, prefixes)
}

func TestNanpaPrefixesFindWithIncludedCountry(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/nanpa_prefixes/6c16d51d-d376-4395-91c4-012321317e48": {status: http.StatusOK, fixture: "nanpa_prefixes/show.json"},
	})

	params := NewQueryParams().Include("country")
	prefix, err := client.NanpaPrefixes().Find(context.Background(), "6c16d51d-d376-4395-91c4-012321317e48", params)
	require.NoError(t, err)

	assert.Equal(t, "6c16d51d-d376-4395-91c4-012321317e48", prefix.ID)
	assert.Equal(t, "864", prefix.NPA)
	assert.Equal(t, "920", prefix.NXX)
	require.NotNil(t, prefix.Country)
	assert.Equal(t, "United States", prefix.Country.Name)
}

func TestNanpaPrefixesFindWithIncludedRegion(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/nanpa_prefixes/1e622e21-c740-4d3f-a615-2a7ef4991922": {status: http.StatusOK, fixture: "nanpa_prefixes/show_with_region.json"},
	})

	params := NewQueryParams().Include("region")
	prefix, err := client.NanpaPrefixes().Find(context.Background(), "1e622e21-c740-4d3f-a615-2a7ef4991922", params)
	require.NoError(t, err)

	assert.Equal(t, "1e622e21-c740-4d3f-a615-2a7ef4991922", prefix.ID)
	assert.Equal(t, "201", prefix.NPA)
	assert.Equal(t, "221", prefix.NXX)

	// Verify included region
	require.NotNil(t, prefix.Region)
	assert.Equal(t, "346e64c8-18c2-4a12-b1e2-20e090043fca", prefix.Region.ID)
	assert.Equal(t, "New Jersey", prefix.Region.Name)
	require.NotNil(t, prefix.Region.ISO)
	assert.Equal(t, "US-NJ", *prefix.Region.ISO)
}
