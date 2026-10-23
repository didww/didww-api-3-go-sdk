package resource

import (
	"encoding/json"
	"time"

	"github.com/didww/didww-api-3-go-sdk/v3/jsonapi"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/smstrunkconfiguration"
)

// SmsTrunk represents an SMS trunk: the delivery destination of inbound
// messages, or the way outbound messages are submitted, depending on its
// configuration type.
//
// ResourceMeta.Meta holds "host", the URL to submit outbound messages to
// (http_out trunks only).
type SmsTrunk struct {
	ID                  string  `json:"-" jsonapi:"sms_trunks"`
	Name                string  `json:"name,omitempty"`
	ExternalReferenceID *string `json:"external_reference_id,omitempty"`
	Blocked             bool    `json:"blocked,omitempty"`
	// Priority is the routing priority within an SMS trunk group, 0-32767; the
	// higher value is tried first. Defaults to 100 and must stay nil on http_out trunks.
	Priority *int `json:"priority,omitempty"`
	// AllowAnyDIDAsSource allows any DID of the account as the message source.
	// Turning it on empties outgoing_dids. Refused on the trunk types that only
	// receive messages (sms_to_email, http_in).
	AllowAnyDIDAsSource bool                                        `json:"allow_any_did_as_source,omitempty"`
	AllowIn             bool                                        `json:"allow_in" api:"readonly"`
	AllowOut            bool                                        `json:"allow_out" api:"readonly"`
	CreatedAt           time.Time                                   `json:"created_at" api:"readonly"`
	UpdatedAt           time.Time                                   `json:"updated_at" api:"readonly"`
	Configuration       smstrunkconfiguration.SmsTrunkConfiguration `json:"-"`
	ResourceMeta
	// Relationship IDs for create/update
	//
	// OutgoingDIDIDs replaces the whole outgoing_dids membership: list every
	// DID to keep. A fetched trunk does not fill it; build it from OutgoingDIDs
	// (include outgoing_dids).
	OutgoingDIDIDs  []string `json:"-" rel:"outgoing_dids,dids"`
	SmsTrunkGroupID string   `json:"-" rel:"sms_trunk_group,sms_trunk_groups"`
	// ClearOutgoingDIDs, when true, sends {"data": []} for the outgoing_dids relationship.
	ClearOutgoingDIDs bool `json:"-"`
	// NullifySmsTrunkGroup, when true, sends {"data": null} for the
	// sms_trunk_group relationship (remove the trunk from its group).
	NullifySmsTrunkGroup bool `json:"-"`
	// Resolved relationships
	OutgoingDIDs          []*DID                  `json:"-" rel:"outgoing_dids"`
	IncomingDIDs          []*DID                  `json:"-" rel:"incoming_dids"`
	SmsTrunkGroup         *SmsTrunkGroup          `json:"-" rel:"sms_trunk_group"`
	SenderIDVerifications []*SenderIDVerification `json:"-" rel:"sender_id_verifications"`
}

// UnmarshalJSON implements custom unmarshaling for SmsTrunk.
func (s *SmsTrunk) UnmarshalJSON(data []byte) error {
	type Alias SmsTrunk
	aux := &struct {
		*Alias
		RawConfig json.RawMessage `json:"configuration"`
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.RawConfig) > 0 {
		config, err := smstrunkconfiguration.Parse(aux.RawConfig)
		if err != nil {
			return err
		}
		s.Configuration = config
	}
	return nil
}

// MarshalJSON implements custom marshaling for SmsTrunk.
func (s SmsTrunk) MarshalJSON() ([]byte, error) { //nolint:gocritic // value receiver required for json.Marshal
	type Alias SmsTrunk
	aux := &struct {
		Alias
		RawConfig json.RawMessage `json:"configuration,omitempty"`
	}{
		Alias: Alias(s),
	}
	if s.Configuration != nil {
		raw, err := smstrunkconfiguration.Marshal(s.Configuration)
		if err != nil {
			return nil, err
		}
		aux.RawConfig = raw
	}
	return json.Marshal(aux)
}

// MarshalRelationships implements RelationshipMarshaler for SmsTrunk.
func (s *SmsTrunk) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if s.ClearOutgoingDIDs {
		rels["outgoing_dids"] = jsonapi.ToManyRelationship([]jsonapi.RelationshipRef{})
	}
	if s.NullifySmsTrunkGroup {
		rels["sms_trunk_group"] = jsonapi.NullRelationship()
	}
	return rels, nil
}

// SmsTrunkGroup represents a group of SMS trunks. Deleting a group also
// deletes the trunks in it.
//
// ResourceMeta.Meta holds "sms_trunks_count".
type SmsTrunkGroup struct {
	ID                  string    `json:"-" jsonapi:"sms_trunk_groups"`
	Name                string    `json:"name,omitempty"`
	ExternalReferenceID *string   `json:"external_reference_id,omitempty"`
	CreatedAt           time.Time `json:"created_at" api:"readonly"`
	UpdatedAt           time.Time `json:"updated_at" api:"readonly"`
	ResourceMeta
	// Relationship IDs for create/update
	//
	// SmsTrunkIDs replaces the whole membership of the group: list every trunk
	// to keep. A fetched group does not fill it; build it from SmsTrunks
	// (include sms_trunks).
	SmsTrunkIDs []string `json:"-" rel:"sms_trunks,sms_trunks"`
	// ClearSmsTrunks, when true, sends {"data": []} for the sms_trunks
	// relationship, which empties the group.
	ClearSmsTrunks bool `json:"-"`
	// Resolved relationships
	SmsTrunks    []*SmsTrunk `json:"-" rel:"sms_trunks"`
	IncomingDIDs []*DID      `json:"-" rel:"incoming_dids"`
}

// MarshalRelationships implements RelationshipMarshaler for SmsTrunkGroup.
func (g *SmsTrunkGroup) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if g.ClearSmsTrunks {
		rels["sms_trunks"] = jsonapi.ToManyRelationship([]jsonapi.RelationshipRef{})
	}
	return rels, nil
}

// SmsTrunkRegenerateCredential regenerates the generated credentials of an
// smpp_esme or http_out SMS trunk. Supported operation: create.
type SmsTrunkRegenerateCredential struct {
	ID string `json:"-" jsonapi:"sms_trunk_regenerate_credentials"`
	// Relationship IDs for create
	SmsTrunkID string `json:"-" rel:"sms_trunk,sms_trunks"`
}
