package orderitem

// ReservationDidOrderItem represents an order item for a reserved DID.
type ReservationDidOrderItem struct {
	DidOrderItem
	DidReservationID string `json:"did_reservation_id,omitempty"`
}

func (i *ReservationDidOrderItem) orderItemType() string { return "did_order_items" }
