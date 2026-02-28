package enums

import "testing"

func TestIdentityType(t *testing.T) {
	tests := []struct {
		name     string
		value    IdentityType
		expected string
	}{
		{"Personal", IdentityTypePersonal, "Personal"},
		{"Business", IdentityTypeBusiness, "Business"},
		{"Any", IdentityTypeAny, "Any"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}
