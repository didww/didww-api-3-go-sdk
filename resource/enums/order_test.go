package enums

import "testing"

func TestOrderStatus(t *testing.T) {
	tests := []struct {
		name     string
		value    OrderStatus
		expected string
	}{
		{"Pending", OrderStatusPending, "pending"},
		{"Canceled", OrderStatusCanceled, "canceled"},
		{"Completed", OrderStatusCompleted, "completed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}
