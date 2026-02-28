package didww

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestCapacityPoolsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/capacity_pools": {status: http.StatusOK, fixture: "capacity_pools/index.json"},
	})

	pools, err := client.CapacityPools().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pools) != 2 {
		t.Fatalf("expected 2 capacity pools, got %d", len(pools))
	}
}

func TestCapacityPoolsFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/capacity_pools/f288d07c-e2fc-4ae6-9837-b18fb469c324": {status: http.StatusOK, fixture: "capacity_pools/show.json"},
	})

	params := NewQueryParams().Include("countries,shared_capacity_groups,qty_based_pricings")
	pool, err := client.CapacityPools().Find(context.Background(), "f288d07c-e2fc-4ae6-9837-b18fb469c324", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ID != "f288d07c-e2fc-4ae6-9837-b18fb469c324" {
		t.Errorf("expected ID 'f288d07c-e2fc-4ae6-9837-b18fb469c324', got %q", pool.ID)
	}
	if pool.Name != "Standard" {
		t.Errorf("expected Name 'Standard', got %q", pool.Name)
	}
	if pool.TotalChannelsCount != 34 {
		t.Errorf("expected TotalChannelsCount 34, got %d", pool.TotalChannelsCount)
	}
	if pool.SetupPrice != "0.0" {
		t.Errorf("expected SetupPrice '0.0', got %q", pool.SetupPrice)
	}
	if pool.MonthlyPrice != "15.0" {
		t.Errorf("expected MonthlyPrice '15.0', got %q", pool.MonthlyPrice)
	}

	// Verify countries are resolved (fixture has many)
	if len(pool.Countries) == 0 {
		t.Error("expected non-empty Countries")
	}
}

func TestCapacityPoolsUpdate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/capacity_pools/f288d07c-e2fc-4ae6-9837-b18fb469c324": {status: http.StatusOK, fixture: "capacity_pools/update.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	pool, err := server.client.CapacityPools().Update(context.Background(), &CapacityPool{
		ID:                 "f288d07c-e2fc-4ae6-9837-b18fb469c324",
		TotalChannelsCount: 25,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ID != "f288d07c-e2fc-4ae6-9837-b18fb469c324" {
		t.Errorf("expected ID 'f288d07c-e2fc-4ae6-9837-b18fb469c324', got %q", pool.ID)
	}

	assertRequestJSON(t, capturedBody, "capacity_pools/update_request.json")
}
