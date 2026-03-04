package jsonapi

import (
	"encoding/json"
	"testing"
)

// --- test helper types ---

type testResource struct {
	ID   string `json:"-" jsonapi:"test_resources"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type testResourceWithReadonly struct {
	ID        string `json:"-"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at" api:"readonly"`
	Status    string `json:"status,omitempty" api:"readonly"`
}

type testResourceWithRels struct {
	ID   string `json:"-"`
	Name string `json:"name"`
	// relationship IDs
	ParentID string `json:"-"`
}

func (r *testResourceWithRels) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if r.ParentID != "" {
		rels["parent"] = ToOneRelationship(RelationshipRef{Type: "parents", ID: r.ParentID})
	}
	return rels, nil
}

type testResourceWithResolver struct {
	ID     string        `json:"-"`
	Name   string        `json:"name"`
	Parent *testResource `json:"-"`
}

func (r *testResourceWithResolver) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if parent, err := ResolveToOne[testResource](included, rels, "parent"); err != nil {
		return err
	} else if parent != nil {
		r.Parent = parent
	}
	return nil
}

// --- Tests ---

func TestGetID(t *testing.T) {
	t.Run("struct with ID field", func(t *testing.T) {
		r := testResource{ID: "abc"}
		if got := GetID(r); got != "abc" {
			t.Errorf("GetID() = %q, want %q", got, "abc")
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		r := &testResource{ID: "xyz"}
		if got := GetID(r); got != "xyz" {
			t.Errorf("GetID() = %q, want %q", got, "xyz")
		}
	})

	t.Run("no ID field", func(t *testing.T) {
		type noID struct {
			Name string
		}
		if got := GetID(noID{Name: "test"}); got != "" {
			t.Errorf("GetID() = %q, want empty", got)
		}
	})

	t.Run("non-struct", func(t *testing.T) {
		if got := GetID("hello"); got != "" {
			t.Errorf("GetID() = %q, want empty", got)
		}
	})
}

func TestSetID(t *testing.T) {
	t.Run("sets ID on pointer", func(t *testing.T) {
		r := &testResource{}
		setID(r, "123")
		if r.ID != "123" {
			t.Errorf("setID() ID = %q, want %q", r.ID, "123")
		}
	})

	t.Run("no-op without ID field", func(t *testing.T) {
		type noID struct {
			Name string
		}
		r := &noID{Name: "test"}
		setID(r, "123") // should not panic
		if r.Name != "test" {
			t.Errorf("setID() modified Name unexpectedly")
		}
	})
}

func TestMarshalWritableAttrs(t *testing.T) {
	t.Run("no readonly fields", func(t *testing.T) {
		r := testResource{Name: "Alice", Age: 30}
		data, err := MarshalWritableAttrs(r)
		if err != nil {
			t.Fatalf("MarshalWritableAttrs() error = %v", err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if _, ok := m["name"]; !ok {
			t.Error("expected 'name' key")
		}
		if _, ok := m["age"]; !ok {
			t.Error("expected 'age' key")
		}
	})

	t.Run("with readonly fields", func(t *testing.T) {
		r := testResourceWithReadonly{Name: "Bob", CreatedAt: "2024-01-01", Status: "active"}
		data, err := MarshalWritableAttrs(r)
		if err != nil {
			t.Fatalf("MarshalWritableAttrs() error = %v", err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if _, ok := m["name"]; !ok {
			t.Error("expected 'name' key")
		}
		if _, ok := m["created_at"]; ok {
			t.Error("expected 'created_at' to be excluded (readonly)")
		}
		if _, ok := m["status"]; ok {
			t.Error("expected 'status' to be excluded (readonly with omitempty)")
		}
	})
}

func TestMarshalResource(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		r := testResource{Name: "Alice", Age: 30}
		data, err := MarshalResource(r, "test_resources")
		if err != nil {
			t.Fatalf("MarshalResource() error = %v", err)
		}
		var doc map[string]json.RawMessage
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		var d map[string]json.RawMessage
		if err := json.Unmarshal(doc["data"], &d); err != nil {
			t.Fatalf("unmarshal data error = %v", err)
		}
		var typ string
		json.Unmarshal(d["type"], &typ)
		if typ != "test_resources" {
			t.Errorf("type = %q, want %q", typ, "test_resources")
		}
		if _, ok := d["id"]; ok {
			t.Error("expected no 'id' when ID is empty")
		}
	})

	t.Run("with ID", func(t *testing.T) {
		r := testResource{ID: "42", Name: "Bob"}
		data, err := MarshalResource(r, "test_resources")
		if err != nil {
			t.Fatalf("MarshalResource() error = %v", err)
		}
		var doc struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if doc.Data.ID != "42" {
			t.Errorf("id = %q, want %q", doc.Data.ID, "42")
		}
	})

	t.Run("with RelationshipMarshaler", func(t *testing.T) {
		r := &testResourceWithRels{Name: "Child", ParentID: "99"}
		data, err := MarshalResource(r, "children")
		if err != nil {
			t.Fatalf("MarshalResource() error = %v", err)
		}
		var doc struct {
			Data struct {
				Relationships map[string]json.RawMessage `json:"relationships"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if _, ok := doc.Data.Relationships["parent"]; !ok {
			t.Error("expected 'parent' relationship")
		}
	})

	t.Run("readonly exclusion", func(t *testing.T) {
		r := testResourceWithReadonly{Name: "Test", CreatedAt: "2024-01-01"}
		data, err := MarshalResource(r, "test_resources")
		if err != nil {
			t.Fatalf("MarshalResource() error = %v", err)
		}
		var doc struct {
			Data struct {
				Attributes json.RawMessage `json:"attributes"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		var attrs map[string]json.RawMessage
		json.Unmarshal(doc.Data.Attributes, &attrs)
		if _, ok := attrs["created_at"]; ok {
			t.Error("expected 'created_at' to be excluded from attributes")
		}
	})
}

func TestUnmarshalOne(t *testing.T) {
	t.Run("single object", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"test","attributes":{"name":"Alice","age":30}}}`
		r, err := UnmarshalOne[testResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}
		if r.ID != "1" {
			t.Errorf("ID = %q, want %q", r.ID, "1")
		}
		if r.Name != "Alice" {
			t.Errorf("Name = %q, want %q", r.Name, "Alice")
		}
		if r.Age != 30 {
			t.Errorf("Age = %d, want %d", r.Age, 30)
		}
	})

	t.Run("array-wrapped singleton", func(t *testing.T) {
		body := `{"data":[{"id":"2","type":"test","attributes":{"name":"Bob","age":25}}]}`
		r, err := UnmarshalOne[testResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}
		if r.ID != "2" {
			t.Errorf("ID = %q, want %q", r.ID, "2")
		}
		if r.Name != "Bob" {
			t.Errorf("Name = %q, want %q", r.Name, "Bob")
		}
	})

	t.Run("null data error", func(t *testing.T) {
		body := `{"data":null}`
		_, err := UnmarshalOne[testResource]([]byte(body))
		if err == nil {
			t.Fatal("expected error for null data")
		}
	})

	t.Run("with included and resolver", func(t *testing.T) {
		body := `{
			"data":{"id":"1","type":"children","attributes":{"name":"Child"},
				"relationships":{"parent":{"data":{"type":"tests","id":"99"}}}},
			"included":[{"id":"99","type":"tests","attributes":{"name":"Parent","age":50}}]
		}`
		r, err := UnmarshalOne[testResourceWithResolver]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}
		if r.Parent == nil {
			t.Fatal("expected Parent to be resolved")
		}
		if r.Parent.Name != "Parent" {
			t.Errorf("Parent.Name = %q, want %q", r.Parent.Name, "Parent")
		}
	})
}

func TestUnmarshalMany(t *testing.T) {
	t.Run("array of objects", func(t *testing.T) {
		body := `{"data":[
			{"id":"1","type":"test","attributes":{"name":"Alice","age":30}},
			{"id":"2","type":"test","attributes":{"name":"Bob","age":25}}
		]}`
		results, err := UnmarshalMany[testResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalMany() error = %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("len = %d, want 2", len(results))
		}
		if results[0].Name != "Alice" {
			t.Errorf("results[0].Name = %q, want %q", results[0].Name, "Alice")
		}
		if results[1].Name != "Bob" {
			t.Errorf("results[1].Name = %q, want %q", results[1].Name, "Bob")
		}
	})

	t.Run("single object as data", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"test","attributes":{"name":"Solo","age":1}}}`
		results, err := UnmarshalMany[testResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalMany() error = %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("len = %d, want 1", len(results))
		}
		if results[0].Name != "Solo" {
			t.Errorf("Name = %q, want %q", results[0].Name, "Solo")
		}
	})

	t.Run("null data returns empty slice", func(t *testing.T) {
		body := `{"data":null}`
		results, err := UnmarshalMany[testResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalMany() error = %v", err)
		}
		if len(results) != 0 {
			t.Errorf("len = %d, want 0", len(results))
		}
	})

	t.Run("with included", func(t *testing.T) {
		body := `{
			"data":[{"id":"1","type":"children","attributes":{"name":"Child"},
				"relationships":{"parent":{"data":{"type":"tests","id":"99"}}}}],
			"included":[{"id":"99","type":"tests","attributes":{"name":"Parent","age":50}}]
		}`
		results, err := UnmarshalMany[testResourceWithResolver]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalMany() error = %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("len = %d, want 1", len(results))
		}
		if results[0].Parent == nil {
			t.Fatal("expected Parent to be resolved")
		}
		if results[0].Parent.Name != "Parent" {
			t.Errorf("Parent.Name = %q, want %q", results[0].Parent.Name, "Parent")
		}
	})
}

func TestToOneRelationship(t *testing.T) {
	ref := RelationshipRef{Type: "users", ID: "1"}
	result := ToOneRelationship(ref)
	data, ok := result["data"]
	if !ok {
		t.Fatal("expected 'data' key")
	}
	got := data.(RelationshipRef)
	if got.Type != "users" || got.ID != "1" {
		t.Errorf("got %+v, want {Type:users ID:1}", got)
	}
}

func TestToManyRelationship(t *testing.T) {
	refs := []RelationshipRef{
		{Type: "tags", ID: "1"},
		{Type: "tags", ID: "2"},
	}
	result := ToManyRelationship(refs)
	data, ok := result["data"]
	if !ok {
		t.Fatal("expected 'data' key")
	}
	got := data.([]RelationshipRef)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestParseToOneRelationship(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{"data":{"type":"users","id":"5"}}`)
		ref, err := ParseToOneRelationship(raw)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if ref == nil || ref.Type != "users" || ref.ID != "5" {
			t.Errorf("got %+v, want {Type:users ID:5}", ref)
		}
	})

	t.Run("null data", func(t *testing.T) {
		raw := json.RawMessage(`{"data":null}`)
		ref, err := ParseToOneRelationship(raw)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if ref != nil {
			t.Errorf("expected nil, got %+v", ref)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{invalid`)
		_, err := ParseToOneRelationship(raw)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestParseToManyRelationship(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{"data":[{"type":"tags","id":"1"},{"type":"tags","id":"2"}]}`)
		refs, err := ParseToManyRelationship(raw)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(refs) != 2 {
			t.Fatalf("len = %d, want 2", len(refs))
		}
		if refs[0].ID != "1" || refs[1].ID != "2" {
			t.Errorf("got %+v", refs)
		}
	})

	t.Run("null data", func(t *testing.T) {
		raw := json.RawMessage(`{"data":null}`)
		refs, err := ParseToManyRelationship(raw)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if refs != nil {
			t.Errorf("expected nil, got %+v", refs)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{bad`)
		_, err := ParseToManyRelationship(raw)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestResolveToOne(t *testing.T) {
	included := IncludedResources{
		"tests:99": json.RawMessage(`{"id":"99","type":"tests","attributes":{"name":"Found","age":10}}`),
	}
	rels := map[string]json.RawMessage{
		"parent": json.RawMessage(`{"data":{"type":"tests","id":"99"}}`),
	}

	t.Run("present and included", func(t *testing.T) {
		r, err := ResolveToOne[testResource](included, rels, "parent")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if r == nil {
			t.Fatal("expected non-nil result")
		}
		if r.Name != "Found" {
			t.Errorf("Name = %q, want %q", r.Name, "Found")
		}
	})

	t.Run("present but missing from included", func(t *testing.T) {
		missingRels := map[string]json.RawMessage{
			"parent": json.RawMessage(`{"data":{"type":"tests","id":"999"}}`),
		}
		r, err := ResolveToOne[testResource](included, missingRels, "parent")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if r != nil {
			t.Errorf("expected nil, got %+v", r)
		}
	})

	t.Run("missing relationship", func(t *testing.T) {
		r, err := ResolveToOne[testResource](included, rels, "nonexistent")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if r != nil {
			t.Errorf("expected nil, got %+v", r)
		}
	})

	t.Run("error on bad JSON", func(t *testing.T) {
		badRels := map[string]json.RawMessage{
			"parent": json.RawMessage(`{bad`),
		}
		_, err := ResolveToOne[testResource](included, badRels, "parent")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestResolveToMany(t *testing.T) {
	included := IncludedResources{
		"tags:1": json.RawMessage(`{"id":"1","type":"tags","attributes":{"name":"A","age":1}}`),
		"tags:2": json.RawMessage(`{"id":"2","type":"tags","attributes":{"name":"B","age":2}}`),
	}
	rels := map[string]json.RawMessage{
		"tags": json.RawMessage(`{"data":[{"type":"tags","id":"1"},{"type":"tags","id":"2"}]}`),
	}

	t.Run("present and included", func(t *testing.T) {
		results, err := ResolveToMany[testResource](included, rels, "tags")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("len = %d, want 2", len(results))
		}
		if results[0].Name != "A" || results[1].Name != "B" {
			t.Errorf("got %+v, %+v", results[0], results[1])
		}
	})

	t.Run("present but missing from included", func(t *testing.T) {
		missingRels := map[string]json.RawMessage{
			"tags": json.RawMessage(`{"data":[{"type":"tags","id":"999"}]}`),
		}
		results, err := ResolveToMany[testResource](included, missingRels, "tags")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected empty, got %d", len(results))
		}
	})

	t.Run("missing relationship", func(t *testing.T) {
		results, err := ResolveToMany[testResource](included, rels, "nonexistent")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if results != nil {
			t.Errorf("expected nil, got %+v", results)
		}
	})

	t.Run("error on bad JSON", func(t *testing.T) {
		badRels := map[string]json.RawMessage{
			"tags": json.RawMessage(`{bad`),
		}
		_, err := ResolveToMany[testResource](included, badRels, "tags")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestReadOnlyKeys(t *testing.T) {
	t.Run("mixed fields", func(t *testing.T) {
		r := testResourceWithReadonly{}
		keys := readOnlyKeys(r)
		if len(keys) != 2 {
			t.Fatalf("len = %d, want 2", len(keys))
		}
		expected := map[string]bool{"created_at": true, "status": true}
		for _, k := range keys {
			if !expected[k] {
				t.Errorf("unexpected key %q", k)
			}
		}
	})

	t.Run("no readonly fields", func(t *testing.T) {
		r := testResource{}
		keys := readOnlyKeys(r)
		if len(keys) != 0 {
			t.Errorf("expected empty, got %v", keys)
		}
	})

	t.Run("json dash excluded", func(t *testing.T) {
		type dashField struct {
			Hidden string `json:"-" api:"readonly"`
		}
		keys := readOnlyKeys(dashField{})
		if len(keys) != 0 {
			t.Errorf("expected empty for json:\"-\", got %v", keys)
		}
	})
}

func TestJsonTagName(t *testing.T) {
	tests := []struct {
		tag  string
		want string
	}{
		{"name,omitempty", "name"},
		{"name", "name"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := jsonTagName(tt.tag); got != tt.want {
			t.Errorf("jsonTagName(%q) = %q, want %q", tt.tag, got, tt.want)
		}
	}
}

// --- tag-based test helper types ---

type testResourceWithRelTags struct {
	ID       string        `json:"-"`
	Name     string        `json:"name"`
	ParentID string        `json:"-" rel:"parent,parents"`
	Parent   *testResource `json:"-" rel:"parent"`
}

type testResourceWithManyRelTags struct {
	ID     string          `json:"-"`
	Name   string          `json:"name"`
	TagIDs []string        `json:"-" rel:"tags,tag_items"`
	Tags   []*testResource `json:"-" rel:"tags"`
}

type testResourceWithNestedRelTags struct {
	ID       string                   `json:"-"`
	Name     string                   `json:"name"`
	ParentID string                   `json:"-" rel:"parent,parents"`
	Parent   *testResourceWithRelTags `json:"-" rel:"parent"`
}

// --- tag-based tests ---

func TestSplitRelTag(t *testing.T) {
	tests := []struct {
		tag     string
		name    string
		apiType string
		hasSep  bool
	}{
		{"country,countries", "country", "countries", true},
		{"country", "country", "", false},
		{"", "", "", false},
		{"a,b", "a", "b", true},
	}
	for _, tt := range tests {
		name, apiType, hasSep := splitRelTag(tt.tag)
		if name != tt.name || apiType != tt.apiType || hasSep != tt.hasSep {
			t.Errorf("splitRelTag(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.tag, name, apiType, hasSep, tt.name, tt.apiType, tt.hasSep)
		}
	}
}

func TestMarshalRelsFromTags(t *testing.T) {
	t.Run("to-one relationship", func(t *testing.T) {
		r := testResourceWithRelTags{Name: "Child", ParentID: "42"}
		rels := marshalRelsFromTags(r)
		if len(rels) != 1 {
			t.Fatalf("expected 1 rel, got %d", len(rels))
		}
		parentRel, ok := rels["parent"]
		if !ok {
			t.Fatal("expected 'parent' relationship")
		}
		m := parentRel.(map[string]any)
		ref := m["data"].(RelationshipRef)
		if ref.Type != "parents" || ref.ID != "42" {
			t.Errorf("got ref %+v, want {Type:parents ID:42}", ref)
		}
	})

	t.Run("to-many relationship", func(t *testing.T) {
		r := testResourceWithManyRelTags{Name: "Item", TagIDs: []string{"1", "2"}}
		rels := marshalRelsFromTags(r)
		if len(rels) != 1 {
			t.Fatalf("expected 1 rel, got %d", len(rels))
		}
		tagsRel, ok := rels["tags"]
		if !ok {
			t.Fatal("expected 'tags' relationship")
		}
		m := tagsRel.(map[string]any)
		refs := m["data"].([]RelationshipRef)
		if len(refs) != 2 {
			t.Fatalf("expected 2 refs, got %d", len(refs))
		}
		if refs[0].Type != "tag_items" || refs[0].ID != "1" {
			t.Errorf("refs[0] = %+v, want {Type:tag_items ID:1}", refs[0])
		}
	})

	t.Run("empty ID skipped", func(t *testing.T) {
		r := testResourceWithRelTags{Name: "Child", ParentID: ""}
		rels := marshalRelsFromTags(r)
		if len(rels) != 0 {
			t.Errorf("expected empty map, got %d entries", len(rels))
		}
	})

	t.Run("empty slice skipped", func(t *testing.T) {
		r := testResourceWithManyRelTags{Name: "Item", TagIDs: nil}
		rels := marshalRelsFromTags(r)
		if len(rels) != 0 {
			t.Errorf("expected empty map, got %d entries", len(rels))
		}
	})

	t.Run("no tags returns nil", func(t *testing.T) {
		r := testResource{Name: "Plain"}
		rels := marshalRelsFromTags(r)
		if rels != nil {
			t.Errorf("expected nil, got %v", rels)
		}
	})

	t.Run("pointer receiver", func(t *testing.T) {
		r := &testResourceWithRelTags{Name: "Child", ParentID: "99"}
		rels := marshalRelsFromTags(r)
		if len(rels) != 1 {
			t.Fatalf("expected 1 rel, got %d", len(rels))
		}
	})
}

func TestResolveRelsFromTags(t *testing.T) {
	t.Run("to-one resolve", func(t *testing.T) {
		included := IncludedResources{
			"parents:42": json.RawMessage(`{"id":"42","type":"parents","attributes":{"name":"Dad","age":40}}`),
		}
		rels := map[string]json.RawMessage{
			"parent": json.RawMessage(`{"data":{"type":"parents","id":"42"}}`),
		}
		r := &testResourceWithRelTags{Name: "Child"}
		err := resolveRelsFromTags(r, included, rels)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if r.Parent == nil {
			t.Fatal("expected Parent to be resolved")
		}
		if r.Parent.Name != "Dad" {
			t.Errorf("Parent.Name = %q, want %q", r.Parent.Name, "Dad")
		}
		if r.Parent.ID != "42" {
			t.Errorf("Parent.ID = %q, want %q", r.Parent.ID, "42")
		}
	})

	t.Run("to-many resolve", func(t *testing.T) {
		included := IncludedResources{
			"tag_items:1": json.RawMessage(`{"id":"1","type":"tag_items","attributes":{"name":"A","age":1}}`),
			"tag_items:2": json.RawMessage(`{"id":"2","type":"tag_items","attributes":{"name":"B","age":2}}`),
		}
		rels := map[string]json.RawMessage{
			"tags": json.RawMessage(`{"data":[{"type":"tag_items","id":"1"},{"type":"tag_items","id":"2"}]}`),
		}
		r := &testResourceWithManyRelTags{Name: "Item"}
		err := resolveRelsFromTags(r, included, rels)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(r.Tags) != 2 {
			t.Fatalf("expected 2 tags, got %d", len(r.Tags))
		}
		if r.Tags[0].Name != "A" || r.Tags[1].Name != "B" {
			t.Errorf("Tags = %+v, %+v", r.Tags[0], r.Tags[1])
		}
	})

	t.Run("missing from included", func(t *testing.T) {
		included := IncludedResources{}
		rels := map[string]json.RawMessage{
			"parent": json.RawMessage(`{"data":{"type":"parents","id":"999"}}`),
		}
		r := &testResourceWithRelTags{Name: "Child"}
		err := resolveRelsFromTags(r, included, rels)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if r.Parent != nil {
			t.Errorf("expected nil Parent, got %+v", r.Parent)
		}
	})

	t.Run("missing rel name", func(t *testing.T) {
		included := IncludedResources{
			"parents:42": json.RawMessage(`{"id":"42","type":"parents","attributes":{"name":"Dad","age":40}}`),
		}
		rels := map[string]json.RawMessage{} // no "parent" key
		r := &testResourceWithRelTags{Name: "Child"}
		err := resolveRelsFromTags(r, included, rels)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if r.Parent != nil {
			t.Errorf("expected nil Parent, got %+v", r.Parent)
		}
	})

	t.Run("nested rel tags", func(t *testing.T) {
		included := IncludedResources{
			"parents:10": json.RawMessage(`{"id":"10","type":"parents","attributes":{"name":"Middle"},
				"relationships":{"parent":{"data":{"type":"parents","id":"20"}}}}`),
			"parents:20": json.RawMessage(`{"id":"20","type":"parents","attributes":{"name":"Grandparent","age":70}}`),
		}
		rels := map[string]json.RawMessage{
			"parent": json.RawMessage(`{"data":{"type":"parents","id":"10"}}`),
		}
		r := &testResourceWithNestedRelTags{Name: "Child"}
		err := resolveRelsFromTags(r, included, rels)
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if r.Parent == nil {
			t.Fatal("expected Parent to be resolved")
		}
		if r.Parent.Name != "Middle" {
			t.Errorf("Parent.Name = %q, want %q", r.Parent.Name, "Middle")
		}
		if r.Parent.Parent == nil {
			t.Fatal("expected nested Parent.Parent to be resolved")
		}
		if r.Parent.Parent.Name != "Grandparent" {
			t.Errorf("Parent.Parent.Name = %q, want %q", r.Parent.Parent.Name, "Grandparent")
		}
	})
}

func TestMarshalResourceWithTags(t *testing.T) {
	t.Run("to-one tag", func(t *testing.T) {
		r := testResourceWithRelTags{Name: "Child", ParentID: "42"}
		data, err := MarshalResource(r, "children")
		if err != nil {
			t.Fatalf("MarshalResource() error = %v", err)
		}
		var doc struct {
			Data struct {
				Relationships map[string]struct {
					Data RelationshipRef `json:"data"`
				} `json:"relationships"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		parent, ok := doc.Data.Relationships["parent"]
		if !ok {
			t.Fatal("expected 'parent' relationship")
		}
		if parent.Data.Type != "parents" || parent.Data.ID != "42" {
			t.Errorf("parent ref = %+v, want {Type:parents ID:42}", parent.Data)
		}
	})

	t.Run("to-many tag", func(t *testing.T) {
		r := testResourceWithManyRelTags{Name: "Item", TagIDs: []string{"1", "2"}}
		data, err := MarshalResource(r, "items")
		if err != nil {
			t.Fatalf("MarshalResource() error = %v", err)
		}
		var doc struct {
			Data struct {
				Relationships map[string]struct {
					Data []RelationshipRef `json:"data"`
				} `json:"relationships"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		tags, ok := doc.Data.Relationships["tags"]
		if !ok {
			t.Fatal("expected 'tags' relationship")
		}
		if len(tags.Data) != 2 {
			t.Fatalf("expected 2 tag refs, got %d", len(tags.Data))
		}
		if tags.Data[0].Type != "tag_items" || tags.Data[0].ID != "1" {
			t.Errorf("tags[0] = %+v, want {Type:tag_items ID:1}", tags.Data[0])
		}
	})

	t.Run("no rels when IDs empty", func(t *testing.T) {
		r := testResourceWithRelTags{Name: "Orphan"}
		data, err := MarshalResource(r, "children")
		if err != nil {
			t.Fatalf("MarshalResource() error = %v", err)
		}
		var doc struct {
			Data struct {
				Relationships json.RawMessage `json:"relationships"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if len(doc.Data.Relationships) > 0 && string(doc.Data.Relationships) != "null" {
			t.Errorf("expected no relationships, got %s", doc.Data.Relationships)
		}
	})
}

func TestUnmarshalOneWithTags(t *testing.T) {
	t.Run("to-one tag resolve", func(t *testing.T) {
		body := `{
			"data":{"id":"1","type":"children","attributes":{"name":"Child"},
				"relationships":{"parent":{"data":{"type":"parents","id":"42"}}}},
			"included":[{"id":"42","type":"parents","attributes":{"name":"Dad","age":40}}]
		}`
		r, err := UnmarshalOne[testResourceWithRelTags]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}
		if r.ID != "1" {
			t.Errorf("ID = %q, want %q", r.ID, "1")
		}
		if r.Name != "Child" {
			t.Errorf("Name = %q, want %q", r.Name, "Child")
		}
		if r.Parent == nil {
			t.Fatal("expected Parent to be resolved")
		}
		if r.Parent.Name != "Dad" {
			t.Errorf("Parent.Name = %q, want %q", r.Parent.Name, "Dad")
		}
		if r.Parent.ID != "42" {
			t.Errorf("Parent.ID = %q, want %q", r.Parent.ID, "42")
		}
	})

	t.Run("to-many tag resolve", func(t *testing.T) {
		body := `{
			"data":{"id":"1","type":"items","attributes":{"name":"Item"},
				"relationships":{"tags":{"data":[
					{"type":"tag_items","id":"1"},
					{"type":"tag_items","id":"2"}
				]}}},
			"included":[
				{"id":"1","type":"tag_items","attributes":{"name":"A","age":1}},
				{"id":"2","type":"tag_items","attributes":{"name":"B","age":2}}
			]
		}`
		r, err := UnmarshalOne[testResourceWithManyRelTags]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}
		if len(r.Tags) != 2 {
			t.Fatalf("expected 2 tags, got %d", len(r.Tags))
		}
		if r.Tags[0].Name != "A" || r.Tags[1].Name != "B" {
			t.Errorf("Tags = %+v, %+v", r.Tags[0], r.Tags[1])
		}
	})

	t.Run("nested tag resolve", func(t *testing.T) {
		body := `{
			"data":{"id":"1","type":"grandchildren","attributes":{"name":"GrandChild"},
				"relationships":{"parent":{"data":{"type":"parents","id":"10"}}}},
			"included":[
				{"id":"10","type":"parents","attributes":{"name":"Middle"},
					"relationships":{"parent":{"data":{"type":"parents","id":"20"}}}},
				{"id":"20","type":"parents","attributes":{"name":"Grandparent","age":70}}
			]
		}`
		r, err := UnmarshalOne[testResourceWithNestedRelTags]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}
		if r.Parent == nil {
			t.Fatal("expected Parent to be resolved")
		}
		if r.Parent.Name != "Middle" {
			t.Errorf("Parent.Name = %q, want %q", r.Parent.Name, "Middle")
		}
		if r.Parent.Parent == nil {
			t.Fatal("expected nested Parent.Parent to be resolved")
		}
		if r.Parent.Parent.Name != "Grandparent" {
			t.Errorf("Parent.Parent.Name = %q, want %q", r.Parent.Parent.Name, "Grandparent")
		}
	})

	t.Run("without included stays nil", func(t *testing.T) {
		body := `{
			"data":{"id":"1","type":"children","attributes":{"name":"Child"},
				"relationships":{"parent":{"data":{"type":"parents","id":"42"}}}}
		}`
		r, err := UnmarshalOne[testResourceWithRelTags]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}
		if r.Parent != nil {
			t.Errorf("expected nil Parent without included, got %+v", r.Parent)
		}
	})
}

func TestUnmarshalManyWithTags(t *testing.T) {
	body := `{
		"data":[
			{"id":"1","type":"children","attributes":{"name":"A"},
				"relationships":{"parent":{"data":{"type":"parents","id":"42"}}}},
			{"id":"2","type":"children","attributes":{"name":"B"},
				"relationships":{"parent":{"data":{"type":"parents","id":"42"}}}}
		],
		"included":[{"id":"42","type":"parents","attributes":{"name":"Shared","age":40}}]
	}`
	results, err := UnmarshalMany[testResourceWithRelTags]([]byte(body))
	if err != nil {
		t.Fatalf("UnmarshalMany() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}
	for i, r := range results {
		if r.Parent == nil {
			t.Fatalf("results[%d].Parent is nil", i)
		}
		if r.Parent.Name != "Shared" {
			t.Errorf("results[%d].Parent.Name = %q, want %q", i, r.Parent.Name, "Shared")
		}
	}
}

// --- ResourceType and Marshal tests ---

func TestResourceType(t *testing.T) {
	t.Run("reads jsonapi tag", func(t *testing.T) {
		got := ResourceType[testResource]()
		if got != "test_resources" {
			t.Errorf("ResourceType[testResource]() = %q, want %q", got, "test_resources")
		}
	})

	t.Run("panics on missing tag", func(t *testing.T) {
		type noTag struct {
			ID   string `json:"-"`
			Name string `json:"name"`
		}
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic for missing jsonapi tag")
			}
		}()
		ResourceType[noTag]()
	})

	t.Run("panics on no ID field", func(t *testing.T) {
		type noID struct {
			Name string `json:"name"`
		}
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic for struct without ID field")
			}
		}()
		ResourceType[noID]()
	})
}

func TestMarshalGeneric(t *testing.T) {
	t.Run("derives type from tag", func(t *testing.T) {
		r := &testResource{Name: "Alice", Age: 30}
		data, err := Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		var doc struct {
			Data struct {
				Type string `json:"type"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if doc.Data.Type != "test_resources" {
			t.Errorf("type = %q, want %q", doc.Data.Type, "test_resources")
		}
	})

	t.Run("includes ID when set", func(t *testing.T) {
		r := &testResource{ID: "42", Name: "Bob", Age: 25}
		data, err := Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		var doc struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if doc.Data.ID != "42" {
			t.Errorf("id = %q, want %q", doc.Data.ID, "42")
		}
		if doc.Data.Type != "test_resources" {
			t.Errorf("type = %q, want %q", doc.Data.Type, "test_resources")
		}
	})
}

func TestResourceTypeFromTag(t *testing.T) {
	t.Run("finds tag on struct value", func(t *testing.T) {
		r := testResource{}
		got := resourceTypeFromTag(r)
		if got != "test_resources" {
			t.Errorf("resourceTypeFromTag() = %q, want %q", got, "test_resources")
		}
	})

	t.Run("finds tag on pointer", func(t *testing.T) {
		r := &testResource{}
		got := resourceTypeFromTag(r)
		if got != "test_resources" {
			t.Errorf("resourceTypeFromTag() = %q, want %q", got, "test_resources")
		}
	})

	t.Run("returns empty for non-struct", func(t *testing.T) {
		got := resourceTypeFromTag("hello")
		if got != "" {
			t.Errorf("resourceTypeFromTag() = %q, want empty", got)
		}
	})

	t.Run("returns empty for missing tag", func(t *testing.T) {
		type noTag struct {
			ID   string `json:"-"`
			Name string `json:"name"`
		}
		got := resourceTypeFromTag(noTag{})
		if got != "" {
			t.Errorf("resourceTypeFromTag() = %q, want empty", got)
		}
	})
}

func TestMarshalResourceAutoDetectsType(t *testing.T) {
	r := testResource{Name: "Alice", Age: 30}
	data, err := MarshalResource(r, "")
	if err != nil {
		t.Fatalf("MarshalResource() error = %v", err)
	}
	var doc struct {
		Data struct {
			Type string `json:"type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if doc.Data.Type != "test_resources" {
		t.Errorf("type = %q, want %q", doc.Data.Type, "test_resources")
	}
}

// --- dirty test helper types ---

type testDirtyResource struct {
	ID          string  `json:"-" jsonapi:"dirty_resources"`
	Name        string  `json:"name"`
	Age         int     `json:"age"`
	Description *string `json:"description"`
}

type testDirtyWithReadonly struct {
	ID        string `json:"-" jsonapi:"dirty_ro_resources"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at" api:"readonly"`
}

type testDirtyWithRel struct {
	ID       string `json:"-" jsonapi:"dirty_rel_resources"`
	Name     string `json:"name"`
	ParentID string `json:"-" rel:"parent,parents"`
}

type testMutualExclusive struct {
	ID      string `json:"-" jsonapi:"mutex_resources"`
	Name    string `json:"name"`
	TrunkID string `json:"-" rel:"trunk,trunks"`
	GroupID string `json:"-" rel:"group,groups"`
}

func (m *testMutualExclusive) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if m.TrunkID != "" {
		rels["group"] = NullRelationship()
	}
	if m.GroupID != "" {
		rels["trunk"] = NullRelationship()
	}
	return rels, nil
}

// --- MarshalPatch tests ---

func TestMarshalPatch(t *testing.T) {
	t.Run("new resource sends only set attribute", func(t *testing.T) {
		r := &testDirtyResource{ID: "1", Name: "Alice"}
		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		if doc.ID != "1" {
			t.Errorf("id = %q, want %q", doc.ID, "1")
		}
		if doc.Type != "dirty_resources" {
			t.Errorf("type = %q, want %q", doc.Type, "dirty_resources")
		}
		// Name changed from "" to "Alice" → dirty
		assertAttrEquals(t, doc.Attrs, "name", `"Alice"`)
		// Age stayed 0 → not dirty
		assertAttrMissing(t, doc.Attrs, "age")
	})

	t.Run("loaded resource only sends changed field", func(t *testing.T) {
		// Simulate loading from API
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30,"description":"hello"}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}

		// Change only age
		r.Age = 31

		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		// Only age changed
		assertAttrEquals(t, doc.Attrs, "age", `31`)
		// Name and description unchanged
		assertAttrMissing(t, doc.Attrs, "name")
		assertAttrMissing(t, doc.Attrs, "description")
	})

	t.Run("set attribute to null sends explicit null", func(t *testing.T) {
		// Simulate loading from API with non-nil description
		desc := "hello"
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30,"description":"hello"}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}
		if r.Description == nil || *r.Description != desc {
			t.Fatalf("expected Description %q, got %v", desc, r.Description)
		}

		// Clear description
		r.Description = nil

		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "description", `null`)
		assertAttrMissing(t, doc.Attrs, "name")
		assertAttrMissing(t, doc.Attrs, "age")
	})

	t.Run("no changes produces empty attributes", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}

		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		if len(doc.Attrs) != 0 {
			t.Errorf("expected empty attributes, got %v", doc.Attrs)
		}
	})

	t.Run("readonly fields excluded from dirty tracking", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"dirty_ro_resources","attributes":{"name":"Alice","created_at":"2024-01-01"}}}`
		r, err := UnmarshalOne[testDirtyWithReadonly]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}

		r.Name = "Bob"
		r.CreatedAt = "2025-01-01" // readonly, should be ignored

		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Bob"`)
		assertAttrMissing(t, doc.Attrs, "created_at")
	})

	t.Run("relationship dirty on new resource", func(t *testing.T) {
		r := &testDirtyWithRel{ID: "1", Name: "Child", ParentID: "99"}
		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Child"`)
		if doc.Rels == nil {
			t.Fatal("expected relationships in patch")
		}
		parentRaw, ok := doc.Rels["parent"]
		if !ok {
			t.Fatal("expected 'parent' relationship")
		}
		ref, err := ParseToOneRelationship(parentRaw)
		if err != nil {
			t.Fatalf("ParseToOneRelationship error = %v", err)
		}
		if ref.Type != "parents" || ref.ID != "99" {
			t.Errorf("parent ref = %+v, want {Type:parents ID:99}", ref)
		}
	})

	t.Run("mutual exclusion trunk set nullifies group", func(t *testing.T) {
		r := &testMutualExclusive{ID: "1", TrunkID: "t1"}
		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		if doc.Rels == nil {
			t.Fatal("expected relationships")
		}
		// trunk should be set
		trunkRaw, ok := doc.Rels["trunk"]
		if !ok {
			t.Fatal("expected 'trunk' relationship")
		}
		ref, _ := ParseToOneRelationship(trunkRaw)
		if ref == nil || ref.ID != "t1" {
			t.Errorf("trunk ref = %+v, want {ID:t1}", ref)
		}
		// group should be explicit null
		groupRaw, ok := doc.Rels["group"]
		if !ok {
			t.Fatal("expected 'group' relationship (null clear)")
		}
		groupRef, _ := ParseToOneRelationship(groupRaw)
		if groupRef != nil {
			t.Errorf("expected null group, got %+v", groupRef)
		}
	})

	t.Run("mutual exclusion group set nullifies trunk", func(t *testing.T) {
		r := &testMutualExclusive{ID: "1", GroupID: "g1"}
		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		if doc.Rels == nil {
			t.Fatal("expected relationships")
		}
		// group should be set
		groupRaw, ok := doc.Rels["group"]
		if !ok {
			t.Fatal("expected 'group' relationship")
		}
		ref, _ := ParseToOneRelationship(groupRaw)
		if ref == nil || ref.ID != "g1" {
			t.Errorf("group ref = %+v, want {ID:g1}", ref)
		}
		// trunk should be explicit null
		trunkRaw, ok := doc.Rels["trunk"]
		if !ok {
			t.Fatal("expected 'trunk' relationship (null clear)")
		}
		trunkRef, _ := ParseToOneRelationship(trunkRaw)
		if trunkRef != nil {
			t.Errorf("expected null trunk, got %+v", trunkRef)
		}
	})

	t.Run("multiple dirty attributes all included", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30,"description":"hello"}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}

		r.Name = "Bob"
		r.Age = 31

		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Bob"`)
		assertAttrEquals(t, doc.Attrs, "age", `31`)
		assertAttrMissing(t, doc.Attrs, "description")
	})

	t.Run("build with fields marks those attrs dirty", func(t *testing.T) {
		desc := "new"
		r := &testDirtyResource{ID: "1", Name: "Alice", Age: 25, Description: &desc}
		data, err := MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Alice"`)
		assertAttrEquals(t, doc.Attrs, "age", `25`)
		assertAttrEquals(t, doc.Attrs, "description", `"new"`)
	})

	t.Run("ForgetCleanState makes resource fully dirty", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		if err != nil {
			t.Fatalf("UnmarshalOne() error = %v", err)
		}

		// Without changes, patch is empty
		data, _ := MarshalPatch(r)
		doc := parsePatchDoc(t, data)
		if len(doc.Attrs) != 0 {
			t.Errorf("expected empty attrs before forget, got %v", doc.Attrs)
		}

		// After forgetting clean state, all non-zero attrs become dirty
		ForgetCleanState(r)
		data, err = MarshalPatch(r)
		if err != nil {
			t.Fatalf("MarshalPatch() error = %v", err)
		}
		doc = parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Alice"`)
		assertAttrEquals(t, doc.Attrs, "age", `30`)
	})
}

// --- MarshalPatch test helpers ---

type patchDoc struct {
	ID    string
	Type  string
	Attrs map[string]json.RawMessage
	Rels  map[string]json.RawMessage
}

func parsePatchDoc(t *testing.T, data []byte) patchDoc {
	t.Helper()
	var doc struct {
		Data struct {
			ID            string                     `json:"id"`
			Type          string                     `json:"type"`
			Attributes    map[string]json.RawMessage `json:"attributes"`
			Relationships map[string]json.RawMessage `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("failed to parse patch doc: %v", err)
	}
	return patchDoc{
		ID:    doc.Data.ID,
		Type:  doc.Data.Type,
		Attrs: doc.Data.Attributes,
		Rels:  doc.Data.Relationships,
	}
}

func assertAttrEquals(t *testing.T, attrs map[string]json.RawMessage, key, want string) {
	t.Helper()
	raw, ok := attrs[key]
	if !ok {
		t.Errorf("expected attribute %q to be present", key)
		return
	}
	if string(raw) != want {
		t.Errorf("attribute %q = %s, want %s", key, raw, want)
	}
}

func assertAttrMissing(t *testing.T, attrs map[string]json.RawMessage, key string) {
	t.Helper()
	if raw, ok := attrs[key]; ok {
		t.Errorf("expected attribute %q to be absent, got %s", key, raw)
	}
}
