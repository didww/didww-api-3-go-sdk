package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestCitiesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/cities": {status: http.StatusOK, fixture: "cities/index.json"},
	})

	cities, err := client.Cities().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cities) == 0 {
		t.Fatal("expected non-empty cities list")
	}
}

func TestCitiesFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/cities/368bf92f-c36e-473f-96fc-d53ed1b4028b": {status: http.StatusOK, fixture: "cities/show.json"},
	})

	params := NewQueryParams().Include("country,region")
	city, err := client.Cities().Find(context.Background(), "368bf92f-c36e-473f-96fc-d53ed1b4028b", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if city.ID != "368bf92f-c36e-473f-96fc-d53ed1b4028b" {
		t.Errorf("expected ID '368bf92f-c36e-473f-96fc-d53ed1b4028b', got %q", city.ID)
	}
	if city.Name != "New York" {
		t.Errorf("expected Name 'New York', got %q", city.Name)
	}
	if city.Country == nil {
		t.Fatal("expected non-nil Country")
	}
	if city.Country.Name != "United States" {
		t.Errorf("expected country name 'United States', got %q", city.Country.Name)
	}
	if city.Region == nil {
		t.Fatal("expected non-nil Region")
	}
	if city.Region.Name != "New York" {
		t.Errorf("expected region name 'New York', got %q", city.Region.Name)
	}
}
