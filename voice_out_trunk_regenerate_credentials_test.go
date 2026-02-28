package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestVoiceOutTrunkRegenerateCredentialsCreate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/voice_out_trunk_regenerate_credentials": {status: http.StatusCreated, fixture: "voice_out_trunk_regenerate_credentials/create.json"},
	})

	cred, err := client.VoiceOutTrunkRegenerateCredentials().Create(context.Background(), &VoiceOutTrunkRegenerateCredential{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cred.ID != "5fc59e7e-79eb-498a-8779-800416b5c68a" {
		t.Errorf("expected ID '5fc59e7e-79eb-498a-8779-800416b5c68a', got %q", cred.ID)
	}
}
