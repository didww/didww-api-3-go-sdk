package enums

// VoiceOutPriceCategory defines the category of a Voice OUT price.
type VoiceOutPriceCategory string

const (
	VoiceOutPriceCategoryInternational VoiceOutPriceCategory = "International"
	VoiceOutPriceCategoryOriginBased   VoiceOutPriceCategory = "Origin based"
	VoiceOutPriceCategoryLocal         VoiceOutPriceCategory = "Local"
)
