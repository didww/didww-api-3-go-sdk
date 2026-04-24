package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
)

// Export represents a CDR export.
//
// Filters map keys for cdr_in / cdr_out exports:
//
//   - "from": ISO 8601 / "YYYY-MM-DD HH:MM:SS" lower bound, INCLUSIVE (time_start >= from).
//   - "to":   ISO 8601 / "YYYY-MM-DD HH:MM:SS" upper bound, EXCLUSIVE (time_start < to).
//   - "did_number": only for cdr_in exports.
//   - "voice_out_trunk_id": only for cdr_out exports.
type Export struct {
	ID                  string                 `json:"-" jsonapi:"exports"`
	Status              enums.ExportStatus     `json:"status" api:"readonly"`
	CreatedAt           time.Time              `json:"created_at" api:"readonly"`
	URL                 *string                `json:"url" api:"readonly"`
	CallbackURL         *string                `json:"callback_url,omitempty"`
	CallbackMethod      *string                `json:"callback_method,omitempty"`
	ExportType          enums.ExportType       `json:"export_type"`
	Filters             map[string]interface{} `json:"filters,omitempty"`
	ExternalReferenceID *string                `json:"external_reference_id,omitempty"`
}

// IsPending returns true when the export status is "pending".
func (e *Export) IsPending() bool { return e.Status == enums.ExportStatusPending }

// IsProcessing returns true when the export status is "processing".
func (e *Export) IsProcessing() bool { return e.Status == enums.ExportStatusProcessing }

// IsCompleted returns true when the export status is "completed".
func (e *Export) IsCompleted() bool { return e.Status == enums.ExportStatusCompleted }
