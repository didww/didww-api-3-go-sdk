package orderitem

// CapacityOrderItem represents a capacity pool order item.
type CapacityOrderItem struct {
	BaseOrderItem
	CapacityPoolID string `json:"capacity_pool_id,omitempty"`
	Qty            int    `json:"qty,omitempty"`
}

func (i *CapacityOrderItem) orderItemType() string { return typeCapacityOrderItems }
