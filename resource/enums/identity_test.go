package enums

import "testing"

func TestIdentityType(t *testing.T) {
	tests := []struct {
		name     string
		value    IdentityType
		expected string
	}{
		{"Personal", IdentityTypePersonal, "personal"},
		{"Business", IdentityTypeBusiness, "business"},
		{"Any", IdentityTypeAny, "any"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}
