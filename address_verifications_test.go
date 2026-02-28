package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func TestAddressVerificationsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/address_verifications": {status: http.StatusOK, fixture: "address_verifications/index.json"},
	})

	avs, err := client.AddressVerifications().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(avs) == 0 {
		t.Fatal("expected non-empty address verifications list")
	}
}

func TestAddressVerificationsCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/address_verifications": {status: http.StatusCreated, fixture: "address_verifications/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	cbURL := "http://example.com"
	cbMethod := "GET"
	av, err := server.client.AddressVerifications().Create(context.Background(), &AddressVerification{
		CallbackURL:    &cbURL,
		CallbackMethod: &cbMethod,
		AddressID:      "d3414687-40f4-4346-a267-c2c65117d28c",
		DIDIDs:         []string{"a9d64c02-4486-4acb-a9a1-be4c81ff0659"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if av.ID != "78182ef2-8377-41cd-89e1-26e8266c9c94" {
		t.Errorf("expected ID '78182ef2-8377-41cd-89e1-26e8266c9c94', got %q", av.ID)
	}
	if av.Status != enums.AddressVerificationStatusPending {
		t.Errorf("expected Status 'Pending', got %q", av.Status)
	}

	// Verify included address
	if av.AddressRel == nil {
		t.Fatal("expected non-nil AddressRel")
	}
	if av.AddressRel.ID != "d3414687-40f4-4346-a267-c2c65117d28c" {
		t.Errorf("expected address ID 'd3414687-40f4-4346-a267-c2c65117d28c', got %q", av.AddressRel.ID)
	}
	if av.AddressRel.CityName != "Chicago" {
		t.Errorf("expected address city 'Chicago', got %q", av.AddressRel.CityName)
	}

	assertRequestJSON(t, capturedBody, "address_verifications/create_request.json")
}

func TestAddressVerificationsFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/address_verifications/c8e004b0-87ec-4987-b4fb-ee89db099f0e": {status: http.StatusOK, fixture: "address_verifications/show.json"},
	})

	av, err := client.AddressVerifications().Find(context.Background(), "c8e004b0-87ec-4987-b4fb-ee89db099f0e")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if av.ID != "c8e004b0-87ec-4987-b4fb-ee89db099f0e" {
		t.Errorf("expected ID 'c8e004b0-87ec-4987-b4fb-ee89db099f0e', got %q", av.ID)
	}
	if av.Status != enums.AddressVerificationStatusApproved {
		t.Errorf("expected Status 'Approved', got %q", av.Status)
	}
	if av.Reference != "SHB-485120" {
		t.Errorf("expected Reference 'SHB-485120', got %q", av.Reference)
	}
}
