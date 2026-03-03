package jsonapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

const jsonNull = "null"

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

type dirtySnapshot struct {
	attributes    map[string]json.RawMessage
	relationships map[string]json.RawMessage
}

type dirtyKey struct {
	ptr uintptr
	typ reflect.Type
}

var (
	dirtyStateMu sync.RWMutex
	dirtyState   = map[dirtyKey]dirtySnapshot{}
)

// RelationshipRef represents a JSON:API relationship linkage ({type, id}).
type RelationshipRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// ToOneRelationship builds a to-one relationship entry.
func ToOneRelationship(ref RelationshipRef) map[string]any {
	return map[string]any{"data": ref}
}

// NullRelationship builds a null to-one relationship entry ({"data": null}).
func NullRelationship() map[string]any {
	return map[string]any{"data": nil}
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
	if len(raw) == 0 || string(raw) == jsonNull {
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

// UnmarshalOne parses a JSON:API document with a single data object into T.
func UnmarshalOne[T any](body []byte) (*T, error) {
	var doc jsonapiDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON:API document: %w", err)
	}

	if len(doc.Data) == 0 || string(doc.Data) == jsonNull {
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

// UnmarshalMany parses a JSON:API document with an array of data objects into []*T.
func UnmarshalMany[T any](body []byte) ([]*T, error) {
	var doc jsonapiDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON:API document: %w", err)
	}

	if len(doc.Data) == 0 || string(doc.Data) == jsonNull {
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
	if len(res.Attributes) > 0 && string(res.Attributes) != jsonNull {
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
		// Resolve from rel tags
		if err := resolveRelsFromTags(&result, included, res.Relationships); err != nil {
			return nil, fmt.Errorf("failed to resolve relationships: %w", err)
		}
		// Then interface (backward compat)
		if rr, ok := any(&result).(RelationshipResolver); ok {
			if err := rr.ResolveRelationships(included, res.Relationships); err != nil {
				return nil, fmt.Errorf("failed to resolve relationships: %w", err)
			}
		}
	}

	_ = rememberCleanState(&result)

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

// GetID uses reflection to get the ID field from a resource struct.
func GetID(resource any) string {
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

// ResourceType returns the JSON:API type string for resource T,
// read from the `jsonapi` struct tag on the ID field (tagged `json:"-"`).
// Panics if no jsonapi tag is found.
func ResourceType[T any]() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("json") == "-" {
			if apiType := f.Tag.Get("jsonapi"); apiType != "" {
				return apiType
			}
		}
	}
	panic(fmt.Sprintf("jsonapi: type %s has no jsonapi struct tag on its ID field", t.Name()))
}

// Marshal serializes a resource to JSON:API format, deriving the type from the struct tag.
func Marshal[T any](resource *T) ([]byte, error) {
	return MarshalResource(resource, ResourceType[T]())
}

// MarshalPatch serializes a resource for PATCH requests using dirty-only fields.
// Only fields modified since loading are included. Pointer fields set to nil produce explicit JSON null.
func MarshalPatch[T any](resource *T) ([]byte, error) {
	current, err := captureSnapshot(resource)
	if err != nil {
		return nil, err
	}

	baseline, err := baselineSnapshot(resource)
	if err != nil {
		return nil, err
	}

	dirtyAttrs := diffRawMaps(current.attributes, baseline.attributes, func(json.RawMessage) json.RawMessage {
		return json.RawMessage(jsonNull)
	})
	if dirtyAttrs == nil {
		dirtyAttrs = map[string]json.RawMessage{}
	}

	dirtyRels := diffRawMaps(current.relationships, baseline.relationships, relationshipClearPayload)

	data := map[string]any{
		"type":       ResourceType[T](),
		"attributes": dirtyAttrs,
	}
	if id := GetID(resource); id != "" {
		data["id"] = id
	}
	if len(dirtyRels) > 0 {
		data["relationships"] = dirtyRels
	}

	return json.Marshal(map[string]any{"data": data})
}

// resourceTypeFromTag extracts the JSON:API type from a resource's struct tag.
// Returns an empty string if no tag is found.
func resourceTypeFromTag(resource any) string {
	t := reflect.TypeOf(resource)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("json") == "-" {
			if apiType := f.Tag.Get("jsonapi"); apiType != "" {
				return apiType
			}
		}
	}
	return ""
}

// MarshalResource serializes a resource into a JSON:API request body.
// Fields tagged with `api:"readonly"` (read-only) are excluded from the serialized attributes.
// If resourceType is empty, it is auto-detected from the struct's jsonapi tag.
func MarshalResource(resource any, resourceType string) ([]byte, error) {
	if resourceType == "" {
		resourceType = resourceTypeFromTag(resource)
	}
	attrs, err := MarshalWritableAttrs(resource)
	if err != nil {
		return nil, err
	}

	id := GetID(resource)

	data := map[string]any{
		"type":       resourceType,
		"attributes": json.RawMessage(attrs),
	}

	if id != "" {
		data["id"] = id
	}

	// 1. Tag-based relationships
	rels := marshalRelsFromTags(resource)

	// 2. Merge interface-based relationships (overrides tags)
	if rm, ok := resource.(RelationshipMarshaler); ok {
		ifaceRels, err := rm.MarshalRelationships()
		if err != nil {
			return nil, err
		}
		for k, v := range ifaceRels {
			if rels == nil {
				rels = make(map[string]any)
			}
			rels[k] = v
		}
	}

	if len(rels) > 0 {
		data["relationships"] = rels
	}

	return json.Marshal(map[string]any{"data": data})
}

// MarshalWritableAttrs serializes a resource to JSON, excluding fields tagged `api:"readonly"`.
// It first marshals normally via json.Marshal, then removes read-only keys.
func MarshalWritableAttrs(resource any) ([]byte, error) {
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
			// Parse "name,omitempty" -> "name"
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

// splitRelTag parses a rel struct tag into name and apiType.
// Tags with a comma (e.g. "country,countries") are marshal tags (hasSep=true).
// Tags without a comma (e.g. "country") are resolve tags (hasSep=false).
func splitRelTag(tag string) (name, apiType string, hasSep bool) {
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			return tag[:i], tag[i+1:], true
		}
	}
	return tag, "", false
}

// marshalRelsFromTags scans a struct for `rel:"name,apitype"` tagged fields
// and builds a relationship map. string → to-one, []string → to-many.
func marshalRelsFromTags(resource any) map[string]any {
	v := reflect.ValueOf(resource)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	var rels map[string]any
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("rel")
		if tag == "" {
			continue
		}
		name, apiType, hasSep := splitRelTag(tag)
		if !hasSep {
			continue // resolve-only tag
		}
		fv := v.Field(i)
		switch f.Type.Kind() {
		case reflect.String:
			id := fv.String()
			if id == "" {
				continue
			}
			if rels == nil {
				rels = make(map[string]any)
			}
			rels[name] = ToOneRelationship(RelationshipRef{Type: apiType, ID: id})
		case reflect.Slice:
			if f.Type.Elem().Kind() != reflect.String {
				continue
			}
			if fv.Len() == 0 {
				continue
			}
			refs := make([]RelationshipRef, fv.Len())
			for j := 0; j < fv.Len(); j++ {
				refs[j] = RelationshipRef{Type: apiType, ID: fv.Index(j).String()}
			}
			if rels == nil {
				rels = make(map[string]any)
			}
			rels[name] = ToManyRelationship(refs)
		}
	}
	return rels
}

func captureSnapshot(resource any) (dirtySnapshot, error) {
	attrs, err := marshalDirtyAttributes(resource)
	if err != nil {
		return dirtySnapshot{}, err
	}
	rels, err := marshalRelationshipsRaw(resource)
	if err != nil {
		return dirtySnapshot{}, err
	}
	return dirtySnapshot{attributes: attrs, relationships: rels}, nil
}

func marshalDirtyAttributes(resource any) (map[string]json.RawMessage, error) {
	fieldAttrs, err := marshalAttrFields(resource)
	if err != nil {
		return nil, err
	}

	serializedAttrs, err := marshalWritableAttrsMap(resource)
	if err != nil {
		return nil, err
	}

	// Include attributes produced by custom MarshalJSON implementations
	// that may not map directly to struct fields.
	for key, raw := range serializedAttrs {
		if _, ok := fieldAttrs[key]; !ok {
			fieldAttrs[key] = cloneRaw(raw)
		}
	}

	return fieldAttrs, nil
}

func marshalAttrFields(resource any) (map[string]json.RawMessage, error) {
	v := reflect.ValueOf(resource)
	if !v.IsValid() {
		return map[string]json.RawMessage{}, nil
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return map[string]json.RawMessage{}, nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return map[string]json.RawMessage{}, nil
	}

	t := v.Type()
	attrs := make(map[string]json.RawMessage)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if f.Tag.Get("api") == "readonly" {
			continue
		}
		jsonTag := f.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := jsonTagName(jsonTag)
		if name == "" || name == "-" {
			continue
		}

		raw, err := json.Marshal(v.Field(i).Interface())
		if err != nil {
			return nil, err
		}
		attrs[name] = raw
	}

	return attrs, nil
}

func marshalWritableAttrsMap(resource any) (map[string]json.RawMessage, error) {
	raw, err := MarshalWritableAttrs(resource)
	if err != nil {
		return nil, err
	}
	var attrs map[string]json.RawMessage
	if err := json.Unmarshal(raw, &attrs); err != nil {
		return nil, err
	}
	if attrs == nil {
		attrs = map[string]json.RawMessage{}
	}
	return attrs, nil
}

func marshalRelationshipsRaw(resource any) (map[string]json.RawMessage, error) {
	rels := marshalRelsFromTags(resource)
	if rm, ok := resource.(RelationshipMarshaler); ok {
		ifaceRels, err := rm.MarshalRelationships()
		if err != nil {
			return nil, err
		}
		for k, v := range ifaceRels {
			if rels == nil {
				rels = make(map[string]any)
			}
			rels[k] = v
		}
	}

	if len(rels) == 0 {
		return map[string]json.RawMessage{}, nil
	}

	rawRels := make(map[string]json.RawMessage, len(rels))
	for k, v := range rels {
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		rawRels[k] = raw
	}
	return rawRels, nil
}

func cleanStateKey(resource any) (dirtyKey, bool) {
	v := reflect.ValueOf(resource)
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() {
		return dirtyKey{}, false
	}
	if v.Elem().Kind() != reflect.Struct {
		return dirtyKey{}, false
	}
	return dirtyKey{ptr: v.Pointer(), typ: v.Type()}, true
}

// ForgetCleanState removes the stored baseline for a resource pointer,
// freeing the associated memory. Safe to call with nil or non-pointer values.
func ForgetCleanState(resource any) {
	if key, ok := cleanStateKey(resource); ok {
		dirtyStateMu.Lock()
		delete(dirtyState, key)
		dirtyStateMu.Unlock()
	}
}

func rememberCleanState(resource any) error {
	key, ok := cleanStateKey(resource)
	if !ok {
		return nil
	}

	snapshot, err := captureSnapshot(resource)
	if err != nil {
		return err
	}

	dirtyStateMu.Lock()
	dirtyState[key] = cloneSnapshot(snapshot)
	dirtyStateMu.Unlock()

	// Register a finalizer to automatically clean up when the resource is GC'd.
	// This prevents memory leaks and stale pointer reuse.
	setCleanupFinalizer(resource, key)

	return nil
}

// setCleanupFinalizer registers a runtime finalizer that removes the dirty
// state entry when the resource pointer is garbage collected.
func setCleanupFinalizer(resource any, key dirtyKey) {
	v := reflect.ValueOf(resource)
	if v.Kind() != reflect.Ptr {
		return
	}
	runtime.SetFinalizer(resource, func(_ any) {
		dirtyStateMu.Lock()
		delete(dirtyState, key)
		dirtyStateMu.Unlock()
	})
}

func baselineSnapshot(resource any) (dirtySnapshot, error) {
	if key, ok := cleanStateKey(resource); ok {
		dirtyStateMu.RLock()
		snapshot, exists := dirtyState[key]
		dirtyStateMu.RUnlock()
		if exists {
			return cloneSnapshot(snapshot), nil
		}
	}

	zero := zeroResource(resource)
	if zero == nil {
		return dirtySnapshot{
			attributes:    map[string]json.RawMessage{},
			relationships: map[string]json.RawMessage{},
		}, nil
	}
	return captureSnapshot(zero)
}

func zeroResource(resource any) any {
	t := reflect.TypeOf(resource)
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Ptr {
		if t.Elem().Kind() != reflect.Struct {
			return nil
		}
		return reflect.New(t.Elem()).Interface()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	return reflect.New(t).Interface()
}

func diffRawMaps(
	current, baseline map[string]json.RawMessage,
	missingValue func(previous json.RawMessage) json.RawMessage,
) map[string]json.RawMessage {
	var dirty map[string]json.RawMessage
	for key, currentValue := range current {
		baselineValue, ok := baseline[key]
		if ok && bytes.Equal(currentValue, baselineValue) {
			continue
		}
		if dirty == nil {
			dirty = make(map[string]json.RawMessage)
		}
		dirty[key] = cloneRaw(currentValue)
	}
	for key, baselineValue := range baseline {
		if _, ok := current[key]; ok {
			continue
		}
		if dirty == nil {
			dirty = make(map[string]json.RawMessage)
		}
		dirty[key] = cloneRaw(missingValue(baselineValue))
	}
	return dirty
}

func relationshipClearPayload(previous json.RawMessage) json.RawMessage {
	var rel struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(previous, &rel); err == nil {
		data := bytes.TrimSpace(rel.Data)
		if len(data) > 0 && data[0] == '[' {
			return json.RawMessage(`{"data":[]}`)
		}
	}
	return json.RawMessage(`{"data":null}`)
}

func cloneSnapshot(snapshot dirtySnapshot) dirtySnapshot {
	return dirtySnapshot{
		attributes:    cloneRawMap(snapshot.attributes),
		relationships: cloneRawMap(snapshot.relationships),
	}
}

func cloneRawMap(rawMap map[string]json.RawMessage) map[string]json.RawMessage {
	if len(rawMap) == 0 {
		return map[string]json.RawMessage{}
	}
	out := make(map[string]json.RawMessage, len(rawMap))
	for key, value := range rawMap {
		out[key] = cloneRaw(value)
	}
	return out
}

func cloneRaw(raw json.RawMessage) json.RawMessage {
	if raw == nil {
		return nil
	}
	out := make([]byte, len(raw))
	copy(out, raw)
	return out
}

// resolveRelsFromTags scans a struct for `rel:"name"` tagged pointer fields
// and resolves them from included resources.
func resolveRelsFromTags(resource any, included IncludedResources, rels map[string]json.RawMessage) error {
	v := reflect.ValueOf(resource)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("rel")
		if tag == "" {
			continue
		}
		name, _, hasSep := splitRelTag(tag)
		if hasSep {
			continue // marshal tag, not resolve
		}
		fv := v.Field(i)
		ft := f.Type

		switch {
		case ft.Kind() == reflect.Ptr && ft.Elem().Kind() == reflect.Struct:
			if err := resolveToOneReflect(fv, ft.Elem(), name, included, rels); err != nil {
				return err
			}
		case ft.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Ptr &&
			ft.Elem().Elem().Kind() == reflect.Struct:
			if err := resolveToManyReflect(fv, ft, name, included, rels); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolveToOneReflect resolves a to-one relationship field via reflection.
func resolveToOneReflect(fv reflect.Value, elemType reflect.Type, name string, included IncludedResources, rels map[string]json.RawMessage) error {
	raw, ok := rels[name]
	if !ok {
		return nil
	}
	ref, err := ParseToOneRelationship(raw)
	if err != nil {
		return err
	}
	if ref == nil {
		return nil
	}
	resRaw, ok := included[ref.Type+":"+ref.ID]
	if !ok {
		return nil
	}
	val, err := unmarshalResourceReflect(resRaw, elemType, included)
	if err != nil {
		return err
	}
	fv.Set(val)
	return nil
}

// resolveToManyReflect resolves a to-many relationship field via reflection.
func resolveToManyReflect(fv reflect.Value, sliceType reflect.Type, name string, included IncludedResources, rels map[string]json.RawMessage) error {
	raw, ok := rels[name]
	if !ok {
		return nil
	}
	refs, err := ParseToManyRelationship(raw)
	if err != nil {
		return err
	}
	if len(refs) == 0 {
		return nil
	}
	elemType := sliceType.Elem().Elem()
	slice := reflect.MakeSlice(sliceType, 0, len(refs))
	for _, ref := range refs {
		resRaw, ok := included[ref.Type+":"+ref.ID]
		if !ok {
			continue
		}
		val, err := unmarshalResourceReflect(resRaw, elemType, included)
		if err != nil {
			return err
		}
		slice = reflect.Append(slice, val)
	}
	if slice.Len() > 0 {
		fv.Set(slice)
	}
	return nil
}

// unmarshalResourceReflect is the reflection-based equivalent of
// unmarshalResourceWithIncluded[T], used when the target type is known
// only at runtime (e.g. from rel tag field types).
func unmarshalResourceReflect(data []byte, elemType reflect.Type, included IncludedResources) (reflect.Value, error) {
	var res jsonapiResource
	if err := json.Unmarshal(data, &res); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to parse resource: %w", err)
	}

	ptr := reflect.New(elemType)
	result := ptr.Interface()

	if len(res.Attributes) > 0 && string(res.Attributes) != jsonNull {
		if err := json.Unmarshal(res.Attributes, result); err != nil {
			return reflect.Value{}, fmt.Errorf("failed to parse attributes: %w", err)
		}
	}

	setID(result, res.ID)

	if len(res.Relationships) > 0 {
		if ru, ok := result.(RelationshipUnmarshaler); ok {
			if err := ru.UnmarshalRelationships(res.Relationships); err != nil {
				return reflect.Value{}, fmt.Errorf("failed to parse relationships: %w", err)
			}
		}
	}

	if len(included) > 0 && len(res.Relationships) > 0 {
		if err := resolveRelsFromTags(result, included, res.Relationships); err != nil {
			return reflect.Value{}, fmt.Errorf("failed to resolve relationships: %w", err)
		}
	}

	_ = rememberCleanState(result)

	return ptr, nil
}
