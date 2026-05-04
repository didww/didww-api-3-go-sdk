package trunkconfiguration

import (
	"fmt"

	"github.com/didww/didww-api-3-go-sdk/v3/jsonapi"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
)

// SIPConfiguration represents a SIP trunk configuration.
//
// Fields tagged `api:"readonly"` (incoming_auth_username, incoming_auth_password)
// are populated from server responses but stripped from POST/PATCH request
// bodies via the custom MarshalJSON below. The server returns 400 Param not
// allowed if a client tries to write them.
type SIPConfiguration struct {
	Username                   string                          `json:"username,omitempty"`
	Host                       string                          `json:"host,omitempty"`
	Port                       int                             `json:"port,omitempty"`
	CodecIDs                   []enums.Codec                   `json:"codec_ids,omitempty"`
	RxDtmfFormatID             enums.RxDtmfFormat              `json:"rx_dtmf_format_id,omitempty"`
	TxDtmfFormatID             enums.TxDtmfFormat              `json:"tx_dtmf_format_id,omitempty"`
	ResolveRuri                bool                            `json:"resolve_ruri,omitempty"`
	AuthEnabled                bool                            `json:"auth_enabled,omitempty"`
	AuthUser                   string                          `json:"auth_user,omitempty"`
	AuthPassword               string                          `json:"auth_password,omitempty"`
	AuthFromUser               string                          `json:"auth_from_user,omitempty"`
	AuthFromDomain             string                          `json:"auth_from_domain,omitempty"`
	SstEnabled                 bool                            `json:"sst_enabled,omitempty"`
	SstMinTimer                int                             `json:"sst_min_timer,omitempty"`
	SstMaxTimer                int                             `json:"sst_max_timer,omitempty"`
	SstAccept501               bool                            `json:"sst_accept_501,omitempty"`
	SipTimerB                  int                             `json:"sip_timer_b,omitempty"`
	DnsSrvFailoverTimer        int                             `json:"dns_srv_failover_timer,omitempty"`
	RtpPing                    bool                            `json:"rtp_ping,omitempty"`
	RtpTimeout                 int                             `json:"rtp_timeout,omitempty"`
	ForceSymmetricRtp          bool                            `json:"force_symmetric_rtp,omitempty"`
	SymmetricRtpIgnoreRtcp     bool                            `json:"symmetric_rtp_ignore_rtcp,omitempty"`
	ReroutingDisconnectCodeIDs []enums.ReroutingDisconnectCode `json:"rerouting_disconnect_code_ids,omitempty"`
	SstSessionExpires          *int                            `json:"sst_session_expires,omitempty"`
	SstRefreshMethodID         enums.SstRefreshMethod          `json:"sst_refresh_method_id,omitempty"`
	TransportProtocolID        enums.TransportProtocol         `json:"transport_protocol_id,omitempty"`
	MaxTransfers               int                             `json:"max_transfers,omitempty"`
	Max30xRedirects            int                             `json:"max_30x_redirects,omitempty"`
	MediaEncryptionMode        enums.MediaEncryptionMode       `json:"media_encryption_mode,omitempty"`
	StirShakenMode             enums.StirShakenMode            `json:"stir_shaken_mode,omitempty"`
	AllowedRtpIPs              []string                        `json:"allowed_rtp_ips,omitempty"`
	DiversionRelayPolicy       enums.DiversionRelayPolicy      `json:"diversion_relay_policy,omitempty"`

	// API 2026-04-16 writable attributes.
	//
	// Server-side validation rules for EnabledSipRegistration:
	//   - When true, the trunk's Host and Port must be left blank
	//     (server returns 422 otherwise).
	//   - When disabling sip registration on an existing trunk, the same
	//     PATCH must also set Host to a non-blank value and UseDIDInRuri
	//     to false, or the server returns 422.
	DiversionInjectMode     enums.DiversionInjectMode     `json:"diversion_inject_mode,omitempty"`
	NetworkProtocolPriority enums.NetworkProtocolPriority `json:"network_protocol_priority,omitempty"`
	// `*bool` (not `bool`) so that explicit false values ARE serialized.
	// The disable-sip_registration PATCH flow requires sending
	// `enabled_sip_registration: false` together with a non-blank `host`
	// and `use_did_in_ruri: false` in the same request — with a plain bool
	// + omitempty, those false values would silently drop from the wire.
	EnabledSipRegistration  *bool                         `json:"enabled_sip_registration,omitempty"`
	UseDIDInRuri            *bool                         `json:"use_did_in_ruri,omitempty"`
	CnamLookup              *bool                         `json:"cnam_lookup,omitempty"`

	// API 2026-04-16 read-only attributes. Server-generated SIP
	// registration credentials, returned only when EnabledSipRegistration is
	// true. The `api:"readonly"` tag makes MarshalJSON strip them from
	// POST/PATCH request bodies (the API rejects writes with 400 Param not
	// allowed).
	IncomingAuthUsername string `json:"incoming_auth_username,omitempty" api:"readonly"`
	IncomingAuthPassword string `json:"incoming_auth_password,omitempty" api:"readonly"`
}

func (c *SIPConfiguration) ConfigurationType() string { return "sip_configurations" }

// String implements fmt.Stringer so default fmt.Sprintf / fmt.Println output
// redacts SIP credential fields. The wire payload is unaffected — MarshalJSON
// above continues to emit the real values (or strip read-only ones via the
// `api:"readonly"` tag).
func (c *SIPConfiguration) String() string {
	mask := func(s string) string {
		if s == "" {
			return ""
		}
		return "[FILTERED]"
	}
	enabled := "<nil>"
	if c.EnabledSipRegistration != nil {
		enabled = fmt.Sprintf("%v", *c.EnabledSipRegistration)
	}
	return fmt.Sprintf("SIPConfiguration{Username:%q Host:%q Port:%d AuthPassword:%q EnabledSipRegistration:%s IncomingAuthUsername:%q IncomingAuthPassword:%q}",
		c.Username, c.Host, c.Port, mask(c.AuthPassword), enabled, mask(c.IncomingAuthUsername), mask(c.IncomingAuthPassword))
}

// GoString mirrors String for the %#v verb.
func (c *SIPConfiguration) GoString() string { return c.String() }

// MarshalJSON serializes SIPConfiguration for outbound POST/PATCH bodies,
// excluding fields tagged `api:"readonly"` (the server-generated
// incoming_auth_* credentials returned in responses).
func (c SIPConfiguration) MarshalJSON() ([]byte, error) { //nolint:gocritic // value receiver required for json.Marshal
	type alias SIPConfiguration
	a := alias(c)
	return jsonapi.MarshalWritableAttrs(a)
}
