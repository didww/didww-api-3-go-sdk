package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/v4/jsonapi"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
)

// SenderIDVerification represents the verification of a sender identity used
// for outbound messages. Supported operations: index, show, update. Only Name
// and the sms_trunk relationship can be changed, and a closed verification
// cannot be updated.
//
// ResourceMeta.Meta holds "setup_price" and "monthly_price", decimal strings.
type SenderIDVerification struct {
	ID         string                           `json:"-" jsonapi:"sender_id_verifications"`
	Name       string                           `json:"name"`
	Status     enums.SenderIDVerificationStatus `json:"status" api:"readonly"`
	SenderType enums.SenderType                 `json:"sender_type" api:"readonly"`
	// SenderIDs are the sender IDs the verification covers; for long_code and
	// toll_free, the numbers of the DIDs.
	SenderIDs []string `json:"sender_ids" api:"readonly"`
	Reference string   `json:"reference" api:"readonly"`
	// RejectReasons and RejectComment are set only when changes are required.
	RejectReasons []string `json:"reject_reasons" api:"readonly"`
	RejectComment *string  `json:"reject_comment" api:"readonly"`
	// RenewDate is the next renewal date (date-only, e.g. "2026-11-22"); nil once closed.
	RenewDate   *string    `json:"renew_date" api:"readonly"`
	CreatedAt   time.Time  `json:"created_at" api:"readonly"`
	ActivatedAt *time.Time `json:"activated_at" api:"readonly"`
	ClosedAt    *time.Time `json:"closed_at" api:"readonly"`
	ResourceMeta
	// Relationship IDs for update. Only http_out, smpp_esme and smpp_smsc trunks are accepted.
	SmsTrunkID string `json:"-" rel:"sms_trunk,sms_trunks"`
	// NullifySmsTrunk, when true, sends {"data": null} for the sms_trunk
	// relationship (unassign the trunk from this verification).
	NullifySmsTrunk bool `json:"-"`
	// Resolved relationships
	SmsTrunk  *SmsTrunk  `json:"-" rel:"sms_trunk"`
	Countries []*Country `json:"-" rel:"countries"`
	DIDs      []*DID     `json:"-" rel:"dids"`
	Order     *Order     `json:"-" rel:"order"`
	Identity  *Identity  `json:"-" rel:"identity"`
	Address   *Address   `json:"-" rel:"address"`
}

// MarshalRelationships implements RelationshipMarshaler for SenderIDVerification.
func (s *SenderIDVerification) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if s.NullifySmsTrunk {
		rels["sms_trunk"] = jsonapi.NullRelationship()
	}
	return rels, nil
}

// IsActive returns true when the verification status is "active".
func (s *SenderIDVerification) IsActive() bool {
	return s.Status == enums.SenderIDVerificationStatusActive
}

// IsChangesRequired returns true when the verification status is "changes_required".
func (s *SenderIDVerification) IsChangesRequired() bool {
	return s.Status == enums.SenderIDVerificationStatusChangesRequired
}

// IsClosed returns true when the verification status is "closed".
func (s *SenderIDVerification) IsClosed() bool {
	return s.Status == enums.SenderIDVerificationStatusClosed
}

// IsClosingInProcess returns true when the verification status is "closing_in_process".
func (s *SenderIDVerification) IsClosingInProcess() bool {
	return s.Status == enums.SenderIDVerificationStatusClosingInProcess
}

// IsInProcess returns true when the verification status is "in_process".
func (s *SenderIDVerification) IsInProcess() bool {
	return s.Status == enums.SenderIDVerificationStatusInProcess
}

// IsNew returns true when the verification status is "new".
func (s *SenderIDVerification) IsNew() bool {
	return s.Status == enums.SenderIDVerificationStatusNew
}
