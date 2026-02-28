package enums

import "testing"

func TestAddressVerificationStatus(t *testing.T) {
	tests := []struct {
		name     string
		value    AddressVerificationStatus
		expected string
	}{
		{"Pending", AddressVerificationStatusPending, "Pending"},
		{"Approved", AddressVerificationStatusApproved, "Approved"},
		{"Rejected", AddressVerificationStatusRejected, "Rejected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestAreaLevel(t *testing.T) {
	tests := []struct {
		name     string
		value    AreaLevel
		expected string
	}{
		{"WorldWide", AreaLevelWorldWide, "WorldWide"},
		{"Country", AreaLevelCountry, "Country"},
		{"Area", AreaLevelArea, "Area"},
		{"City", AreaLevelCity, "City"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}
