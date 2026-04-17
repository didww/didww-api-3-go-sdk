package authenticationmethod

import (
	"encoding/json"
	"fmt"
)

// AuthenticationMethod is the interface for polymorphic authentication methods
// on VoiceOutTrunk resources.
type AuthenticationMethod interface {
	AuthenticationType() string
}

// IpOnly is a read-only authentication method.
// It can only be configured manually by DIDWW staff upon request
// and cannot be set via the API on create or update.
// Trunks with ip_only authentication can still be read and their
// non-auth attributes updated via the API.
type IpOnly struct {
	AllowedSipIPs []string `json:"allowed_sip_ips,omitempty"`
	TechPrefix    string   `json:"tech_prefix,omitempty"`
}

func (a *IpOnly) AuthenticationType() string { return "ip_only" }

// CredentialsAndIp uses credentials plus IP-based authentication.
// Username and Password are server-generated and returned in responses only.
type CredentialsAndIp struct {
	AllowedSipIPs []string `json:"allowed_sip_ips,omitempty"`
	TechPrefix    string   `json:"tech_prefix,omitempty"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password,omitempty"`
}

func (a *CredentialsAndIp) AuthenticationType() string { return "credentials_and_ip" }

// Twilio uses Twilio SIP trunking authentication.
type Twilio struct {
	TwilioAccountSid string `json:"twilio_account_sid,omitempty"`
}

func (a *Twilio) AuthenticationType() string { return "twilio" }

// Generic wraps an unknown authentication method type for forward-compatibility.
type Generic struct {
	Type       string                 `json:"-"`
	Attributes map[string]interface{} `json:"-"`
}

func (a *Generic) AuthenticationType() string { return a.Type }

// MarshalJSON marshals an authentication method as { type: ..., attributes: ... }.
func MarshalJSON(am AuthenticationMethod) ([]byte, error) {
	if am == nil {
		return []byte("null"), nil
	}
	if g, ok := am.(*Generic); ok {
		return json.Marshal(map[string]interface{}{
			"type":       g.Type,
			"attributes": g.Attributes,
		})
	}
	attrBytes, err := json.Marshal(am)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]json.RawMessage{
		"type":       json.RawMessage(fmt.Sprintf("%q", am.AuthenticationType())),
		"attributes": attrBytes,
	})
}

// UnmarshalJSON unmarshals an authentication method from { type: ..., attributes: ... }.
func UnmarshalJSON(data []byte) (AuthenticationMethod, error) {
	if string(data) == "null" {
		return nil, nil
	}
	var wrapper struct {
		Type       string          `json:"type"`
		Attributes json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	switch wrapper.Type {
	case "ip_only":
		var am IpOnly
		if err := json.Unmarshal(wrapper.Attributes, &am); err != nil {
			return nil, err
		}
		return &am, nil
	case "credentials_and_ip":
		var am CredentialsAndIp
		if err := json.Unmarshal(wrapper.Attributes, &am); err != nil {
			return nil, err
		}
		return &am, nil
	case "twilio":
		var am Twilio
		if err := json.Unmarshal(wrapper.Attributes, &am); err != nil {
			return nil, err
		}
		return &am, nil
	default:
		var attrs map[string]interface{}
		if wrapper.Attributes != nil {
			if err := json.Unmarshal(wrapper.Attributes, &attrs); err != nil {
				return nil, err
			}
		}
		return &Generic{Type: wrapper.Type, Attributes: attrs}, nil
	}
}
