package didww

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/smstrunkconfiguration"
)

func TestSmsTrunkGroupsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sms_trunk_groups": {status: http.StatusOK, fixture: "sms_trunk_groups/index.json"},
	})

	groups, err := client.SmsTrunkGroups().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, groups, 2)

	assert.Equal(t, testSmsTrunkGroupID, groups[0].ID)
	assert.Equal(t, "Inbound SMS", groups[0].Name)
	assert.Nil(t, groups[0].ExternalReferenceID)
	assert.Equal(t, time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC), groups[0].CreatedAt)
	assert.Equal(t, time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC), groups[0].UpdatedAt)
	assert.Equal(t, "2", groups[0].Meta["sms_trunks_count"])
	assert.Equal(t, "Outbound SMS", groups[1].Name)
	assert.Equal(t, "0", groups[1].Meta["sms_trunks_count"])
}

func TestSmsTrunkGroupsFindWithIncludedSmsTrunks(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sms_trunk_groups/" + testSmsTrunkGroupID: {status: http.StatusOK, fixture: "sms_trunk_groups/show_with_sms_trunks.json"},
	})

	params := NewQueryParams().Include("sms_trunks")
	group, err := client.SmsTrunkGroups().Find(context.Background(), testSmsTrunkGroupID, params)
	require.NoError(t, err)

	require.Len(t, group.SmsTrunks, 2)
	assert.Equal(t, testSmsTrunkHTTPInID, group.SmsTrunks[0].ID)
	assert.IsType(t, &smstrunkconfiguration.HTTPIn{}, group.SmsTrunks[0].Configuration)
	assert.Equal(t, "1a2b3c4d-0005-4a5b-8c6d-7e8f9a0b1c05", group.SmsTrunks[1].ID)
	assert.IsType(t, &smstrunkconfiguration.SmppSmsc{}, group.SmsTrunks[1].Configuration)
	assert.Equal(t, "2", group.Meta["sms_trunks_count"])
}

func TestSmsTrunkGroupsFindNotFound(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sms_trunk_groups/" + testSmsTrunkGroupID: {status: http.StatusNotFound, fixture: "sms_trunk_groups/not_found.json"},
	})

	_, err := client.SmsTrunkGroups().Find(context.Background(), testSmsTrunkGroupID)
	require.Error(t, err)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr), "expected *APIError")
	assert.Equal(t, http.StatusNotFound, apiErr.HTTPStatus)
}

func TestSmsTrunkGroupsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/sms_trunk_groups": {status: http.StatusCreated, fixture: "sms_trunk_groups/create.json"},
	})

	group, err := server.client.SmsTrunkGroups().Create(context.Background(), &resource.SmsTrunkGroup{
		Name:        "Inbound SMS",
		SmsTrunkIDs: []string{testSmsTrunkHTTPInID},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunk_groups/create_request.json")
	assert.Equal(t, testSmsTrunkGroupID, group.ID)
	assert.Equal(t, "Inbound SMS", group.Name)
	assert.Equal(t, "1", group.Meta["sms_trunks_count"])
}

func TestSmsTrunkGroupsUpdateExternalReferenceID(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sms_trunk_groups/" + testSmsTrunkGroupID: {status: http.StatusOK, fixture: "sms_trunk_groups/update_external_reference_id.json"},
	})

	group, err := server.client.SmsTrunkGroups().Update(context.Background(), &resource.SmsTrunkGroup{
		ID:                  testSmsTrunkGroupID,
		ExternalReferenceID: Ptr("updated-ext-ref-99"),
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunk_groups/update_external_reference_id_request.json")
	require.NotNil(t, group.ExternalReferenceID)
	assert.Equal(t, "updated-ext-ref-99", *group.ExternalReferenceID)
}

func TestSmsTrunkGroupsUpdateSmsTrunks(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sms_trunk_groups/" + testSmsTrunkGroupID: {status: http.StatusOK, fixture: "sms_trunk_groups/update_sms_trunks.json"},
	})

	group, err := server.client.SmsTrunkGroups().Update(context.Background(), &resource.SmsTrunkGroup{
		ID:          testSmsTrunkGroupID,
		SmsTrunkIDs: []string{testSmsTrunkHTTPInID, "1a2b3c4d-0005-4a5b-8c6d-7e8f9a0b1c05"},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunk_groups/update_sms_trunks_request.json")
	assert.Equal(t, "2", group.Meta["sms_trunks_count"])
}

func TestSmsTrunkGroupsUpdateClearSmsTrunks(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sms_trunk_groups/" + testSmsTrunkGroupID: {status: http.StatusOK, fixture: "sms_trunk_groups/update_sms_trunks.json"},
	})

	_, err := server.client.SmsTrunkGroups().Update(context.Background(), &resource.SmsTrunkGroup{
		ID:             testSmsTrunkGroupID,
		ClearSmsTrunks: true,
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunk_groups/update_clear_sms_trunks_request.json")
}

func TestSmsTrunkGroupsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/sms_trunk_groups/" + testSmsTrunkGroupID: {status: http.StatusNoContent},
	})

	err := client.SmsTrunkGroups().Delete(context.Background(), testSmsTrunkGroupID)
	require.NoError(t, err)
}

func TestSmsTrunkGroupsDeleteInUse(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/sms_trunk_groups/" + testSmsTrunkGroupID: {status: http.StatusUnprocessableEntity, fixture: "sms_trunk_groups/delete_error.json"},
	})

	err := client.SmsTrunkGroups().Delete(context.Background(), testSmsTrunkGroupID)
	require.Error(t, err)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr), "expected *APIError")
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.HTTPStatus)
}
