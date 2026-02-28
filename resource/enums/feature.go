package enums

// Feature defines a DID feature capability.
type Feature string

const (
	FeatureVoiceIn  Feature = "voice_in"
	FeatureVoiceOut Feature = "voice_out"
	FeatureT38      Feature = "t38"
	FeatureSmsIn    Feature = "sms_in"
	FeatureSmsOut   Feature = "sms_out"
)
