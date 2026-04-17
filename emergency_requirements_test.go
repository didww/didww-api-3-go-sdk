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

	require.Len(t, reqs, 1)
	assert.Equal(t, "c1d2e3f4-a5b6-7890-1234-567890abcdef", reqs[0].ID)
	assert.Equal(t, "Any", reqs[0].IdentityType)
	assert.Equal(t, "City", reqs[0].AddressAreaLevel)
	assert.Equal(t, []string{"city", "postal_code"}, reqs[0].AddressMandatoryFields)
	assert.Equal(t, []string{"first_name", "last_name"}, reqs[0].PersonalMandatoryFields)
	assert.Equal(t, "7-14 days", reqs[0].EstimateSetupTime)
}
