package enums

// ExportType defines the type of export.
type ExportType string

const (
	ExportTypeCdrIn  ExportType = "cdr_in"
	ExportTypeCdrOut ExportType = "cdr_out"
	ExportTypeSmsIn  ExportType = "sms_in"
	ExportTypeSmsOut ExportType = "sms_out"
)

// Values of the "trunk_type" filter of sms_in exports (Export.Filters).
// The filter takes one value or a slice of them.
const (
	ExportTrunkTypeSmppSmsc = "SMPP SMSC"
	ExportTrunkTypeSmppEsme = "SMPP ESME"
	ExportTrunkTypeEmail    = "Email"
	ExportTrunkTypeHTTPIn   = "HTTP IN"
)

// ExportStatus defines the processing status of an export.
type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "pending"
	ExportStatusProcessing ExportStatus = "processing"
	ExportStatusCompleted  ExportStatus = "completed"
)

// CallbackMethod defines the HTTP method used for webhook callbacks.
type CallbackMethod string

const (
	CallbackMethodPOST CallbackMethod = "post"
	CallbackMethodGET  CallbackMethod = "get"
)
