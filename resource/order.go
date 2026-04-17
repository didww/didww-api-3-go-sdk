package resource

import (
	"encoding/json"
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/resource/orderitem"
)

// Order represents a DIDWW order.
type Order struct {
	ID                string                `json:"-" jsonapi:"orders"`
	Amount            string                `json:"amount" api:"readonly"`
	Status            enums.OrderStatus     `json:"status" api:"readonly"`
	CreatedAt         time.Time             `json:"created_at" api:"readonly"`
	Description       string                `json:"description" api:"readonly"`
	Reference         string                `json:"reference" api:"readonly"`
	Items             []orderitem.OrderItem `json:"items"`
	AllowBackOrdering bool                  `json:"allow_back_ordering,omitempty"`
	CallbackURL       *string               `json:"callback_url,omitempty"`
	CallbackMethod      *string               `json:"callback_method,omitempty"`
	ExternalReferenceID *string               `json:"external_reference_id,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling for Order.
func (o *Order) UnmarshalJSON(data []byte) error {
	type Alias Order
	aux := &struct {
		*Alias
		RawItems []json.RawMessage `json:"items"`
	}{
		Alias: (*Alias)(o),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	o.Items = make([]orderitem.OrderItem, 0, len(aux.RawItems))
	for _, raw := range aux.RawItems {
		item, err := orderitem.Parse(raw)
		if err != nil {
			return err
		}
		if item != nil {
			o.Items = append(o.Items, item)
		}
	}
	return nil
}

// MarshalJSON implements custom marshaling for Order.
func (o Order) MarshalJSON() ([]byte, error) { //nolint:gocritic // value receiver required for json.Marshal
	type Alias Order
	rawItems := make([]json.RawMessage, 0, len(o.Items))
	for _, item := range o.Items {
		raw, err := orderitem.MarshalItem(item)
		if err != nil {
			return nil, err
		}
		rawItems = append(rawItems, raw)
	}
	aux := &struct {
		Alias
		RawItems []json.RawMessage `json:"items,omitempty"`
	}{
		Alias:    Alias(o),
		RawItems: rawItems,
	}
	return json.Marshal(aux)
}
