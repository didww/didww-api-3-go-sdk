package resource

import "time"

// Address represents a customer address.
type Address struct {
	ID          string    `json:"-" jsonapi:"addresses"`
	CityName    string    `json:"city_name"`
	PostalCode  string    `json:"postal_code"`
	Address     string    `json:"address"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at" api:"readonly"`
	Verified    bool      `json:"verified" api:"readonly"`
	// Relationship IDs for create/update
	IdentityID string `json:"-" rel:"identity,identities"`
	CountryID  string `json:"-" rel:"country,countries"`
	// Resolved relationships
	Country  *Country  `json:"-" rel:"country"`
	Identity *Identity `json:"-" rel:"identity"`
	Proofs   []*Proof  `json:"-" rel:"proofs"`
}
