package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressRequirementsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/address_requirements": {status: http.StatusOK, fixture: "address_requirements/index.json"},
	})

	requirements, err := client.AddressRequirements().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, requirements)
}

func TestAddressRequirementsFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/address_requirements/25d12afe-1ec6-4fe3-9621-b250dd1fb959": {status: http.StatusOK, fixture: "address_requirements/show.json"},
	})

	params := NewQueryParams().Include("country,did_group_type,personal_permanent_document,business_permanent_document,personal_onetime_document,business_onetime_document,personal_proof_types,business_proof_types,address_proof_types")
	req, err := client.AddressRequirements().Find(context.Background(), "25d12afe-1ec6-4fe3-9621-b250dd1fb959", params)
	require.NoError(t, err)

	assert.Equal(t, "25d12afe-1ec6-4fe3-9621-b250dd1fb959", req.ID)
	assert.Equal(t, "any", req.IdentityType)
	assert.Equal(t, 1, req.PersonalProofQty)
	assert.True(t, req.ServiceDescriptionRequired)

	// Verify resolved country
	require.NotNil(t, req.Country)
	assert.Equal(t, "Spain", req.Country.Name)

	// Verify resolved DIDGroupType
	require.NotNil(t, req.DIDGroupType)
	assert.Equal(t, "Local", req.DIDGroupType.Name)

	// Verify resolved document templates
	require.NotNil(t, req.PersonalPermanentDocument)
	assert.Equal(t, "Belgium Registration Form", req.PersonalPermanentDocument.Name)
	require.NotNil(t, req.PersonalOnetimeDocument)
	assert.Equal(t, "Generic LOI", req.PersonalOnetimeDocument.Name)

	// Verify resolved proof types
	require.Len(t, req.PersonalProofTypes, 1)
	assert.Equal(t, "Drivers License", req.PersonalProofTypes[0].Name)
	require.Len(t, req.BusinessProofTypes, 7)
	require.Len(t, req.AddressProofTypes, 1)
	assert.Equal(t, "Copy of Phone Bill", req.AddressProofTypes[0].Name)
}
