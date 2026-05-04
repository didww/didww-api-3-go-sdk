// Package redact provides shared helpers for masking sensitive credential
// values in default fmt / debug / log output. The wire format is never
// affected — only Stringer / GoStringer / repr-style methods route through
// these helpers.
package redact

// Mask returns "[FILTERED]" for any non-empty input and "" otherwise. The
// empty-string passthrough avoids leaking "this field was set" information
// when the value is genuinely unset.
func Mask(s string) string {
	if s == "" {
		return ""
	}
	return "[FILTERED]"
}
