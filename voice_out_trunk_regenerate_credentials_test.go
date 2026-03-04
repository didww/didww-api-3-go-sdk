package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVoiceOutTrunkRegenerateCredentialsCreate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/voice_out_trunk_regenerate_credentials": {status: http.StatusCreated, fixture: "voice_out_trunk_regenerate_credentials/create.json"},
	})

	cred, err := client.VoiceOutTrunkRegenerateCredentials().Create(context.Background(), &VoiceOutTrunkRegenerateCredential{})
	require.NoError(t, err)

	assert.Equal(t, "5fc59e7e-79eb-498a-8779-800416b5c68a", cred.ID)
}
