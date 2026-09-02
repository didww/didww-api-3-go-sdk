package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmergencyRequirementsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/emergency_requirements": {status: http.StatusOK, fixture: "emergency_requirements/index.json"},
	})

	reqs, err := client.EmergencyRequirements().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, reqs, 2)
	assert.Equal(t, "c1d2e3f4-a5b6-7890-1234-567890abcdef", reqs[0].ID)
	assert.Equal(t, "any", reqs[0].IdentityType)
	assert.Equal(t, "city", reqs[0].AddressAreaLevel)
	assert.Equal(t, "world_wide", reqs[0].PersonalAreaLevel)
	assert.Equal(t, "country", reqs[0].BusinessAreaLevel)
	assert.Equal(t, []string{"city", "postal_code"}, reqs[0].AddressMandatoryFields)
	assert.Equal(t, []string{"first_name", "last_name"}, reqs[0].PersonalMandatoryFields)
	assert.Equal(t, "7-14 days", reqs[0].EstimateSetupTime)

	// Meta fields
	require.NotNil(t, reqs[0].Meta)
	assert.Equal(t, "0.0", reqs[0].Meta["setup_price"])
	assert.Equal(t, "1.5", reqs[0].Meta["monthly_price"])

	// A country that accepts business identities only leaves the personal level unset.
	assert.Equal(t, "business", reqs[1].IdentityType)
	assert.Empty(t, reqs[1].PersonalAreaLevel)
	assert.Equal(t, "world_wide", reqs[1].BusinessAreaLevel)
	assert.Equal(t, "0.0", reqs[1].Meta["setup_price"])
	assert.Equal(t, "2.5", reqs[1].Meta["monthly_price"])
}
