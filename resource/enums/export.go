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
	ExportStatusPending    ExportStatus = "Pending"
	ExportStatusProcessing ExportStatus = "Processing"
	ExportStatusCompleted  ExportStatus = "Completed"
)

// CallbackMethod defines the HTTP method used for webhook callbacks.
type CallbackMethod string

const (
	CallbackMethodPOST CallbackMethod = "POST"
	CallbackMethodGET  CallbackMethod = "GET"
)
