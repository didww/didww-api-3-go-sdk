package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func TestDIDsFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show.json"},
	})

	did, err := client.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b")
	require.NoError(t, err)

	assert.Equal(t, "9df99644-f1a5-4a3c-99a4-559d758eb96b", did.ID)
	assert.Equal(t, "16091609123456797", did.Number)
	assert.False(t, did.Blocked)
	assert.False(t, did.Terminated)
}

func TestDIDsListWithIncludedOrder(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids": {status: http.StatusOK, fixture: "dids/index.json"},
	})

	params := NewQueryParams().Include("order")
	dids, err := client.DIDs().List(context.Background(), params)
	require.NoError(t, err)

	require.Len(t, dids, 3)

	// First DID should have resolved order
	first := dids[0]
	require.NotNil(t, first.Order)
	assert.Equal(t, "11b3fba2-96e2-452e-bed8-5124ed351af3", first.Order.ID)
	assert.Equal(t, "0.37", first.Order.Amount)
}

func TestDIDsFindWithAddressVerificationAndDIDGroup(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/21d0b02c-b556-4d3e-acbf-504b78295dbe": {status: http.StatusOK, fixture: "dids/show_with_address_verification_and_did_group.json"},
	})

	params := NewQueryParams().Include("address_verification,did_group")
	did, err := client.DIDs().Find(context.Background(), "21d0b02c-b556-4d3e-acbf-504b78295dbe", params)
	require.NoError(t, err)

	assert.Equal(t, "21d0b02c-b556-4d3e-acbf-504b78295dbe", did.ID)
	assert.Equal(t, "61488943592", did.Number)

	// Verify address verification
	require.NotNil(t, did.AddressVerification)
	assert.Equal(t, "75dc8d39-5e17-4470-a6f3-df42642c975f", did.AddressVerification.ID)
	assert.Equal(t, enums.AddressVerificationStatus("Approved"), did.AddressVerification.Status)

	// Verify DID group
	require.NotNil(t, did.DIDGroup)
	assert.Equal(t, "2b60bb9a-d382-4d35-84c6-61689f45f2f5", did.DIDGroup.ID)
	assert.Equal(t, "Mobile", did.DIDGroup.AreaName)
}

func TestDIDsUpdate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show.json"},
	})

	desc := "updated"
	_, err := client.DIDs().Update(context.Background(), &DID{
		ID:          "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		Description: &desc,
	})
	require.NoError(t, err)
}

func TestDIDsUpdateDescription(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/update_success.json"},
	})

	desc := "something"
	did, err := client.DIDs().Update(context.Background(), &DID{
		ID:          "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		Description: &desc,
	})
	require.NoError(t, err)

	require.NotNil(t, did.Description)
	assert.Equal(t, "something", *did.Description)
	assert.Equal(t, "2019-01-27T10:00:04.755Z", did.ExpiresAt)
}

func TestDIDsFindBlockedTerminated(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/update_blocked_terminated.json"},
	})

	did, err := client.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b")
	require.NoError(t, err)

	assert.True(t, did.Blocked)
	assert.True(t, did.Terminated)
	require.NotNil(t, did.BillingCyclesCount)
	assert.Equal(t, 0, *did.BillingCyclesCount)
}

func TestDIDsUpdateTerminated(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/update_blocked_terminated.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	did, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:         "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		Terminated: true,
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "dids/update_terminated_request.json")

	assert.Equal(t, "9df99644-f1a5-4a3c-99a4-559d758eb96b", did.ID)
	assert.True(t, did.Blocked)
	assert.True(t, did.Terminated)
	require.NotNil(t, did.BillingCyclesCount)
	assert.Equal(t, 0, *did.BillingCyclesCount)
}

func TestDIDsUpdateInvalidParam(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusBadRequest, fixture: "dids/update_error_invalid_param.json"},
	})

	_, err := client.DIDs().Update(context.Background(), &DID{
		ID: "9df99644-f1a5-4a3c-99a4-559d758eb96b",
	})
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError")
	assert.Equal(t, http.StatusBadRequest, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "105", apiErr.Errors[0].Code)
}

func TestDIDsUpdateInvalidTrunkGroup(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusUnprocessableEntity, fixture: "dids/update_error_invalid_trunk_group.json"},
	})

	_, err := client.DIDs().Update(context.Background(), &DID{
		ID:                  "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		VoiceInTrunkGroupID: "invalid-id",
	})
	require.Error(t, err)

	apiErr, ok := err.(*APIError)
	require.True(t, ok, "expected *APIError")
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "voice_in_trunk_group - is invalid", apiErr.Errors[0].Detail)
}

func TestDIDsUpdateRequiresID(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{})

	_, err := client.DIDs().Update(context.Background(), &DID{})
	require.Error(t, err)
}

func TestDIDsUpdateAssignTrunk(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show_with_trunk.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	did, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:             "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		VoiceInTrunkID: "41b94706-325e-4704-a433-d65105758836",
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "dids/update_assign_trunk_request.json")

	require.NotNil(t, did.VoiceInTrunk)
	assert.Equal(t, "41b94706-325e-4704-a433-d65105758836", did.VoiceInTrunk.ID)
	assert.Equal(t, "hello, test pstn trunk", did.VoiceInTrunk.Name)
}

func TestDIDsUpdateAssignTrunkGroup(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show_with_trunk_group.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	did, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:                  "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		VoiceInTrunkGroupID: "b2319703-ce6c-480d-bb53-614e7abcfc96",
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "dids/update_assign_trunk_group_request.json")

	require.NotNil(t, did.VoiceInTrunkGroup)
	assert.Equal(t, "b2319703-ce6c-480d-bb53-614e7abcfc96", did.VoiceInTrunkGroup.ID)
	assert.Equal(t, "trunk group sample with 2 trunks", did.VoiceInTrunkGroup.Name)
}

func TestDIDsFindWithTrunkResolved(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show_with_trunk.json"},
	})

	params := NewQueryParams().Include("voice_in_trunk")
	did, err := client.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b", params)
	require.NoError(t, err)

	require.NotNil(t, did.VoiceInTrunk)
	assert.Equal(t, "41b94706-325e-4704-a433-d65105758836", did.VoiceInTrunk.ID)
	assert.Equal(t, "hello, test pstn trunk", did.VoiceInTrunk.Name)
	assert.Nil(t, did.VoiceInTrunkGroup)
}

func TestDIDsFindWithTrunkGroupResolved(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show_with_trunk_group.json"},
	})

	params := NewQueryParams().Include("voice_in_trunk_group")
	did, err := client.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b", params)
	require.NoError(t, err)

	require.NotNil(t, did.VoiceInTrunkGroup)
	assert.Equal(t, "b2319703-ce6c-480d-bb53-614e7abcfc96", did.VoiceInTrunkGroup.ID)
	assert.Equal(t, "trunk group sample with 2 trunks", did.VoiceInTrunkGroup.Name)
	assert.Nil(t, did.VoiceInTrunk)
}
