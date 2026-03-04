package didww

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const testDIDID = "9df99644-f1a5-4a3c-99a4-559d758eb96b"

// TestDirtyPatch_NewResourceOnlyDirtyFields verifies that creating a DID
// with only an ID and one attribute sends just that attribute.
func TestDirtyPatch_NewResourceOnlyDirtyFields(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	desc := "updated"
	_, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:          testDIDID,
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc := parsePatchBody(t, capturedBody)

	// Only description should be in attributes
	if len(doc.Attrs) != 1 {
		t.Errorf("expected 1 attribute, got %d: %v", len(doc.Attrs), attrKeys(doc.Attrs))
	}
	assertAttr(t, doc.Attrs, "description", `"updated"`)

	// No relationships should be present (no trunk IDs set)
	if len(doc.Rels) != 0 {
		t.Errorf("expected no relationships, got %v", doc.Rels)
	}
}

// TestDirtyPatch_NullAttributeClear verifies that setting a non-nil field
// to nil on a loaded resource produces explicit null in the PATCH.
func TestDirtyPatch_NullAttributeClear(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/dids/" + testDIDID:   {status: http.StatusOK, fixture: "dids/show.json"},
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		if r.Method == http.MethodPatch {
			capturedBody, _ = io.ReadAll(r.Body)
		}
	})

	// Load from API - description is "something", capacity_limit is 2
	did, err := server.client.DIDs().Find(context.Background(), testDIDID)
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}
	if did.Description == nil || *did.Description != "something" {
		t.Fatalf("expected Description 'something', got %v", did.Description)
	}

	// Clear description to nil
	did.Description = nil

	_, err = server.client.DIDs().Update(context.Background(), did)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	doc := parsePatchBody(t, capturedBody)

	// description should be explicit null
	assertAttr(t, doc.Attrs, "description", "null")
	// capacity_limit and dedicated_channels_count should NOT be present (unchanged)
	assertAttrAbsent(t, doc.Attrs, "capacity_limit")
	assertAttrAbsent(t, doc.Attrs, "dedicated_channels_count")
}

// TestDirtyPatch_LoadedResourceOnlyChangedField verifies that after loading
// from the API, setting one field sends only that field.
func TestDirtyPatch_LoadedResourceOnlyChangedField(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/dids/" + testDIDID:   {status: http.StatusOK, fixture: "dids/show.json"},
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		if r.Method == http.MethodPatch {
			capturedBody, _ = io.ReadAll(r.Body)
		}
	})

	// Load from API
	did, err := server.client.DIDs().Find(context.Background(), testDIDID)
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}

	// Change only dedicated_channels_count
	did.DedicatedChannelsCount = 5

	_, err = server.client.DIDs().Update(context.Background(), did)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	doc := parsePatchBody(t, capturedBody)

	// Only the changed field
	if len(doc.Attrs) != 1 {
		t.Errorf("expected 1 attribute, got %d: %v", len(doc.Attrs), attrKeys(doc.Attrs))
	}
	assertAttr(t, doc.Attrs, "dedicated_channels_count", "5")
}

// TestDirtyPatch_SetVoiceInTrunkNullifiesTrunkGroup verifies mutual
// exclusion: setting voice_in_trunk sends explicit null for voice_in_trunk_group.
func TestDirtyPatch_SetVoiceInTrunkNullifiesTrunkGroup(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show_with_trunk.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:             testDIDID,
		VoiceInTrunkID: "trunk-1",
	})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	doc := parsePatchBody(t, capturedBody)

	// Attributes should be empty (no attr changes)
	if len(doc.Attrs) != 0 {
		t.Errorf("expected empty attributes, got %v", attrKeys(doc.Attrs))
	}

	// voice_in_trunk should be set
	assertRelSet(t, doc.Rels, "voice_in_trunk", "voice_in_trunks", "trunk-1")
	// voice_in_trunk_group should be null
	assertRelNull(t, doc.Rels, "voice_in_trunk_group")
}

// TestDirtyPatch_SetVoiceInTrunkGroupNullifiesTrunk verifies mutual
// exclusion: setting voice_in_trunk_group sends explicit null for voice_in_trunk.
func TestDirtyPatch_SetVoiceInTrunkGroupNullifiesTrunk(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show_with_trunk_group.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:                  testDIDID,
		VoiceInTrunkGroupID: "group-1",
	})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	doc := parsePatchBody(t, capturedBody)

	// Attributes should be empty (no attr changes)
	if len(doc.Attrs) != 0 {
		t.Errorf("expected empty attributes, got %v", attrKeys(doc.Attrs))
	}

	// voice_in_trunk_group should be set
	assertRelSet(t, doc.Rels, "voice_in_trunk_group", "voice_in_trunk_groups", "group-1")
	// voice_in_trunk should be null
	assertRelNull(t, doc.Rels, "voice_in_trunk")
}

// TestDirtyPatch_CreateUnchanged verifies that Create still sends all fields.
func TestDirtyPatch_CreateUnchanged(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/dids": {status: http.StatusCreated, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	desc := "new"
	cl := 5
	_, err := server.client.DIDs().Create(context.Background(), &DID{
		CapacityLimit:          &cl,
		Description:            &desc,
		DedicatedChannelsCount: 3,
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// Create should use Marshal (full), not MarshalPatch
	var doc struct {
		Data struct {
			Attrs map[string]json.RawMessage `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(capturedBody, &doc); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	// All writable attributes should be present
	if _, ok := doc.Data.Attrs["capacity_limit"]; !ok {
		t.Error("Create should include capacity_limit")
	}
	if _, ok := doc.Data.Attrs["description"]; !ok {
		t.Error("Create should include description")
	}
	if _, ok := doc.Data.Attrs["dedicated_channels_count"]; !ok {
		t.Error("Create should include dedicated_channels_count")
	}
}

// TestDirtyPatch_ResponseClearsState verifies that after an update,
// the returned resource has a clean baseline.
func TestDirtyPatch_ResponseClearsState(t *testing.T) {
	var callCount int
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		callCount++
		if callCount == 2 {
			capturedBody, _ = io.ReadAll(r.Body)
		}
	})

	desc := "first"
	did, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:          testDIDID,
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("first update error: %v", err)
	}

	// The returned DID should have a clean baseline.
	// Updating it without changes should produce empty attributes.
	did2, err := server.client.DIDs().Update(context.Background(), did)
	if err != nil {
		t.Fatalf("second update error: %v", err)
	}
	_ = did2

	doc := parsePatchBody(t, capturedBody)
	if len(doc.Attrs) != 0 {
		t.Errorf("expected empty attributes on second update, got %v", attrKeys(doc.Attrs))
	}
}

// TestDirtyPatch_UpdateBuiltSingleAttr verifies that building a DID with ID
// and setting only capacity_limit sends just that attribute.
func TestDirtyPatch_UpdateBuiltSingleAttr(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	cl := 10
	_, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:            testDIDID,
		CapacityLimit: &cl,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "dids/update_built_single_attr_request.json")
}

// TestDirtyPatch_UpdateClearDescription verifies that building a DID with ID
// and setting Description=nil sends explicit null.
func TestDirtyPatch_UpdateClearDescription(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/dids/" + testDIDID:   {status: http.StatusOK, fixture: "dids/show.json"},
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		if r.Method == http.MethodPatch {
			capturedBody, _ = io.ReadAll(r.Body)
		}
	})

	did, err := server.client.DIDs().Find(context.Background(), testDIDID)
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}

	did.Description = nil

	_, err = server.client.DIDs().Update(context.Background(), did)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "dids/update_clear_description_request.json")
}

// TestDirtyPatch_UpdateTerminated verifies that building a DID with ID
// and setting Terminated=true sends only terminated in the PATCH.
func TestDirtyPatch_UpdateTerminated(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.DIDs().Update(context.Background(), &DID{
		ID:         testDIDID,
		Terminated: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "dids/update_terminated_request.json")
}

// TestDirtyPatch_UpdateFromLoadedSetVoiceInTrunk verifies that after loading
// from the API, setting VoiceInTrunkID sends only the relationship change.
func TestDirtyPatch_UpdateFromLoadedSetVoiceInTrunk(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/dids/" + testDIDID:   {status: http.StatusOK, fixture: "dids/show.json"},
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show_with_trunk.json"},
	}, func(r *http.Request) {
		if r.Method == http.MethodPatch {
			capturedBody, _ = io.ReadAll(r.Body)
		}
	})

	did, err := server.client.DIDs().Find(context.Background(), testDIDID)
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}

	did.VoiceInTrunkID = "41b94706-325e-4704-a433-d65105758836"

	_, err = server.client.DIDs().Update(context.Background(), did)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "dids/update_from_loaded_set_voice_in_trunk_request.json")
}

// TestDirtyPatch_FindWithIncludedHasNoDirtyFlags verifies that loading a DID
// with included voice_in_trunk produces no dirty fields on re-update.
func TestDirtyPatch_FindWithIncludedHasNoDirtyFlags(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/dids/" + testDIDID:   {status: http.StatusOK, fixture: "dids/show_with_trunk.json"},
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show_with_trunk.json"},
	}, func(r *http.Request) {
		if r.Method == http.MethodPatch {
			capturedBody, _ = io.ReadAll(r.Body)
		}
	})

	params := NewQueryParams().Include("voice_in_trunk")
	did, err := server.client.DIDs().Find(context.Background(), testDIDID, params)
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}
	if did.VoiceInTrunk == nil {
		t.Fatal("expected non-nil VoiceInTrunk after include")
	}

	_, err = server.client.DIDs().Update(context.Background(), did)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	doc := parsePatchBody(t, capturedBody)
	if len(doc.Attrs) != 0 {
		t.Errorf("expected empty attributes, got %v", attrKeys(doc.Attrs))
	}
	if len(doc.Rels) != 0 {
		t.Errorf("expected no relationships, got %v", doc.Rels)
	}
}

// TestDirtyPatch_UpdateFailRetainsCleanState verifies that a failed update
// preserves the clean state so a retry sends the correct (not over-inclusive) PATCH.
func TestDirtyPatch_UpdateFailRetainsCleanState(t *testing.T) {
	var patchCount int
	var capturedBody []byte
	ts := newTestServerWithDynamicPatch(t, map[string]testRoute{
		"GET /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		if r.Method == http.MethodPatch {
			patchCount++
			capturedBody, _ = io.ReadAll(r.Body)
		}
	}, func(patchCall int) testRoute {
		if patchCall == 1 {
			return testRoute{status: http.StatusUnprocessableEntity, fixture: "dids/update_error_invalid_trunk_group.json"}
		}
		return testRoute{status: http.StatusOK, fixture: "dids/show.json"}
	})

	did, err := ts.client.DIDs().Find(context.Background(), testDIDID)
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}

	did.DedicatedChannelsCount = 5

	// First update fails (422)
	_, err = ts.client.DIDs().Update(context.Background(), did)
	if err == nil {
		t.Fatal("expected error for first update")
	}

	// Retry — should still send only the changed field, not become over-inclusive
	_, err = ts.client.DIDs().Update(context.Background(), did)
	if err != nil {
		t.Fatalf("retry update error: %v", err)
	}

	if patchCount != 2 {
		t.Fatalf("expected 2 PATCH calls, got %d", patchCount)
	}
	doc := parsePatchBody(t, capturedBody)
	if len(doc.Attrs) != 1 {
		t.Errorf("expected 1 attribute on retry, got %d: %v", len(doc.Attrs), attrKeys(doc.Attrs))
	}
	assertAttr(t, doc.Attrs, "dedicated_channels_count", "5")
}

// TestDirtyPatch_UpdateFromLoadedChangedDescription verifies that after loading
// from the API, setting Description sends only that field, validated against fixture.
func TestDirtyPatch_UpdateFromLoadedChangedDescription(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/dids/" + testDIDID:   {status: http.StatusOK, fixture: "dids/show.json"},
		"PATCH /v3/dids/" + testDIDID: {status: http.StatusOK, fixture: "dids/show.json"},
	}, func(r *http.Request) {
		if r.Method == http.MethodPatch {
			capturedBody, _ = io.ReadAll(r.Body)
		}
	})

	did, err := server.client.DIDs().Find(context.Background(), testDIDID)
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}

	desc := "patched from loaded resource"
	did.Description = &desc

	_, err = server.client.DIDs().Update(context.Background(), did)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "dids/update_from_loaded_request.json")
}

// --- test helpers ---

type patchBodyDoc struct {
	ID    string
	Type  string
	Attrs map[string]json.RawMessage
	Rels  map[string]json.RawMessage
}

func parsePatchBody(t *testing.T, body []byte) patchBodyDoc {
	t.Helper()
	var doc struct {
		Data struct {
			ID            string                     `json:"id"`
			Type          string                     `json:"type"`
			Attributes    map[string]json.RawMessage `json:"attributes"`
			Relationships map[string]json.RawMessage `json:"relationships"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("failed to parse patch body: %v\nbody: %s", err, body)
	}
	return patchBodyDoc{
		ID:    doc.Data.ID,
		Type:  doc.Data.Type,
		Attrs: doc.Data.Attributes,
		Rels:  doc.Data.Relationships,
	}
}

func assertAttr(t *testing.T, attrs map[string]json.RawMessage, key, want string) {
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

func assertAttrAbsent(t *testing.T, attrs map[string]json.RawMessage, key string) {
	t.Helper()
	if raw, ok := attrs[key]; ok {
		t.Errorf("expected attribute %q to be absent, got %s", key, raw)
	}
}

func assertRelSet(t *testing.T, rels map[string]json.RawMessage, name, wantType, wantID string) {
	t.Helper()
	raw, ok := rels[name]
	if !ok {
		t.Fatalf("expected relationship %q to be present", name)
	}
	var wrapper struct {
		Data *struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		t.Fatalf("failed to parse relationship %q: %v", name, err)
	}
	if wrapper.Data == nil {
		t.Fatalf("expected relationship %q data to be non-null", name)
	}
	if wrapper.Data.Type != wantType || wrapper.Data.ID != wantID {
		t.Errorf("relationship %q = {%s, %s}, want {%s, %s}", name, wrapper.Data.Type, wrapper.Data.ID, wantType, wantID)
	}
}

func assertRelNull(t *testing.T, rels map[string]json.RawMessage, name string) {
	t.Helper()
	raw, ok := rels[name]
	if !ok {
		t.Fatalf("expected relationship %q to be present (as null)", name)
	}
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		t.Fatalf("failed to parse relationship %q: %v", name, err)
	}
	if string(wrapper.Data) != "null" {
		t.Errorf("expected relationship %q data to be null, got %s", name, wrapper.Data)
	}
}

func attrKeys(attrs map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	return keys
}
