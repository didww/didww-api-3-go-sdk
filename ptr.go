package didww

// Ptr returns a pointer to the value v. Useful for constructing API request
// payloads where a field's "omit when nil, send when non-nil" semantics
// matter — most commonly *bool, where the zero value `false` would otherwise
// be dropped by `json:",omitempty"`. Example:
//
//	cfg := &trunkconfiguration.SIPConfiguration{
//	    EnabledSipRegistration: didww.Ptr(false),
//	    UseDIDInRuri:           didww.Ptr(false),
//	    Host:                   "203.0.113.10",
//	}
func Ptr[T any](v T) *T { return &v }
