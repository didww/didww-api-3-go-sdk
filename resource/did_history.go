package resource

import (
	"encoding/json"
	"time"
)

// DIDHistory represents a DID ownership history record.
// Introduced in API 2026-04-16. Records are retained for the last 90 days.
//
// When Action is "billing_cycles_count_changed", the JSON:API resource-level
// meta block contains "from" and "to" string fields indicating the previous
// and new billing_cycles_count values. Meta is nil for all other actions.
type DIDHistory struct {
	ID        string            `json:"-" jsonapi:"did_history"`
	DIDNumber string            `json:"did_number" api:"readonly"`
	Action    string            `json:"action" api:"readonly"`
	Method    string            `json:"method" api:"readonly"`
	CreatedAt time.Time         `json:"created_at" api:"readonly"`
	Meta      map[string]string `json:"-"`
}

// UnmarshalMeta parses the resource-level JSON:API meta block into a generic map.
func (d *DIDHistory) UnmarshalMeta(raw json.RawMessage) error {
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	d.Meta = m
	return nil
}
