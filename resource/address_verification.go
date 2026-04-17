package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// AddressVerification represents an address verification request.
type AddressVerification struct {
	ID                 string                          `json:"-" jsonapi:"address_verifications"`
	ServiceDescription *string                         `json:"service_description,omitempty"`
	CallbackURL        *string                         `json:"callback_url,omitempty"`
	CallbackMethod     *string                         `json:"callback_method,omitempty"`
	Status             enums.AddressVerificationStatus `json:"status" api:"readonly"`
	RejectReasons      []string                        `json:"reject_reasons" api:"readonly"`
	CreatedAt          time.Time                       `json:"created_at" api:"readonly"`
	Reference           string                          `json:"reference" api:"readonly"`
	RejectComment       string                          `json:"reject_comment" api:"readonly"`
	ExternalReferenceID *string                         `json:"external_reference_id,omitempty"`
	// Relationship IDs for create/update
	AddressID string   `json:"-" rel:"address,addresses"`
	DIDIDs    []string `json:"-" rel:"dids,dids"`
	// Resolved relationships
	AddressRel *Address `json:"-" rel:"address"`
}

// IsPending returns true when the verification status is "pending".
func (a *AddressVerification) IsPending() bool {
	return a.Status == enums.AddressVerificationStatusPending
}

// IsApproved returns true when the verification status is "approved".
func (a *AddressVerification) IsApproved() bool {
	return a.Status == enums.AddressVerificationStatusApproved
}

// IsRejected returns true when the verification status is "rejected".
func (a *AddressVerification) IsRejected() bool {
	return a.Status == enums.AddressVerificationStatusRejected
}
