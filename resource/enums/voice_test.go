package enums

import (
	"encoding/json"
	"testing"
)

func TestCliFormat(t *testing.T) {
	tests := []struct {
		name     string
		value    CliFormat
		expected string
	}{
		{"Raw", CliFormatRaw, "raw"},
		{"E164", CliFormatE164, "e164"},
		{"Local", CliFormatLocal, "local"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestCliFormatJSON(t *testing.T) {
	testStringEnumJSON(t, CliFormatE164, `"e164"`)
}

func TestOnCliMismatchAction(t *testing.T) {
	tests := []struct {
		name     string
		value    OnCliMismatchAction
		expected string
	}{
		{"SendOriginalCli", OnCliMismatchActionSendOriginalCli, "send_original_cli"},
		{"RejectCall", OnCliMismatchActionRejectCall, "reject_call"},
		{"ReplaceCli", OnCliMismatchActionReplaceCli, "replace_cli"},
		{"RandomizeCli", OnCliMismatchActionRandomizeCli, "randomize_cli"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestMediaEncryptionMode(t *testing.T) {
	tests := []struct {
		name     string
		value    MediaEncryptionMode
		expected string
	}{
		{"Disabled", MediaEncryptionModeDisabled, "disabled"},
		{"SrtpSdes", MediaEncryptionModeSrtpSdes, "srtp_sdes"},
		{"SrtpDtls", MediaEncryptionModeSrtpDtls, "srtp_dtls"},
		{"Zrtp", MediaEncryptionModeZrtp, "zrtp"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestDefaultDstAction(t *testing.T) {
	tests := []struct {
		name     string
		value    DefaultDstAction
		expected string
	}{
		{"AllowAll", DefaultDstActionAllowAll, "allow_all"},
		{"RejectAll", DefaultDstActionRejectAll, "reject_all"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestVoiceOutTrunkStatus(t *testing.T) {
	tests := []struct {
		name     string
		value    VoiceOutTrunkStatus
		expected string
	}{
		{"Active", VoiceOutTrunkStatusActive, "active"},
		{"Blocked", VoiceOutTrunkStatusBlocked, "blocked"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestStirShakenMode(t *testing.T) {
	tests := []struct {
		name     string
		value    StirShakenMode
		expected string
	}{
		{"Disabled", StirShakenModeDisabled, "disabled"},
		{"Original", StirShakenModeOriginal, "original"},
		{"Pai", StirShakenModePai, "pai"},
		{"OriginalPai", StirShakenModeOriginalPai, "original_pai"},
		{"Verstat", StirShakenModeVerstat, "verstat"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestTransportProtocol(t *testing.T) {
	tests := []struct {
		name     string
		value    TransportProtocol
		expected int
	}{
		{"UDP", TransportProtocolUDP, 1},
		{"TCP", TransportProtocolTCP, 2},
		{"TLS", TransportProtocolTLS, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.value) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, int(tt.value))
			}
		})
	}
}

func TestTransportProtocolJSON(t *testing.T) {
	testIntEnumJSON(t, TransportProtocolUDP, `1`)
}

func TestRxDtmfFormat(t *testing.T) {
	tests := []struct {
		name     string
		value    RxDtmfFormat
		expected int
	}{
		{"RFC2833", RxDtmfFormatRFC2833, 1},
		{"SIPInfo", RxDtmfFormatSIPInfo, 2},
		{"RFC2833OrSIPInfo", RxDtmfFormatRFC2833OrSIPInfo, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.value) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, int(tt.value))
			}
		})
	}
}

func TestTxDtmfFormat(t *testing.T) {
	tests := []struct {
		name     string
		value    TxDtmfFormat
		expected int
	}{
		{"Disabled", TxDtmfFormatDisabled, 1},
		{"RFC2833", TxDtmfFormatRFC2833, 2},
		{"SIPInfoRelay", TxDtmfFormatSIPInfoRelay, 3},
		{"SIPInfoDtmf", TxDtmfFormatSIPInfoDtmf, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.value) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, int(tt.value))
			}
		})
	}
}

func TestSstRefreshMethod(t *testing.T) {
	tests := []struct {
		name     string
		value    SstRefreshMethod
		expected int
	}{
		{"Invite", SstRefreshMethodInvite, 1},
		{"Update", SstRefreshMethodUpdate, 2},
		{"UpdateFallbackInvite", SstRefreshMethodUpdateFallbackInvite, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.value) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, int(tt.value))
			}
		})
	}
}

func TestCodecValues(t *testing.T) {
	tests := []struct {
		name     string
		value    Codec
		expected string
	}{
		{"TelephoneEvent", CodecTelephoneEvent, "telephone_event"},
		{"G723", CodecG723, "g723"},
		{"G729", CodecG729, "g729"},
		{"PCMU", CodecPCMU, "pcmu"},
		{"PCMA", CodecPCMA, "pcma"},
		{"Speex", CodecSpeex, "speex"},
		{"GSM", CodecGSM, "gsm"},
		{"G726_32", CodecG726_32, "g726_32"},
		{"G721", CodecG721, "g721"},
		{"G726_24", CodecG726_24, "g726_24"},
		{"G726_40", CodecG726_40, "g726_40"},
		{"G726_16", CodecG726_16, "g726_16"},
		{"L16", CodecL16, "l16"},
		{"AMR", CodecAMR, "amr"},
		{"AMRWB", CodecAMRWB, "amr_wb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestCodecJSON(t *testing.T) {
	testStringEnumJSON(t, CodecAMRWB, `"amr_wb"`)
}

func TestCodecJSONArray(t *testing.T) {
	codecs := []Codec{CodecPCMU, CodecPCMA, CodecTelephoneEvent}
	data, err := json.Marshal(codecs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `["pcmu","pcma","telephone_event"]` {
		t.Errorf("unexpected JSON %s", data)
	}
	var decoded []Codec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(decoded) != 3 || decoded[0] != CodecPCMU || decoded[1] != CodecPCMA || decoded[2] != CodecTelephoneEvent {
		t.Errorf("unexpected codecs %v", decoded)
	}
}

func TestVoiceOutCodecValues(t *testing.T) {
	tests := []struct {
		name     string
		value    VoiceOutCodec
		expected string
	}{
		{"TelephoneEvent", VoiceOutCodecTelephoneEvent, "telephone_event"},
		{"G723", VoiceOutCodecG723, "g723"},
		{"G729", VoiceOutCodecG729, "g729"},
		{"PCMU", VoiceOutCodecPCMU, "pcmu"},
		{"PCMA", VoiceOutCodecPCMA, "pcma"},
		{"GSM", VoiceOutCodecGSM, "gsm"},
		{"G722", VoiceOutCodecG722, "g722"},
		{"Opus", VoiceOutCodecOpus, "opus"},
		{"AMR", VoiceOutCodecAMR, "amr"},
		{"AMRWB", VoiceOutCodecAMRWB, "amr_wb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestVoiceOutCodecJSON(t *testing.T) {
	testStringEnumJSON(t, VoiceOutCodecOpus, `"opus"`)
}

func TestTrunkMediaModes(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"IceModeDisabled", string(IceModeDisabled), "disabled"},
		{"IceModeAccept", string(IceModeAccept), "accept"},
		{"IceModeInitiate", string(IceModeInitiate), "initiate"},
		{"RtcpMuxModeDisabled", string(RtcpMuxModeDisabled), "disabled"},
		{"RtcpMuxModeAccept", string(RtcpMuxModeAccept), "accept"},
		{"RtcpMuxModeInitiate", string(RtcpMuxModeInitiate), "initiate"},
		{"RtcpFeedbackModeDisabled", string(RtcpFeedbackModeDisabled), "disabled"},
		{"RtcpFeedbackModeAccept", string(RtcpFeedbackModeAccept), "accept"},
		{"RtcpFeedbackModeInitiate", string(RtcpFeedbackModeInitiate), "initiate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.value)
			}
		})
	}
}

func TestPop(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"NYC", PopNYC, "nyc"},
		{"FRA", PopFRA, "fra"},
		{"LAC", PopLAC, "lac"},
		{"MIA", PopMIA, "mia"},
		{"SG", PopSG, "sg"},
		{"HK", PopHK, "hk"},
		{"AMS", PopAMS, "ams"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.value)
			}
		})
	}
}
