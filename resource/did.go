package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/jsonapi"
)

// DID represents a DID (phone number) resource.
type DID struct {
	ID                     string     `json:"-" jsonapi:"dids"`
	Blocked                bool       `json:"blocked" api:"readonly"`
	CapacityLimit          *int       `json:"capacity_limit"`
	Description            *string    `json:"description"`
	Terminated             bool       `json:"terminated"`
	AwaitingRegistration   bool       `json:"awaiting_registration" api:"readonly"`
	CreatedAt              time.Time  `json:"created_at" api:"readonly"`
	BillingCyclesCount     *int       `json:"billing_cycles_count" api:"readonly"`
	Number                 string     `json:"number" api:"readonly"`
	ExpiresAt              *time.Time `json:"expires_at" api:"readonly"`
	ChannelsIncludedCount  int        `json:"channels_included_count" api:"readonly"`
	DedicatedChannelsCount int        `json:"dedicated_channels_count"`
	EmergencyEnabled       bool       `json:"emergency_enabled" api:"readonly"`
	// Relationship IDs for create/update
	VoiceInTrunkID        string `json:"-" rel:"voice_in_trunk,voice_in_trunks"`
	VoiceInTrunkGroupID   string `json:"-" rel:"voice_in_trunk_group,voice_in_trunk_groups"`
	CapacityPoolID            string `json:"-" rel:"capacity_pool,capacity_pools"`
	SharedCapacityGroupID     string `json:"-" rel:"shared_capacity_group,shared_capacity_groups"`
	EmergencyCallingServiceID string `json:"-" rel:"emergency_calling_service,emergency_calling_services"`
	EmergencyVerificationID   string `json:"-" rel:"emergency_verification,emergency_verifications"`
	// Resolved relationships
	Order               *Order               `json:"-" rel:"order"`
	AddressVerification *AddressVerification `json:"-" rel:"address_verification"`
	DIDGroup            *DIDGroup            `json:"-" rel:"did_group"`
	VoiceInTrunk        *VoiceInTrunk        `json:"-" rel:"voice_in_trunk"`
	VoiceInTrunkGroup   *VoiceInTrunkGroup   `json:"-" rel:"voice_in_trunk_group"`
	CapacityPool            *CapacityPool            `json:"-" rel:"capacity_pool"`
	SharedCapacityGroup     *SharedCapacityGroup     `json:"-" rel:"shared_capacity_group"`
	EmergencyCallingService *EmergencyCallingService `json:"-" rel:"emergency_calling_service"`
	EmergencyVerification   *EmergencyVerification   `json:"-" rel:"emergency_verification"`
}

// MarshalRelationships implements RelationshipMarshaler for DID.
// Ensures mutual exclusivity: setting a trunk nullifies the trunk group and vice versa.
func (d *DID) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if d.VoiceInTrunkID != "" {
		rels["voice_in_trunk_group"] = jsonapi.NullRelationship()
	}
	if d.VoiceInTrunkGroupID != "" {
		rels["voice_in_trunk"] = jsonapi.NullRelationship()
	}
	if d.CapacityPoolID != "" {
		rels["shared_capacity_group"] = jsonapi.NullRelationship()
	}
	if d.SharedCapacityGroupID != "" {
		rels["capacity_pool"] = jsonapi.NullRelationship()
	}
	return rels, nil
}
