package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
)

func TestAddressRequirementValidationsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/address_requirement_validations": {status: http.StatusCreated, fixture: "address_requirement_validations/create.json"},
	})

	rv, err := server.client.AddressRequirementValidations().Create(context.Background(), &resource.AddressRequirementValidation{
		AddressID:            "d3414687-40f4-4346-a267-c2c65117d28c",
		AddressRequirementID: "aea92b24-a044-4864-9740-89d3e15b65c7",
	})
	require.NoError(t, err)

	assert.NotEmpty(t, rv.ID)

	assertRequestJSON(t, *capturedBodyPtr, "address_requirement_validations/create_request.json")
}

func TestAddressRequirementValidationsCreateError(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/address_requirement_validations": {status: http.StatusUnprocessableEntity, fixture: "address_requirement_validations/create_error_validation.json"},
	})

	_, err := server.client.AddressRequirementValidations().Create(context.Background(), &resource.AddressRequirementValidation{
		IdentityID:           "5e9df058-50d2-4e34-b0d4-d1746b86f41a",
		AddressID:            "d3414687-40f4-4346-a267-c2c65117d28c",
		AddressRequirementID: "2efc3427-8ba6-4d50-875d-f2de4a068de8",
	})
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError")
	require.Len(t, apiErr.Errors, 3)

	assertRequestJSON(t, *capturedBodyPtr, "address_requirement_validations/create_request_failed.json")
}
