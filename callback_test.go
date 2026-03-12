package didww

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestValidator_Sandbox(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	url := "http://example.com/callback.php?id=7ae7c48f-d48a-499f-9dc1-c9217014b457&reject_reason=&status=approved&type=address_verifications"
	payload := map[string]string{
		"status":        "approved",
		"id":            "7ae7c48f-d48a-499f-9dc1-c9217014b457",
		"type":          "address_verifications",
		"reject_reason": "",
	}
	signature := "18050028b6b22d0ed516706fba1c1af8d6a8f9d5"
	assert.True(t, validator.Validate(url, payload, signature))
}

func TestRequestValidator_ValidRequest(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	url := "http://example.com/callbacks"
	payload := map[string]string{
		"status": "completed",
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"type":   "orders",
	}
	signature := "fe99e416c3547f2f59002403ec856ea386d05b2f"
	assert.True(t, validator.Validate(url, payload, signature))
}

func TestRequestValidator_ValidRequestWithQueryAndFragment(t *testing.T) {
	validator := NewRequestValidator("OTHERAPIKEY")
	url := "http://example.com/callbacks?foo=bar#baz"
	payload := map[string]string{
		"status": "completed",
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"type":   "orders",
	}
	signature := "32754ba93ac1207e540c0cf90371e7786b3b1cde"
	assert.True(t, validator.Validate(url, payload, signature))
}

func TestRequestValidator_EmptySignature(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	url := "http://example.com/callbacks"
	payload := map[string]string{
		"status": "completed",
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"type":   "orders",
	}
	assert.False(t, validator.Validate(url, payload, ""))
}

func TestRequestValidator_InvalidSignature(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	url := "http://example.com/callbacks"
	payload := map[string]string{
		"status": "completed",
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"type":   "orders",
	}
	assert.False(t, validator.Validate(url, payload, "fbdb1d1b18aa6c08324b7d64b71fb76370690e1d"))
}

// TestRequestValidator_DocumentationExample verifies the working example from the official DIDWW API documentation:
// https://doc.didww.com/api3/2022-05-10/callbacks-details.html#algorithm-implementation-details
func TestRequestValidator_DocumentationExample(t *testing.T) {
	validator := NewRequestValidator("szrdgh6547umt7tht7xbqhj6g9gdbyp7") // NOSONAR
	url := "https://mycompany.com/didww_callbacks?opaque=123"
	payload := map[string]string{
		"id":     "bf2cee72-6caa-4ae2-917e-bea01945691e",
		"status": "completed",
		"type":   "orders",
	}
	signature := "30f66e9d72eb5e193051fd02952f70d8e934b4ff"
	assert.True(t, validator.Validate(url, payload, signature))
}

func TestRequestValidator_URLNormalization(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	payload := map[string]string{
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"status": "completed",
		"type":   "orders",
	}

	tests := []struct {
		name      string
		url       string
		signature string
	}{
		{"http default", "http://foo.com/bar", "4d1ce2be656d20d064183bec2ab98a2ff3981f73"},
		{"http explicit port 80", "http://foo.com:80/bar", "4d1ce2be656d20d064183bec2ab98a2ff3981f73"},
		{"http port 443", "http://foo.com:443/bar", "904eaa65c0759afac0e4d8912de424e2dfb96ea1"},
		{"http non-standard port", "http://foo.com:8182/bar", "eb8fcfb3d7ed4b4c2265d73cf93c31ba614384d1"},

		{"http with query", "http://foo.com/bar?baz=boo", "78b00717a86ce9df06abf45ff818aa94537e1729"},
		{"http with userinfo", "http://user:pass@foo.com/bar", "88615a11a78c021c1da2e1e0bfb8cc165170afc5"}, // NOSONAR
		{"http with fragment", "http://foo.com/bar#test", "b1c4391fcdab7c0521bb5b9eb4f41f08529b8418"},
		{"https default", "https://foo.com/bar", "f26a771c302319a7094accbe2989bad67fff2928"},
		{"https explicit port 443", "https://foo.com:443/bar", "f26a771c302319a7094accbe2989bad67fff2928"},
		{"https port 80", "https://foo.com:80/bar", "bd45af5253b72f6383c6af7dc75250f12b73a4e1"},
		{"https non-standard port", "https://foo.com:8384/bar", "9c9fec4b7ebd6e1c461cb8e4ffe4f2987a19a5d3"},
		{"https with query", "https://foo.com/bar?qwe=asd", "4a0e98ddf286acadd1d5be1b0ed85a4e541c3137"},
		{"https with userinfo", "https://qwe:asd@foo.com/bar", "7a8cd4a6c349910dfecaf9807e56a63787250bbd"}, // NOSONAR
		{"https with fragment", "https://foo.com/bar#baz", "5024919770ea5ca2e3ccc07cb940323d79819508"},

		{"ipv6 http default port", "http://[::1]/bar", "e0e9b83e4046d097f54b3ae64b08cbb4a539f601"},
		{"ipv6 http explicit port 80", "http://[::1]:80/bar", "e0e9b83e4046d097f54b3ae64b08cbb4a539f601"},
		{"ipv6 http custom port", "http://[::1]:9090/bar", "ebec110ec5debd0e0fd086ff2f02e48ca665b543"},
		{"ipv6 https default port", "https://[::1]/bar", "f3cfe6f523fdf1d4eaadc310fcd3ed92e1e324b0"},

		{"percent-encoded path", "http://foo.com/hello%20world", "eb64035b2e8f356ff1442898a39ec94d5c3e2fc8"},
		{"percent-encoded slash in path", "http://foo.com/foo%2Fbar", "db24428442b012fa0972a453ba1ba98e755bba10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, validator.Validate(tt.url, payload, tt.signature),
				"URL %q should produce signature %s", tt.url, tt.signature)
		})
	}
}
