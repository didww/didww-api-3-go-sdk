package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/authenticationmethod"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVoiceOutTrunksList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_out_trunks": {status: http.StatusOK, fixture: "voice_out_trunks/index.json"},
	})

	trunks, err := client.VoiceOutTrunks().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, trunks, 2)

	trunk := trunks[0]
	assert.Equal(t, "425ce763-a3a9-49b4-af5b-ada1a65c8864", trunk.ID)
	assert.Equal(t, "test", trunk.Name)
	assert.Equal(t, enums.VoiceOutTrunkStatusBlocked, trunk.Status)

	// Verify authentication_method is parsed as credentials_and_ip
	require.NotNil(t, trunk.AuthenticationMethod)
	credAM, ok := trunk.AuthenticationMethod.(*authenticationmethod.CredentialsAndIp)
	require.True(t, ok, "expected CredentialsAndIp authentication method")
	assert.Equal(t, "dpjgwbbac9", credAM.Username)
	assert.Equal(t, "z0hshvbcy7", credAM.Password)
	assert.Equal(t, []string{"203.0.113.1/32"}, credAM.AllowedSipIPs)
}

func TestVoiceOutTrunksFindWithIncludedDids(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_out_trunks/425ce763-a3a9-49b4-af5b-ada1a65c8864": {status: http.StatusOK, fixture: "voice_out_trunks/show.json"},
	})

	params := NewQueryParams().Include("dids,default_did")
	trunk, err := client.VoiceOutTrunks().Find(context.Background(), "425ce763-a3a9-49b4-af5b-ada1a65c8864", params)
	require.NoError(t, err)

	assert.Equal(t, "425ce763-a3a9-49b4-af5b-ada1a65c8864", trunk.ID)
	assert.Equal(t, "test", trunk.Name)
	assert.Equal(t, enums.MediaEncryptionModeSrtpSdes, trunk.MediaEncryptionMode)
	assert.True(t, trunk.ForceSymmetricRtp)
	assert.True(t, trunk.RtpPing)

	// Verify authentication_method
	require.NotNil(t, trunk.AuthenticationMethod)
	credAM, ok := trunk.AuthenticationMethod.(*authenticationmethod.CredentialsAndIp)
	require.True(t, ok, "expected CredentialsAndIp authentication method")
	assert.Equal(t, "dpjgwbbac9", credAM.Username)
	assert.Equal(t, "z0hshvbcy7", credAM.Password)

	// Verify included default_did
	require.NotNil(t, trunk.DefaultDID)
	assert.Equal(t, "7de7f718-4042-4d74-9fe9-863fa1777520", trunk.DefaultDID.ID)
	assert.Equal(t, "37061498222", trunk.DefaultDID.Number)

	// Verify included dids
	require.Len(t, trunk.DIDs, 2)
}

func TestVoiceOutTrunksFindIpOnly(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_out_trunks/23fd58f9-9094-406c-bfd9-f4d25bda13c6": {status: http.StatusOK, fixture: "voice_out_trunks/show_ip_only.json"},
	})

	trunk, err := client.VoiceOutTrunks().Find(context.Background(), "23fd58f9-9094-406c-bfd9-f4d25bda13c6", nil)
	require.NoError(t, err)

	assert.Equal(t, "23fd58f9-9094-406c-bfd9-f4d25bda13c6", trunk.ID)
	assert.Equal(t, "SDK Test credentials_and_ip", trunk.Name)
	assert.Equal(t, enums.VoiceOutTrunkStatusActive, trunk.Status)

	// Verify authentication_method is parsed as IpOnly, not CredentialsAndIp
	require.NotNil(t, trunk.AuthenticationMethod)
	ipAM, ok := trunk.AuthenticationMethod.(*authenticationmethod.IpOnly)
	require.True(t, ok, "expected IpOnly authentication method, got %T", trunk.AuthenticationMethod)
	assert.Equal(t, []string{"203.0.113.1/32"}, ipAM.AllowedSipIPs)

	// Must NOT be CredentialsAndIp
	_, notCred := trunk.AuthenticationMethod.(*authenticationmethod.CredentialsAndIp)
	assert.False(t, notCred, "authentication_method should not be CredentialsAndIp")
}

func TestVoiceOutTrunksFindTwilio(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_out_trunks/b5e701f4-ea15-4f9d-8f35-6a0bdce04385": {status: http.StatusOK, fixture: "voice_out_trunks/show_twilio.json"},
	})

	trunk, err := client.VoiceOutTrunks().Find(context.Background(), "b5e701f4-ea15-4f9d-8f35-6a0bdce04385", nil)
	require.NoError(t, err)

	assert.Equal(t, "b5e701f4-ea15-4f9d-8f35-6a0bdce04385", trunk.ID)
	assert.Equal(t, "SDK Test twilio", trunk.Name)
	assert.Equal(t, enums.VoiceOutTrunkStatusActive, trunk.Status)

	// Verify authentication_method is parsed as Twilio
	require.NotNil(t, trunk.AuthenticationMethod)
	twilioAM, ok := trunk.AuthenticationMethod.(*authenticationmethod.Twilio)
	require.True(t, ok, "expected Twilio authentication method, got %T", trunk.AuthenticationMethod)
	assert.Equal(t, "AC22222222222222222222222222222222", twilioAM.TwilioAccountSid)
}

func TestVoiceOutTrunksCreateTwilio(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/voice_out_trunks": {status: http.StatusCreated, fixture: "voice_out_trunks/create_twilio.json"},
	})

	trunk, err := server.client.VoiceOutTrunks().Create(context.Background(), &resource.VoiceOutTrunk{
		Name:                "SDK Test twilio create",
		OnCliMismatchAction: enums.OnCliMismatchActionRejectCall,
		AuthenticationMethod: &authenticationmethod.Twilio{
			TwilioAccountSid: "AC33333333333333333333333333333333",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "507fa5a2-fd58-4c4d-a231-efba27f67c3a", trunk.ID)
	assert.Equal(t, "SDK Test twilio create", trunk.Name)

	// Verify authentication_method in response
	require.NotNil(t, trunk.AuthenticationMethod)
	twilioAM, ok := trunk.AuthenticationMethod.(*authenticationmethod.Twilio)
	require.True(t, ok, "expected Twilio authentication method")
	assert.Equal(t, "AC33333333333333333333333333333333", twilioAM.TwilioAccountSid)

	assertRequestJSON(t, *capturedBodyPtr, "voice_out_trunks/create_twilio_request.json")
}

func TestVoiceOutTrunksCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/voice_out_trunks": {status: http.StatusCreated, fixture: "voice_out_trunks/create.json"},
	})

	trunk, err := server.client.VoiceOutTrunks().Create(context.Background(), &resource.VoiceOutTrunk{
		Name:                "java-test",
		OnCliMismatchAction: enums.OnCliMismatchActionReplaceCli,
		AuthenticationMethod: &authenticationmethod.CredentialsAndIp{
			AllowedSipIPs: []string{"203.0.113.0/24"},
		},
		DefaultDIDID: "7a028c32-e6b6-4c86-bf01-90f901b37012",
		DIDIDs:       []string{"7a028c32-e6b6-4c86-bf01-90f901b37012"},
	})
	require.NoError(t, err)

	assert.Equal(t, "b60201c1-21f0-4d9a-aafa-0e6d1e12f22e", trunk.ID)

	// Verify authentication_method in response
	require.NotNil(t, trunk.AuthenticationMethod)
	credAM, ok := trunk.AuthenticationMethod.(*authenticationmethod.CredentialsAndIp)
	require.True(t, ok, "expected CredentialsAndIp authentication method")
	assert.Equal(t, []string{"203.0.113.0/24"}, credAM.AllowedSipIPs)
	assert.Equal(t, "dLPa6JbLTeMjKjl5", credAM.Username)
	assert.Equal(t, "BZj1YvP45yWvX5Ic", credAM.Password)

	assertRequestJSON(t, *capturedBodyPtr, "voice_out_trunks/create_request.json")
}

func TestVoiceOutTrunksUpdate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/voice_out_trunks/425ce763-a3a9-49b4-af5b-ada1a65c8864": {status: http.StatusOK, fixture: "voice_out_trunks/update.json"},
	})

	trunk, err := client.VoiceOutTrunks().Update(context.Background(), &resource.VoiceOutTrunk{
		ID:            "425ce763-a3a9-49b4-af5b-ada1a65c8864",
		CapacityLimit: intPtr(123),
	})
	require.NoError(t, err)

	assert.Equal(t, "425ce763-a3a9-49b4-af5b-ada1a65c8864", trunk.ID)
	assert.Equal(t, "test", trunk.Name)
	require.NotNil(t, trunk.CapacityLimit)
	assert.Equal(t, 123, *trunk.CapacityLimit)
	assert.True(t, trunk.ForceSymmetricRtp)
	assert.True(t, trunk.RtpPing)
}

func TestVoiceOutTrunksUpdateAuthenticationMethod(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/voice_out_trunks/425ce763-a3a9-49b4-af5b-ada1a65c8864": {status: http.StatusOK, fixture: "voice_out_trunks/update.json"},
	})

	trunk, err := server.client.VoiceOutTrunks().Update(context.Background(), &resource.VoiceOutTrunk{
		ID: "425ce763-a3a9-49b4-af5b-ada1a65c8864",
		AuthenticationMethod: &authenticationmethod.CredentialsAndIp{
			AllowedSipIPs: []string{"192.0.2.10/32"},
			TechPrefix:    "99",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "425ce763-a3a9-49b4-af5b-ada1a65c8864", trunk.ID)

	assertRequestJSON(t, *capturedBodyPtr, "voice_out_trunks/update_auth_method_request.json")
}

func TestVoiceOutTrunksUpdateEmergencyEnableAll(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/voice_out_trunks/01234567-89ab-cdef-0123-456789abcdef": {status: http.StatusOK, fixture: "voice_out_trunks/update_emergency_dids.json"},
	})

	trunk, err := server.client.VoiceOutTrunks().Update(context.Background(), &resource.VoiceOutTrunk{
		ID:                 "01234567-89ab-cdef-0123-456789abcdef",
		EmergencyEnableAll: true,
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "voice_out_trunks/update_emergency_enable_all_request.json")

	assert.Equal(t, "01234567-89ab-cdef-0123-456789abcdef", trunk.ID)
}

func TestVoiceOutTrunksUpdateEmergencyDIDs(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/voice_out_trunks/01234567-89ab-cdef-0123-456789abcdef": {status: http.StatusOK, fixture: "voice_out_trunks/update_emergency_dids.json"},
	})

	trunk, err := server.client.VoiceOutTrunks().Update(context.Background(), &resource.VoiceOutTrunk{
		ID: "01234567-89ab-cdef-0123-456789abcdef",
		EmergencyDIDIDs: []string{
			"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "voice_out_trunks/update_emergency_dids_request.json")

	assert.Equal(t, "01234567-89ab-cdef-0123-456789abcdef", trunk.ID)
}

func TestVoiceOutTrunksUpdateClearEmergencyDIDs(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/voice_out_trunks/01234567-89ab-cdef-0123-456789abcdef": {status: http.StatusOK, fixture: "voice_out_trunks/update_emergency_dids.json"},
	})

	trunk, err := server.client.VoiceOutTrunks().Update(context.Background(), &resource.VoiceOutTrunk{
		ID:                 "01234567-89ab-cdef-0123-456789abcdef",
		ClearEmergencyDIDs: true,
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "voice_out_trunks/update_emergency_dids_clear_request.json")

	assert.Equal(t, "01234567-89ab-cdef-0123-456789abcdef", trunk.ID)
}

func TestVoiceOutTrunksDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/voice_out_trunks/425ce763-a3a9-49b4-af5b-ada1a65c8864": {status: http.StatusNoContent},
	})

	err := client.VoiceOutTrunks().Delete(context.Background(), "425ce763-a3a9-49b4-af5b-ada1a65c8864")
	require.NoError(t, err)
}
