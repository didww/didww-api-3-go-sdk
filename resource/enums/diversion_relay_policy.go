package enums

// DiversionRelayPolicy defines the diversion header relay policy for SIP INVITE.
type DiversionRelayPolicy string

const (
	DiversionRelayPolicyNone DiversionRelayPolicy = "none"
	DiversionRelayPolicyAsIs DiversionRelayPolicy = "as_is"
	DiversionRelayPolicySIP  DiversionRelayPolicy = "sip"
	DiversionRelayPolicyTel  DiversionRelayPolicy = "tel"
)
