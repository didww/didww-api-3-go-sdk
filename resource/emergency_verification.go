package resource

import "time"

// EmergencyVerification represents a verification record for an emergency calling service.
// Supported operations: index, show, create. Introduced in API 2026-04-16.
type EmergencyVerification struct {
	ID                 string    `json:"-" jsonapi:"emergency_verifications"`
	Reference          string    `json:"reference" api:"readonly"`
	Status             string    `json:"status" api:"readonly"`
	RejectReasons      []string  `json:"reject_reasons" api:"readonly"`
	RejectComment      string    `json:"reject_comment" api:"readonly"`
	CallbackURL        *string   `json:"callback_url,omitempty"`
	CallbackMethod     *string   `json:"callback_method,omitempty"`
	ExternalReferenceID *string  `json:"external_reference_id,omitempty"`
	CreatedAt          time.Time `json:"created_at" api:"readonly"`
	// Relationship IDs for create/update
	AddressID                 string   `json:"-" rel:"address,addresses"`
	EmergencyCallingServiceID string   `json:"-" rel:"emergency_calling_service,emergency_calling_services"`
	DIDIDs                    []string `json:"-" rel:"dids,dids"`
	// Resolved relationships
	AddressRel              *Address                 `json:"-" rel:"address"`
	EmergencyCallingService *EmergencyCallingService  `json:"-" rel:"emergency_calling_service"`
	DIDs                    []*DID                   `json:"-" rel:"dids"`
}

// Emergency verification status constants (lowercase, per API).
const (
	EmergencyVerificationStatusPending  = "pending"
	EmergencyVerificationStatusApproved = "approved"
	EmergencyVerificationStatusRejected = "rejected"
)

// IsPending returns true when the verification status is "pending".
func (e *EmergencyVerification) IsPending() bool {
	return e.Status == EmergencyVerificationStatusPending
}

// IsApproved returns true when the verification status is "approved".
func (e *EmergencyVerification) IsApproved() bool {
	return e.Status == EmergencyVerificationStatusApproved
}

// IsRejected returns true when the verification status is "rejected".
func (e *EmergencyVerification) IsRejected() bool {
	return e.Status == EmergencyVerificationStatusRejected
}
