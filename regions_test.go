package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestRegionsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/regions": {status: http.StatusOK, fixture: "regions/index.json"},
	})

	regions, err := client.Regions().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(regions) == 0 {
		t.Fatal("expected non-empty regions list")
	}
}

func TestRegionsFindWithIncludedCountry(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/regions/c11b1f34-16cf-4ba6-8497-f305b53d5b01": {status: http.StatusOK, fixture: "regions/show.json"},
	})

	params := NewQueryParams().Include("country")
	region, err := client.Regions().Find(context.Background(), "c11b1f34-16cf-4ba6-8497-f305b53d5b01", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if region.ID != "c11b1f34-16cf-4ba6-8497-f305b53d5b01" {
		t.Errorf("expected ID 'c11b1f34-16cf-4ba6-8497-f305b53d5b01', got %q", region.ID)
	}
	if region.Name != "California" {
		t.Errorf("expected Name 'California', got %q", region.Name)
	}
	if region.Country == nil {
		t.Fatal("expected non-nil Country")
	}
	if region.Country.Name != "United States" {
		t.Errorf("expected country name 'United States', got %q", region.Country.Name)
	}
	if region.Country.Prefix != "1" {
		t.Errorf("expected country prefix '1', got %q", region.Country.Prefix)
	}
	if region.Country.ISO != "US" {
		t.Errorf("expected country ISO 'US', got %q", region.Country.ISO)
	}
}
