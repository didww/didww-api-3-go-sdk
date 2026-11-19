// Validates the signature of a DIDWW callback with didww.RequestValidator.
//
// DIDWW signs every callback with the callback secret that is enabled in the
// DIDWW User Panel and sends the signature in the X-DIDWW-Signature header.
// The example builds a sample callback (URL and payload), signs it the way
// DIDWW does, and validates it in an HTTP handler: the signed callback prints
// "valid: true", the same callback with a tampered payload "valid: false".
//
// Usage: go run ./examples/webhook/ (no API key needed, no request is sent)
package main

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // DIDWW signs callbacks with HMAC-SHA1
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
)

const (
	// callbackSecret is the callback secret enabled in the DIDWW User Panel.
	callbackSecret = "your-callback-secret"
	// callbackURL is the callback_url set on the resource, here an order.
	callbackURL = "https://example.com/didww/callbacks?opaque=123"
)

func main() {
	validator := didww.NewRequestValidator(callbackSecret)

	payload := map[string]string{
		"id":     "1dd7a68b-e235-402b-8912-fe73ee14243a",
		"status": "completed",
		"type":   "orders",
	}
	// DIDWW signs the callback URL with its port spelled out
	signature := sign(callbackSecret, "https://example.com:443/didww/callbacks?opaque=123", payload)
	fmt.Printf("%s: %s\n", didww.SignatureHeaderName, signature)

	fmt.Printf("valid: %t\n", validateCallback(validator, newCallbackRequest(payload, signature)))

	tampered := map[string]string{}
	for k, v := range payload {
		tampered[k] = v
	}
	tampered["status"] = "canceled"
	fmt.Printf("valid: %t\n", validateCallback(validator, newCallbackRequest(tampered, signature)))
}

// validateCallback is what a webhook handler does: it reads the form payload
// and the signature header of the request and validates them against the
// callback URL.
func validateCallback(validator *didww.RequestValidator, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		return false
	}
	payload := make(map[string]string, len(r.PostForm))
	for key := range r.PostForm {
		payload[key] = r.PostForm.Get(key)
	}
	return validator.Validate(callbackURL, payload, r.Header.Get(didww.SignatureHeaderName))
}

// newCallbackRequest builds the POST request DIDWW sends to the callback URL.
func newCallbackRequest(payload map[string]string, signature string) *http.Request {
	form := url.Values{}
	for k, v := range payload {
		form.Set(k, v)
	}
	r := httptest.NewRequest(http.MethodPost, callbackURL, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set(didww.SignatureHeaderName, signature)
	return r
}

// sign computes a callback signature the way DIDWW does: the HMAC-SHA1, keyed
// with the callback secret, of the callback URL followed by every payload key
// and value, sorted by key, as a lowercase hex string.
func sign(secret, urlWithPort string, payload map[string]string) string {
	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var data strings.Builder
	data.WriteString(urlWithPort)
	for _, k := range keys {
		data.WriteString(k)
		data.WriteString(payload[k])
	}

	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write([]byte(data.String()))
	return hex.EncodeToString(mac.Sum(nil))
}
