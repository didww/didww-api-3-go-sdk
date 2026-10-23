package didww

import (
	"regexp"
	"strings"
)

// Environment represents a DIDWW API environment: Sandbox or Production.
// NewClient rejects any other value; use WithBaseURL, WithAuthURL and the
// prices URL options to reach other hosts.
type Environment string

const (
	Sandbox    Environment = "https://sandbox-api.didww.com/v3"
	Production Environment = "https://api.didww.com/v3"
)

// DefaultAPIVersion is the API version requested when no other is configured.
// Earlier versions are not supported by this major version of the SDK.
const DefaultAPIVersion = "2026-10-22"

// Scope is the scope requested for access tokens.
type Scope string

const (
	// ScopeRead requests read-only tokens: every method but GET and HEAD is refused.
	ScopeRead Scope = "read"
	// ScopeFull requests tokens allowing every method.
	ScopeFull Scope = "full"
)

// Token audiences: the production hostname of each service, in every environment.
const (
	audienceAPI      = "api.didww.com"
	audienceVoiceIn  = "api-vin.didww.com"
	audienceVoiceOut = "api-vout.didww.com"
	audienceSMS      = "api-sms.didww.com"
)

// endpoints holds the token service and price host URLs of an environment.
type endpoints struct {
	auth     string
	voiceIn  string
	voiceOut string
	sms      string
}

var environmentEndpoints = map[Environment]endpoints{
	Sandbox: {
		auth:     "https://auth-sandbox.didww.com",
		voiceIn:  "https://api-vin-sandbox.didww.com/v3",
		voiceOut: "https://api-vout-sandbox.didww.com/v3",
		sms:      "https://api-sms-sandbox.didww.com/v3",
	},
	Production: {
		auth:     "https://auth.didww.com",
		voiceIn:  "https://api-vin.didww.com/v3",
		voiceOut: "https://api-vout.didww.com/v3",
		sms:      "https://api-sms.didww.com/v3",
	},
}

var apiVersionPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// joinV3 joins base and path, appending /v3 to base when it does not end with it.
func joinV3(base, path string) string {
	base = strings.TrimSuffix(base, "/")
	if !strings.HasSuffix(base, "/v3") {
		base += "/v3"
	}
	return base + "/" + strings.TrimPrefix(path, "/")
}
