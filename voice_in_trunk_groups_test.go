package didww

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func intPtr(v int) *int {
	return &v
}

func TestVoiceInTrunkGroupsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_in_trunk_groups": {status: http.StatusOK, fixture: "voice_in_trunk_groups/index.json"},
	})

	groups, err := client.VoiceInTrunkGroups().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(groups) == 0 {
		t.Fatal("expected non-empty trunk groups list")
	}
}

func TestVoiceInTrunkGroupsCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/voice_in_trunk_groups": {status: http.StatusCreated, fixture: "voice_in_trunk_groups/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	group, err := server.client.VoiceInTrunkGroups().Create(context.Background(), &VoiceInTrunkGroup{
		Name:            "trunk group sample with 2 trunks",
		CapacityLimit:   intPtr(1000),
		VoiceInTrunkIDs: []string{"7c15bca2-7f17-46fb-9486-7e2a17158c7e", "b07a4cab-48c6-4b3a-9670-11b90b81bdef"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if group.ID != "b2319703-ce6c-480d-bb53-614e7abcfc96" {
		t.Errorf("expected ID 'b2319703-ce6c-480d-bb53-614e7abcfc96', got %q", group.ID)
	}
	if group.Name != "trunk group sample with 2 trunks" {
		t.Errorf("expected Name 'trunk group sample with 2 trunks', got %q", group.Name)
	}

	// Verify included voice_in_trunks
	if len(group.VoiceInTrunks) != 2 {
		t.Fatalf("expected 2 voice in trunks, got %d", len(group.VoiceInTrunks))
	}
	if group.VoiceInTrunks[0].Name != "test custom11" {
		t.Errorf("expected first trunk name 'test custom11', got %q", group.VoiceInTrunks[0].Name)
	}

	assertRequestJSON(t, capturedBody, "voice_in_trunk_groups/create_request.json")
}

func TestVoiceInTrunkGroupsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/voice_in_trunk_groups/b2319703-ce6c-480d-bb53-614e7abcfc96": {status: http.StatusNoContent},
	})

	err := client.VoiceInTrunkGroups().Delete(context.Background(), "b2319703-ce6c-480d-bb53-614e7abcfc96")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
