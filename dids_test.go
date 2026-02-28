package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestDIDsFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show.json"},
	})

	did, err := client.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if did.ID != "9df99644-f1a5-4a3c-99a4-559d758eb96b" {
		t.Errorf("expected ID '9df99644-f1a5-4a3c-99a4-559d758eb96b', got %q", did.ID)
	}
	if did.Number != "16091609123456797" {
		t.Errorf("expected Number '16091609123456797', got %q", did.Number)
	}
	if did.Blocked {
		t.Error("expected Blocked to be false")
	}
	if did.Terminated {
		t.Error("expected Terminated to be false")
	}
}

func TestDIDsListWithIncludedOrder(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids": {status: http.StatusOK, fixture: "dids/index.json"},
	})

	params := NewQueryParams().Include("order")
	dids, err := client.DIDs().List(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dids) != 3 {
		t.Fatalf("expected 3 dids, got %d", len(dids))
	}

	// First DID should have resolved order
	first := dids[0]
	if first.Order == nil {
		t.Fatal("expected non-nil Order on first DID")
	}
	if first.Order.ID != "11b3fba2-96e2-452e-bed8-5124ed351af3" {
		t.Errorf("expected order ID '11b3fba2-96e2-452e-bed8-5124ed351af3', got %q", first.Order.ID)
	}
	if first.Order.Amount != "0.37" {
		t.Errorf("expected order amount '0.37', got %q", first.Order.Amount)
	}
}

func TestDIDsFindWithAddressVerificationAndDIDGroup(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/21d0b02c-b556-4d3e-acbf-504b78295dbe": {status: http.StatusOK, fixture: "dids/show_with_address_verification_and_did_group.json"},
	})

	params := NewQueryParams().Include("address_verification,did_group")
	did, err := client.DIDs().Find(context.Background(), "21d0b02c-b556-4d3e-acbf-504b78295dbe", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if did.ID != "21d0b02c-b556-4d3e-acbf-504b78295dbe" {
		t.Errorf("expected ID '21d0b02c-b556-4d3e-acbf-504b78295dbe', got %q", did.ID)
	}
	if did.Number != "61488943592" {
		t.Errorf("expected Number '61488943592', got %q", did.Number)
	}

	// Verify address verification
	if did.AddressVerification == nil {
		t.Fatal("expected non-nil AddressVerification")
	}
	if did.AddressVerification.ID != "75dc8d39-5e17-4470-a6f3-df42642c975f" {
		t.Errorf("expected AV ID '75dc8d39-5e17-4470-a6f3-df42642c975f', got %q", did.AddressVerification.ID)
	}
	if did.AddressVerification.Status != "Approved" {
		t.Errorf("expected AV Status 'Approved', got %q", did.AddressVerification.Status)
	}

	// Verify DID group
	if did.DIDGroup == nil {
		t.Fatal("expected non-nil DIDGroup")
	}
	if did.DIDGroup.ID != "2b60bb9a-d382-4d35-84c6-61689f45f2f5" {
		t.Errorf("expected DIDGroup ID '2b60bb9a-d382-4d35-84c6-61689f45f2f5', got %q", did.DIDGroup.ID)
	}
	if did.DIDGroup.AreaName != "Mobile" {
		t.Errorf("expected DIDGroup AreaName 'Mobile', got %q", did.DIDGroup.AreaName)
	}
}

func TestDIDsUpdate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show.json"},
	})

	desc := "updated"
	_, err := client.DIDs().Update(context.Background(), &DID{
		ID:          "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDIDsUpdateRequiresID(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{})

	_, err := client.DIDs().Update(context.Background(), &DID{})
	if err == nil {
		t.Fatal("expected error when updating without ID")
	}
}
