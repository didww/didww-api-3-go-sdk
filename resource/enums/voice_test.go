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
		expected int
	}{
		{"TelephoneEvent", CodecTelephoneEvent, 6},
		{"G723", CodecG723, 7},
		{"G729", CodecG729, 8},
		{"PCMU", CodecPCMU, 9},
		{"PCMA", CodecPCMA, 10},
		{"Speex", CodecSpeex, 12},
		{"GSM", CodecGSM, 13},
		{"G726_32", CodecG726_32, 14},
		{"G721", CodecG721, 15},
		{"G726_24", CodecG726_24, 16},
		{"G726_40", CodecG726_40, 17},
		{"G726_16", CodecG726_16, 18},
		{"L16", CodecL16, 19},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.value) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, int(tt.value))
			}
		})
	}
}

func TestCodecJSONMarshal(t *testing.T) {
	testIntEnumJSON(t, CodecPCMU, `9`)
}

func TestCodecJSONArrayMarshal(t *testing.T) {
	codecs := []Codec{CodecPCMU, CodecPCMA, CodecG729}
	data, err := json.Marshal(codecs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `[9,10,8]`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestCodecJSONArrayUnmarshal(t *testing.T) {
	input := `[9,10,8]`
	var codecs []Codec
	err := json.Unmarshal([]byte(input), &codecs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(codecs) != 3 {
		t.Fatalf("expected 3 codecs, got %d", len(codecs))
	}
	if codecs[0] != CodecPCMU {
		t.Errorf("expected CodecPCMU, got %d", codecs[0])
	}
	if codecs[1] != CodecPCMA {
		t.Errorf("expected CodecPCMA, got %d", codecs[1])
	}
	if codecs[2] != CodecG729 {
		t.Errorf("expected CodecG729, got %d", codecs[2])
	}
}
