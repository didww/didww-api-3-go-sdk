package resource

import "time"

// EmergencyCallingService represents a customer-owned subscription to emergency calling.
// Supported operations: index, show, destroy. Introduced in API 2026-04-16.
type EmergencyCallingService struct {
	ID string `json:"-" jsonapi:"emergency_calling_services"`
	// Name is the human-readable label for the service.
	Name string `json:"name" api:"readonly"`
	// Reference is the server-assigned reference code (e.g. "ECS-0042").
	Reference string `json:"reference" api:"readonly"`
	// Status is the current lifecycle status ("pending", "active", "canceled", etc.).
	Status string `json:"status" api:"readonly"`
	// ActivatedAt is when the service became active (nil if not yet activated).
	ActivatedAt *time.Time `json:"activated_at" api:"readonly"`
	// CanceledAt is when the service was canceled (nil if still active).
	CanceledAt *time.Time `json:"canceled_at" api:"readonly"`
	// CreatedAt is when the service was created.
	CreatedAt time.Time `json:"created_at" api:"readonly"`
	// RenewDate is the next renewal date for the service subscription (date-only, e.g. "2026-05-22").
	RenewDate string `json:"renew_date" api:"readonly"`
	// Resolved relationships
	Country               *Country               `json:"-" rel:"country"`
	DIDGroupType          *DIDGroupType           `json:"-" rel:"did_group_type"`
	Order                 *Order                  `json:"-" rel:"order"`
	Address               *Address                `json:"-" rel:"address"`
	EmergencyRequirement  *EmergencyRequirement   `json:"-" rel:"emergency_requirement"`
	EmergencyVerification *EmergencyVerification  `json:"-" rel:"emergency_verification"`
	DIDs                  []*DID                  `json:"-" rel:"dids"`
}
