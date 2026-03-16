package orderitem

// AvailableDidOrderItem represents an order item for a specific available DID.
type AvailableDidOrderItem struct {
	DidOrderItem
	AvailableDidID string `json:"available_did_id,omitempty"`
}

func (i *AvailableDidOrderItem) orderItemType() string { return "did_order_items" }
