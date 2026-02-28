package enums

// OrderStatus defines the status of an order.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "Pending"
	OrderStatusCanceled  OrderStatus = "Canceled"
	OrderStatusCompleted OrderStatus = "Completed"
)
