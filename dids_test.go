package didww

import (
	"context"
	"io"
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

func TestDIDsUpdateDescription(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/update_success.json"},
	})

	desc := "something"
	did, err := client.DIDs().Update(context.Background(), &DID{
		ID:          "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if did.Description == nil || *did.Description != "something" {
		t.Errorf("expected Description 'something', got %v", did.Description)
	}
	if did.ExpiresAt != "2019-01-27T10:00:04.755Z" {
		t.Errorf("expected ExpiresAt '2019-01-27T10:00:04.755Z', got %q", did.ExpiresAt)
	}
}

func TestDIDsFindBlockedTerminated(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/update_blocked_terminated.json"},
	})

	did, err := client.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !did.Blocked {
		t.Error("expected Blocked to be true")
	}
	if !did.Terminated {
		t.Error("expected Terminated to be true")
	}
	if did.BillingCyclesCount == nil || *did.BillingCyclesCount != 0 {
		t.Errorf("expected BillingCyclesCount 0, got %v", did.BillingCyclesCount)
	}
}

func TestDIDsUpdateTerminated(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/update_blocked_terminated.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	did, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:         "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		Terminated: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "dids/update_terminated_request.json")

	if did.ID != "9df99644-f1a5-4a3c-99a4-559d758eb96b" {
		t.Errorf("expected ID '9df99644-f1a5-4a3c-99a4-559d758eb96b', got %q", did.ID)
	}
	if !did.Blocked {
		t.Error("expected Blocked to be true")
	}
	if !did.Terminated {
		t.Error("expected Terminated to be true")
	}
	if did.BillingCyclesCount == nil || *did.BillingCyclesCount != 0 {
		t.Errorf("expected BillingCyclesCount 0, got %v", did.BillingCyclesCount)
	}
}

func TestDIDsUpdateInvalidParam(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusBadRequest, fixture: "dids/update_error_invalid_param.json"},
	})

	_, err := client.DIDs().Update(context.Background(), &DID{
		ID: "9df99644-f1a5-4a3c-99a4-559d758eb96b",
	})
	if err == nil {
		t.Fatal("expected error for invalid param")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.HTTPStatus != http.StatusBadRequest {
		t.Errorf("expected HTTP status 400, got %d", apiErr.HTTPStatus)
	}
	if len(apiErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(apiErr.Errors))
	}
	if apiErr.Errors[0].Code != "105" {
		t.Errorf("expected error code '105', got %q", apiErr.Errors[0].Code)
	}
}

func TestDIDsUpdateInvalidTrunkGroup(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusUnprocessableEntity, fixture: "dids/update_error_invalid_trunk_group.json"},
	})

	_, err := client.DIDs().Update(context.Background(), &DID{
		ID:                  "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		VoiceInTrunkGroupID: "invalid-id",
	})
	if err == nil {
		t.Fatal("expected error for invalid trunk group")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.HTTPStatus != http.StatusUnprocessableEntity {
		t.Errorf("expected HTTP status 422, got %d", apiErr.HTTPStatus)
	}
	if len(apiErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(apiErr.Errors))
	}
	if apiErr.Errors[0].Detail != "voice_in_trunk_group - is invalid" {
		t.Errorf("expected detail 'voice_in_trunk_group - is invalid', got %q", apiErr.Errors[0].Detail)
	}
}

func TestDIDsUpdateRequiresID(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{})

	_, err := client.DIDs().Update(context.Background(), &DID{})
	if err == nil {
		t.Fatal("expected error when updating without ID")
	}
}

func TestDIDsUpdateAssignTrunk(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show_with_trunk.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	did, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:             "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		VoiceInTrunkID: "41b94706-325e-4704-a433-d65105758836",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "dids/update_assign_trunk_request.json")

	if did.VoiceInTrunk == nil {
		t.Fatal("expected non-nil VoiceInTrunk")
	}
	if did.VoiceInTrunk.ID != "41b94706-325e-4704-a433-d65105758836" {
		t.Errorf("expected VoiceInTrunk ID '41b94706-325e-4704-a433-d65105758836', got %q", did.VoiceInTrunk.ID)
	}
	if did.VoiceInTrunk.Name != "hello, test pstn trunk" {
		t.Errorf("expected VoiceInTrunk Name 'hello, test pstn trunk', got %q", did.VoiceInTrunk.Name)
	}
}

func TestDIDsUpdateAssignTrunkGroup(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show_with_trunk_group.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	did, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:                  "9df99644-f1a5-4a3c-99a4-559d758eb96b",
		VoiceInTrunkGroupID: "b2319703-ce6c-480d-bb53-614e7abcfc96",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "dids/update_assign_trunk_group_request.json")

	if did.VoiceInTrunkGroup == nil {
		t.Fatal("expected non-nil VoiceInTrunkGroup")
	}
	if did.VoiceInTrunkGroup.ID != "b2319703-ce6c-480d-bb53-614e7abcfc96" {
		t.Errorf("expected VoiceInTrunkGroup ID 'b2319703-ce6c-480d-bb53-614e7abcfc96', got %q", did.VoiceInTrunkGroup.ID)
	}
	if did.VoiceInTrunkGroup.Name != "trunk group sample with 2 trunks" {
		t.Errorf("expected VoiceInTrunkGroup Name 'trunk group sample with 2 trunks', got %q", did.VoiceInTrunkGroup.Name)
	}
}

func TestDIDsFindWithTrunkResolved(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show_with_trunk.json"},
	})

	params := NewQueryParams().Include("voice_in_trunk")
	did, err := client.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if did.VoiceInTrunk == nil {
		t.Fatal("expected non-nil VoiceInTrunk")
	}
	if did.VoiceInTrunk.ID != "41b94706-325e-4704-a433-d65105758836" {
		t.Errorf("expected VoiceInTrunk ID '41b94706-325e-4704-a433-d65105758836', got %q", did.VoiceInTrunk.ID)
	}
	if did.VoiceInTrunk.Name != "hello, test pstn trunk" {
		t.Errorf("expected VoiceInTrunk Name 'hello, test pstn trunk', got %q", did.VoiceInTrunk.Name)
	}
	if did.VoiceInTrunkGroup != nil {
		t.Error("expected nil VoiceInTrunkGroup (mutual exclusivity)")
	}
}

func TestDIDsFindWithTrunkGroupResolved(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/dids/9df99644-f1a5-4a3c-99a4-559d758eb96b": {status: http.StatusOK, fixture: "dids/show_with_trunk_group.json"},
	})

	params := NewQueryParams().Include("voice_in_trunk_group")
	did, err := client.DIDs().Find(context.Background(), "9df99644-f1a5-4a3c-99a4-559d758eb96b", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if did.VoiceInTrunkGroup == nil {
		t.Fatal("expected non-nil VoiceInTrunkGroup")
	}
	if did.VoiceInTrunkGroup.ID != "b2319703-ce6c-480d-bb53-614e7abcfc96" {
		t.Errorf("expected VoiceInTrunkGroup ID 'b2319703-ce6c-480d-bb53-614e7abcfc96', got %q", did.VoiceInTrunkGroup.ID)
	}
	if did.VoiceInTrunkGroup.Name != "trunk group sample with 2 trunks" {
		t.Errorf("expected VoiceInTrunkGroup Name 'trunk group sample with 2 trunks', got %q", did.VoiceInTrunkGroup.Name)
	}
	if did.VoiceInTrunk != nil {
		t.Error("expected nil VoiceInTrunk (mutual exclusivity)")
	}
}
