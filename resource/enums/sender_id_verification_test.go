package enums

import "testing"

func TestSenderIDVerificationStatus(t *testing.T) {
	tests := []struct {
		name     string
		value    SenderIDVerificationStatus
		expected string
	}{
		{"Active", SenderIDVerificationStatusActive, "active"},
		{"ChangesRequired", SenderIDVerificationStatusChangesRequired, "changes_required"},
		{"Closed", SenderIDVerificationStatusClosed, "closed"},
		{"ClosingInProcess", SenderIDVerificationStatusClosingInProcess, "closing_in_process"},
		{"InProcess", SenderIDVerificationStatusInProcess, "in_process"},
		{"New", SenderIDVerificationStatusNew, "new"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestSenderType(t *testing.T) {
	tests := []struct {
		name     string
		value    SenderType
		expected string
	}{
		{"Alphanumeric", SenderTypeAlphanumeric, "alphanumeric"},
		{"Carrier", SenderTypeCarrier, "carrier"},
		{"LongCode", SenderTypeLongCode, "long_code"},
		{"ShortCode", SenderTypeShortCode, "short_code"},
		{"TollFree", SenderTypeTollFree, "toll_free"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestSenderIDVerificationEnumsJSON(t *testing.T) {
	testStringEnumJSON(t, SenderIDVerificationStatusClosingInProcess, `"closing_in_process"`)
	testStringEnumJSON(t, SenderTypeShortCode, `"short_code"`)
}
