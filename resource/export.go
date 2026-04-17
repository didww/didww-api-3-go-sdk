package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// Export represents a CDR export.
type Export struct {
	ID             string                 `json:"-" jsonapi:"exports"`
	Status         enums.ExportStatus     `json:"status" api:"readonly"`
	CreatedAt      time.Time              `json:"created_at" api:"readonly"`
	URL            *string                `json:"url" api:"readonly"`
	CallbackURL    *string                `json:"callback_url,omitempty"`
	CallbackMethod *string                `json:"callback_method,omitempty"`
	ExportType     enums.ExportType       `json:"export_type"`
	Filters             map[string]interface{} `json:"filters,omitempty"`
	ExternalReferenceID *string                `json:"external_reference_id,omitempty"`
}
