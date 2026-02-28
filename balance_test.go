package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestBalanceFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/balance": {status: http.StatusOK, fixture: "balance/index.json"},
	})

	balance, err := client.Balance().Find(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if balance.ID != "4c39e0bf-683b-4697-9322-5abaf4011883" {
		t.Errorf("expected ID '4c39e0bf-683b-4697-9322-5abaf4011883', got %q", balance.ID)
	}
	if balance.TotalBalance != "60.00" {
		t.Errorf("expected TotalBalance '60.00', got %q", balance.TotalBalance)
	}
	if balance.Credit != "10.00" {
		t.Errorf("expected Credit '10.00', got %q", balance.Credit)
	}
	if balance.Balance != "50.00" {
		t.Errorf("expected Balance '50.00', got %q", balance.Balance)
	}
}
