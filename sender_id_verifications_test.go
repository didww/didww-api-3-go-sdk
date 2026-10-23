package didww

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/smstrunkconfiguration"
)

const testSenderIDVerificationID = "e5a4e2f1-3b26-4a5c-9a0a-2f7c1f3e8d10"

func TestSenderIDVerificationsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sender_id_verifications": {status: http.StatusOK, fixture: "sender_id_verifications/index.json"},
	})

	verifications, err := client.SenderIDVerifications().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, verifications, 2)

	active := verifications[0]
	assert.Equal(t, testSenderIDVerificationID, active.ID)
	assert.Equal(t, "ACME alerts", active.Name)
	assert.Equal(t, enums.SenderIDVerificationStatusActive, active.Status)
	assert.True(t, active.IsActive())
	assert.Equal(t, enums.SenderTypeLongCode, active.SenderType)
	assert.Equal(t, []string{"12125550100", "12125550101"}, active.SenderIDs)
	assert.Equal(t, "SID-1234", active.Reference)
	assert.Nil(t, active.RejectReasons)
	assert.Nil(t, active.RejectComment)
	require.NotNil(t, active.RenewDate)
	assert.Equal(t, "2026-11-22", *active.RenewDate)
	assert.Equal(t, time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC), active.CreatedAt)
	require.NotNil(t, active.ActivatedAt)
	assert.Equal(t, time.Date(2026, 10, 3, 12, 30, 0, 0, time.UTC), *active.ActivatedAt)
	assert.Nil(t, active.ClosedAt)
	assert.Equal(t, "10.0", active.Meta["setup_price"])
	assert.Equal(t, "15.0", active.Meta["monthly_price"])

	pending := verifications[1]
	assert.True(t, pending.IsChangesRequired())
	assert.False(t, pending.IsActive())
	assert.Equal(t, enums.SenderTypeAlphanumeric, pending.SenderType)
	assert.Equal(t, []string{"Company documents are missing"}, pending.RejectReasons)
	require.NotNil(t, pending.RejectComment)
	assert.Equal(t, "Upload the certificate of incorporation.", *pending.RejectComment)
	assert.Nil(t, pending.RenewDate)
	assert.Nil(t, pending.ActivatedAt)
}

func TestSenderIDVerificationsFindWithIncludedSmsTrunk(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sender_id_verifications/" + testSenderIDVerificationID: {status: http.StatusOK, fixture: "sender_id_verifications/show_with_sms_trunk.json"},
	})

	params := NewQueryParams().Include("sms_trunk")
	siv, err := client.SenderIDVerifications().Find(context.Background(), testSenderIDVerificationID, params)
	require.NoError(t, err)

	require.NotNil(t, siv.SmsTrunk)
	assert.Equal(t, testSmsTrunkHTTPOutID, siv.SmsTrunk.ID)
	assert.Equal(t, "Outbound HTTP", siv.SmsTrunk.Name)
	assert.IsType(t, &smstrunkconfiguration.HTTPOut{}, siv.SmsTrunk.Configuration)
}

func TestSenderIDVerificationsUpdateName(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sender_id_verifications/" + testSenderIDVerificationID: {status: http.StatusOK, fixture: "sender_id_verifications/update_name.json"},
	})

	siv, err := server.client.SenderIDVerifications().Update(context.Background(), &resource.SenderIDVerification{
		ID:   testSenderIDVerificationID,
		Name: "ACME notifications",
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sender_id_verifications/update_name_request.json")
	assert.Equal(t, "ACME notifications", siv.Name)
}

func TestSenderIDVerificationsUpdateFromLoadedSendsOnlyName(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"GET /v3/sender_id_verifications/" + testSenderIDVerificationID:   {status: http.StatusOK, fixture: "sender_id_verifications/show_with_sms_trunk.json"},
		"PATCH /v3/sender_id_verifications/" + testSenderIDVerificationID: {status: http.StatusOK, fixture: "sender_id_verifications/update_name.json"},
	})

	siv, err := server.client.SenderIDVerifications().Find(context.Background(), testSenderIDVerificationID)
	require.NoError(t, err)

	siv.Name = "ACME notifications"
	// Read-only attributes are never sent, even when changed locally.
	siv.Status = enums.SenderIDVerificationStatusClosed
	siv.SenderIDs = []string{"12125550199"}

	_, err = server.client.SenderIDVerifications().Update(context.Background(), siv)
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sender_id_verifications/update_name_request.json")
}

func TestSenderIDVerificationsAssignSmsTrunk(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sender_id_verifications/" + testSenderIDVerificationID: {status: http.StatusOK, fixture: "sender_id_verifications/assign_sms_trunk.json"},
	})

	_, err := server.client.SenderIDVerifications().Update(context.Background(), &resource.SenderIDVerification{
		ID:         testSenderIDVerificationID,
		SmsTrunkID: testSmsTrunkHTTPOutID,
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sender_id_verifications/assign_sms_trunk_request.json")
}

func TestSenderIDVerificationsUnassignSmsTrunk(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sender_id_verifications/" + testSenderIDVerificationID: {status: http.StatusOK, fixture: "sender_id_verifications/unassign_sms_trunk.json"},
	})

	_, err := server.client.SenderIDVerifications().Update(context.Background(), &resource.SenderIDVerification{
		ID:              testSenderIDVerificationID,
		NullifySmsTrunk: true,
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sender_id_verifications/unassign_sms_trunk_request.json")
}

func TestSenderIDVerificationsUpdateClosed(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/sender_id_verifications/" + testSenderIDVerificationID: {status: http.StatusUnprocessableEntity, fixture: "sender_id_verifications/update_closed_error.json"},
	})

	_, err := client.SenderIDVerifications().Update(context.Background(), &resource.SenderIDVerification{
		ID:   testSenderIDVerificationID,
		Name: "ACME notifications",
	})
	require.Error(t, err)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr), "expected *APIError")
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "Sender ID Verification is closed and can not be changed", apiErr.Errors[0].Detail)
	assert.Equal(t, "/data", apiErr.Errors[0].Source.Pointer)
}

func TestSenderIDVerificationStatusPredicates(t *testing.T) {
	tests := []struct {
		status                                                              enums.SenderIDVerificationStatus
		active, changesRequired, closed, closingInProcess, inProcess, isNew bool
	}{
		{enums.SenderIDVerificationStatusActive, true, false, false, false, false, false},
		{enums.SenderIDVerificationStatusChangesRequired, false, true, false, false, false, false},
		{enums.SenderIDVerificationStatusClosed, false, false, true, false, false, false},
		{enums.SenderIDVerificationStatusClosingInProcess, false, false, false, true, false, false},
		{enums.SenderIDVerificationStatusInProcess, false, false, false, false, true, false},
		{enums.SenderIDVerificationStatusNew, false, false, false, false, false, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			siv := &resource.SenderIDVerification{Status: tt.status}
			assert.Equal(t, tt.active, siv.IsActive())
			assert.Equal(t, tt.changesRequired, siv.IsChangesRequired())
			assert.Equal(t, tt.closed, siv.IsClosed())
			assert.Equal(t, tt.closingInProcess, siv.IsClosingInProcess())
			assert.Equal(t, tt.inProcess, siv.IsInProcess())
			assert.Equal(t, tt.isNew, siv.IsNew())
		})
	}
}
