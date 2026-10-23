package enums

import "testing"

func TestNumberType(t *testing.T) {
	tests := []struct {
		name     string
		value    NumberType
		expected string
	}{
		{"Local", NumberTypeLocal, "local"},
		{"National", NumberTypeNational, "national"},
		{"Mobile", NumberTypeMobile, "mobile"},
		{"TollFree", NumberTypeTollFree, "toll_free"},
		{"SharedCost", NumberTypeSharedCost, "shared_cost"},
		{"Global", NumberTypeGlobal, "global"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestNumberTypeJSON(t *testing.T) {
	testStringEnumJSON(t, NumberTypeTollFree, `"toll_free"`)
}
