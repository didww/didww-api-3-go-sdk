// Package smstrunkconfiguration holds the delivery configurations of an SMS trunk.
//
// The configuration is sent as a whole object: fields left at their zero value
// are omitted from a request, and attributes left out keep their stored values.
// On update the configuration type must match the stored one.
//
// So the typed structs cannot clear a nullable attribute, such as the BodyType
// and Body of an HTTPIn when switching it to get, or the AllowedIPs of an
// HTTPOut. To send explicit nulls, use a Generic configuration of the same
// type, whose nil attribute values are sent as null:
//
//	&Generic{Type: TypeHTTPIn, Attributes: map[string]any{
//		"http_method": HTTPMethodGet, "body_type": nil, "body": nil,
//	}}
package smstrunkconfiguration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/didww/didww-api-3-go-sdk/v4/internal/redact"
)

// Configuration type values, as returned by ConfigurationType.
const (
	TypeSmsToEmail = "sms_to_email"
	TypeHTTPIn     = "http_in"
	TypeHTTPOut    = "http_out"
	TypeSmppEsme   = "smpp_esme"
	TypeSmppSmsc   = "smpp_smsc"
)

// HTTP methods of an HTTPIn configuration.
const (
	HTTPMethodGet  = "get"
	HTTPMethodPost = "post"
	HTTPMethodPut  = "put"
)

// Body types of an HTTPIn configuration.
const (
	BodyTypeJSON           = "json"
	BodyTypeHTMLMultipart  = "html_multipart"
	BodyTypeHTMLURLEncoded = "html_url_encoded"
	BodyTypeRaw            = "raw"
)

// DIDWW SMPP server locations. Production accepts nyc and ams; the sandbox accepts nyc only.
const (
	ServerLocationNYC = "nyc"
	ServerLocationAMS = "ams"
)

// SmsTrunkConfiguration is the interface for SMS trunk configurations.
type SmsTrunkConfiguration interface {
	ConfigurationType() string
}

// Parse deserializes a nested configuration object { type, attributes }.
// An unknown type is returned as *Generic, so it survives a read-modify-write.
// A JSON null yields a nil configuration.
func Parse(data []byte) (SmsTrunkConfiguration, error) {
	if string(data) == "null" {
		return nil, nil
	}
	var env struct {
		Type       string          `json:"type"`
		Attributes json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	var cfg SmsTrunkConfiguration
	switch env.Type {
	case TypeSmsToEmail:
		cfg = &SmsToEmail{}
	case TypeHTTPIn:
		cfg = &HTTPIn{}
	case TypeHTTPOut:
		cfg = &HTTPOut{}
	case TypeSmppEsme:
		cfg = &SmppEsme{}
	case TypeSmppSmsc:
		cfg = &SmppSmsc{}
	default:
		g := &Generic{Type: env.Type}
		if len(env.Attributes) > 0 {
			dec := json.NewDecoder(bytes.NewReader(env.Attributes))
			dec.UseNumber()
			if err := dec.Decode(&g.Attributes); err != nil {
				return nil, err
			}
		}
		return g, nil
	}
	if len(env.Attributes) > 0 {
		if err := json.Unmarshal(env.Attributes, cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// Marshal serializes a configuration as { type, attributes } for a request body.
// Read-only attributes are left out.
func Marshal(c SmsTrunkConfiguration) ([]byte, error) {
	attrs, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type":       c.ConfigurationType(),
		"attributes": json.RawMessage(attrs),
	})
}

// Generic wraps a configuration type this SDK version does not know yet.
// It keeps the type and the attributes as received; numbers are decoded as
// json.Number, so they keep their exact value.
type Generic struct {
	Type       string
	Attributes map[string]any
}

func (c *Generic) ConfigurationType() string { return c.Type }

// String implements fmt.Stringer and redacts the attribute values, which may
// carry credentials. Only the attribute names and explicit nulls are shown.
func (c *Generic) String() string {
	names := slices.Sorted(maps.Keys(c.Attributes))
	attrs := make([]string, len(names))
	for i, name := range names {
		value := nilString
		if v := c.Attributes[name]; v != nil {
			value = redact.Mask(fmt.Sprint(v))
		}
		attrs[i] = name + ":" + value
	}
	return fmt.Sprintf("Generic{Type:%q Attributes:map[%s]}", c.Type, strings.Join(attrs, " "))
}

// GoString mirrors String for the %#v verb.
func (c *Generic) GoString() string { return c.String() }

// MarshalJSON serializes the attributes of a Generic configuration.
func (c Generic) MarshalJSON() ([]byte, error) {
	if c.Attributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(c.Attributes)
}
