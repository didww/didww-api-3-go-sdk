package enums

// NetworkProtocolPriority defines the SIP network protocol priority preference.
// Available in API 2026-04-16.
type NetworkProtocolPriority string

const (
	NetworkProtocolPriorityForceIPv4  NetworkProtocolPriority = "force_ipv4"
	NetworkProtocolPriorityForceIPv6  NetworkProtocolPriority = "force_ipv6"
	NetworkProtocolPriorityAny        NetworkProtocolPriority = "any"
	NetworkProtocolPriorityPreferIPv4 NetworkProtocolPriority = "prefer_ipv4"
	NetworkProtocolPriorityPreferIPv6 NetworkProtocolPriority = "prefer_ipv6"
)
