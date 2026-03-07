package callback

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSandbox(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	url := "http://example.com/callback.php?id=7ae7c48f-d48a-499f-9dc1-c9217014b457&reject_reason=&status=approved&type=address_verifications"
	payload := map[string]string{
		"status":        "approved",
		"id":            "7ae7c48f-d48a-499f-9dc1-c9217014b457",
		"type":          "address_verifications",
		"reject_reason": "",
	}
	assert.True(t, validator.Validate(url, payload, "18050028b6b22d0ed516706fba1c1af8d6a8f9d5"))
}

func TestValidateValidRequest(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	payload := map[string]string{
		"status": "completed",
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"type":   "orders",
	}
	assert.True(t, validator.Validate("http://example.com/callbacks", payload, "fe99e416c3547f2f59002403ec856ea386d05b2f"))
}

func TestValidateValidRequestWithQueryAndFragment(t *testing.T) {
	validator := NewRequestValidator("OTHERAPIKEY")
	payload := map[string]string{
		"status": "completed",
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"type":   "orders",
	}
	assert.True(t, validator.Validate("http://example.com/callbacks?foo=bar#baz", payload, "32754ba93ac1207e540c0cf90371e7786b3b1cde"))
}

func TestValidateEmptySignature(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	payload := map[string]string{
		"status": "completed",
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"type":   "orders",
	}
	assert.False(t, validator.Validate("http://example.com/callbacks", payload, ""))
}

func TestValidateInvalidSignature(t *testing.T) {
	validator := NewRequestValidator("SOMEAPIKEY")
	payload := map[string]string{
		"status": "completed",
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"type":   "orders",
	}
	assert.False(t, validator.Validate("http://example.com/callbacks", payload, "fbdb1d1b18aa6c08324b7d64b71fb76370690e1d"))
}

func TestComputeSignatureSortsKeys(t *testing.T) {
	validator := NewRequestValidator("testkey")
	sig1 := validator.ComputeSignature("http://example.com", map[string]string{"b": "2", "a": "1"})
	sig2 := validator.ComputeSignature("http://example.com", map[string]string{"a": "1", "b": "2"})
	assert.Equal(t, sig1, sig2)
}
