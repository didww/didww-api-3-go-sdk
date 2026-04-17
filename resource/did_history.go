package resource

import "time"

// DIDHistory represents a DID ownership history record.
// Introduced in API 2026-04-16. Records are retained for the last 90 days.
type DIDHistory struct {
	ID        string    `json:"-" jsonapi:"did_history"`
	DIDNumber string    `json:"did_number" api:"readonly"`
	Action    string    `json:"action" api:"readonly"`
	Method    string    `json:"method" api:"readonly"`
	CreatedAt time.Time `json:"created_at" api:"readonly"`
}
