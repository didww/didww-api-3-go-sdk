package didww

import (
	"net/url"
	"strings"
	"testing"
)

func TestQueryParamsEmpty(t *testing.T) {
	params := NewQueryParams()
	encoded := params.Encode()
	if encoded != "" {
		t.Errorf("expected empty string, got %q", encoded)
	}
}

func TestQueryParamsSingleFilter(t *testing.T) {
	params := NewQueryParams().Filter("country.iso", "US")
	encoded := params.Encode()
	assertContains(t, encoded, "filter[country.iso]=US")
}

func TestQueryParamsMultipleFilters(t *testing.T) {
	params := NewQueryParams().
		Filter("country.iso", "US").
		Filter("name", "test")
	encoded := params.Encode()
	assertContains(t, encoded, "filter[country.iso]=US")
	assertContains(t, encoded, "filter[name]=test")
}

func TestQueryParamsSortSingleField(t *testing.T) {
	params := NewQueryParams().Sort("name")
	encoded := params.Encode()
	assertContains(t, encoded, "sort=name")
}

func TestQueryParamsSortMultipleFields(t *testing.T) {
	params := NewQueryParams().Sort("-created_at", "name")
	encoded := params.Encode()
	assertContains(t, encoded, "sort=-created_at%2Cname")
}

func TestQueryParamsSortDescending(t *testing.T) {
	params := NewQueryParams().Sort("-created_at")
	encoded := params.Encode()
	assertContains(t, encoded, "sort=-created_at")
}

func TestQueryParamsIncludeSingle(t *testing.T) {
	params := NewQueryParams().Include("country")
	encoded := params.Encode()
	assertContains(t, encoded, "include=country")
}

func TestQueryParamsIncludeMultiple(t *testing.T) {
	params := NewQueryParams().Include("country", "region")
	encoded := params.Encode()
	assertContains(t, encoded, "include=country%2Cregion")
}

func TestQueryParamsPageNumber(t *testing.T) {
	params := NewQueryParams().Page(2, 25)
	encoded := params.Encode()
	assertContains(t, encoded, "page[number]=2")
	assertContains(t, encoded, "page[size]=25")
}

func TestQueryParamsPageDefaults(t *testing.T) {
	params := NewQueryParams().Page(1, 50)
	encoded := params.Encode()
	assertContains(t, encoded, "page[number]=1")
	assertContains(t, encoded, "page[size]=50")
}

func TestQueryParamsFields(t *testing.T) {
	params := NewQueryParams().Fields("countries", "name", "iso")
	encoded := params.Encode()
	assertContains(t, encoded, "fields[countries]=name%2Ciso")
}

func TestQueryParamsFieldsMultipleTypes(t *testing.T) {
	params := NewQueryParams().
		Fields("countries", "name", "iso").
		Fields("regions", "name")
	encoded := params.Encode()
	assertContains(t, encoded, "fields[countries]=name%2Ciso")
	assertContains(t, encoded, "fields[regions]=name")
}

func TestQueryParamsCombined(t *testing.T) {
	params := NewQueryParams().
		Filter("country.iso", "US").
		Sort("-created_at").
		Include("country", "region").
		Page(2, 25).
		Fields("countries", "name", "iso")

	encoded := params.Encode()

	assertContains(t, encoded, "filter[country.iso]=US")
	assertContains(t, encoded, "sort=-created_at")
	assertContains(t, encoded, "include=country%2Cregion")
	assertContains(t, encoded, "page[number]=2")
	assertContains(t, encoded, "page[size]=25")
	assertContains(t, encoded, "fields[countries]=name%2Ciso")
}

func TestQueryParamsEncodesToValidURLQuery(t *testing.T) {
	params := NewQueryParams().
		Filter("name", "hello world").
		Sort("name")

	encoded := params.Encode()

	// Should be parseable as URL query
	_, err := url.ParseQuery(encoded)
	if err != nil {
		t.Fatalf("expected valid URL query, got error: %v", err)
	}
}

func TestQueryParamsFilterWithSpecialCharacters(t *testing.T) {
	params := NewQueryParams().Filter("name", "hello&world=yes")
	encoded := params.Encode()

	// Should properly encode special characters
	values, err := url.ParseQuery(encoded)
	if err != nil {
		t.Fatalf("expected valid URL query, got error: %v", err)
	}
	if values.Get("filter[name]") != "hello&world=yes" {
		t.Errorf("expected filter value to be preserved after encoding, got %q", values.Get("filter[name]"))
	}
}

func TestQueryParamsChaining(t *testing.T) {
	// Ensure method chaining returns the same QueryParams
	params := NewQueryParams()
	result := params.Filter("a", "b").Sort("c").Include("d").Page(1, 10).Fields("e", "f")
	if result == nil {
		t.Fatal("expected chained result to be non-nil")
	}
}

// assertContains is a helper that checks if s contains substr
func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}
