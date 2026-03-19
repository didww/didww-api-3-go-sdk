package orderitem

// GenericOrderItem represents a generic order item.
type GenericOrderItem struct {
	BaseOrderItem
	Qty int `json:"qty,omitempty"`
}

func (i *GenericOrderItem) orderItemType() string { return typeGenericOrderItems }
