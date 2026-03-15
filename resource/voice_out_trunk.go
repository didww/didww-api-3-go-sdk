package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// VoiceOutTrunk represents a voice outbound trunk.
type VoiceOutTrunk struct {
	ID                  string                    `json:"-" jsonapi:"voice_out_trunks"`
	AllowedSipIPs       []string                  `json:"allowed_sip_ips,omitempty"`
	AllowedRtpIPs       []string                  `json:"allowed_rtp_ips,omitempty"`
	AllowAnyDidAsCli    bool                      `json:"allow_any_did_as_cli,omitempty"`
	Status              enums.VoiceOutTrunkStatus `json:"status" api:"readonly"`
	OnCliMismatchAction enums.OnCliMismatchAction `json:"on_cli_mismatch_action,omitempty"`
	Name                string                    `json:"name,omitempty"`
	CapacityLimit       *int                      `json:"capacity_limit,omitempty"`
	Username            string                    `json:"username" api:"readonly"`
	Password            string                    `json:"password" api:"readonly"`
	CreatedAt           time.Time                 `json:"created_at" api:"readonly"`
	ThresholdReached    bool                      `json:"threshold_reached" api:"readonly"`
	ThresholdAmount     *string                   `json:"threshold_amount,omitempty"`
	MediaEncryptionMode enums.MediaEncryptionMode `json:"media_encryption_mode,omitempty"`
	DefaultDstAction    enums.DefaultDstAction    `json:"default_dst_action,omitempty"`
	DstPrefixes         []string                  `json:"dst_prefixes,omitempty"`
	ForceSymmetricRtp   bool                      `json:"force_symmetric_rtp,omitempty"`
	RtpPing             bool                      `json:"rtp_ping,omitempty"`
	CallbackURL         *string                   `json:"callback_url,omitempty"`
	// Relationship IDs for create/update
	DefaultDIDID string   `json:"-" rel:"default_did,dids"`
	DIDIDs       []string `json:"-" rel:"dids,dids"`
	// Resolved relationships
	DefaultDID *DID   `json:"-" rel:"default_did"`
	DIDs       []*DID `json:"-" rel:"dids"`
}

// VoiceOutTrunkRegenerateCredential represents a credential regeneration for voice out trunks.
type VoiceOutTrunkRegenerateCredential struct {
	ID string `json:"-" jsonapi:"voice_out_trunk_regenerate_credentials"`
}
