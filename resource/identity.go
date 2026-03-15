package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// Identity represents a customer identity.
type Identity struct {
	ID                  string             `json:"-" jsonapi:"identities"`
	FirstName           string             `json:"first_name"`
	LastName            string             `json:"last_name"`
	PhoneNumber         string             `json:"phone_number"`
	IDNumber            *string            `json:"id_number"`
	BirthDate           string             `json:"birth_date"`
	CompanyName         *string            `json:"company_name"`
	CompanyRegNumber    *string            `json:"company_reg_number"`
	VatID               *string            `json:"vat_id"`
	Description         *string            `json:"description"`
	PersonalTaxID       *string            `json:"personal_tax_id"`
	IdentityType        enums.IdentityType `json:"identity_type"`
	CreatedAt           time.Time          `json:"created_at" api:"readonly"`
	ExternalReferenceID *string            `json:"external_reference_id"`
	Verified            bool               `json:"verified" api:"readonly"`
	ContactEmail        *string            `json:"contact_email"`
	// Relationship IDs for create/update
	CountryID string `json:"-" rel:"country,countries"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
}
