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
	// Status is the current lifecycle status ("active", "canceled", "new", "changes_required", "in_process", "pending_update").
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
	DIDGroupType          *DIDGroupType          `json:"-" rel:"did_group_type"`
	Order                 *Order                 `json:"-" rel:"order"`
	Address               *Address               `json:"-" rel:"address"`
	EmergencyRequirement  *EmergencyRequirement  `json:"-" rel:"emergency_requirement"`
	EmergencyVerification *EmergencyVerification `json:"-" rel:"emergency_verification"`
	DIDs                  []*DID                 `json:"-" rel:"dids"`
}

// EmergencyCallingService status constants.
const (
	ECSStatusActive          = "active"
	ECSStatusCanceled        = "canceled"
	ECSStatusChangesRequired = "changes_required"
	ECSStatusInProcess       = "in_process"
	ECSStatusNew             = "new"
	ECSStatusPendingUpdate   = "pending_update"
)

// IsActive returns true when the service status is "active".
func (e *EmergencyCallingService) IsActive() bool { return e.Status == ECSStatusActive }

// IsCanceled returns true when the service status is "canceled".
func (e *EmergencyCallingService) IsCanceled() bool { return e.Status == ECSStatusCanceled }

// IsChangesRequired returns true when the service status is "changes_required".
func (e *EmergencyCallingService) IsChangesRequired() bool {
	return e.Status == ECSStatusChangesRequired
}

// IsInProcess returns true when the service status is "in_process".
func (e *EmergencyCallingService) IsInProcess() bool { return e.Status == ECSStatusInProcess }

// IsNew returns true when the service status is "new".
func (e *EmergencyCallingService) IsNew() bool { return e.Status == ECSStatusNew }

// IsPendingUpdate returns true when the service status is "pending_update".
func (e *EmergencyCallingService) IsPendingUpdate() bool { return e.Status == ECSStatusPendingUpdate }
