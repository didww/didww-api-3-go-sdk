package orderitem

// EmergencyOrderItem represents an emergency service order item.
type EmergencyOrderItem struct {
	BaseOrderItem
	EmergencyCallingServiceID string `json:"emergency_calling_service_id,omitempty"`
	Qty                       int    `json:"qty,omitempty"`
}

func (i *EmergencyOrderItem) orderItemType() string { return typeEmergencyOrderItems }
