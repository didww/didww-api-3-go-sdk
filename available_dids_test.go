package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestAvailableDIDsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/available_dids": {status: http.StatusOK, fixture: "available_dids/index.json"},
	})

	dids, err := client.AvailableDIDs().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dids) == 0 {
		t.Fatal("expected non-empty available dids list")
	}
}

func TestAvailableDIDsFindWithIncludedDIDGroup(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/available_dids/0b76223b-9625-412f-b0f3-330551473e7e": {status: http.StatusOK, fixture: "available_dids/show.json"},
	})

	params := NewQueryParams().Include("did_group")
	did, err := client.AvailableDIDs().Find(context.Background(), "0b76223b-9625-412f-b0f3-330551473e7e", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if did.ID != "0b76223b-9625-412f-b0f3-330551473e7e" {
		t.Errorf("expected ID '0b76223b-9625-412f-b0f3-330551473e7e', got %q", did.ID)
	}
	if did.Number != "16169886810" {
		t.Errorf("expected Number '16169886810', got %q", did.Number)
	}
	if did.DIDGroup == nil {
		t.Fatal("expected non-nil DIDGroup")
	}
	if did.DIDGroup.ID != "a9e3d346-d7bc-4a85-adb0-8ef1119cf237" {
		t.Errorf("expected DIDGroup ID 'a9e3d346-d7bc-4a85-adb0-8ef1119cf237', got %q", did.DIDGroup.ID)
	}
	if did.DIDGroup.AreaName != "Grand Rapids" {
		t.Errorf("expected DIDGroup AreaName 'Grand Rapids', got %q", did.DIDGroup.AreaName)
	}
}

func TestAvailableDIDsListWithNanpaPrefix(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/available_dids": {status: http.StatusOK, fixture: "available_dids/index_with_nanpa.json"},
	})

	dids, err := client.AvailableDIDs().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dids) != 1 {
		t.Fatalf("expected 1 available did, got %d", len(dids))
	}
	if dids[0].ID != "aa13b01c-36c8-405c-b5a8-1427aa7966ea" {
		t.Errorf("expected ID 'aa13b01c-36c8-405c-b5a8-1427aa7966ea', got %q", dids[0].ID)
	}
	if dids[0].Number != "18649204444" {
		t.Errorf("expected Number '18649204444', got %q", dids[0].Number)
	}
}

func TestAvailableDIDsFindWithNanpaPrefix(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/available_dids/ID": {status: http.StatusOK, fixture: "available_dids/show_with_nanpa_prefix.json"},
	})

	params := NewQueryParams().Include("nanpa_prefix")
	did, err := client.AvailableDIDs().Find(context.Background(), "ID", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if did.NanpaPrefix == nil {
		t.Fatal("expected non-nil NanpaPrefix")
	}
	if did.NanpaPrefix.NPA != "201" {
		t.Errorf("expected NPA '201', got %q", did.NanpaPrefix.NPA)
	}
	if did.NanpaPrefix.NXX != "221" {
		t.Errorf("expected NXX '221', got %q", did.NanpaPrefix.NXX)
	}
}
