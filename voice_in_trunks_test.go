package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/resource/trunkconfiguration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVoiceInTrunksList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_in_trunks": {status: http.StatusOK, fixture: "voice_in_trunks/index.json"},
	})

	trunks, err := client.VoiceInTrunks().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, trunks, 2)

	// First trunk is PSTN
	pstn := trunks[0]
	assert.Equal(t, "2b4b1fcf-fe6a-4de9-8a58-7df46820ba13", pstn.ID)
	assert.Equal(t, "sample trunk pstn", pstn.Name)
	assert.Equal(t, enums.CliFormatE164, pstn.CliFormat)
	pstnCfg, ok := pstn.Configuration.(*trunkconfiguration.PSTNConfiguration)
	require.True(t, ok, "expected PSTN configuration")
	assert.Equal(t, "442080995011", pstnCfg.Dst)

	// Second trunk is SIP
	sip := trunks[1]
	assert.Equal(t, "Sip trunk sample", sip.Name)
	sipCfg, ok := sip.Configuration.(*trunkconfiguration.SIPConfiguration)
	require.True(t, ok, "expected SIP configuration")
	assert.Equal(t, "216.58.215.78", sipCfg.Host)
}

func TestVoiceInTrunksCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/voice_in_trunks": {status: http.StatusCreated, fixture: "voice_in_trunks/create.json"},
	})

	trunk, err := server.client.VoiceInTrunks().Create(context.Background(), &resource.VoiceInTrunk{
		Name: "hello, test pstn trunk",
		Configuration: &trunkconfiguration.PSTNConfiguration{
			Dst: "558540420024",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "41b94706-325e-4704-a433-d65105758836", trunk.ID)

	assertRequestJSON(t, *capturedBodyPtr, "voice_in_trunks/create_request.json")
}

func TestVoiceInTrunksCreateSipWithReroutingCodes(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/voice_in_trunks": {status: http.StatusCreated, fixture: "voice_in_trunks/create.json"},
	})

	_, err := server.client.VoiceInTrunks().Create(context.Background(), &resource.VoiceInTrunk{
		Name: "hello, test sip trunk",
		Configuration: &trunkconfiguration.SIPConfiguration{
			Username:           "username",
			Host:               "216.58.215.110",
			SstRefreshMethodID: enums.SstRefreshMethodInvite,
			Port:               5060,
			CodecIDs: []enums.Codec{
				enums.CodecPCMU, enums.CodecPCMA, enums.CodecG729, enums.CodecG723, enums.CodecTelephoneEvent,
			},
			ReroutingDisconnectCodeIDs: []enums.ReroutingDisconnectCode{
				enums.DCSIP400BadRequest, enums.DCSIP402PaymentRequired, enums.DCSIP403Forbidden,
				enums.DCSIP404NotFound, enums.DCSIP408RequestTimeout, enums.DCSIP409Conflict,
				enums.DCSIP410Gone, enums.DCSIP412ConditionalRequestFail, enums.DCSIP413RequestEntityTooLarge,
				enums.DCSIP414RequestURITooLong, enums.DCSIP415UnsupportedMediaType, enums.DCSIP416UnsupportedURIScheme,
				enums.DCSIP417UnknownResourcePriority, enums.DCSIP420BadExtension, enums.DCSIP421ExtensionRequired,
				enums.DCSIP422SessionIntervalTooSmall, enums.DCSIP423IntervalTooBrief, enums.DCSIP424BadLocationInformation,
				enums.DCSIP428UseIdentityHeader, enums.DCSIP429ProvideReferrerIdentity,
				enums.DCSIP433AnonymityDisallowed, enums.DCSIP436BadIdentityInfo, enums.DCSIP437UnsupportedCertificate,
				enums.DCSIP438InvalidIdentityHeader, enums.DCSIP480TemporarilyUnavailable, enums.DCSIP482LoopDetected,
				enums.DCSIP483TooManyHops, enums.DCSIP484AddressIncomplete, enums.DCSIP485Ambiguous,
				enums.DCSIP486BusyHere, enums.DCSIP487RequestTerminated, enums.DCSIP488NotAcceptableHere,
				enums.DCSIP494SecurityAgreementReq, enums.DCSIP500ServerInternalError, enums.DCSIP501NotImplemented,
				enums.DCSIP502BadGateway, enums.DCSIP504ServerTimeout, enums.DCSIP505VersionNotSupported,
				enums.DCSIP513MessageTooLarge, enums.DCSIP580PreconditionFailure, enums.DCSIP600BusyEverywhere,
				enums.DCSIP603Decline, enums.DCSIP604DoesNotExistAnywhere, enums.DCSIP606NotAcceptable,
				enums.DCRingingTimeout,
			},
			MediaEncryptionMode: enums.MediaEncryptionModeZrtp,
			StirShakenMode:      enums.StirShakenModePai,
			AllowedRtpIPs:       []string{"127.0.0.1"},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "voice_in_trunks/create_sip_request.json")
}

func TestVoiceInTrunksCreateSip(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/voice_in_trunks": {status: http.StatusCreated, fixture: "voice_in_trunks/create_sip.json"},
	})

	trunk, err := client.VoiceInTrunks().Create(context.Background(), &resource.VoiceInTrunk{
		Name: "hello, test sip trunk",
		Configuration: &trunkconfiguration.SIPConfiguration{
			Username: "username",
			Host:     "216.58.215.110",
			Port:     5060,
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "a80006b6-4183-4865-8b99-7ebbd359a762", trunk.ID)
	assert.Equal(t, "hello, test sip trunk", trunk.Name)
	sipCfg, ok := trunk.Configuration.(*trunkconfiguration.SIPConfiguration)
	require.True(t, ok, "expected SIP configuration")
	assert.Equal(t, "username", sipCfg.Username)
	assert.Equal(t, "216.58.215.110", sipCfg.Host)
}

func TestVoiceInTrunksUpdatePstn(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/voice_in_trunks/41b94706-325e-4704-a433-d65105758836": {status: http.StatusOK, fixture: "voice_in_trunks/update_pstn.json"},
	})

	trunk, err := client.VoiceInTrunks().Update(context.Background(), &resource.VoiceInTrunk{
		ID:   "41b94706-325e-4704-a433-d65105758836",
		Name: "hello, updated test pstn trunk",
		Configuration: &trunkconfiguration.PSTNConfiguration{
			Dst: "558540420025",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "41b94706-325e-4704-a433-d65105758836", trunk.ID)
	assert.Equal(t, "hello, updated test pstn trunk", trunk.Name)
	pstnCfg, ok := trunk.Configuration.(*trunkconfiguration.PSTNConfiguration)
	require.True(t, ok, "expected PSTN configuration")
	assert.Equal(t, "558540420025", pstnCfg.Dst)
}

func TestVoiceInTrunksUpdateSip(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/voice_in_trunks/a80006b6-4183-4865-8b99-7ebbd359a762": {status: http.StatusOK, fixture: "voice_in_trunks/update_sip.json"},
	})

	desc := "just a description"
	trunk, err := client.VoiceInTrunks().Update(context.Background(), &resource.VoiceInTrunk{
		ID:          "a80006b6-4183-4865-8b99-7ebbd359a762",
		Name:        "hello, updated test sip trunk",
		Description: &desc,
		Configuration: &trunkconfiguration.SIPConfiguration{
			Username:     "new-username",
			Host:         "216.58.215.110",
			MaxTransfers: 5,
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "a80006b6-4183-4865-8b99-7ebbd359a762", trunk.ID)
	assert.Equal(t, "hello, updated test sip trunk", trunk.Name)
	sipCfg, ok := trunk.Configuration.(*trunkconfiguration.SIPConfiguration)
	require.True(t, ok, "expected SIP configuration")
	assert.Equal(t, "new-username", sipCfg.Username)
	assert.Equal(t, 5, sipCfg.MaxTransfers)
}

func TestVoiceInTrunksDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/voice_in_trunks/2b4b1fcf-fe6a-4de9-8a58-7df46820ba13": {status: http.StatusNoContent},
	})

	err := client.VoiceInTrunks().Delete(context.Background(), "2b4b1fcf-fe6a-4de9-8a58-7df46820ba13")
	require.NoError(t, err)
}
