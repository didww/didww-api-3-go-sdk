package didww

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // SHA-1 required by DIDWW callback signature protocol
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// SignatureHeaderName is the HTTP header name used by DIDWW for callback signatures.
const SignatureHeaderName = "X-DIDWW-Signature"

// RequestValidator validates DIDWW callback request signatures using HMAC-SHA1.
type RequestValidator struct {
	apiKey string
}

// NewRequestValidator creates a new RequestValidator with the given API key.
func NewRequestValidator(apiKey string) *RequestValidator {
	return &RequestValidator{apiKey: apiKey}
}

// Validate checks whether the provided signature matches the expected HMAC-SHA1
// signature for the given URL and payload.
func (rv *RequestValidator) Validate(rawURL string, payload map[string]string, signature string) bool {
	if signature == "" {
		return false
	}
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	expected := rv.computeSignatureBytes(rawURL, payload)
	return hmac.Equal(sigBytes, expected)
}

// ComputeSignature computes the HMAC-SHA1 hex digest for the given URL and payload.
func (rv *RequestValidator) ComputeSignature(rawURL string, payload map[string]string) string {
	return hex.EncodeToString(rv.computeSignatureBytes(rawURL, payload))
}

// computeSignatureBytes computes the raw HMAC-SHA1 digest for the given URL and payload.
func (rv *RequestValidator) computeSignatureBytes(rawURL string, payload map[string]string) []byte {
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

	mac := hmac.New(sha1.New, []byte(rv.apiKey))
	_, _ = mac.Write([]byte(data.String()))
	return mac.Sum(nil)
}

func normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
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

	var b strings.Builder
	b.WriteString(scheme)
	b.WriteString("://")
	if parsed.User != nil {
		b.WriteString(parsed.User.String())
		b.WriteByte('@')
	}
	if strings.Contains(host, ":") {
		b.WriteByte('[')
		b.WriteString(host)
		b.WriteByte(']')
	} else {
		b.WriteString(host)
	}
	b.WriteByte(':')
	b.WriteString(port)
	path := parsed.EscapedPath()
	b.WriteString(path)
	if parsed.RawQuery != "" {
		b.WriteByte('?')
		b.WriteString(parsed.RawQuery)
	}
	if parsed.Fragment != "" {
		b.WriteByte('#')
		b.WriteString(parsed.Fragment)
	}

	return b.String()
}
