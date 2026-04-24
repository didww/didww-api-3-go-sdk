package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmergencyCallingServicesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/emergency_calling_services": {status: http.StatusOK, fixture: "emergency_calling_services/index.json"},
	})

	services, err := client.EmergencyCallingServices().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, services, 1)
	assert.Equal(t, "ecs-001-id", services[0].ID)
	assert.Equal(t, "E911 Service US", services[0].Name)
	assert.Equal(t, "ECS-12345", services[0].Reference)
	assert.Equal(t, "active", services[0].Status)
	assert.False(t, services[0].CreatedAt.IsZero())
	require.NotNil(t, services[0].ActivatedAt)
	assert.Nil(t, services[0].CanceledAt)
	assert.Equal(t, "2026-05-01", services[0].RenewDate)

	// Meta fields
	require.NotNil(t, services[0].Meta)
	assert.Equal(t, "0.0", services[0].Meta["setup_price"])
	assert.Equal(t, "1.5", services[0].Meta["monthly_price"])
}

func TestEmergencyCallingServicesFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/emergency_calling_services/ecs-001-id": {status: http.StatusOK, fixture: "emergency_calling_services/show_with_includes.json"},
	})

	params := NewQueryParams().Include("emergency_requirement,emergency_verification")
	svc, err := client.EmergencyCallingServices().Find(context.Background(), "ecs-001-id", params)
	require.NoError(t, err)

	assert.Equal(t, "ecs-001-id", svc.ID)
	assert.Equal(t, "E911 Service US", svc.Name)
	assert.Equal(t, "active", svc.Status)

	// Meta fields
	require.NotNil(t, svc.Meta)
	assert.Equal(t, "0.0", svc.Meta["setup_price"])
	assert.Equal(t, "1.5", svc.Meta["monthly_price"])

	// Verify included emergency_requirement
	require.NotNil(t, svc.EmergencyRequirement)
	assert.Equal(t, "ereq-001-id", svc.EmergencyRequirement.ID)
	assert.Equal(t, "personal", svc.EmergencyRequirement.IdentityType)
	assert.Equal(t, "city", svc.EmergencyRequirement.AddressAreaLevel)
	assert.Equal(t, []string{"city", "postal_code"}, svc.EmergencyRequirement.AddressMandatoryFields)

	// Verify included emergency_verification
	require.NotNil(t, svc.EmergencyVerification)
	assert.Equal(t, "ever-001-id", svc.EmergencyVerification.ID)
	assert.Equal(t, "EVR-54321", svc.EmergencyVerification.Reference)
	assert.Equal(t, "approved", svc.EmergencyVerification.Status)
}
