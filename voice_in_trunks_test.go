package didww

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/trunkconfiguration"

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
	assert.Equal(t, "203.0.113.78", sipCfg.Host)
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
			Host:               "203.0.113.110",
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
			AllowedRtpIPs:       []string{"203.0.113.1"},
			// API 2026-04-16 writable attributes.
			//
			// Note: `EnabledSipRegistration` and `UseDIDInRuri` are bool
			// fields with `omitempty`, so leaving them at their zero value
			// (false) keeps them out of the wire body — which is what the
			// API expects for a non-registered SIP trunk.  The dedicated
			// SIP-registration test exercises the true case.
			DiversionRelayPolicy:    enums.DiversionRelayPolicyAsIs,
			DiversionInjectMode:     enums.DiversionInjectModeDIDNumber,
			NetworkProtocolPriority: enums.NetworkProtocolPriorityForceIPv4,
			CnamLookup:              Ptr(true),
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
			Host:     "203.0.113.110",
			Port:     5060,
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "a80006b6-4183-4865-8b99-7ebbd359a762", trunk.ID)
	assert.Equal(t, "hello, test sip trunk", trunk.Name)
	sipCfg, ok := trunk.Configuration.(*trunkconfiguration.SIPConfiguration)
	require.True(t, ok, "expected SIP configuration")
	assert.Equal(t, "username", sipCfg.Username)
	assert.Equal(t, "203.0.113.110", sipCfg.Host)
	assert.Equal(t, enums.DiversionRelayPolicyAsIs, sipCfg.DiversionRelayPolicy)
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
			Host:         "203.0.113.110",
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

// API 2026-04-16 SIP-registration attributes.
//
// Verifies that the new SipConfiguration fields round-trip correctly and
// that the read-only incoming_auth_* credentials returned by the server are
// stripped from POST/PATCH request bodies (the API returns 400 Param not
// allowed if a client tries to write them).
func TestSIPConfigurationRegistrationWritableAttributesSerialize(t *testing.T) {
	// Host intentionally absent: sip_registration enabled + host present is
	// invalid per the server, and the MarshalJSON cascade would reset
	// EnabledSipRegistration to false to keep the wire consistent.
	cfg := trunkconfiguration.SIPConfiguration{
		EnabledSipRegistration:  Ptr(true),
		UseDIDInRuri:            Ptr(true),
		CnamLookup:              Ptr(true),
		DiversionInjectMode:     enums.DiversionInjectModeDIDNumber,
		NetworkProtocolPriority: enums.NetworkProtocolPriorityPreferIPv4,
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, `"enabled_sip_registration":true`)
	assert.Contains(t, out, `"use_did_in_ruri":true`)
	assert.Contains(t, out, `"cnam_lookup":true`)
	assert.Contains(t, out, `"diversion_inject_mode":"did_number"`)
	assert.Contains(t, out, `"network_protocol_priority":"prefer_ipv4"`)
}

func TestSIPConfigurationStripsReadOnlyCredentialsOnSerialize(t *testing.T) {
	cfg := trunkconfiguration.SIPConfiguration{
		EnabledSipRegistration: Ptr(true),
		UseDIDInRuri:           Ptr(true),
		IncomingAuthUsername:   "sipreg-user-1",
		IncomingAuthPassword:   "s3cret-Pa55",
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, `"enabled_sip_registration":true`)
	assert.Contains(t, out, `"use_did_in_ruri":true`)
	assert.NotContains(t, out, "incoming_auth_username")
	assert.NotContains(t, out, "incoming_auth_password")
}

// Regression test for the disable-sip_registration PATCH flow.
//
// `EnabledSipRegistration`, `UseDIDInRuri`, and `CnamLookup` are *bool
// (pointers) — a plain `bool + json:",omitempty"` would silently drop
// any `false` value, breaking the documented disable PATCH flow that
// has to send `enabled_sip_registration: false` together with
// `use_did_in_ruri: false` and a non-blank `host` in the same body.
//
// If anyone reverts these fields to plain `bool`, this test fails because
// the explicit `false` values are dropped from the JSON output by
// `json:",omitempty"`.
func TestSIPConfigurationDisableFlowSerializesExplicitFalse(t *testing.T) {
	cfg := trunkconfiguration.SIPConfiguration{
		Host:                   "203.0.113.10",
		EnabledSipRegistration: Ptr(false),
		UseDIDInRuri:           Ptr(false),
		CnamLookup:             Ptr(false),
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, `"enabled_sip_registration":false`,
		"explicit false must be serialized; reverting to plain bool drops it")
	assert.Contains(t, out, `"use_did_in_ruri":false`,
		"explicit false must be serialized; reverting to plain bool drops it")
	assert.Contains(t, out, `"cnam_lookup":false`,
		"explicit false must be serialized; reverting to plain bool drops it")
	assert.Contains(t, out, `"host":"203.0.113.10"`)
}

// Default fmt output (Sprintf / Println / %v / %#v) MUST redact SIP
// credentials — log lines / panics / debugger inspection should never
// show plaintext credentials. The wire format is unaffected: MarshalJSON
// continues to emit the real values (or strip read-only ones via the
// `api:"readonly"` tag).
func TestSIPConfigurationStringRedactsCredentials(t *testing.T) {
	cfg := &trunkconfiguration.SIPConfiguration{
		Username:               "alice",
		Host:                   "sip.example.com",
		AuthPassword:           "s3cret-Pa55",
		EnabledSipRegistration: Ptr(true),
		IncomingAuthUsername:   "srv-user-xyz",
		IncomingAuthPassword:   "srv-pass-xyz",
	}
	out := cfg.String()
	assert.Contains(t, out, "alice")
	assert.Contains(t, out, "sip.example.com")
	assert.NotContains(t, out, "s3cret-Pa55")
	assert.NotContains(t, out, "srv-user-xyz")
	assert.NotContains(t, out, "srv-pass-xyz")
	assert.Contains(t, out, "[FILTERED]")

	// %v and %#v go through Stringer / GoStringer.
	verboseV := fmt.Sprintf("%v", cfg)
	verboseHash := fmt.Sprintf("%#v", cfg)
	assert.NotContains(t, verboseV, "s3cret-Pa55")
	assert.NotContains(t, verboseHash, "s3cret-Pa55")

	// Wire format still includes the real values.
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"auth_password":"s3cret-Pa55"`)
}

// Auto-cascade tests: MarshalJSON normalises server-enforced field
// dependencies on the wire so callers do not have to enumerate them.

func TestSIPConfigurationMarshalJSONCascadesEnabledSipRegistrationOnHost(t *testing.T) {
	// Setting Host implies sip_registration is disabled (server-side rule);
	// the cascade adds enabled_sip_registration: false and use_did_in_ruri:
	// false to the wire even when the caller did not set them.
	cfg := trunkconfiguration.SIPConfiguration{Host: "sip.example.com"}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, `"host":"sip.example.com"`)
	assert.Contains(t, out, `"enabled_sip_registration":false`)
	assert.Contains(t, out, `"use_did_in_ruri":false`)
}

func TestSIPConfigurationMarshalJSONCascadesUseDidInRuriOnDisable(t *testing.T) {
	// EnabledSipRegistration: false forces use_did_in_ruri: false on the wire.
	cfg := trunkconfiguration.SIPConfiguration{EnabledSipRegistration: Ptr(false)}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, `"enabled_sip_registration":false`)
	assert.Contains(t, out, `"use_did_in_ruri":false`)
}

func TestSIPConfigurationMarshalJSONLeavesUseDidInRuriOnEnable(t *testing.T) {
	// EnabledSipRegistration: true does not touch use_did_in_ruri — the
	// server allows either value when sip_registration is enabled.
	cfg := trunkconfiguration.SIPConfiguration{EnabledSipRegistration: Ptr(true)}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, `"enabled_sip_registration":true`)
	assert.NotContains(t, out, `"use_did_in_ruri"`)
}

func TestSIPConfigurationMarshalJSONOnFreshConfigEmitsHostAndPortAsNullOnWire(t *testing.T) {
	// Regression: PATCH against an existing trunk that already has a
	// host/port persisted server-side. The local SIPConfiguration starts
	// empty (Host/Port zero-valued), so the cascade must still emit
	// "host": null and "port": null on the wire — otherwise the server
	// merges the new EnabledSipRegistration=true with the persisted host
	// and rejects with 422.
	cfg := trunkconfiguration.SIPConfiguration{EnabledSipRegistration: Ptr(true)}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, `"host":null`)
	assert.Contains(t, out, `"port":null`)
	assert.Contains(t, out, `"enabled_sip_registration":true`)
}

func TestSIPConfigurationMarshalJSONDoesNotMutateInput(t *testing.T) {
	// The cascade applies to a local copy inside MarshalJSON, so the
	// caller's struct must remain untouched. Without this guarantee the
	// SDK would surprise users by retroactively flipping fields after
	// they had passed the value to a serializer.
	cfg := trunkconfiguration.SIPConfiguration{
		Host:                   "sip.example.com",
		EnabledSipRegistration: Ptr(true),
		UseDIDInRuri:           Ptr(true),
	}
	_, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Equal(t, "sip.example.com", cfg.Host, "input Host must not be mutated by Marshal")
	require.NotNil(t, cfg.EnabledSipRegistration)
	assert.True(t, *cfg.EnabledSipRegistration, "input EnabledSipRegistration must not be mutated")
	require.NotNil(t, cfg.UseDIDInRuri)
	assert.True(t, *cfg.UseDIDInRuri, "input UseDIDInRuri must not be mutated")
}

func TestSIPConfigurationDeserializeServerResponseBypassesCascade(t *testing.T) {
	// json.Unmarshal writes directly into struct fields; the cascade lives
	// in MarshalJSON only. Server-returned shapes (regular SIP trunk with
	// host: present + sip_registration: false + use_did_in_ruri: true)
	// deserialize as-is.
	body := `{
		"host": "sip.example.com",
		"port": 5060,
		"enabled_sip_registration": false,
		"use_did_in_ruri": true
	}`
	var cfg trunkconfiguration.SIPConfiguration
	require.NoError(t, json.Unmarshal([]byte(body), &cfg))
	assert.Equal(t, "sip.example.com", cfg.Host)
	assert.Equal(t, 5060, cfg.Port)
	require.NotNil(t, cfg.EnabledSipRegistration)
	assert.False(t, *cfg.EnabledSipRegistration)
	require.NotNil(t, cfg.UseDIDInRuri)
	assert.True(t, *cfg.UseDIDInRuri, "deserialization must not cascade UseDIDInRuri to false")
}

// Companion: when neither host nor the toggle pointers are set, omitempty
// kicks in and the keys are absent — that's how callers express "leave
// these fields alone" in a partial PATCH. (When Host is set, the cascade
// in MarshalJSON populates enabled_sip_registration / use_did_in_ruri on
// the wire — see TestSIPConfigurationMarshalJSONCascadesOnHost.)
func TestSIPConfigurationOmitsBoolPointersWhenNil(t *testing.T) {
	cfg := trunkconfiguration.SIPConfiguration{
		Username: "alice",
		// Host / EnabledSipRegistration / UseDIDInRuri / CnamLookup deliberately empty/nil.
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	out := string(data)
	assert.NotContains(t, out, "enabled_sip_registration")
	assert.NotContains(t, out, "use_did_in_ruri")
	assert.NotContains(t, out, "cnam_lookup")
}

// End-to-end PATCH /voice_in_trunks/:id wire-format check for the disable
// flow.  The server returns 422 for any request that flips
// EnabledSipRegistration to false without simultaneously providing a
// non-blank Host and UseDIDInRuri = false, so the wire body must carry
// all three fields together.
// The capturedBody comparison fails if anyone reverts EnabledSip-
// Registration / UseDIDInRuri to plain bool — the explicit `false`
// values silently drop from `json:",omitempty"` output and the captured
// body no longer matches the fixture.
func TestVoiceInTrunksDisableSipRegistrationPatchSerializesAllThreeFields(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/voice_in_trunks/57a939dd-1600-41a6-80b1-f624e22a1f4c": {
			status:  http.StatusOK,
			fixture: "voice_in_trunks/disable_sip_registration.json",
		},
	})

	trunk, err := server.client.VoiceInTrunks().Update(context.Background(), &resource.VoiceInTrunk{
		ID: "57a939dd-1600-41a6-80b1-f624e22a1f4c",
		Configuration: &trunkconfiguration.SIPConfiguration{
			Host:                   "203.0.113.10",
			EnabledSipRegistration: Ptr(false),
			UseDIDInRuri:           Ptr(false),
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "voice_in_trunks/disable_sip_registration_request.json")

	sipCfg, ok := trunk.Configuration.(*trunkconfiguration.SIPConfiguration)
	require.True(t, ok, "expected SIP configuration")
	require.NotNil(t, sipCfg.EnabledSipRegistration)
	assert.False(t, *sipCfg.EnabledSipRegistration)
	require.NotNil(t, sipCfg.UseDIDInRuri)
	assert.False(t, *sipCfg.UseDIDInRuri)
	assert.Equal(t, "203.0.113.10", sipCfg.Host)
	assert.Empty(t, sipCfg.IncomingAuthUsername)
	assert.Empty(t, sipCfg.IncomingAuthPassword)
}

func TestSIPConfigurationDeserializesIncomingAuthCredentials(t *testing.T) {
	// Real wire shape captured from sandbox: when sip_registration is
	// enabled, host/port/username come back as null and the API rejects
	// any attempt to set them.
	body := `{
		"username": null,
		"host": null,
		"port": null,
		"enabled_sip_registration": true,
		"incoming_auth_username": "sipreg-user-1",
		"incoming_auth_password": "s3cret-Pa55"
	}`
	var cfg trunkconfiguration.SIPConfiguration
	require.NoError(t, json.Unmarshal([]byte(body), &cfg))
	assert.NotNil(t, cfg.EnabledSipRegistration)
	assert.True(t, *cfg.EnabledSipRegistration)
	assert.Equal(t, "sipreg-user-1", cfg.IncomingAuthUsername)
	assert.Equal(t, "s3cret-Pa55", cfg.IncomingAuthPassword)
}

// End-to-end: when the SDK sends `enabled_sip_registration: true`, the
// server returns 201 with server-generated `incoming_auth_username` and
// `incoming_auth_password`. The SDK must surface those populated values to
// the caller, not nil/empty strings. The captureRequestBody helper also
// asserts the outgoing wire body matches the fixture, so a regression
// that drops EnabledSipRegistration / UseDIDInRuri / CnamLookup from the
// POST body fails the request-body diff.
func TestVoiceInTrunksCreateWithEnabledSipRegistrationReturnsPopulatedIncomingAuth(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/voice_in_trunks": {status: http.StatusCreated, fixture: "voice_in_trunks/create_with_sip_registration.json"},
	})

	ringingTimeout := 30
	created, err := server.client.VoiceInTrunks().Create(context.Background(), &resource.VoiceInTrunk{
		Name:           "sip-registration",
		Priority:       1,
		Weight:         100,
		CliFormat:      enums.CliFormatE164,
		RingingTimeout: &ringingTimeout,
		Configuration: &trunkconfiguration.SIPConfiguration{
			EnabledSipRegistration:  Ptr(true),
			UseDIDInRuri:            Ptr(true),
			CnamLookup:              Ptr(true),
			DiversionRelayPolicy:    enums.DiversionRelayPolicyAsIs,
			DiversionInjectMode:     enums.DiversionInjectModeDIDNumber,
			NetworkProtocolPriority: enums.NetworkProtocolPriorityPreferIPv4,
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "voice_in_trunks/create_with_sip_registration_request.json")

	cfg, ok := created.Configuration.(*trunkconfiguration.SIPConfiguration)
	require.True(t, ok, "expected SIP configuration")
	assert.NotNil(t, cfg.EnabledSipRegistration)
	assert.True(t, *cfg.EnabledSipRegistration)
	// Server-generated credentials are populated, not empty.
	assert.NotEmpty(t, cfg.IncomingAuthUsername, "expected incoming_auth_username to be populated")
	assert.NotEmpty(t, cfg.IncomingAuthPassword, "expected incoming_auth_password to be populated")
}

func TestVoiceInTrunksDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/voice_in_trunks/2b4b1fcf-fe6a-4de9-8a58-7df46820ba13": {status: http.StatusNoContent},
	})

	err := client.VoiceInTrunks().Delete(context.Background(), "2b4b1fcf-fe6a-4de9-8a58-7df46820ba13")
	require.NoError(t, err)
}
