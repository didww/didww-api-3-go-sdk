package enums

import (
	"encoding/json"
	"testing"
)

func TestReroutingDisconnectCodeValues(t *testing.T) {
	tests := []struct {
		name     string
		value    ReroutingDisconnectCode
		expected int
	}{
		{"SIP400", DCSIP400BadRequest, 56},
		{"SIP401", DCSIP401Unauthorized, 57},
		{"SIP402", DCSIP402PaymentRequired, 58},
		{"SIP403", DCSIP403Forbidden, 59},
		{"SIP404", DCSIP404NotFound, 60},
		{"SIP408", DCSIP408RequestTimeout, 64},
		{"SIP409", DCSIP409Conflict, 65},
		{"SIP410", DCSIP410Gone, 66},
		{"SIP412", DCSIP412ConditionalRequestFail, 67},
		{"SIP413", DCSIP413RequestEntityTooLarge, 68},
		{"SIP414", DCSIP414RequestURITooLong, 69},
		{"SIP415", DCSIP415UnsupportedMediaType, 70},
		{"SIP416", DCSIP416UnsupportedURIScheme, 71},
		{"SIP417", DCSIP417UnknownResourcePriority, 72},
		{"SIP420", DCSIP420BadExtension, 73},
		{"SIP421", DCSIP421ExtensionRequired, 74},
		{"SIP422", DCSIP422SessionIntervalTooSmall, 75},
		{"SIP423", DCSIP423IntervalTooBrief, 76},
		{"SIP424", DCSIP424BadLocationInformation, 77},
		{"SIP428", DCSIP428UseIdentityHeader, 78},
		{"SIP429", DCSIP429ProvideReferrerIdentity, 79},
		{"SIP433", DCSIP433AnonymityDisallowed, 80},
		{"SIP436", DCSIP436BadIdentityInfo, 81},
		{"SIP437", DCSIP437UnsupportedCertificate, 82},
		{"SIP438", DCSIP438InvalidIdentityHeader, 83},
		{"SIP480", DCSIP480TemporarilyUnavailable, 84},
		{"SIP482", DCSIP482LoopDetected, 86},
		{"SIP483", DCSIP483TooManyHops, 87},
		{"SIP484", DCSIP484AddressIncomplete, 88},
		{"SIP485", DCSIP485Ambiguous, 89},
		{"SIP486", DCSIP486BusyHere, 90},
		{"SIP487", DCSIP487RequestTerminated, 91},
		{"SIP488", DCSIP488NotAcceptableHere, 92},
		{"SIP494", DCSIP494SecurityAgreementReq, 96},
		{"SIP500", DCSIP500ServerInternalError, 97},
		{"SIP501", DCSIP501NotImplemented, 98},
		{"SIP502", DCSIP502BadGateway, 99},
		{"SIP503", DCSIP503ServiceUnavailable, 100},
		{"SIP504", DCSIP504ServerTimeout, 101},
		{"SIP505", DCSIP505VersionNotSupported, 102},
		{"SIP513", DCSIP513MessageTooLarge, 103},
		{"SIP580", DCSIP580PreconditionFailure, 104},
		{"SIP600", DCSIP600BusyEverywhere, 105},
		{"SIP603", DCSIP603Decline, 106},
		{"SIP604", DCSIP604DoesNotExistAnywhere, 107},
		{"SIP606", DCSIP606NotAcceptable, 108},
		{"RingingTimeout", DCRingingTimeout, 1505},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.value) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, int(tt.value))
			}
		})
	}
}

func TestReroutingDisconnectCodeJSONArray(t *testing.T) {
	codes := []ReroutingDisconnectCode{
		DCSIP486BusyHere,
		DCRingingTimeout,
	}
	data, err := json.Marshal(codes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `[90,1505]`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}
