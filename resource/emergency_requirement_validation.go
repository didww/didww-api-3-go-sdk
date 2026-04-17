package resource

// EmergencyRequirementValidation validates a prospective emergency calling service
// order against an EmergencyRequirement. A successful POST returns 201 Created
// with the validation resource (id mirrors the submitted emergency_requirement_id).
// Introduced in API 2026-04-16.
type EmergencyRequirementValidation struct {
	ID string `json:"-" jsonapi:"emergency_requirement_validations"`
	// Relationship IDs for create
	EmergencyRequirementID string `json:"-" rel:"emergency_requirement,emergency_requirements"`
	AddressID              string `json:"-" rel:"address,addresses"`
	IdentityID             string `json:"-" rel:"identity,identities"`
}
