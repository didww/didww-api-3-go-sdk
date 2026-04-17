package resource

import "time"

// EmergencyCallingService represents a customer-owned subscription to emergency calling.
// Supported operations: index, show, destroy. Introduced in API 2026-04-16.
type EmergencyCallingService struct {
	ID          string     `json:"-" jsonapi:"emergency_calling_services"`
	Name        string     `json:"name" api:"readonly"`
	Reference   string     `json:"reference" api:"readonly"`
	Status      string     `json:"status" api:"readonly"`
	ActivatedAt *time.Time `json:"activated_at" api:"readonly"`
	CanceledAt  *time.Time `json:"canceled_at" api:"readonly"`
	CreatedAt   time.Time  `json:"created_at" api:"readonly"`
	RenewDate   string     `json:"renew_date" api:"readonly"`
	// Resolved relationships
	Country               *Country               `json:"-" rel:"country"`
	DIDGroupType          *DIDGroupType           `json:"-" rel:"did_group_type"`
	Order                 *Order                  `json:"-" rel:"order"`
	Address               *Address                `json:"-" rel:"address"`
	EmergencyRequirement  *EmergencyRequirement   `json:"-" rel:"emergency_requirement"`
	EmergencyVerification *EmergencyVerification  `json:"-" rel:"emergency_verification"`
	DIDs                  []*DID                  `json:"-" rel:"dids"`
}
