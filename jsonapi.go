package didww

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// jsonapiDocument represents a JSON:API response document.
type jsonapiDocument struct {
	Data     json.RawMessage `json:"data"`
	Included json.RawMessage `json:"included,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

// jsonapiResource represents a single JSON:API resource object.
type jsonapiResource struct {
	ID            string                     `json:"id"`
	Type          string                     `json:"type"`
	Attributes    json.RawMessage            `json:"attributes"`
	Relationships map[string]json.RawMessage `json:"relationships,omitempty"`
}

// IncludedResources maps "type:id" to the raw JSON:API resource object.
type IncludedResources map[string]json.RawMessage

// RelationshipMarshaler is implemented by resources that serialize JSON:API relationships.
type RelationshipMarshaler interface {
	MarshalRelationships() (map[string]any, error)
}

// RelationshipUnmarshaler is implemented by resources that parse JSON:API relationships.
type RelationshipUnmarshaler interface {
	UnmarshalRelationships(rels map[string]json.RawMessage) error
}

// RelationshipResolver is implemented by resources that resolve included relationships.
type RelationshipResolver interface {
	ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error
}

// RelationshipRef represents a JSON:API relationship linkage ({type, id}).
type RelationshipRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// ToOneRelationship builds a to-one relationship entry.
func ToOneRelationship(ref RelationshipRef) map[string]any {
	return map[string]any{"data": ref}
}

// ToManyRelationship builds a to-many relationship entry.
func ToManyRelationship(refs []RelationshipRef) map[string]any {
	return map[string]any{"data": refs}
}

// ParseToOneRelationship extracts a to-one relationship reference from raw JSON.
func ParseToOneRelationship(raw json.RawMessage) (*RelationshipRef, error) {
	var wrapper struct {
		Data *RelationshipRef `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data, nil
}

// ParseToManyRelationship extracts to-many relationship references from raw JSON.
func ParseToManyRelationship(raw json.RawMessage) ([]RelationshipRef, error) {
	var wrapper struct {
		Data []RelationshipRef `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data, nil
}

// ResolveToOne resolves a to-one relationship from included resources.
func ResolveToOne[T any](included IncludedResources, rels map[string]json.RawMessage, name string) (*T, error) {
	raw, ok := rels[name]
	if !ok {
		return nil, nil
	}
	ref, err := ParseToOneRelationship(raw)
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, nil
	}
	key := ref.Type + ":" + ref.ID
	resRaw, ok := included[key]
	if !ok {
		return nil, nil
	}
	return unmarshalResourceWithIncluded[T](resRaw, included)
}

// ResolveToMany resolves a to-many relationship from included resources.
func ResolveToMany[T any](included IncludedResources, rels map[string]json.RawMessage, name string) ([]*T, error) {
	raw, ok := rels[name]
	if !ok {
		return nil, nil
	}
	refs, err := ParseToManyRelationship(raw)
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return nil, nil
	}
	results := make([]*T, 0, len(refs))
	for _, ref := range refs {
		key := ref.Type + ":" + ref.ID
		resRaw, ok := included[key]
		if !ok {
			continue
		}
		item, err := unmarshalResourceWithIncluded[T](resRaw, included)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

// parseIncluded builds an IncludedResources map from the raw included array.
func parseIncluded(raw json.RawMessage) (IncludedResources, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var resources []json.RawMessage
	if err := json.Unmarshal(raw, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse included array: %w", err)
	}
	included := make(IncludedResources, len(resources))
	for _, resRaw := range resources {
		var res struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(resRaw, &res); err != nil {
			return nil, fmt.Errorf("failed to parse included resource: %w", err)
		}
		key := res.Type + ":" + res.ID
		included[key] = resRaw
	}
	return included, nil
}

// unmarshalOne parses a JSON:API document with a single data object into T.
func unmarshalOne[T any](body []byte) (*T, error) {
	var doc jsonapiDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON:API document: %w", err)
	}

	if len(doc.Data) == 0 || string(doc.Data) == "null" {
		return nil, fmt.Errorf("no data in response")
	}

	included, err := parseIncluded(doc.Included)
	if err != nil {
		return nil, err
	}

	// Check if data is an array (for singleton endpoints that return arrays)
	if doc.Data[0] == '[' {
		return unmarshalFirstFromArrayWithIncluded[T](doc.Data, included)
	}

	return unmarshalResourceWithIncluded[T](doc.Data, included)
}

// unmarshalMany parses a JSON:API document with an array of data objects into []*T.
func unmarshalMany[T any](body []byte) ([]*T, error) {
	var doc jsonapiDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON:API document: %w", err)
	}

	if len(doc.Data) == 0 || string(doc.Data) == "null" {
		return []*T{}, nil
	}

	included, err := parseIncluded(doc.Included)
	if err != nil {
		return nil, err
	}

	// Handle single object wrapped as data (not array)
	if doc.Data[0] == '{' {
		item, err := unmarshalResourceWithIncluded[T](doc.Data, included)
		if err != nil {
			return nil, err
		}
		return []*T{item}, nil
	}

	var rawResources []json.RawMessage
	if err := json.Unmarshal(doc.Data, &rawResources); err != nil {
		return nil, fmt.Errorf("failed to parse data array: %w", err)
	}

	results := make([]*T, 0, len(rawResources))
	for _, raw := range rawResources {
		item, err := unmarshalResourceWithIncluded[T](raw, included)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// unmarshalResourceWithIncluded parses a single JSON:API resource object into T, with included resolution.
func unmarshalResourceWithIncluded[T any](data []byte, included IncludedResources) (*T, error) {
	var res jsonapiResource
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("failed to parse resource: %w", err)
	}

	var result T
	if len(res.Attributes) > 0 && string(res.Attributes) != "null" {
		if err := json.Unmarshal(res.Attributes, &result); err != nil {
			return nil, fmt.Errorf("failed to parse attributes: %w", err)
		}
	}

	setID(&result, res.ID)

	// Parse relationships if the resource type supports it
	if len(res.Relationships) > 0 {
		if ru, ok := any(&result).(RelationshipUnmarshaler); ok {
			if err := ru.UnmarshalRelationships(res.Relationships); err != nil {
				return nil, fmt.Errorf("failed to parse relationships: %w", err)
			}
		}
	}

	// Resolve included relationships
	if len(included) > 0 && len(res.Relationships) > 0 {
		if rr, ok := any(&result).(RelationshipResolver); ok {
			if err := rr.ResolveRelationships(included, res.Relationships); err != nil {
				return nil, fmt.Errorf("failed to resolve relationships: %w", err)
			}
		}
	}

	return &result, nil
}

// unmarshalFirstFromArrayWithIncluded parses a JSON:API data array and returns the first element.
func unmarshalFirstFromArrayWithIncluded[T any](data []byte, included IncludedResources) (*T, error) {
	var rawResources []json.RawMessage
	if err := json.Unmarshal(data, &rawResources); err != nil {
		return nil, fmt.Errorf("failed to parse data array: %w", err)
	}

	if len(rawResources) == 0 {
		return nil, fmt.Errorf("empty data array")
	}

	return unmarshalResourceWithIncluded[T](rawResources[0], included)
}

// setID uses reflection to set the ID field on a resource struct.
func setID(resource any, id string) {
	v := reflect.ValueOf(resource)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	f := v.FieldByName("ID")
	if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
		f.SetString(id)
	}
}

// getID uses reflection to get the ID field from a resource struct.
func getID(resource any) string {
	v := reflect.ValueOf(resource)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	f := v.FieldByName("ID")
	if f.IsValid() && f.Kind() == reflect.String {
		return f.String()
	}
	return ""
}

// marshalResource serializes a resource into a JSON:API request body.
// Fields tagged with `api:"readonly"` (read-only) are excluded from the serialized attributes.
func marshalResource(resource any, resourceType string) ([]byte, error) {
	attrs, err := marshalWritableAttrs(resource)
	if err != nil {
		return nil, err
	}

	id := getID(resource)

	data := map[string]any{
		"type":       resourceType,
		"attributes": json.RawMessage(attrs),
	}

	if id != "" {
		data["id"] = id
	}

	// Include relationships if the resource marshals them
	if rm, ok := resource.(RelationshipMarshaler); ok {
		rels, err := rm.MarshalRelationships()
		if err != nil {
			return nil, err
		}
		if len(rels) > 0 {
			data["relationships"] = rels
		}
	}

	return json.Marshal(map[string]any{"data": data})
}

// marshalWritableAttrs serializes a resource to JSON, excluding fields tagged `api:"readonly"`.
// It first marshals normally via json.Marshal, then removes read-only keys.
func marshalWritableAttrs(resource any) ([]byte, error) {
	// Collect read-only JSON keys by inspecting struct tags.
	roKeys := readOnlyKeys(resource)

	if len(roKeys) == 0 {
		// No read-only fields; fast path.
		return json.Marshal(resource)
	}

	// Marshal to a generic map, remove read-only keys, re-marshal.
	raw, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	for _, key := range roKeys {
		delete(m, key)
	}
	return json.Marshal(m)
}

// readOnlyKeys returns the JSON key names of fields tagged `api:"readonly"`.
func readOnlyKeys(resource any) []string {
	v := reflect.ValueOf(resource)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return nil
	}
	var keys []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("api") == "readonly" {
			jsonTag := f.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			// Parse "name,omitempty" → "name"
			name := jsonTag
			if idx := len(name); idx > 0 {
				if comma := jsonTagName(jsonTag); comma != "" {
					name = comma
				}
			}
			keys = append(keys, name)
		}
	}
	return keys
}

// jsonTagName extracts the field name from a json struct tag value.
func jsonTagName(tag string) string {
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			return tag[:i]
		}
	}
	return tag
}
