package resource

import "time"

// CapacityPool represents a capacity pool.
type CapacityPool struct {
	ID                    string `json:"-" jsonapi:"capacity_pools"`
	Name                  string `json:"name,omitempty"`
	RenewDate             string `json:"renew_date" api:"readonly"`
	TotalChannelsCount    int    `json:"total_channels_count"`
	AssignedChannelsCount int    `json:"assigned_channels_count" api:"readonly"`
	MinimumLimit          int    `json:"minimum_limit" api:"readonly"`
	MinimumQtyPerOrder    int    `json:"minimum_qty_per_order" api:"readonly"`
	SetupPrice            string `json:"setup_price" api:"readonly"`
	MonthlyPrice          string `json:"monthly_price" api:"readonly"`
	MeteredRate           string `json:"metered_rate" api:"readonly"`
	// Resolved relationships
	Countries            []*Country             `json:"-" rel:"countries"`
	SharedCapacityGroups []*SharedCapacityGroup `json:"-" rel:"shared_capacity_groups"`
	QtyBasedPricings     []*QtyBasedPricing     `json:"-" rel:"qty_based_pricings"`
}

// SharedCapacityGroup represents a shared capacity group.
type SharedCapacityGroup struct {
	ID                   string    `json:"-" jsonapi:"shared_capacity_groups"`
	Name                 string    `json:"name"`
	SharedChannelsCount  int       `json:"shared_channels_count"`
	CreatedAt            time.Time `json:"created_at" api:"readonly"`
	MeteredChannelsCount int       `json:"metered_channels_count"`
	ExternalReferenceID  *string   `json:"external_reference_id,omitempty"`
	// Relationship IDs for create/update
	CapacityPoolID string `json:"-" rel:"capacity_pool,capacity_pools"`
	// Resolved relationships
	CapacityPool *CapacityPool `json:"-" rel:"capacity_pool"`
	DIDs         []*DID        `json:"-" rel:"dids"`
}
