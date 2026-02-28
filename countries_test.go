package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestCountriesFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/countries/7eda11bb-0e66-4146-98e7-57a5281f56c8": {status: http.StatusOK, fixture: "countries/show.json"},
	})

	country, err := client.Countries().Find(context.Background(), "7eda11bb-0e66-4146-98e7-57a5281f56c8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if country.ID != "7eda11bb-0e66-4146-98e7-57a5281f56c8" {
		t.Errorf("expected ID '7eda11bb-0e66-4146-98e7-57a5281f56c8', got %q", country.ID)
	}
	if country.Name != "United Kingdom" {
		t.Errorf("expected Name 'United Kingdom', got %q", country.Name)
	}
	if country.Prefix != "44" {
		t.Errorf("expected Prefix '44', got %q", country.Prefix)
	}
	if country.ISO != "GB" {
		t.Errorf("expected ISO 'GB', got %q", country.ISO)
	}
}

func TestCountriesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/countries": {status: http.StatusOK, fixture: "countries/index.json"},
	})

	countries, err := client.Countries().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(countries) == 0 {
		t.Fatal("expected non-empty countries list")
	}

	first := countries[0]
	if first.ID == "" {
		t.Error("expected non-empty ID")
	}
	if first.Name == "" {
		t.Error("expected non-empty Name")
	}
	if first.Prefix == "" {
		t.Error("expected non-empty Prefix")
	}
	if first.ISO == "" {
		t.Error("expected non-empty ISO")
	}
}

func TestCountriesFindWithIncludedRegions(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/countries/661d8448-8897-4765-acda-00cc1740148d": {status: http.StatusOK, fixture: "countries/show_with_regions.json"},
	})

	params := NewQueryParams().Include("regions")
	country, err := client.Countries().Find(context.Background(), "661d8448-8897-4765-acda-00cc1740148d", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if country.ID != "661d8448-8897-4765-acda-00cc1740148d" {
		t.Errorf("expected ID '661d8448-8897-4765-acda-00cc1740148d', got %q", country.ID)
	}
	if country.Name != "Lithuania" {
		t.Errorf("expected Name 'Lithuania', got %q", country.Name)
	}
	if country.Prefix != "370" {
		t.Errorf("expected Prefix '370', got %q", country.Prefix)
	}
	if country.ISO != "LT" {
		t.Errorf("expected ISO 'LT', got %q", country.ISO)
	}
	if len(country.Regions) != 10 {
		t.Fatalf("expected 10 regions, got %d", len(country.Regions))
	}
	if country.Regions[0].Name != "Alytaus Apskritis" {
		t.Errorf("expected first region name 'Alytaus Apskritis', got %q", country.Regions[0].Name)
	}
}
