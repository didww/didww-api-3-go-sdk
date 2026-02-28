package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func TestVoiceOutTrunksList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_out_trunks": {status: http.StatusOK, fixture: "voice_out_trunks/index.json"},
	})

	trunks, err := client.VoiceOutTrunks().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(trunks) != 2 {
		t.Fatalf("expected 2 voice out trunks, got %d", len(trunks))
	}

	trunk := trunks[0]
	if trunk.ID != "425ce763-a3a9-49b4-af5b-ada1a65c8864" {
		t.Errorf("expected ID '425ce763-a3a9-49b4-af5b-ada1a65c8864', got %q", trunk.ID)
	}
	if trunk.Name != "test" {
		t.Errorf("expected Name 'test', got %q", trunk.Name)
	}
	if trunk.Status != enums.VoiceOutTrunkStatusBlocked {
		t.Errorf("expected Status 'blocked', got %q", trunk.Status)
	}
}

func TestVoiceOutTrunksFindWithIncludedDids(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/voice_out_trunks/425ce763-a3a9-49b4-af5b-ada1a65c8864": {status: http.StatusOK, fixture: "voice_out_trunks/show.json"},
	})

	params := NewQueryParams().Include("dids,default_did")
	trunk, err := client.VoiceOutTrunks().Find(context.Background(), "425ce763-a3a9-49b4-af5b-ada1a65c8864", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trunk.ID != "425ce763-a3a9-49b4-af5b-ada1a65c8864" {
		t.Errorf("expected ID '425ce763-a3a9-49b4-af5b-ada1a65c8864', got %q", trunk.ID)
	}
	if trunk.Name != "test" {
		t.Errorf("expected Name 'test', got %q", trunk.Name)
	}
	if trunk.Username != "dpjgwbbac9" {
		t.Errorf("expected Username 'dpjgwbbac9', got %q", trunk.Username)
	}
	if trunk.Password != "z0hshvbcy7" {
		t.Errorf("expected Password 'z0hshvbcy7', got %q", trunk.Password)
	}
	if trunk.MediaEncryptionMode != enums.MediaEncryptionModeSrtpSdes {
		t.Errorf("expected MediaEncryptionMode 'srtp_sdes', got %q", trunk.MediaEncryptionMode)
	}
	if !trunk.ForceSymmetricRtp {
		t.Error("expected ForceSymmetricRtp to be true")
	}
	if !trunk.RtpPing {
		t.Error("expected RtpPing to be true")
	}

	// Verify included default_did
	if trunk.DefaultDID == nil {
		t.Fatal("expected non-nil DefaultDID")
	}
	if trunk.DefaultDID.ID != "7de7f718-4042-4d74-9fe9-863fa1777520" {
		t.Errorf("expected DefaultDID ID '7de7f718-4042-4d74-9fe9-863fa1777520', got %q", trunk.DefaultDID.ID)
	}
	if trunk.DefaultDID.Number != "37061498222" {
		t.Errorf("expected DefaultDID Number '37061498222', got %q", trunk.DefaultDID.Number)
	}

	// Verify included dids
	if len(trunk.DIDs) != 2 {
		t.Fatalf("expected 2 DIDs, got %d", len(trunk.DIDs))
	}
}

func TestVoiceOutTrunksCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/voice_out_trunks": {status: http.StatusCreated, fixture: "voice_out_trunks/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	trunk, err := server.client.VoiceOutTrunks().Create(context.Background(), &VoiceOutTrunk{
		Name:                "java-test",
		AllowedSipIPs:       []string{"0.0.0.0/0"},
		OnCliMismatchAction: enums.OnCliMismatchActionReplaceCli,
		DefaultDIDID:        "7a028c32-e6b6-4c86-bf01-90f901b37012",
		DIDIDs:              []string{"7a028c32-e6b6-4c86-bf01-90f901b37012"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trunk.ID != "b60201c1-21f0-4d9a-aafa-0e6d1e12f22e" {
		t.Errorf("expected ID 'b60201c1-21f0-4d9a-aafa-0e6d1e12f22e', got %q", trunk.ID)
	}

	assertRequestJSON(t, capturedBody, "voice_out_trunks/create_request.json")
}

func TestVoiceOutTrunksUpdate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/voice_out_trunks/425ce763-a3a9-49b4-af5b-ada1a65c8864": {status: http.StatusOK, fixture: "voice_out_trunks/update.json"},
	})

	trunk, err := client.VoiceOutTrunks().Update(context.Background(), &VoiceOutTrunk{
		ID:            "425ce763-a3a9-49b4-af5b-ada1a65c8864",
		AllowedSipIPs: []string{"10.11.12.13/32"},
		CapacityLimit: intPtr(123),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if trunk.ID != "425ce763-a3a9-49b4-af5b-ada1a65c8864" {
		t.Errorf("expected ID '425ce763-a3a9-49b4-af5b-ada1a65c8864', got %q", trunk.ID)
	}
	if trunk.Name != "test" {
		t.Errorf("expected Name 'test', got %q", trunk.Name)
	}
	if trunk.CapacityLimit == nil || *trunk.CapacityLimit != 123 {
		t.Errorf("expected CapacityLimit 123, got %v", trunk.CapacityLimit)
	}
	if len(trunk.AllowedSipIPs) != 1 || trunk.AllowedSipIPs[0] != "10.11.12.13/32" {
		t.Errorf("expected AllowedSipIPs ['10.11.12.13/32'], got %v", trunk.AllowedSipIPs)
	}
	if !trunk.ForceSymmetricRtp {
		t.Error("expected ForceSymmetricRtp to be true")
	}
	if !trunk.RtpPing {
		t.Error("expected RtpPing to be true")
	}
}

func TestVoiceOutTrunksDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/voice_out_trunks/425ce763-a3a9-49b4-af5b-ada1a65c8864": {status: http.StatusNoContent},
	})

	err := client.VoiceOutTrunks().Delete(context.Background(), "425ce763-a3a9-49b4-af5b-ada1a65c8864")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
