package enums

import "testing"

func TestFeature(t *testing.T) {
	tests := []struct {
		name     string
		value    Feature
		expected string
	}{
		{"VoiceIn", FeatureVoiceIn, "voice_in"},
		{"VoiceOut", FeatureVoiceOut, "voice_out"},
		{"T38", FeatureT38, "t38"},
		{"SmsIn", FeatureSmsIn, "sms_in"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}
