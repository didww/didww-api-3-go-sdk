package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/resource"
)

func TestAddressesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/addresses": {status: http.StatusOK, fixture: "addresses/index.json"},
	})

	addresses, err := client.Addresses().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, addresses)
}

func TestAddressesCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/addresses": {status: http.StatusCreated, fixture: "addresses/create.json"},
	})

	address, err := server.client.Addresses().Create(context.Background(), &resource.Address{
		CityName:    "New York",
		PostalCode:  "123",
		Address:     "some street",
		Description: "test address",
		CountryID:   "1f6fc2bd-f081-4202-9b1a-d9cb88d942b9",
		IdentityID:  "5e9df058-50d2-4e34-b0d4-d1746b86f41a",
	})
	require.NoError(t, err)

	assert.Equal(t, "bf69bc70-e1c2-442c-9f30-335ee299b663", address.ID)
	assert.Equal(t, "New York", address.CityName)

	// Verify included country
	require.NotNil(t, address.Country)
	assert.Equal(t, "United States", address.Country.Name)

	assertRequestJSON(t, *capturedBodyPtr, "addresses/create_request.json")
}

func TestAddressesUpdate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/addresses/bf69bc70-e1c2-442c-9f30-335ee299b663": {status: http.StatusOK, fixture: "addresses/update.json"},
	})

	address, err := client.Addresses().Update(context.Background(), &resource.Address{
		ID:       "bf69bc70-e1c2-442c-9f30-335ee299b663",
		CityName: "Chicago",
	})
	require.NoError(t, err)

	assert.Equal(t, "bf69bc70-e1c2-442c-9f30-335ee299b663", address.ID)
	assert.Equal(t, "Chicago", address.CityName)
	assert.Equal(t, "1234", address.PostalCode)
}

func TestAddressesDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/addresses/bf69bc70-e1c2-442c-9f30-335ee299b663": {status: http.StatusNoContent},
	})

	err := client.Addresses().Delete(context.Background(), "bf69bc70-e1c2-442c-9f30-335ee299b663")
	require.NoError(t, err)
}
