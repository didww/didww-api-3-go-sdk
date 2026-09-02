package resource

import "encoding/json"

// ResourceMeta is embedded by resources that carry a resource-level JSON:API
// meta block. The promoted UnmarshalMeta method satisfies jsonapi.MetaUnmarshaler.
type ResourceMeta struct {
	// Meta holds the resource-level JSON:API meta block as a string map.
	Meta map[string]string `json:"-"`
}

// UnmarshalMeta parses the resource-level JSON:API meta block into a string map.
func (r *ResourceMeta) UnmarshalMeta(raw json.RawMessage) error {
	meta, err := unmarshalStringMeta(raw)
	if err != nil {
		return err
	}
	r.Meta = meta
	return nil
}

// unmarshalStringMeta decodes a resource-level JSON:API meta block into a string map.
// A value that is not a JSON string keeps its literal JSON text, verbatim, so that one
// such value does not fail the whole response and a large number stays exact.
func unmarshalStringMeta(raw json.RawMessage) (map[string]string, error) {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}

	meta := make(map[string]string, len(values))
	for key, value := range values {
		// A JSON null decodes into a string as the empty string, as it did when
		// the whole map was decoded as map[string]string.
		var s string
		if err := json.Unmarshal(value, &s); err == nil {
			meta[key] = s
			continue
		}
		meta[key] = string(value)
	}

	return meta, nil
}
