package enums

// OrderStatus defines the status of an order.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCanceled  OrderStatus = "canceled"
	OrderStatusCompleted OrderStatus = "completed"
)
