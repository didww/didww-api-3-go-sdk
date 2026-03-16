package orderitem

import "encoding/json"

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
	case "did_order_items":
		var item DidOrderItem
		if err := json.Unmarshal(env.Attributes, &item); err != nil {
			return nil, err
		}
		return &item, nil
	case "capacity_order_items":
		var item CapacityOrderItem
		if err := json.Unmarshal(env.Attributes, &item); err != nil {
			return nil, err
		}
		return &item, nil
	case "generic_order_items":
		var item GenericOrderItem
		if err := json.Unmarshal(env.Attributes, &item); err != nil {
			return nil, err
		}
		return &item, nil
	default:
		return nil, nil
	}
}

// MarshalItem serializes an OrderItem to its JSON:API {type, attributes} envelope.
func MarshalItem(item OrderItem) ([]byte, error) {
	attrs, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type":       item.orderItemType(),
		"attributes": json.RawMessage(attrs),
	})
}
