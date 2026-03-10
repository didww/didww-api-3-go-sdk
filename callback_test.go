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
