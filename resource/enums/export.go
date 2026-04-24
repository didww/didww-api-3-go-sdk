package enums

// ExportType defines the type of CDR export.
type ExportType string

const (
	ExportTypeCdrIn  ExportType = "cdr_in"
	ExportTypeCdrOut ExportType = "cdr_out"
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
