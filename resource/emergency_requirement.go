package resource

import "github.com/didww/didww-api-3-go-sdk/v4/resource/enums"

// EmergencyRequirement represents the regulatory requirements for ordering
// an emergency calling service. Introduced in API 2026-04-16.
type EmergencyRequirement struct {
	ID string `json:"-" jsonapi:"emergency_requirements"`
	// IdentityType is the required identity type ("personal", "business", or "any").
	IdentityType string `json:"identity_type" api:"readonly"`
	// AddressAreaLevel is the minimum geographic precision for the address ("country", "city", "street", etc.).
	AddressAreaLevel string `json:"address_area_level" api:"readonly"`
	// PersonalAreaLevel is the minimum geographic precision for personal identity addresses.
	// Empty when the country does not accept a personal identity for emergency calling.
	PersonalAreaLevel string `json:"personal_area_level" api:"readonly"`
	// BusinessAreaLevel is the minimum geographic precision for business identity addresses.
	// Empty when the country does not accept a business identity for emergency calling.
	BusinessAreaLevel string `json:"business_area_level" api:"readonly"`
	// AddressMandatoryFields lists the address fields that must be filled in
	// ("area"). Empty when none are.
	AddressMandatoryFields []string `json:"address_mandatory_fields" api:"readonly"`
	// PersonalMandatoryFields lists the fields that must be filled in for a
	// personal identity, among "birth_date", "country", "id_number",
	// "personal_tax_id" and "contact_email". Empty when none are.
	PersonalMandatoryFields []string `json:"personal_mandatory_fields" api:"readonly"`
	// BusinessMandatoryFields lists the fields that must be filled in for a
	// business identity, among "id_number", "tax_id" (Identity.VatID),
	// "country", "company_reg_number", "personal_tax_id", "contact_email" and
	// "birth_date". Empty when none are.
	BusinessMandatoryFields []string `json:"business_mandatory_fields" api:"readonly"`
	// EstimateSetupTime is the estimated time before emergency calling is enabled (e.g. "7-14 days").
	EstimateSetupTime string `json:"estimate_setup_time" api:"readonly"`
	// RequirementRestrictionMessage is a human-readable restriction message. May be empty.
	RequirementRestrictionMessage string `json:"requirement_restriction_message" api:"readonly"`
	// NumberType is the kind of number the requirement applies to.
	NumberType enums.NumberType `json:"number_type" api:"readonly"`
	// ResourceMeta.Meta holds setup_price and monthly_price, decimal strings.
	ResourceMeta
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
}
