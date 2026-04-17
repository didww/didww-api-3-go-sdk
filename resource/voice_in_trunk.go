package resource

import (
	"encoding/json"
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/resource/trunkconfiguration"
)

// VoiceInTrunk represents a voice inbound trunk.
type VoiceInTrunk struct {
	ID             string                                `json:"-" jsonapi:"voice_in_trunks"`
	Priority       int                                   `json:"priority,omitempty"`
	CapacityLimit  *int                                  `json:"capacity_limit,omitempty"`
	Weight         int                                   `json:"weight,omitempty"`
	Name           string                                `json:"name,omitempty"`
	CliFormat      enums.CliFormat                       `json:"cli_format,omitempty"`
	CliPrefix      *string                               `json:"cli_prefix,omitempty"`
	Description    *string                               `json:"description,omitempty"`
	RingingTimeout *int                                  `json:"ringing_timeout,omitempty"`
	Configuration       trunkconfiguration.TrunkConfiguration `json:"-"`
	CreatedAt           time.Time                             `json:"created_at" api:"readonly"`
	ExternalReferenceID *string                               `json:"external_reference_id,omitempty"`
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
		config, err := trunkconfiguration.Parse(aux.RawConfig)
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
			"type": v.Configuration.ConfigurationType(),
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
	ID                  string    `json:"-" jsonapi:"voice_in_trunk_groups"`
	Name                string    `json:"name,omitempty"`
	CapacityLimit       *int      `json:"capacity_limit,omitempty"`
	CreatedAt           time.Time `json:"created_at" api:"readonly"`
	ExternalReferenceID *string   `json:"external_reference_id,omitempty"`
	// Relationship IDs for create/update
	VoiceInTrunkIDs []string `json:"-" rel:"voice_in_trunks,voice_in_trunks"`
	// Resolved relationships
	VoiceInTrunks []*VoiceInTrunk `json:"-" rel:"voice_in_trunks"`
}
