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
