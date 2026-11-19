package enums

import "testing"

func TestVoiceOutPriceCategory(t *testing.T) {
	tests := []struct {
		name     string
		value    VoiceOutPriceCategory
		expected string
	}{
		{"International", VoiceOutPriceCategoryInternational, "International"},
		{"OriginBased", VoiceOutPriceCategoryOriginBased, "Origin based"},
		{"Local", VoiceOutPriceCategoryLocal, "Local"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestVoiceOutPriceCategoryJSON(t *testing.T) {
	testStringEnumJSON(t, VoiceOutPriceCategoryOriginBased, `"Origin based"`)
}
