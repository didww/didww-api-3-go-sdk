package enums

import "testing"

func TestAddressVerificationStatus(t *testing.T) {
	tests := []struct {
		name     string
		value    AddressVerificationStatus
		expected string
	}{
		{"Pending", AddressVerificationStatusPending, "pending"},
		{"Approved", AddressVerificationStatusApproved, "approved"},
		{"Rejected", AddressVerificationStatusRejected, "rejected"},
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
		{"WorldWide", AreaLevelWorldWide, "world_wide"},
		{"Country", AreaLevelCountry, "country"},
		{"Area", AreaLevelArea, "area"},
		{"City", AreaLevelCity, "city"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}
