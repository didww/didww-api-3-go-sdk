package didww

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // SHA-1 required by DIDWW callback signature protocol
	"crypto/subtle"
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
	expected := rv.ComputeSignature(rawURL, payload)
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) == 1
}

// ComputeSignature computes the HMAC-SHA1 hex digest for the given URL and payload.
func (rv *RequestValidator) ComputeSignature(rawURL string, payload map[string]string) string {
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
	mac.Write([]byte(data.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

func normalizeURL(rawURL string) string {
	// If no scheme is present, prepend "http://" so url.Parse correctly
	// identifies the host (otherwise "foo.com/bar" is treated as a path).
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "http"
	}

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
	b.WriteString(host)
	b.WriteByte(':')
	b.WriteString(port)
	b.WriteString(parsed.Path)
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
