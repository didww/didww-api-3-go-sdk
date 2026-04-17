package resource

// EmergencyRequirement represents the regulatory requirements for ordering
// an emergency calling service. Introduced in API 2026-04-16.
type EmergencyRequirement struct {
	ID                      string   `json:"-" jsonapi:"emergency_requirements"`
	IdentityType            string   `json:"identity_type" api:"readonly"`
	AddressAreaLevel        string   `json:"address_area_level" api:"readonly"`
	PersonalAreaLevel       string   `json:"personal_area_level" api:"readonly"`
	BusinessAreaLevel       string   `json:"business_area_level" api:"readonly"`
	AddressMandatoryFields  []string `json:"address_mandatory_fields" api:"readonly"`
	PersonalMandatoryFields []string `json:"personal_mandatory_fields" api:"readonly"`
	BusinessMandatoryFields []string `json:"business_mandatory_fields" api:"readonly"`
	// EstimateSetupTime is the estimated time before emergency calling is enabled (e.g. "7-14 days").
	EstimateSetupTime string `json:"estimate_setup_time" api:"readonly"`
	// RequirementRestrictionMessage is a human-readable restriction message. May be empty.
	RequirementRestrictionMessage string `json:"requirement_restriction_message" api:"readonly"`
	// Resolved relationships
	Country      *Country      `json:"-" rel:"country"`
	DIDGroupType *DIDGroupType `json:"-" rel:"did_group_type"`
}
