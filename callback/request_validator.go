package callback

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // SHA-1 required by DIDWW webhook signature protocol
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// RequestValidator validates DIDWW webhook callback signatures.
type RequestValidator struct {
	apiKey string
}

// NewRequestValidator creates a new RequestValidator with the given API key.
func NewRequestValidator(apiKey string) *RequestValidator {
	return &RequestValidator{apiKey: apiKey}
}

// Validate checks whether the signature matches the expected HMAC-SHA1 of the URL and payload.
func (v *RequestValidator) Validate(rawURL string, payload map[string]string, signature string) bool {
	if signature == "" {
		return false
	}
	expected := v.ComputeSignature(rawURL, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ComputeSignature computes the expected HMAC-SHA1 hex digest for the given URL and payload.
func (v *RequestValidator) ComputeSignature(rawURL string, payload map[string]string) string {
	normalized := normalizeURL(rawURL)

	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var data strings.Builder
	data.WriteString(normalized)
	for _, k := range keys {
		data.WriteString(k)
		data.WriteString(payload[k])
	}

	mac := hmac.New(sha1.New, []byte(v.apiKey))
	mac.Write([]byte(data.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

// normalizeURL normalizes a URL to include explicit port numbers.
func normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	scheme := parsed.Scheme
	host := parsed.Hostname()

	port := parsed.Port()
	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	userInfo := ""
	if parsed.User != nil {
		userInfo = parsed.User.String() + "@"
	}

	path := parsed.Path

	query := ""
	if parsed.RawQuery != "" {
		query = "?" + parsed.RawQuery
	}

	fragment := ""
	if parsed.Fragment != "" {
		fragment = "#" + parsed.Fragment
	}

	return fmt.Sprintf("%s://%s%s:%s%s%s%s", scheme, userInfo, host, port, path, query, fragment)
}
