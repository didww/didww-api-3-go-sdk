package orderitem

// GenericOrderItem represents a generic order item.
type GenericOrderItem struct {
	Qty int `json:"qty,omitempty"`
}

func (i *GenericOrderItem) orderItemType() string { return "generic_order_items" }
