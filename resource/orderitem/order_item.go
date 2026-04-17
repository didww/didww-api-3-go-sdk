package orderitem

import (
	"encoding/json"

	"github.com/didww/didww-api-3-go-sdk/jsonapi"
)

const (
	typeDidOrderItems       = "did_order_items"
	typeCapacityOrderItems  = "capacity_order_items"
	typeGenericOrderItems   = "generic_order_items"
	typeEmergencyOrderItems = "emergency_order_items"
)

// OrderItem is the interface for all order item types.
type OrderItem interface {
	orderItemType() string
}

// Parse deserializes a single order item JSON object (with "type" and "attributes" fields).
func Parse(data []byte) (OrderItem, error) {
	var env struct {
		Type       string          `json:"type"`
		Attributes json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Type {
	case typeDidOrderItems:
		var item DidOrderItem
		if err := json.Unmarshal(env.Attributes, &item); err != nil {
			return nil, err
		}
		return &item, nil
	case typeCapacityOrderItems:
		var item CapacityOrderItem
		if err := json.Unmarshal(env.Attributes, &item); err != nil {
			return nil, err
		}
		return &item, nil
	case typeGenericOrderItems:
		var item GenericOrderItem
		if err := json.Unmarshal(env.Attributes, &item); err != nil {
			return nil, err
		}
		return &item, nil
	case typeEmergencyOrderItems:
		var item EmergencyOrderItem
		if err := json.Unmarshal(env.Attributes, &item); err != nil {
			return nil, err
		}
		return &item, nil
	default:
		return nil, nil
	}
}

// MarshalItem serializes an OrderItem to its JSON:API {type, attributes} envelope.
// Read-only fields (tagged `api:"readonly"`) are excluded from the output.
func MarshalItem(item OrderItem) ([]byte, error) {
	attrs, err := jsonapi.MarshalWritableAttrs(item)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type":       item.orderItemType(),
		"attributes": json.RawMessage(attrs),
	})
}
