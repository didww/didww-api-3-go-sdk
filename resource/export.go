package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
)

// Export represents a CDR or SMS export.
//
// Filters map keys (dotted keys are sent as written):
//
//   - every type, required: "from" and "to", "YYYY-MM-DD HH:MM:SS" UTC;
//     "from" is INCLUSIVE (time_start >= from), "to" is EXCLUSIVE (time_start < to).
//   - cdr_in: "did_number", "source_number", "source_number_out", "source_country_iso".
//   - cdr_out: "voice_out_trunk.id", "source", "cli", "destination".
//   - sms_in and sms_out: "source_address_eq", "source_address_cont",
//     "destination_address_eq", "destination_address_cont", "status", "sms_id",
//     "sms_trunk.id" (one id or a slice of ids).
//   - sms_in only: "trunk_type", one of the enums.ExportTrunkType* values or a slice of them.
//   - sms_out only: "service_type" ("a2p" or "p2p", any letter case),
//     "sender_id_verification.id" (one id or a slice of ids).
//
// Exports created at earlier API versions may also return "year", "month" and "day".
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
