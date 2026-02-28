package enums

import "testing"

func TestOrderStatus(t *testing.T) {
	tests := []struct {
		name     string
		value    OrderStatus
		expected string
	}{
		{"Pending", OrderStatusPending, "Pending"},
		{"Canceled", OrderStatusCanceled, "Canceled"},
		{"Completed", OrderStatusCompleted, "Completed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}
