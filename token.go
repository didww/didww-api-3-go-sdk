package didww

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/didww/didww-api-3-go-sdk/v4/internal/redact"
)

// Token is an access token granted by the token service. It marshals to JSON
// with all fields, so that a TokenStore can persist it, while String, GoString
// and LogValue never show AccessToken.
type Token struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	Audience    []string  `json:"aud"`
	Scope       string    `json:"scope"`
	APIVersion  string    `json:"api_version"`
	MintedAt    time.Time `json:"minted_at"`
}

// String returns a description of the token without the access token.
func (t Token) String() string { //nolint:gocritic // value receiver so that Token values are formatted too
	return fmt.Sprintf("Token{AccessToken: %s, TokenType: %s, ExpiresIn: %d, Audience: %v, Scope: %s, APIVersion: %s, MintedAt: %s}",
		redact.Mask(t.AccessToken), t.TokenType, t.ExpiresIn, t.Audience, t.Scope, t.APIVersion, t.MintedAt.Format(time.RFC3339))
}

// GoString returns a Go-syntax description of the token without the access token.
func (t Token) GoString() string { //nolint:gocritic // value receiver so that Token values are formatted too
	return fmt.Sprintf("didww.Token{AccessToken:%q, TokenType:%q, ExpiresIn:%d, Audience:%#v, Scope:%q, APIVersion:%q, MintedAt:%q}",
		redact.Mask(t.AccessToken), t.TokenType, t.ExpiresIn, t.Audience, t.Scope, t.APIVersion, t.MintedAt.Format(time.RFC3339))
}

// LogValue implements slog.LogValuer without the access token.
func (t Token) LogValue() slog.Value { //nolint:gocritic // value receiver so that Token values are logged too
	return slog.GroupValue(
		slog.String("access_token", redact.Mask(t.AccessToken)),
		slog.String("token_type", t.TokenType),
		slog.Int("expires_in", t.ExpiresIn),
		slog.Any("aud", t.Audience),
		slog.String("scope", t.Scope),
		slog.String("api_version", t.APIVersion),
		slog.Time("minted_at", t.MintedAt),
	)
}
