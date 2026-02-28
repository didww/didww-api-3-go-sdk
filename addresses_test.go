package didww

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestAddressesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/addresses": {status: http.StatusOK, fixture: "addresses/index.json"},
	})

	addresses, err := client.Addresses().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(addresses) == 0 {
		t.Fatal("expected non-empty addresses list")
	}
}

func TestAddressesCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/addresses": {status: http.StatusCreated, fixture: "addresses/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	address, err := server.client.Addresses().Create(context.Background(), &Address{
		CityName:    "New York",
		PostalCode:  "123",
		Address:     "some street",
		Description: "test address",
		CountryID:   "1f6fc2bd-f081-4202-9b1a-d9cb88d942b9",
		IdentityID:  "5e9df058-50d2-4e34-b0d4-d1746b86f41a",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if address.ID != "bf69bc70-e1c2-442c-9f30-335ee299b663" {
		t.Errorf("expected ID 'bf69bc70-e1c2-442c-9f30-335ee299b663', got %q", address.ID)
	}
	if address.CityName != "New York" {
		t.Errorf("expected CityName 'New York', got %q", address.CityName)
	}

	// Verify included country
	if address.Country == nil {
		t.Fatal("expected non-nil Country")
	}
	if address.Country.Name != "United States" {
		t.Errorf("expected country name 'United States', got %q", address.Country.Name)
	}

	assertRequestJSON(t, capturedBody, "addresses/create_request.json")
}

func TestAddressesDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/addresses/bf69bc70-e1c2-442c-9f30-335ee299b663": {status: http.StatusNoContent},
	})

	err := client.Addresses().Delete(context.Background(), "bf69bc70-e1c2-442c-9f30-335ee299b663")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
