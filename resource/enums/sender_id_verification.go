package enums

// SenderIDVerificationStatus defines the lifecycle status of a sender ID verification.
type SenderIDVerificationStatus string

const (
	SenderIDVerificationStatusActive           SenderIDVerificationStatus = "active"
	SenderIDVerificationStatusChangesRequired  SenderIDVerificationStatus = "changes_required"
	SenderIDVerificationStatusClosed           SenderIDVerificationStatus = "closed"
	SenderIDVerificationStatusClosingInProcess SenderIDVerificationStatus = "closing_in_process"
	SenderIDVerificationStatusInProcess        SenderIDVerificationStatus = "in_process"
	SenderIDVerificationStatusNew              SenderIDVerificationStatus = "new"
)

// SenderType defines the kind of sender a sender ID verification covers.
type SenderType string

const (
	SenderTypeAlphanumeric SenderType = "alphanumeric"
	SenderTypeCarrier      SenderType = "carrier"
	SenderTypeLongCode     SenderType = "long_code"
	SenderTypeShortCode    SenderType = "short_code"
	SenderTypeTollFree     SenderType = "toll_free"
)
