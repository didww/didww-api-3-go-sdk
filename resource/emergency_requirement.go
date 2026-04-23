package resource

import "encoding/json"

// EmergencyRequirement represents the regulatory requirements for ordering
// an emergency calling service. Introduced in API 2026-04-16.
type EmergencyRequirement struct {
	ID string `json:"-" jsonapi:"emergency_requirements"`
	// IdentityType is the required identity type ("personal", "business", or "any").
	IdentityType string `json:"identity_type" api:"readonly"`
	// AddressAreaLevel is the minimum geographic precision for the address ("country", "city", "street", etc.).
	AddressAreaLevel string `json:"address_area_level" api:"readonly"`
	// PersonalAreaLevel is the minimum geographic precision for personal identity addresses.
	PersonalAreaLevel string `json:"personal_area_level" api:"readonly"`
	// BusinessAreaLevel is the minimum geographic precision for business identity addresses.
	BusinessAreaLevel string `json:"business_area_level" api:"readonly"`
	// AddressMandatoryFields lists address fields required for this requirement.
	AddressMandatoryFields []string `json:"address_mandatory_fields" api:"readonly"`
	// PersonalMandatoryFields lists identity fields required for personal identities.
	PersonalMandatoryFields []string `json:"personal_mandatory_fields" api:"readonly"`
	// BusinessMandatoryFields lists identity fields required for business identities.
	BusinessMandatoryFields []string `json:"business_mandatory_fields" api:"readonly"`
	// EstimateSetupTime is the estimated time before emergency calling is enabled (e.g. "7-14 days").
	EstimateSetupTime string `json:"estimate_setup_time" api:"readonly"`
	// RequirementRestrictionMessage is a human-readable restriction message. May be empty.
	RequirementRestrictionMessage string `json:"requirement_restriction_message" api:"readonly"`
	// Meta holds resource-level JSON:API meta (e.g. setup_price, monthly_price).
	Meta map[string]string `json:"-"`
	// Resolved relationships
	Country      *Country      `json:"-" rel:"country"`
	DIDGroupType *DIDGroupType `json:"-" rel:"did_group_type"`
}

// UnmarshalMeta parses the resource-level JSON:API meta block into a generic map.
func (e *EmergencyRequirement) UnmarshalMeta(raw json.RawMessage) error {
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	e.Meta = m
	return nil
}
