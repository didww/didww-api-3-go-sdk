package trunkconfiguration

import "encoding/json"

// TrunkConfiguration is an interface for voice in trunk configurations.
type TrunkConfiguration interface {
	ConfigurationType() string
}

// Parse deserializes a nested configuration object.
func Parse(data []byte) (TrunkConfiguration, error) {
	var env struct {
		Type       string          `json:"type"`
		Attributes json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Type {
	case "pstn_configurations":
		var cfg PSTNConfiguration
		if err := json.Unmarshal(env.Attributes, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	case "sip_configurations":
		var cfg SIPConfiguration
		if err := json.Unmarshal(env.Attributes, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	default:
		// Skip unsupported legacy configuration types (iax2, h323, etc.)
		return nil, nil
	}
}
