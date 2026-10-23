package didww

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testToken() Token {
	return Token{
		AccessToken: "secret-access-token",
		TokenType:   "Bearer",
		ExpiresIn:   900,
		Audience:    []string{"api.didww.com"},
		Scope:       "full",
		APIVersion:  "2026-10-22",
		MintedAt:    time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC),
	}
}

func TestTokenFormattingHidesAccessToken(t *testing.T) {
	tok := testToken()

	for _, format := range []string{"%v", "%+v", "%s", "%#v"} {
		out := fmt.Sprintf(format, tok)
		assert.NotContains(t, out, "secret-access-token", format)
		assert.Contains(t, out, "[FILTERED]", format)
		assert.Contains(t, out, "api.didww.com", format)
	}
	assert.NotContains(t, fmt.Sprintf("%v", &tok), "secret-access-token")
	assert.NotContains(t, fmt.Sprintf("%v", []Token{tok}), "secret-access-token")
}

func TestTokenLogValueHidesAccessToken(t *testing.T) {
	buf := &syncBuffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))

	logger.Info("token", "token", testToken())

	assert.NotContains(t, buf.String(), "secret-access-token")
	assert.Contains(t, buf.String(), `"aud":["api.didww.com"]`)
	assert.Contains(t, buf.String(), `"api_version":"2026-10-22"`)
}

func TestTokenJSONRoundTrip(t *testing.T) {
	tok := testToken()

	data, err := json.Marshal(tok)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"access_token":"secret-access-token"`)

	var decoded Token
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, tok, decoded)
}

func TestTokenFromFixture(t *testing.T) {
	var tok Token
	require.NoError(t, json.Unmarshal(loadFixture(t, "token/create.json"), &tok))

	assert.Equal(t, "test-access-token", tok.AccessToken)
	assert.Equal(t, "Bearer", tok.TokenType)
	assert.Equal(t, 900, tok.ExpiresIn)
	assert.Equal(t, []string{"api.didww.com"}, tok.Audience)
	assert.Equal(t, "full", tok.Scope)
	assert.Equal(t, "2026-10-22", tok.APIVersion)
}
