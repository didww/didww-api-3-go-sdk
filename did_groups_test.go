package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestDIDGroupsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_groups": {status: http.StatusOK, fixture: "did_groups/index.json"},
	})

	groups, err := client.DIDGroups().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(groups) == 0 {
		t.Fatal("expected non-empty did groups list")
	}
}

func TestDIDGroupsFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_groups/2187c36d-28fb-436f-8861-5a0f5b5a3ee1": {status: http.StatusOK, fixture: "did_groups/show.json"},
	})

	params := NewQueryParams().Include("country,city,did_group_type,stock_keeping_units")
	group, err := client.DIDGroups().Find(context.Background(), "2187c36d-28fb-436f-8861-5a0f5b5a3ee1", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if group.ID != "2187c36d-28fb-436f-8861-5a0f5b5a3ee1" {
		t.Errorf("expected ID '2187c36d-28fb-436f-8861-5a0f5b5a3ee1', got %q", group.ID)
	}
	if group.Prefix != "241" {
		t.Errorf("expected Prefix '241', got %q", group.Prefix)
	}
	if group.AreaName != "Aachen" {
		t.Errorf("expected AreaName 'Aachen', got %q", group.AreaName)
	}
	if !group.AllowAdditionalChannels {
		t.Error("expected AllowAdditionalChannels to be true")
	}

	// Verify included country
	if group.Country == nil {
		t.Fatal("expected non-nil Country")
	}
	if group.Country.Name != "Germany" {
		t.Errorf("expected country name 'Germany', got %q", group.Country.Name)
	}
	if group.Country.ISO != "DE" {
		t.Errorf("expected country ISO 'DE', got %q", group.Country.ISO)
	}

	// Verify included city
	if group.City == nil {
		t.Fatal("expected non-nil City")
	}
	if group.City.Name != "Aachen" {
		t.Errorf("expected city name 'Aachen', got %q", group.City.Name)
	}

	// Verify included DID group type
	if group.DIDGroupType == nil {
		t.Fatal("expected non-nil DIDGroupType")
	}
	if group.DIDGroupType.Name != "Local" {
		t.Errorf("expected did_group_type name 'Local', got %q", group.DIDGroupType.Name)
	}

	// Verify included stock keeping units
	if len(group.StockKeepingUnits) != 2 {
		t.Fatalf("expected 2 stock keeping units, got %d", len(group.StockKeepingUnits))
	}
	if group.StockKeepingUnits[0].SetupPrice != "0.4" {
		t.Errorf("expected first SKU setup price '0.4', got %q", group.StockKeepingUnits[0].SetupPrice)
	}
	if group.StockKeepingUnits[0].MonthlyPrice != "0.8" {
		t.Errorf("expected first SKU monthly price '0.8', got %q", group.StockKeepingUnits[0].MonthlyPrice)
	}
	if group.StockKeepingUnits[0].ChannelsIncludedCount != 0 {
		t.Errorf("expected first SKU channels 0, got %d", group.StockKeepingUnits[0].ChannelsIncludedCount)
	}
	if group.StockKeepingUnits[1].ChannelsIncludedCount != 2 {
		t.Errorf("expected second SKU channels 2, got %d", group.StockKeepingUnits[1].ChannelsIncludedCount)
	}
}
