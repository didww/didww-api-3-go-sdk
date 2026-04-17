package resource

import (
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func TestAddressVerificationStatusPredicates(t *testing.T) {
	tests := []struct {
		status                          enums.AddressVerificationStatus
		expectPending, expectApproved, expectRejected bool
	}{
		{enums.AddressVerificationStatusPending, true, false, false},
		{enums.AddressVerificationStatusApproved, false, true, false},
		{enums.AddressVerificationStatusRejected, false, false, true},
	}
	for _, tt := range tests {
		av := &AddressVerification{Status: tt.status}
		if got := av.IsPending(); got != tt.expectPending {
			t.Errorf("IsPending() for %q = %v, want %v", tt.status, got, tt.expectPending)
		}
		if got := av.IsApproved(); got != tt.expectApproved {
			t.Errorf("IsApproved() for %q = %v, want %v", tt.status, got, tt.expectApproved)
		}
		if got := av.IsRejected(); got != tt.expectRejected {
			t.Errorf("IsRejected() for %q = %v, want %v", tt.status, got, tt.expectRejected)
		}
	}
}

func TestEmergencyVerificationStatusPredicates(t *testing.T) {
	tests := []struct {
		status                                        string
		expectPending, expectApproved, expectRejected bool
	}{
		{"pending", true, false, false},
		{"approved", false, true, false},
		{"rejected", false, false, true},
	}
	for _, tt := range tests {
		ev := &EmergencyVerification{Status: tt.status}
		if got := ev.IsPending(); got != tt.expectPending {
			t.Errorf("IsPending() for %q = %v, want %v", tt.status, got, tt.expectPending)
		}
		if got := ev.IsApproved(); got != tt.expectApproved {
			t.Errorf("IsApproved() for %q = %v, want %v", tt.status, got, tt.expectApproved)
		}
		if got := ev.IsRejected(); got != tt.expectRejected {
			t.Errorf("IsRejected() for %q = %v, want %v", tt.status, got, tt.expectRejected)
		}
	}
}

func TestOrderStatusPredicates(t *testing.T) {
	tests := []struct {
		status                                           enums.OrderStatus
		expectPending, expectCompleted, expectCancelled bool
	}{
		{enums.OrderStatusPending, true, false, false},
		{enums.OrderStatusCompleted, false, true, false},
		{enums.OrderStatusCanceled, false, false, true},
	}
	for _, tt := range tests {
		o := &Order{Status: tt.status}
		if got := o.IsPending(); got != tt.expectPending {
			t.Errorf("IsPending() for %q = %v, want %v", tt.status, got, tt.expectPending)
		}
		if got := o.IsCompleted(); got != tt.expectCompleted {
			t.Errorf("IsCompleted() for %q = %v, want %v", tt.status, got, tt.expectCompleted)
		}
		if got := o.IsCancelled(); got != tt.expectCancelled {
			t.Errorf("IsCancelled() for %q = %v, want %v", tt.status, got, tt.expectCancelled)
		}
	}
}
