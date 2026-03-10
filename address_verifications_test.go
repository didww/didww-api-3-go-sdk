package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressVerificationsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/address_verifications": {status: http.StatusOK, fixture: "address_verifications/index.json"},
	})

	avs, err := client.AddressVerifications().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, avs)
}

func TestAddressVerificationsCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/address_verifications": {status: http.StatusCreated, fixture: "address_verifications/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	cbURL := "http://example.com"
	cbMethod := "GET"
	av, err := server.client.AddressVerifications().Create(context.Background(), &AddressVerification{
		CallbackURL:    &cbURL,
		CallbackMethod: &cbMethod,
		AddressID:      "d3414687-40f4-4346-a267-c2c65117d28c",
		DIDIDs:         []string{"a9d64c02-4486-4acb-a9a1-be4c81ff0659"},
	})
	require.NoError(t, err)

	assert.Equal(t, "78182ef2-8377-41cd-89e1-26e8266c9c94", av.ID)
	assert.Equal(t, enums.AddressVerificationStatusPending, av.Status)

	// Verify included address
	require.NotNil(t, av.AddressRel)
	assert.Equal(t, "d3414687-40f4-4346-a267-c2c65117d28c", av.AddressRel.ID)
	assert.Equal(t, "Chicago", av.AddressRel.CityName)

	assertRequestJSON(t, capturedBody, "address_verifications/create_request.json")
}

func TestAddressVerificationsFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/address_verifications/c8e004b0-87ec-4987-b4fb-ee89db099f0e": {status: http.StatusOK, fixture: "address_verifications/show.json"},
	})

	av, err := client.AddressVerifications().Find(context.Background(), "c8e004b0-87ec-4987-b4fb-ee89db099f0e")
	require.NoError(t, err)

	assert.Equal(t, "c8e004b0-87ec-4987-b4fb-ee89db099f0e", av.ID)
	assert.Equal(t, enums.AddressVerificationStatusApproved, av.Status)
	assert.Equal(t, "SHB-485120", av.Reference)
}

func TestAddressVerificationsFindRejected(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/address_verifications/429e6d4e-2ee9-4953-aa98-0b3ac07f0f96": {status: http.StatusOK, fixture: "address_verifications/show_rejected.json"},
	})

	av, err := client.AddressVerifications().Find(context.Background(), "429e6d4e-2ee9-4953-aa98-0b3ac07f0f96")
	require.NoError(t, err)

	assert.Equal(t, "429e6d4e-2ee9-4953-aa98-0b3ac07f0f96", av.ID)
	assert.Equal(t, enums.AddressVerificationStatusRejected, av.Status)
	require.NotNil(t, av.RejectReasons)
	assert.Equal(t, "Address cannot be validated; Proof of address should be not older than of 6 months", *av.RejectReasons)
	assert.Equal(t, []string{"Address cannot be validated", "Proof of address should be not older than of 6 months"}, av.RejectReasonsList())
	assert.Equal(t, "ODW-879912", av.Reference)
}
