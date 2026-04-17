package resource

import (
	"encoding/json"
	"time"

	"github.com/didww/didww-api-3-go-sdk/jsonapi"
	"github.com/didww/didww-api-3-go-sdk/resource/authenticationmethod"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// VoiceOutTrunk represents a voice outbound trunk.
type VoiceOutTrunk struct {
	ID                   string                                    `json:"-" jsonapi:"voice_out_trunks"`
	AllowedRtpIPs        []string                                  `json:"allowed_rtp_ips,omitempty"`
	AllowAnyDidAsCli     bool                                      `json:"allow_any_did_as_cli,omitempty"`
	Status               enums.VoiceOutTrunkStatus                 `json:"status" api:"readonly"`
	OnCliMismatchAction  enums.OnCliMismatchAction                 `json:"on_cli_mismatch_action,omitempty"`
	Name                 string                                    `json:"name,omitempty"`
	CapacityLimit        *int                                      `json:"capacity_limit,omitempty"`
	CreatedAt            time.Time                                 `json:"created_at" api:"readonly"`
	ThresholdReached     bool                                      `json:"threshold_reached" api:"readonly"`
	ThresholdAmount      *string                                   `json:"threshold_amount,omitempty"`
	MediaEncryptionMode  enums.MediaEncryptionMode                 `json:"media_encryption_mode,omitempty"`
	DefaultDstAction     enums.DefaultDstAction                    `json:"default_dst_action,omitempty"`
	DstPrefixes          []string                                  `json:"dst_prefixes,omitempty"`
	ForceSymmetricRtp    bool                                      `json:"force_symmetric_rtp,omitempty"`
	RtpPing              bool                                      `json:"rtp_ping,omitempty"`
	CallbackURL          *string                                   `json:"callback_url,omitempty"`
	ExternalReferenceID  *string                                   `json:"external_reference_id,omitempty"`
	EmergencyEnableAll   bool                                      `json:"emergency_enable_all,omitempty"`
	RtpTimeout           *int                                      `json:"rtp_timeout,omitempty"`
	AuthenticationMethod authenticationmethod.AuthenticationMethod `json:"-"`
	// Relationship IDs for create/update
	DefaultDIDID    string   `json:"-" rel:"default_did,dids"`
	DIDIDs          []string `json:"-" rel:"dids,dids"`
	EmergencyDIDIDs []string `json:"-" rel:"emergency_dids,dids"`
	// ClearEmergencyDIDs, when true, sends {"data": []} for the
	// emergency_dids relationship (remove all emergency DIDs from this trunk).
	ClearEmergencyDIDs bool `json:"-"`
	// Resolved relationships
	DefaultDID    *DID   `json:"-" rel:"default_did"`
	DIDs          []*DID `json:"-" rel:"dids"`
	EmergencyDIDs []*DID `json:"-" rel:"emergency_dids"`
}

// MarshalJSON handles custom serialization for VoiceOutTrunk.
func (v *VoiceOutTrunk) MarshalJSON() ([]byte, error) {
	type Alias VoiceOutTrunk
	aux := struct {
		Alias
		AuthenticationMethod json.RawMessage `json:"authentication_method,omitempty"`
	}{
		Alias: Alias(*v),
	}
	if v.AuthenticationMethod != nil {
		am, err := authenticationmethod.MarshalJSON(v.AuthenticationMethod)
		if err != nil {
			return nil, err
		}
		aux.AuthenticationMethod = am
	}
	return json.Marshal(aux)
}

// UnmarshalJSON handles custom deserialization for VoiceOutTrunk.
func (v *VoiceOutTrunk) UnmarshalJSON(data []byte) error {
	type Alias VoiceOutTrunk
	aux := &struct {
		*Alias
		AuthenticationMethod json.RawMessage `json:"authentication_method"`
	}{
		Alias: (*Alias)(v),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.AuthenticationMethod) > 0 && string(aux.AuthenticationMethod) != "null" {
		am, err := authenticationmethod.UnmarshalJSON(aux.AuthenticationMethod)
		if err != nil {
			return err
		}
		v.AuthenticationMethod = am
	}
	return nil
}

// MarshalRelationships implements RelationshipMarshaler for VoiceOutTrunk.
// When ClearEmergencyDIDs is true, emits {"data": []} for emergency_dids.
func (v *VoiceOutTrunk) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if v.ClearEmergencyDIDs {
		rels["emergency_dids"] = jsonapi.ToManyRelationship([]jsonapi.RelationshipRef{})
	}
	return rels, nil
}

// IsActive returns true when the trunk status is "active".
func (v *VoiceOutTrunk) IsActive() bool { return v.Status == enums.VoiceOutTrunkStatusActive }

// IsBlocked returns true when the trunk status is "blocked".
func (v *VoiceOutTrunk) IsBlocked() bool { return v.Status == enums.VoiceOutTrunkStatusBlocked }

// VoiceOutTrunkRegenerateCredential represents a credential regeneration for voice out trunks.
type VoiceOutTrunkRegenerateCredential struct {
	ID string `json:"-" jsonapi:"voice_out_trunk_regenerate_credentials"`
}
