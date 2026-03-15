package resource

import "time"

// AvailableDID represents a DID available for purchase.
type AvailableDID struct {
	ID     string `json:"-" jsonapi:"available_dids"`
	Number string `json:"number"`
	// Resolved relationships
	DIDGroup    *DIDGroup    `json:"-" rel:"did_group"`
	NanpaPrefix *NanpaPrefix `json:"-" rel:"nanpa_prefix"`
}

// DIDReservation represents a reserved DID.
type DIDReservation struct {
	ID          string    `json:"-" jsonapi:"did_reservations"`
	ExpireAt    time.Time `json:"expire_at" api:"readonly"`
	CreatedAt   time.Time `json:"created_at" api:"readonly"`
	Description string    `json:"description"`
	// Relationship IDs for create/update
	AvailableDIDID string `json:"-" rel:"available_did,available_dids"`
	// Resolved relationships
	AvailableDID *AvailableDID `json:"-" rel:"available_did"`
}
