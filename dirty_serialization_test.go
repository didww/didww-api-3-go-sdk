package didww

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	doc := parsePatchBody(t, capturedBody)

	// Only description should be in attributes
	require.Len(t, doc.Attrs, 1)
	assertAttr(t, doc.Attrs, "description", `"updated"`)

	// No relationships should be present (no trunk IDs set)
	assert.Empty(t, doc.Rels)
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
	require.NoError(t, err)
	require.NotNil(t, did.Description)
	require.Equal(t, "something", *did.Description)

	// Clear description to nil
	did.Description = nil

	_, err = server.client.DIDs().Update(context.Background(), did)
	require.NoError(t, err)

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
	require.NoError(t, err)

	// Change only dedicated_channels_count
	did.DedicatedChannelsCount = 5

	_, err = server.client.DIDs().Update(context.Background(), did)
	require.NoError(t, err)

	doc := parsePatchBody(t, capturedBody)

	// Only the changed field
	require.Len(t, doc.Attrs, 1)
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
	require.NoError(t, err)

	doc := parsePatchBody(t, capturedBody)

	// Attributes should be empty (no attr changes)
	assert.Empty(t, doc.Attrs)

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
	require.NoError(t, err)

	doc := parsePatchBody(t, capturedBody)

	// Attributes should be empty (no attr changes)
	assert.Empty(t, doc.Attrs)

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
	require.NoError(t, err)

	// Create should use Marshal (full), not MarshalPatch
	var doc struct {
		Data struct {
			Attrs map[string]json.RawMessage `json:"attributes"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(capturedBody, &doc))

	// All writable attributes should be present
	assert.Contains(t, doc.Data.Attrs, "capacity_limit")
	assert.Contains(t, doc.Data.Attrs, "description")
	assert.Contains(t, doc.Data.Attrs, "dedicated_channels_count")
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
	require.NoError(t, err)

	// The returned DID should have a clean baseline.
	// Updating it without changes should produce empty attributes.
	did2, err := server.client.DIDs().Update(context.Background(), did)
	require.NoError(t, err)
	_ = did2

	doc := parsePatchBody(t, capturedBody)
	assert.Empty(t, doc.Attrs)
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
	require.NoError(t, err)

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
	require.NoError(t, err)

	did.Description = nil

	_, err = server.client.DIDs().Update(context.Background(), did)
	require.NoError(t, err)

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
	require.NoError(t, err)

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
	require.NoError(t, err)

	did.VoiceInTrunkID = "41b94706-325e-4704-a433-d65105758836"

	_, err = server.client.DIDs().Update(context.Background(), did)
	require.NoError(t, err)

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
	require.NoError(t, err)
	require.NotNil(t, did.VoiceInTrunk)

	_, err = server.client.DIDs().Update(context.Background(), did)
	require.NoError(t, err)

	doc := parsePatchBody(t, capturedBody)
	assert.Empty(t, doc.Attrs)
	assert.Empty(t, doc.Rels)
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
	require.NoError(t, err)

	did.DedicatedChannelsCount = 5

	// First update fails (422)
	_, err = ts.client.DIDs().Update(context.Background(), did)
	require.Error(t, err)

	// Retry — should still send only the changed field, not become over-inclusive
	_, err = ts.client.DIDs().Update(context.Background(), did)
	require.NoError(t, err)

	require.Equal(t, 2, patchCount)
	doc := parsePatchBody(t, capturedBody)
	require.Len(t, doc.Attrs, 1)
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
	require.NoError(t, err)

	desc := "patched from loaded resource"
	did.Description = &desc

	_, err = server.client.DIDs().Update(context.Background(), did)
	require.NoError(t, err)

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
	require.NoError(t, json.Unmarshal(body, &doc), "failed to parse patch body: %s", body)
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
	if !assert.True(t, ok, "expected attribute %q to be present", key) {
		return
	}
	assert.Equal(t, want, string(raw), "attribute %q", key)
}

func assertAttrAbsent(t *testing.T, attrs map[string]json.RawMessage, key string) {
	t.Helper()
	assert.NotContains(t, attrs, key, "expected attribute %q to be absent", key)
}

func assertRelSet(t *testing.T, rels map[string]json.RawMessage, name, wantType, wantID string) {
	t.Helper()
	raw, ok := rels[name]
	require.True(t, ok, "expected relationship %q to be present", name)
	var wrapper struct {
		Data *struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &wrapper), "failed to parse relationship %q", name)
	require.NotNil(t, wrapper.Data, "expected relationship %q data to be non-null", name)
	assert.Equal(t, wantType, wrapper.Data.Type, "relationship %q type", name)
	assert.Equal(t, wantID, wrapper.Data.ID, "relationship %q id", name)
}

func assertRelNull(t *testing.T, rels map[string]json.RawMessage, name string) {
	t.Helper()
	raw, ok := rels[name]
	require.True(t, ok, "expected relationship %q to be present (as null)", name)
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &wrapper), "failed to parse relationship %q", name)
	assert.Equal(t, "null", string(wrapper.Data), "expected relationship %q data to be null", name)
}
