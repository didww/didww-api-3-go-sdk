package trunkconfiguration

import "github.com/didww/didww-api-3-go-sdk/resource/enums"

// SIPConfiguration represents a SIP trunk configuration.
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
}

func (c *SIPConfiguration) ConfigurationType() string { return "sip_configurations" }
