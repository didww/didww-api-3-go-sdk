package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/resource"
)

func TestEmergencyRequirementValidationsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/emergency_requirement_validations": {status: http.StatusCreated, fixture: "emergency_requirement_validations/create.json"},
	})

	rv, err := server.client.EmergencyRequirementValidations().Create(context.Background(), &resource.EmergencyRequirementValidation{
		EmergencyRequirementID: "c1d2e3f4-a5b6-7890-1234-567890abcdef",
		AddressID:              "d3414687-40f4-4346-a267-c2c65117d28c",
		IdentityID:             "5e9df058-50d2-4e34-b0d4-d1746b86f41a",
	})
	require.NoError(t, err)
	assert.Equal(t, "c1d2e3f4-a5b6-7890-1234-567890abcdef", rv.ID)

	assertRequestJSON(t, *capturedBodyPtr, "emergency_requirement_validations/create_request.json")
}
