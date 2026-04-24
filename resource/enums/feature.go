package enums

// Feature defines a DID feature capability.
type Feature string

const (
	FeatureVoiceIn   Feature = "voice_in"
	FeatureVoiceOut  Feature = "voice_out"
	FeatureT38       Feature = "t38"
	FeatureSmsIn     Feature = "sms_in"
	FeatureP2P       Feature = "p2p"
	FeatureA2P       Feature = "a2p"
	FeatureEmergency Feature = "emergency"
	FeatureCnamOut   Feature = "cnam_out"
)
