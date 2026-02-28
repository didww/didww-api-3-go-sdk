package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func TestVoiceInTrunksList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_in_trunks": {status: http.StatusOK, fixture: "voice_in_trunks/index.json"},
	})

	trunks, err := client.VoiceInTrunks().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(trunks) != 2 {
		t.Fatalf("expected 2 voice in trunks, got %d", len(trunks))
	}

	// First trunk is PSTN
	pstn := trunks[0]
	if pstn.ID != "2b4b1fcf-fe6a-4de9-8a58-7df46820ba13" {
		t.Errorf("expected ID '2b4b1fcf-fe6a-4de9-8a58-7df46820ba13', got %q", pstn.ID)
	}
	if pstn.Name != "sample trunk pstn" {
		t.Errorf("expected Name 'sample trunk pstn', got %q", pstn.Name)
	}
	if pstn.CliFormat != enums.CliFormatE164 {
		t.Errorf("expected CliFormat 'e164', got %q", pstn.CliFormat)
	}
	pstnCfg, ok := pstn.Configuration.(*PSTNConfiguration)
	if !ok {
		t.Fatal("expected PSTN configuration")
	}
	if pstnCfg.Dst != "442080995011" {
		t.Errorf("expected Dst '442080995011', got %q", pstnCfg.Dst)
	}

	// Second trunk is SIP
	sip := trunks[1]
	if sip.Name != "Sip trunk sample" {
		t.Errorf("expected Name 'Sip trunk sample', got %q", sip.Name)
	}
	sipCfg, ok := sip.Configuration.(*SIPConfiguration)
	if !ok {
		t.Fatal("expected SIP configuration")
	}
	if sipCfg.Host != "216.58.215.78" {
		t.Errorf("expected Host '216.58.215.78', got %q", sipCfg.Host)
	}
}

func TestVoiceInTrunksCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/voice_in_trunks": {status: http.StatusCreated, fixture: "voice_in_trunks/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	trunk, err := server.client.VoiceInTrunks().Create(context.Background(), &VoiceInTrunk{
		Name: "hello, test pstn trunk",
		Configuration: &PSTNConfiguration{
			Dst: "558540420024",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trunk.ID != "41b94706-325e-4704-a433-d65105758836" {
		t.Errorf("expected ID '41b94706-325e-4704-a433-d65105758836', got %q", trunk.ID)
	}

	assertRequestJSON(t, capturedBody, "voice_in_trunks/create_request.json")
}

func TestVoiceInTrunksCreateSipWithReroutingCodes(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/voice_in_trunks": {status: http.StatusCreated, fixture: "voice_in_trunks/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.VoiceInTrunks().Create(context.Background(), &VoiceInTrunk{
		Name: "hello, test sip trunk",
		Configuration: &SIPConfiguration{
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "voice_in_trunks/create_sip_request.json")
}

func TestVoiceInTrunksCreateSip(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/voice_in_trunks": {status: http.StatusCreated, fixture: "voice_in_trunks/create_sip.json"},
	})

	trunk, err := client.VoiceInTrunks().Create(context.Background(), &VoiceInTrunk{
		Name: "hello, test sip trunk",
		Configuration: &SIPConfiguration{
			Username: "username",
			Host:     "216.58.215.110",
			Port:     5060,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trunk.ID != "a80006b6-4183-4865-8b99-7ebbd359a762" {
		t.Errorf("expected ID 'a80006b6-4183-4865-8b99-7ebbd359a762', got %q", trunk.ID)
	}
	if trunk.Name != "hello, test sip trunk" {
		t.Errorf("expected Name 'hello, test sip trunk', got %q", trunk.Name)
	}
	sipCfg, ok := trunk.Configuration.(*SIPConfiguration)
	if !ok {
		t.Fatal("expected SIP configuration")
	}
	if sipCfg.Username != "username" {
		t.Errorf("expected Username 'username', got %q", sipCfg.Username)
	}
	if sipCfg.Host != "216.58.215.110" {
		t.Errorf("expected Host '216.58.215.110', got %q", sipCfg.Host)
	}
}

func TestVoiceInTrunksUpdatePstn(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/voice_in_trunks/41b94706-325e-4704-a433-d65105758836": {status: http.StatusOK, fixture: "voice_in_trunks/update_pstn.json"},
	})

	trunk, err := client.VoiceInTrunks().Update(context.Background(), &VoiceInTrunk{
		ID:   "41b94706-325e-4704-a433-d65105758836",
		Name: "hello, updated test pstn trunk",
		Configuration: &PSTNConfiguration{
			Dst: "558540420025",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trunk.ID != "41b94706-325e-4704-a433-d65105758836" {
		t.Errorf("expected ID '41b94706-325e-4704-a433-d65105758836', got %q", trunk.ID)
	}
	if trunk.Name != "hello, updated test pstn trunk" {
		t.Errorf("expected Name 'hello, updated test pstn trunk', got %q", trunk.Name)
	}
	pstnCfg, ok := trunk.Configuration.(*PSTNConfiguration)
	if !ok {
		t.Fatal("expected PSTN configuration")
	}
	if pstnCfg.Dst != "558540420025" {
		t.Errorf("expected Dst '558540420025', got %q", pstnCfg.Dst)
	}
}

func TestVoiceInTrunksUpdateSip(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/voice_in_trunks/a80006b6-4183-4865-8b99-7ebbd359a762": {status: http.StatusOK, fixture: "voice_in_trunks/update_sip.json"},
	})

	desc := "just a description"
	trunk, err := client.VoiceInTrunks().Update(context.Background(), &VoiceInTrunk{
		ID:          "a80006b6-4183-4865-8b99-7ebbd359a762",
		Name:        "hello, updated test sip trunk",
		Description: &desc,
		Configuration: &SIPConfiguration{
			Username:     "new-username",
			Host:         "216.58.215.110",
			MaxTransfers: 5,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trunk.ID != "a80006b6-4183-4865-8b99-7ebbd359a762" {
		t.Errorf("expected ID 'a80006b6-4183-4865-8b99-7ebbd359a762', got %q", trunk.ID)
	}
	if trunk.Name != "hello, updated test sip trunk" {
		t.Errorf("expected Name 'hello, updated test sip trunk', got %q", trunk.Name)
	}
	sipCfg, ok := trunk.Configuration.(*SIPConfiguration)
	if !ok {
		t.Fatal("expected SIP configuration")
	}
	if sipCfg.Username != "new-username" {
		t.Errorf("expected Username 'new-username', got %q", sipCfg.Username)
	}
	if sipCfg.MaxTransfers != 5 {
		t.Errorf("expected MaxTransfers 5, got %d", sipCfg.MaxTransfers)
	}
}

func TestVoiceInTrunksDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/voice_in_trunks/2b4b1fcf-fe6a-4de9-8a58-7df46820ba13": {status: http.StatusNoContent},
	})

	err := client.VoiceInTrunks().Delete(context.Background(), "2b4b1fcf-fe6a-4de9-8a58-7df46820ba13")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
