package resource

import (
	"time"

	"github.com/didww/didww-api-3-go-sdk/jsonapi"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// OrderItemAttributes contains the attributes of an order item.
type OrderItemAttributes struct {
	Qty                int     `json:"qty,omitempty"`
	Nrc                string  `json:"nrc,omitempty" api:"readonly"`
	Mrc                string  `json:"mrc,omitempty" api:"readonly"`
	ProratedMrc        bool    `json:"prorated_mrc" api:"readonly"`
	BilledFrom         *string `json:"billed_from" api:"readonly"`
	BilledTo           *string `json:"billed_to" api:"readonly"`
	SetupPrice         string  `json:"setup_price,omitempty" api:"readonly"`
	MonthlyPrice       string  `json:"monthly_price,omitempty" api:"readonly"`
	DIDGroupID         string  `json:"did_group_id,omitempty"`
	SkuID              string  `json:"sku_id,omitempty"`
	AvailableDidID     string  `json:"available_did_id,omitempty"`
	DidReservationID   string  `json:"did_reservation_id,omitempty"`
	CapacityPoolID     string  `json:"capacity_pool_id,omitempty"`
	BillingCyclesCount *int    `json:"billing_cycles_count,omitempty"`
	NanpaPrefixID      string  `json:"nanpa_prefix_id,omitempty"`
}

// MarshalJSON implements custom marshaling for OrderItemAttributes to exclude read-only fields.
func (a OrderItemAttributes) MarshalJSON() ([]byte, error) { //nolint:gocritic // value receiver required for json.Marshal
	type Alias OrderItemAttributes
	return jsonapi.MarshalWritableAttrs(Alias(a))
}

// OrderItem represents an item within an order.
type OrderItem struct {
	Type       string              `json:"type"`
	Attributes OrderItemAttributes `json:"attributes"`
}

// Order represents a DIDWW order.
type Order struct {
	ID                string            `json:"-" jsonapi:"orders"`
	Amount            string            `json:"amount" api:"readonly"`
	Status            enums.OrderStatus `json:"status" api:"readonly"`
	CreatedAt         time.Time         `json:"created_at" api:"readonly"`
	Description       string            `json:"description" api:"readonly"`
	Reference         string            `json:"reference" api:"readonly"`
	Items             []OrderItem       `json:"items"`
	AllowBackOrdering bool              `json:"allow_back_ordering,omitempty"`
	CallbackURL       *string           `json:"callback_url,omitempty"`
	CallbackMethod    *string           `json:"callback_method,omitempty"`
}
