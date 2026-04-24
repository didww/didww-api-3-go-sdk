package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentitiesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/identities": {status: http.StatusOK, fixture: "identities/index.json"},
	})

	identities, err := client.Identities().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, identities)

	first := identities[0]
	assert.Equal(t, "John", first.FirstName)
	assert.Equal(t, "Doe", first.LastName)
	assert.Equal(t, enums.IdentityTypePersonal, first.IdentityType)
}

func TestIdentitiesCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/identities": {status: http.StatusCreated, fixture: "identities/create.json"},
	})

	idNumber := "ABC1234"
	companyName := "Test Company Limited"
	companyRegNumber := "543221"
	vatID := "GB1234"
	description := "test identity"
	personalTaxID := "987654321"
	externalRefID := "111"
	contactEmail := "john.doe@example.com"
	identity, err := server.client.Identities().Create(context.Background(), &resource.Identity{
		FirstName:           "John",
		LastName:            "Doe",
		PhoneNumber:         "123456789",
		IDNumber:            &idNumber,
		BirthDate:           "1970-01-01",
		CompanyName:         &companyName,
		CompanyRegNumber:    &companyRegNumber,
		VatID:               &vatID,
		Description:         &description,
		PersonalTaxID:       &personalTaxID,
		IdentityType:        enums.IdentityTypeBusiness,
		ExternalReferenceID: &externalRefID,
		ContactEmail:        &contactEmail,
		CountryID:           "1f6fc2bd-f081-4202-9b1a-d9cb88d942b9",
	})
	require.NoError(t, err)

	assert.Equal(t, "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae", identity.ID)
	assert.Equal(t, "John", identity.FirstName)
	assert.Equal(t, enums.IdentityTypeBusiness, identity.IdentityType)

	// Verify included country
	require.NotNil(t, identity.Country)
	assert.Equal(t, "United States", identity.Country.Name)

	assertRequestJSON(t, *capturedBodyPtr, "identities/create_request.json")
}

func TestIdentitiesCreatePersonal(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/identities": {status: http.StatusCreated, fixture: "identities/create_personal.json"},
	})

	identity, err := client.Identities().Create(context.Background(), &resource.Identity{
		FirstName:    "John",
		LastName:     "Doe",
		PhoneNumber:  "123456789",
		IdentityType: enums.IdentityTypePersonal,
		BirthDate:    "1970-01-01",
		CountryID:    "1f6fc2bd-f081-4202-9b1a-d9cb88d942b9",
	})
	require.NoError(t, err)

	assert.Equal(t, "9728ea13-cb5d-41fb-8a7f-796a005b0a13", identity.ID)
	assert.Equal(t, enums.IdentityTypePersonal, identity.IdentityType)
	assert.Equal(t, "John", identity.FirstName)
	require.NotNil(t, identity.PersonalTaxID)
	assert.Equal(t, "987654321", *identity.PersonalTaxID)
}

func TestIdentitiesUpdate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/identities/e96ae7d1-11d5-42bc-a5c5-211f3c3788ae": {status: http.StatusOK, fixture: "identities/update.json"},
	})

	companyName := "Some Company Limited"
	companyRegNumber := "1222776"
	vatID := "GB1235"
	description := "test"
	personalTaxID := "983217654"
	externalRefID := "112"
	contactEmail := "jake.johnson@example.com"
	identity, err := server.client.Identities().Update(context.Background(), &resource.Identity{
		ID:                  "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae",
		FirstName:           "Jake",
		LastName:            "Johnson",
		PhoneNumber:         "1111111",
		BirthDate:           "1979-01-01",
		CompanyName:         &companyName,
		CompanyRegNumber:    &companyRegNumber,
		VatID:               &vatID,
		Description:         &description,
		PersonalTaxID:       &personalTaxID,
		IdentityType:        enums.IdentityTypeBusiness,
		ExternalReferenceID: &externalRefID,
		ContactEmail:        &contactEmail,
	})
	require.NoError(t, err)

	assert.Equal(t, "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae", identity.ID)
	assert.Equal(t, "Jake", identity.FirstName)
	assert.Equal(t, "Johnson", identity.LastName)
	require.NotNil(t, identity.CompanyName)
	assert.Equal(t, "Some Company Limited", *identity.CompanyName)
	require.NotNil(t, identity.ContactEmail)
	assert.Equal(t, "jake.johnson@example.com", *identity.ContactEmail)

	assertRequestJSON(t, *capturedBodyPtr, "identities/update_request.json")
}

func TestIdentitiesFindWithContactEmail(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/identities/e96ae7d1-11d5-42bc-a5c5-211f3c3788ae": {status: http.StatusOK, fixture: "identities/show_with_contact_email.json"},
	})

	identity, err := client.Identities().Find(context.Background(), "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae")
	require.NoError(t, err)

	assert.Equal(t, "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae", identity.ID)
	assert.Equal(t, "John", identity.FirstName)
	assert.Equal(t, enums.IdentityTypeBusiness, identity.IdentityType)
	require.NotNil(t, identity.ContactEmail)
	assert.Equal(t, "john.doe@example.com", *identity.ContactEmail)
}

func TestIdentitiesFindWithBirthCountry(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/identities/e96ae7d1-11d5-42bc-a5c5-211f3c3788ae": {status: http.StatusOK, fixture: "identities/show_with_birth_country.json"},
	})

	identity, err := client.Identities().Find(context.Background(), "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae")
	require.NoError(t, err)

	assert.Equal(t, "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae", identity.ID)
	assert.Equal(t, "John", identity.FirstName)
	assert.Equal(t, enums.IdentityTypePersonal, identity.IdentityType)

	// Verify included country
	require.NotNil(t, identity.Country)
	assert.Equal(t, "United States", identity.Country.Name)
	assert.Equal(t, "US", identity.Country.ISO)

	// Verify included birth_country
	require.NotNil(t, identity.BirthCountry)
	assert.Equal(t, "a2b3c4d5-e6f7-8901-abcd-ef1234567890", identity.BirthCountry.ID)
	assert.Equal(t, "Canada", identity.BirthCountry.Name)
	assert.Equal(t, "CA", identity.BirthCountry.ISO)
}

func TestIdentitiesDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/identities/e96ae7d1-11d5-42bc-a5c5-211f3c3788ae": {status: http.StatusNoContent},
	})

	err := client.Identities().Delete(context.Background(), "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae")
	require.NoError(t, err)
}
