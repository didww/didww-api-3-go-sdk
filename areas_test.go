package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestAreasList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/areas": {status: http.StatusOK, fixture: "areas/index.json"},
	})

	areas, err := client.Areas().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(areas) == 0 {
		t.Fatal("expected non-empty areas list")
	}
}

func TestAreasFindWithIncludedCountry(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/areas/ab2adc18-7c94-42d9-bdde-b28dfc373a22": {status: http.StatusOK, fixture: "areas/show.json"},
	})

	params := NewQueryParams().Include("country")
	area, err := client.Areas().Find(context.Background(), "ab2adc18-7c94-42d9-bdde-b28dfc373a22", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if area.ID != "ab2adc18-7c94-42d9-bdde-b28dfc373a22" {
		t.Errorf("expected ID 'ab2adc18-7c94-42d9-bdde-b28dfc373a22', got %q", area.ID)
	}
	if area.Name != "Tuscany" {
		t.Errorf("expected Name 'Tuscany', got %q", area.Name)
	}
	if area.Country == nil {
		t.Fatal("expected non-nil Country")
	}
	if area.Country.Name != "Italy" {
		t.Errorf("expected country name 'Italy', got %q", area.Country.Name)
	}
	if area.Country.Prefix != "39" {
		t.Errorf("expected country prefix '39', got %q", area.Country.Prefix)
	}
	if area.Country.ISO != "IT" {
		t.Errorf("expected country ISO 'IT', got %q", area.Country.ISO)
	}
}
