package resource

import "github.com/didww/didww-api-3-go-sdk/resource/enums"

// DIDGroup represents a DID group.
type DIDGroup struct {
	ID                      string          `json:"-" jsonapi:"did_groups"`
	Prefix                  string          `json:"prefix"`
	Features                []enums.Feature `json:"features"`
	IsMetered               bool            `json:"is_metered"`
	AreaName                string          `json:"area_name"`
	AllowAdditionalChannels bool            `json:"allow_additional_channels"`
	// Resolved relationships
	Country           *Country            `json:"-" rel:"country"`
	City              *City               `json:"-" rel:"city"`
	Region            *Region             `json:"-" rel:"region"`
	DIDGroupType      *DIDGroupType       `json:"-" rel:"did_group_type"`
	StockKeepingUnits []*StockKeepingUnit `json:"-" rel:"stock_keeping_units"`
	AddressRequirement *AddressRequirement  `json:"-" rel:"address_requirement"`
}

// DIDGroupType represents a type of DID group.
type DIDGroupType struct {
	ID   string `json:"-" jsonapi:"did_group_types"`
	Name string `json:"name"`
}

// StockKeepingUnit represents an SKU for DID pricing.
type StockKeepingUnit struct {
	ID                    string `json:"-"`
	SetupPrice            string `json:"setup_price"`
	MonthlyPrice          string `json:"monthly_price"`
	ChannelsIncludedCount int    `json:"channels_included_count"`
}

// QtyBasedPricing represents quantity-based pricing for capacity pools.
type QtyBasedPricing struct {
	ID           string `json:"-"`
	SetupPrice   string `json:"setup_price"`
	MonthlyPrice string `json:"monthly_price"`
	Qty          int    `json:"qty"`
}
