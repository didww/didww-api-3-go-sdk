package resource

// AddressRequirement represents a regulatory address requirement.
type AddressRequirement struct {
	ID                         string   `json:"-" jsonapi:"address_requirements"`
	IdentityType               string   `json:"identity_type"`
	PersonalAreaLevel          string   `json:"personal_area_level"`
	BusinessAreaLevel          string   `json:"business_area_level"`
	AddressAreaLevel           string   `json:"address_area_level"`
	PersonalProofQty           int      `json:"personal_proof_qty"`
	BusinessProofQty           int      `json:"business_proof_qty"`
	AddressProofQty            int      `json:"address_proof_qty"`
	PersonalMandatoryFields    []string `json:"personal_mandatory_fields"`
	BusinessMandatoryFields    []string `json:"business_mandatory_fields"`
	ServiceDescriptionRequired bool     `json:"service_description_required"`
	RestrictionMessage         string   `json:"restriction_message"`
	// Resolved relationships
	Country                   *Country                    `json:"-" rel:"country"`
	DIDGroupType              *DIDGroupType               `json:"-" rel:"did_group_type"`
	PersonalPermanentDocument *SupportingDocumentTemplate `json:"-" rel:"personal_permanent_document"`
	BusinessPermanentDocument *SupportingDocumentTemplate `json:"-" rel:"business_permanent_document"`
	PersonalOnetimeDocument   *SupportingDocumentTemplate `json:"-" rel:"personal_onetime_document"`
	BusinessOnetimeDocument   *SupportingDocumentTemplate `json:"-" rel:"business_onetime_document"`
	PersonalProofTypes        []*ProofType                `json:"-" rel:"personal_proof_types"`
	BusinessProofTypes        []*ProofType                `json:"-" rel:"business_proof_types"`
	AddressProofTypes         []*ProofType                `json:"-" rel:"address_proof_types"`
}

// AddressRequirementValidation represents an address requirement validation result.
type AddressRequirementValidation struct {
	ID string `json:"-" jsonapi:"address_requirement_validations"`
	// Relationship IDs for create
	AddressID     string `json:"-" rel:"address,addresses"`
	IdentityID    string `json:"-" rel:"identity,identities"`
	RequirementID string `json:"-" rel:"requirement,address_requirements"`
}
