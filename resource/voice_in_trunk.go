package resource

import (
	"encoding/json"
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// TrunkConfiguration is an interface for voice in trunk configurations.
type TrunkConfiguration interface {
	configurationType() string
}

// PSTNConfiguration represents a PSTN trunk configuration.
type PSTNConfiguration struct {
	Dst string `json:"dst"`
}

func (c *PSTNConfiguration) configurationType() string { return "pstn_configurations" }

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
}

func (c *SIPConfiguration) configurationType() string { return "sip_configurations" }

// parseTrunkConfiguration deserializes a nested configuration object.
func parseTrunkConfiguration(data []byte) (TrunkConfiguration, error) {
	var env struct {
		Type       string          `json:"type"`
		Attributes json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Type {
	case "pstn_configurations":
		var cfg PSTNConfiguration
		if err := json.Unmarshal(env.Attributes, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	case "sip_configurations":
		var cfg SIPConfiguration
		if err := json.Unmarshal(env.Attributes, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	default:
		// Skip unsupported legacy configuration types (iax2, h323, etc.)
		return nil, nil
	}
}

// VoiceInTrunk represents a voice inbound trunk.
type VoiceInTrunk struct {
	ID             string             `json:"-" jsonapi:"voice_in_trunks"`
	Priority       int                `json:"priority,omitempty"`
	CapacityLimit  *int               `json:"capacity_limit,omitempty"`
	Weight         int                `json:"weight,omitempty"`
	Name           string             `json:"name,omitempty"`
	CliFormat      enums.CliFormat    `json:"cli_format,omitempty"`
	CliPrefix      *string            `json:"cli_prefix,omitempty"`
	Description    *string            `json:"description,omitempty"`
	RingingTimeout *int               `json:"ringing_timeout,omitempty"`
	Configuration  TrunkConfiguration `json:"-"`
	CreatedAt      time.Time          `json:"created_at" api:"readonly"`
	// Resolved relationships
	Pop               *Pop               `json:"-" rel:"pop"`
	VoiceInTrunkGroup *VoiceInTrunkGroup `json:"-" rel:"voice_in_trunk_group"`
}

// UnmarshalJSON implements custom unmarshaling for VoiceInTrunk.
func (v *VoiceInTrunk) UnmarshalJSON(data []byte) error {
	type Alias VoiceInTrunk
	aux := &struct {
		*Alias
		RawConfig json.RawMessage `json:"configuration"`
	}{
		Alias: (*Alias)(v),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.RawConfig) > 0 && string(aux.RawConfig) != "null" {
		config, err := parseTrunkConfiguration(aux.RawConfig)
		if err != nil {
			return err
		}
		v.Configuration = config
	}
	return nil
}

// MarshalJSON implements custom marshaling for VoiceInTrunk.
func (v VoiceInTrunk) MarshalJSON() ([]byte, error) { //nolint:gocritic // value receiver required for json.Marshal
	type Alias VoiceInTrunk
	aux := &struct {
		Alias
		RawConfig json.RawMessage `json:"configuration,omitempty"`
	}{
		Alias: Alias(v),
	}
	if v.Configuration != nil {
		configData := map[string]any{
			"type": v.Configuration.configurationType(),
		}
		attrs, err := json.Marshal(v.Configuration)
		if err != nil {
			return nil, err
		}
		configData["attributes"] = json.RawMessage(attrs)
		raw, err := json.Marshal(configData)
		if err != nil {
			return nil, err
		}
		aux.RawConfig = raw
	}
	return json.Marshal(aux)
}

// VoiceInTrunkGroup represents a group of voice inbound trunks.
type VoiceInTrunkGroup struct {
	ID            string    `json:"-" jsonapi:"voice_in_trunk_groups"`
	Name          string    `json:"name,omitempty"`
	CapacityLimit *int      `json:"capacity_limit,omitempty"`
	CreatedAt     time.Time `json:"created_at" api:"readonly"`
	// Relationship IDs for create/update
	VoiceInTrunkIDs []string `json:"-" rel:"voice_in_trunks,voice_in_trunks"`
	// Resolved relationships
	VoiceInTrunks []*VoiceInTrunk `json:"-" rel:"voice_in_trunks"`
}
