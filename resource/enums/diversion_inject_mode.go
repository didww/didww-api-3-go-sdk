package enums

// DiversionInjectMode defines how the Diversion header is injected on SIP INVITE.
// Available in API 2026-04-16.
type DiversionInjectMode string

const (
	DiversionInjectModeNone      DiversionInjectMode = "none"
	DiversionInjectModeDIDNumber DiversionInjectMode = "did_number"
)
