package jsonapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Equal(t, "abc", GetID(r))
	})

	t.Run("pointer to struct", func(t *testing.T) {
		r := &testResource{ID: "xyz"}
		assert.Equal(t, "xyz", GetID(r))
	})

	t.Run("no ID field", func(t *testing.T) {
		type noID struct {
			Name string
		}
		assert.Equal(t, "", GetID(noID{Name: "test"}))
	})

	t.Run("non-struct", func(t *testing.T) {
		assert.Equal(t, "", GetID("hello"))
	})
}

func TestSetID(t *testing.T) {
	t.Run("sets ID on pointer", func(t *testing.T) {
		r := &testResource{}
		setID(r, "123")
		assert.Equal(t, "123", r.ID)
	})

	t.Run("no-op without ID field", func(t *testing.T) {
		type noID struct {
			Name string
		}
		r := &noID{Name: "test"}
		setID(r, "123") // should not panic
		assert.Equal(t, "test", r.Name)
	})
}

func TestMarshalWritableAttrs(t *testing.T) {
	t.Run("no readonly fields", func(t *testing.T) {
		r := testResource{Name: "Alice", Age: 30}
		data, err := MarshalWritableAttrs(r)
		require.NoError(t, err)
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &m))
		assert.Contains(t, m, "name")
		assert.Contains(t, m, "age")
	})

	t.Run("with readonly fields", func(t *testing.T) {
		r := testResourceWithReadonly{Name: "Bob", CreatedAt: "2024-01-01", Status: "active"}
		data, err := MarshalWritableAttrs(r)
		require.NoError(t, err)
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &m))
		assert.Contains(t, m, "name")
		assert.NotContains(t, m, "created_at")
		assert.NotContains(t, m, "status")
	})
}

func TestMarshalResource(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		r := testResource{Name: "Alice", Age: 30}
		data, err := MarshalResource(r, "test_resources")
		require.NoError(t, err)
		var doc map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &doc))
		var d map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(doc["data"], &d))
		var typ string
		json.Unmarshal(d["type"], &typ)
		assert.Equal(t, "test_resources", typ)
		assert.NotContains(t, d, "id")
	})

	t.Run("with ID", func(t *testing.T) {
		r := testResource{ID: "42", Name: "Bob"}
		data, err := MarshalResource(r, "test_resources")
		require.NoError(t, err)
		var doc struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(data, &doc))
		assert.Equal(t, "42", doc.Data.ID)
	})

	t.Run("with RelationshipMarshaler", func(t *testing.T) {
		r := &testResourceWithRels{Name: "Child", ParentID: "99"}
		data, err := MarshalResource(r, "children")
		require.NoError(t, err)
		var doc struct {
			Data struct {
				Relationships map[string]json.RawMessage `json:"relationships"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(data, &doc))
		assert.Contains(t, doc.Data.Relationships, "parent")
	})

	t.Run("readonly exclusion", func(t *testing.T) {
		r := testResourceWithReadonly{Name: "Test", CreatedAt: "2024-01-01"}
		data, err := MarshalResource(r, "test_resources")
		require.NoError(t, err)
		var doc struct {
			Data struct {
				Attributes json.RawMessage `json:"attributes"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(data, &doc))
		var attrs map[string]json.RawMessage
		json.Unmarshal(doc.Data.Attributes, &attrs)
		assert.NotContains(t, attrs, "created_at")
	})
}

func TestUnmarshalOne(t *testing.T) {
	t.Run("single object", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"test","attributes":{"name":"Alice","age":30}}}`
		r, err := UnmarshalOne[testResource]([]byte(body))
		require.NoError(t, err)
		assert.Equal(t, "1", r.ID)
		assert.Equal(t, "Alice", r.Name)
		assert.Equal(t, 30, r.Age)
	})

	t.Run("array-wrapped singleton", func(t *testing.T) {
		body := `{"data":[{"id":"2","type":"test","attributes":{"name":"Bob","age":25}}]}`
		r, err := UnmarshalOne[testResource]([]byte(body))
		require.NoError(t, err)
		assert.Equal(t, "2", r.ID)
		assert.Equal(t, "Bob", r.Name)
	})

	t.Run("null data error", func(t *testing.T) {
		body := `{"data":null}`
		_, err := UnmarshalOne[testResource]([]byte(body))
		require.Error(t, err)
	})

	t.Run("with included and resolver", func(t *testing.T) {
		body := `{
			"data":{"id":"1","type":"children","attributes":{"name":"Child"},
				"relationships":{"parent":{"data":{"type":"tests","id":"99"}}}},
			"included":[{"id":"99","type":"tests","attributes":{"name":"Parent","age":50}}]
		}`
		r, err := UnmarshalOne[testResourceWithResolver]([]byte(body))
		require.NoError(t, err)
		require.NotNil(t, r.Parent)
		assert.Equal(t, "Parent", r.Parent.Name)
	})
}

func TestUnmarshalMany(t *testing.T) {
	t.Run("array of objects", func(t *testing.T) {
		body := `{"data":[
			{"id":"1","type":"test","attributes":{"name":"Alice","age":30}},
			{"id":"2","type":"test","attributes":{"name":"Bob","age":25}}
		]}`
		results, err := UnmarshalMany[testResource]([]byte(body))
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, "Alice", results[0].Name)
		assert.Equal(t, "Bob", results[1].Name)
	})

	t.Run("single object as data", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"test","attributes":{"name":"Solo","age":1}}}`
		results, err := UnmarshalMany[testResource]([]byte(body))
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "Solo", results[0].Name)
	})

	t.Run("null data returns empty slice", func(t *testing.T) {
		body := `{"data":null}`
		results, err := UnmarshalMany[testResource]([]byte(body))
		require.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("with included", func(t *testing.T) {
		body := `{
			"data":[{"id":"1","type":"children","attributes":{"name":"Child"},
				"relationships":{"parent":{"data":{"type":"tests","id":"99"}}}}],
			"included":[{"id":"99","type":"tests","attributes":{"name":"Parent","age":50}}]
		}`
		results, err := UnmarshalMany[testResourceWithResolver]([]byte(body))
		require.NoError(t, err)
		require.Len(t, results, 1)
		require.NotNil(t, results[0].Parent)
		assert.Equal(t, "Parent", results[0].Parent.Name)
	})
}

func TestToOneRelationship(t *testing.T) {
	ref := RelationshipRef{Type: "users", ID: "1"}
	result := ToOneRelationship(ref)
	require.Contains(t, result, "data")
	got := result["data"].(RelationshipRef)
	assert.Equal(t, "users", got.Type)
	assert.Equal(t, "1", got.ID)
}

func TestToManyRelationship(t *testing.T) {
	refs := []RelationshipRef{
		{Type: "tags", ID: "1"},
		{Type: "tags", ID: "2"},
	}
	result := ToManyRelationship(refs)
	require.Contains(t, result, "data")
	got := result["data"].([]RelationshipRef)
	require.Len(t, got, 2)
}

func TestParseToOneRelationship(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{"data":{"type":"users","id":"5"}}`)
		ref, err := ParseToOneRelationship(raw)
		require.NoError(t, err)
		require.NotNil(t, ref)
		assert.Equal(t, "users", ref.Type)
		assert.Equal(t, "5", ref.ID)
	})

	t.Run("null data", func(t *testing.T) {
		raw := json.RawMessage(`{"data":null}`)
		ref, err := ParseToOneRelationship(raw)
		require.NoError(t, err)
		assert.Nil(t, ref)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{invalid`)
		_, err := ParseToOneRelationship(raw)
		require.Error(t, err)
	})
}

func TestParseToManyRelationship(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{"data":[{"type":"tags","id":"1"},{"type":"tags","id":"2"}]}`)
		refs, err := ParseToManyRelationship(raw)
		require.NoError(t, err)
		require.Len(t, refs, 2)
		assert.Equal(t, "1", refs[0].ID)
		assert.Equal(t, "2", refs[1].ID)
	})

	t.Run("null data", func(t *testing.T) {
		raw := json.RawMessage(`{"data":null}`)
		refs, err := ParseToManyRelationship(raw)
		require.NoError(t, err)
		assert.Nil(t, refs)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{bad`)
		_, err := ParseToManyRelationship(raw)
		require.Error(t, err)
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
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "Found", r.Name)
	})

	t.Run("present but missing from included", func(t *testing.T) {
		missingRels := map[string]json.RawMessage{
			"parent": json.RawMessage(`{"data":{"type":"tests","id":"999"}}`),
		}
		r, err := ResolveToOne[testResource](included, missingRels, "parent")
		require.NoError(t, err)
		assert.Nil(t, r)
	})

	t.Run("missing relationship", func(t *testing.T) {
		r, err := ResolveToOne[testResource](included, rels, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, r)
	})

	t.Run("error on bad JSON", func(t *testing.T) {
		badRels := map[string]json.RawMessage{
			"parent": json.RawMessage(`{bad`),
		}
		_, err := ResolveToOne[testResource](included, badRels, "parent")
		require.Error(t, err)
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
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.Equal(t, "A", results[0].Name)
		assert.Equal(t, "B", results[1].Name)
	})

	t.Run("present but missing from included", func(t *testing.T) {
		missingRels := map[string]json.RawMessage{
			"tags": json.RawMessage(`{"data":[{"type":"tags","id":"999"}]}`),
		}
		results, err := ResolveToMany[testResource](included, missingRels, "tags")
		require.NoError(t, err)
		assert.Equal(t, 0, len(results))
	})

	t.Run("missing relationship", func(t *testing.T) {
		results, err := ResolveToMany[testResource](included, rels, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, results)
	})

	t.Run("error on bad JSON", func(t *testing.T) {
		badRels := map[string]json.RawMessage{
			"tags": json.RawMessage(`{bad`),
		}
		_, err := ResolveToMany[testResource](included, badRels, "tags")
		require.Error(t, err)
	})
}

func TestReadOnlyKeys(t *testing.T) {
	t.Run("mixed fields", func(t *testing.T) {
		r := testResourceWithReadonly{}
		keys := readOnlyKeys(r)
		require.Len(t, keys, 2)
		expected := map[string]bool{"created_at": true, "status": true}
		for _, k := range keys {
			assert.True(t, expected[k], "unexpected key %q", k)
		}
	})

	t.Run("no readonly fields", func(t *testing.T) {
		r := testResource{}
		keys := readOnlyKeys(r)
		assert.Len(t, keys, 0)
	})

	t.Run("json dash excluded", func(t *testing.T) {
		type dashField struct {
			Hidden string `json:"-" api:"readonly"`
		}
		keys := readOnlyKeys(dashField{})
		assert.Len(t, keys, 0)
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
		assert.Equal(t, tt.want, jsonTagName(tt.tag), "jsonTagName(%q)", tt.tag)
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
		assert.Equal(t, tt.name, name, "splitRelTag(%q) name", tt.tag)
		assert.Equal(t, tt.apiType, apiType, "splitRelTag(%q) apiType", tt.tag)
		assert.Equal(t, tt.hasSep, hasSep, "splitRelTag(%q) hasSep", tt.tag)
	}
}

func TestMarshalRelsFromTags(t *testing.T) {
	t.Run("to-one relationship", func(t *testing.T) {
		r := testResourceWithRelTags{Name: "Child", ParentID: "42"}
		rels := marshalRelsFromTags(r)
		require.Len(t, rels, 1)
		require.Contains(t, rels, "parent")
		parentRel := rels["parent"]
		m := parentRel.(map[string]any)
		ref := m["data"].(RelationshipRef)
		assert.Equal(t, "parents", ref.Type)
		assert.Equal(t, "42", ref.ID)
	})

	t.Run("to-many relationship", func(t *testing.T) {
		r := testResourceWithManyRelTags{Name: "Item", TagIDs: []string{"1", "2"}}
		rels := marshalRelsFromTags(r)
		require.Len(t, rels, 1)
		require.Contains(t, rels, "tags")
		tagsRel := rels["tags"]
		m := tagsRel.(map[string]any)
		refs := m["data"].([]RelationshipRef)
		require.Len(t, refs, 2)
		assert.Equal(t, "tag_items", refs[0].Type)
		assert.Equal(t, "1", refs[0].ID)
	})

	t.Run("empty ID skipped", func(t *testing.T) {
		r := testResourceWithRelTags{Name: "Child", ParentID: ""}
		rels := marshalRelsFromTags(r)
		assert.Len(t, rels, 0)
	})

	t.Run("empty slice skipped", func(t *testing.T) {
		r := testResourceWithManyRelTags{Name: "Item", TagIDs: nil}
		rels := marshalRelsFromTags(r)
		assert.Len(t, rels, 0)
	})

	t.Run("no tags returns nil", func(t *testing.T) {
		r := testResource{Name: "Plain"}
		rels := marshalRelsFromTags(r)
		assert.Nil(t, rels)
	})

	t.Run("pointer receiver", func(t *testing.T) {
		r := &testResourceWithRelTags{Name: "Child", ParentID: "99"}
		rels := marshalRelsFromTags(r)
		require.Len(t, rels, 1)
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
		require.NoError(t, err)
		require.NotNil(t, r.Parent)
		assert.Equal(t, "Dad", r.Parent.Name)
		assert.Equal(t, "42", r.Parent.ID)
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
		require.NoError(t, err)
		require.Len(t, r.Tags, 2)
		assert.Equal(t, "A", r.Tags[0].Name)
		assert.Equal(t, "B", r.Tags[1].Name)
	})

	t.Run("missing from included", func(t *testing.T) {
		included := IncludedResources{}
		rels := map[string]json.RawMessage{
			"parent": json.RawMessage(`{"data":{"type":"parents","id":"999"}}`),
		}
		r := &testResourceWithRelTags{Name: "Child"}
		err := resolveRelsFromTags(r, included, rels)
		require.NoError(t, err)
		assert.Nil(t, r.Parent)
	})

	t.Run("missing rel name", func(t *testing.T) {
		included := IncludedResources{
			"parents:42": json.RawMessage(`{"id":"42","type":"parents","attributes":{"name":"Dad","age":40}}`),
		}
		rels := map[string]json.RawMessage{} // no "parent" key
		r := &testResourceWithRelTags{Name: "Child"}
		err := resolveRelsFromTags(r, included, rels)
		require.NoError(t, err)
		assert.Nil(t, r.Parent)
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
		require.NoError(t, err)
		require.NotNil(t, r.Parent)
		assert.Equal(t, "Middle", r.Parent.Name)
		require.NotNil(t, r.Parent.Parent)
		assert.Equal(t, "Grandparent", r.Parent.Parent.Name)
	})
}

func TestMarshalResourceWithTags(t *testing.T) {
	t.Run("to-one tag", func(t *testing.T) {
		r := testResourceWithRelTags{Name: "Child", ParentID: "42"}
		data, err := MarshalResource(r, "children")
		require.NoError(t, err)
		var doc struct {
			Data struct {
				Relationships map[string]struct {
					Data RelationshipRef `json:"data"`
				} `json:"relationships"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(data, &doc))
		parent, ok := doc.Data.Relationships["parent"]
		require.True(t, ok, "expected 'parent' relationship")
		assert.Equal(t, "parents", parent.Data.Type)
		assert.Equal(t, "42", parent.Data.ID)
	})

	t.Run("to-many tag", func(t *testing.T) {
		r := testResourceWithManyRelTags{Name: "Item", TagIDs: []string{"1", "2"}}
		data, err := MarshalResource(r, "items")
		require.NoError(t, err)
		var doc struct {
			Data struct {
				Relationships map[string]struct {
					Data []RelationshipRef `json:"data"`
				} `json:"relationships"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(data, &doc))
		tags, ok := doc.Data.Relationships["tags"]
		require.True(t, ok, "expected 'tags' relationship")
		require.Len(t, tags.Data, 2)
		assert.Equal(t, "tag_items", tags.Data[0].Type)
		assert.Equal(t, "1", tags.Data[0].ID)
	})

	t.Run("no rels when IDs empty", func(t *testing.T) {
		r := testResourceWithRelTags{Name: "Orphan"}
		data, err := MarshalResource(r, "children")
		require.NoError(t, err)
		var doc struct {
			Data struct {
				Relationships json.RawMessage `json:"relationships"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(data, &doc))
		assert.True(t, len(doc.Data.Relationships) == 0 || string(doc.Data.Relationships) == "null",
			"expected no relationships, got %s", doc.Data.Relationships)
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
		require.NoError(t, err)
		assert.Equal(t, "1", r.ID)
		assert.Equal(t, "Child", r.Name)
		require.NotNil(t, r.Parent)
		assert.Equal(t, "Dad", r.Parent.Name)
		assert.Equal(t, "42", r.Parent.ID)
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
		require.NoError(t, err)
		require.Len(t, r.Tags, 2)
		assert.Equal(t, "A", r.Tags[0].Name)
		assert.Equal(t, "B", r.Tags[1].Name)
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
		require.NoError(t, err)
		require.NotNil(t, r.Parent)
		assert.Equal(t, "Middle", r.Parent.Name)
		require.NotNil(t, r.Parent.Parent)
		assert.Equal(t, "Grandparent", r.Parent.Parent.Name)
	})

	t.Run("without included stays nil", func(t *testing.T) {
		body := `{
			"data":{"id":"1","type":"children","attributes":{"name":"Child"},
				"relationships":{"parent":{"data":{"type":"parents","id":"42"}}}}
		}`
		r, err := UnmarshalOne[testResourceWithRelTags]([]byte(body))
		require.NoError(t, err)
		assert.Nil(t, r.Parent)
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
	require.NoError(t, err)
	require.Len(t, results, 2)
	for i, r := range results {
		require.NotNil(t, r.Parent, "results[%d].Parent is nil", i)
		assert.Equal(t, "Shared", r.Parent.Name, "results[%d].Parent.Name", i)
	}
}

// --- ResourceType and Marshal tests ---

func TestResourceType(t *testing.T) {
	t.Run("reads jsonapi tag", func(t *testing.T) {
		got := ResourceType[testResource]()
		assert.Equal(t, "test_resources", got)
	})

	t.Run("panics on missing tag", func(t *testing.T) {
		type noTag struct {
			ID   string `json:"-"`
			Name string `json:"name"`
		}
		defer func() {
			r := recover()
			require.NotNil(t, r)
		}()
		ResourceType[noTag]()
	})

	t.Run("panics on no ID field", func(t *testing.T) {
		type noID struct {
			Name string `json:"name"`
		}
		defer func() {
			r := recover()
			require.NotNil(t, r)
		}()
		ResourceType[noID]()
	})
}

func TestMarshalGeneric(t *testing.T) {
	t.Run("derives type from tag", func(t *testing.T) {
		r := &testResource{Name: "Alice", Age: 30}
		data, err := Marshal(r)
		require.NoError(t, err)
		var doc struct {
			Data struct {
				Type string `json:"type"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(data, &doc))
		assert.Equal(t, "test_resources", doc.Data.Type)
	})

	t.Run("includes ID when set", func(t *testing.T) {
		r := &testResource{ID: "42", Name: "Bob", Age: 25}
		data, err := Marshal(r)
		require.NoError(t, err)
		var doc struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(data, &doc))
		assert.Equal(t, "42", doc.Data.ID)
		assert.Equal(t, "test_resources", doc.Data.Type)
	})
}

func TestResourceTypeFromTag(t *testing.T) {
	t.Run("finds tag on struct value", func(t *testing.T) {
		r := testResource{}
		assert.Equal(t, "test_resources", resourceTypeFromTag(r))
	})

	t.Run("finds tag on pointer", func(t *testing.T) {
		r := &testResource{}
		assert.Equal(t, "test_resources", resourceTypeFromTag(r))
	})

	t.Run("returns empty for non-struct", func(t *testing.T) {
		assert.Equal(t, "", resourceTypeFromTag("hello"))
	})

	t.Run("returns empty for missing tag", func(t *testing.T) {
		type noTag struct {
			ID   string `json:"-"`
			Name string `json:"name"`
		}
		assert.Equal(t, "", resourceTypeFromTag(noTag{}))
	})
}

func TestMarshalResourceAutoDetectsType(t *testing.T) {
	r := testResource{Name: "Alice", Age: 30}
	data, err := MarshalResource(r, "")
	require.NoError(t, err)
	var doc struct {
		Data struct {
			Type string `json:"type"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(data, &doc))
	assert.Equal(t, "test_resources", doc.Data.Type)
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
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		assert.Equal(t, "1", doc.ID)
		assert.Equal(t, "dirty_resources", doc.Type)
		// Name changed from "" to "Alice" -> dirty
		assertAttrEquals(t, doc.Attrs, "name", `"Alice"`)
		// Age stayed 0 -> not dirty
		assertAttrMissing(t, doc.Attrs, "age")
	})

	t.Run("loaded resource only sends changed field", func(t *testing.T) {
		// Simulate loading from API
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30,"description":"hello"}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		require.NoError(t, err)

		// Change only age
		r.Age = 31

		data, err := MarshalPatch(r)
		require.NoError(t, err)

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
		require.NoError(t, err)
		require.NotNil(t, r.Description)
		assert.Equal(t, desc, *r.Description)

		// Clear description
		r.Description = nil

		data, err := MarshalPatch(r)
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "description", `null`)
		assertAttrMissing(t, doc.Attrs, "name")
		assertAttrMissing(t, doc.Attrs, "age")
	})

	t.Run("no changes produces empty attributes", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		require.NoError(t, err)

		data, err := MarshalPatch(r)
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		assert.Len(t, doc.Attrs, 0)
	})

	t.Run("readonly fields excluded from dirty tracking", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"dirty_ro_resources","attributes":{"name":"Alice","created_at":"2024-01-01"}}}`
		r, err := UnmarshalOne[testDirtyWithReadonly]([]byte(body))
		require.NoError(t, err)

		r.Name = "Bob"
		r.CreatedAt = "2025-01-01" // readonly, should be ignored

		data, err := MarshalPatch(r)
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Bob"`)
		assertAttrMissing(t, doc.Attrs, "created_at")
	})

	t.Run("relationship dirty on new resource", func(t *testing.T) {
		r := &testDirtyWithRel{ID: "1", Name: "Child", ParentID: "99"}
		data, err := MarshalPatch(r)
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Child"`)
		require.NotNil(t, doc.Rels)
		require.Contains(t, doc.Rels, "parent")
		parentRaw := doc.Rels["parent"]
		ref, err := ParseToOneRelationship(parentRaw)
		require.NoError(t, err)
		assert.Equal(t, "parents", ref.Type)
		assert.Equal(t, "99", ref.ID)
	})

	t.Run("mutual exclusion trunk set nullifies group", func(t *testing.T) {
		r := &testMutualExclusive{ID: "1", TrunkID: "t1"}
		data, err := MarshalPatch(r)
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		require.NotNil(t, doc.Rels)
		// trunk should be set
		require.Contains(t, doc.Rels, "trunk")
		trunkRaw := doc.Rels["trunk"]
		ref, _ := ParseToOneRelationship(trunkRaw)
		require.NotNil(t, ref)
		assert.Equal(t, "t1", ref.ID)
		// group should be explicit null
		require.Contains(t, doc.Rels, "group")
		groupRaw := doc.Rels["group"]
		groupRef, _ := ParseToOneRelationship(groupRaw)
		assert.Nil(t, groupRef)
	})

	t.Run("mutual exclusion group set nullifies trunk", func(t *testing.T) {
		r := &testMutualExclusive{ID: "1", GroupID: "g1"}
		data, err := MarshalPatch(r)
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		require.NotNil(t, doc.Rels)
		// group should be set
		require.Contains(t, doc.Rels, "group")
		groupRaw := doc.Rels["group"]
		ref, _ := ParseToOneRelationship(groupRaw)
		require.NotNil(t, ref)
		assert.Equal(t, "g1", ref.ID)
		// trunk should be explicit null
		require.Contains(t, doc.Rels, "trunk")
		trunkRaw := doc.Rels["trunk"]
		trunkRef, _ := ParseToOneRelationship(trunkRaw)
		assert.Nil(t, trunkRef)
	})

	t.Run("multiple dirty attributes all included", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30,"description":"hello"}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		require.NoError(t, err)

		r.Name = "Bob"
		r.Age = 31

		data, err := MarshalPatch(r)
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Bob"`)
		assertAttrEquals(t, doc.Attrs, "age", `31`)
		assertAttrMissing(t, doc.Attrs, "description")
	})

	t.Run("build with fields marks those attrs dirty", func(t *testing.T) {
		desc := "new"
		r := &testDirtyResource{ID: "1", Name: "Alice", Age: 25, Description: &desc}
		data, err := MarshalPatch(r)
		require.NoError(t, err)

		doc := parsePatchDoc(t, data)
		assertAttrEquals(t, doc.Attrs, "name", `"Alice"`)
		assertAttrEquals(t, doc.Attrs, "age", `25`)
		assertAttrEquals(t, doc.Attrs, "description", `"new"`)
	})

	t.Run("ForgetCleanState makes resource fully dirty", func(t *testing.T) {
		body := `{"data":{"id":"1","type":"dirty_resources","attributes":{"name":"Alice","age":30}}}`
		r, err := UnmarshalOne[testDirtyResource]([]byte(body))
		require.NoError(t, err)

		// Without changes, patch is empty
		data, _ := MarshalPatch(r)
		doc := parsePatchDoc(t, data)
		assert.Len(t, doc.Attrs, 0)

		// After forgetting clean state, all non-zero attrs become dirty
		ForgetCleanState(r)
		data, err = MarshalPatch(r)
		require.NoError(t, err)
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
	require.NoError(t, json.Unmarshal(data, &doc))
	return patchDoc{
		ID:    doc.Data.ID,
		Type:  doc.Data.Type,
		Attrs: doc.Data.Attributes,
		Rels:  doc.Data.Relationships,
	}
}

func assertAttrEquals(t *testing.T, attrs map[string]json.RawMessage, key, want string) {
	t.Helper()
	require.Contains(t, attrs, key)
	assert.Equal(t, want, string(attrs[key]))
}

func assertAttrMissing(t *testing.T, attrs map[string]json.RawMessage, key string) {
	t.Helper()
	assert.NotContains(t, attrs, key)
}
