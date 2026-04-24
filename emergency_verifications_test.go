package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmergencyVerificationsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/emergency_verifications": {status: http.StatusOK, fixture: "emergency_verifications/index.json"},
	})

	evs, err := client.EmergencyVerifications().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, evs, 1)
	assert.Equal(t, "ev-001-id", evs[0].ID)
	assert.Equal(t, "EV-123", evs[0].Reference)
	assert.Equal(t, "pending", evs[0].Status)
	assert.False(t, evs[0].CreatedAt.IsZero())
}

func TestEmergencyVerificationsUpdateExternalReferenceID(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/emergency_verifications/ev-001-id": {status: http.StatusOK, fixture: "emergency_verifications/update.json"},
	})

	extRef := "ev-ext-ref"
	ev, err := server.client.EmergencyVerifications().Update(context.Background(), &resource.EmergencyVerification{
		ID:                  "ev-001-id",
		ExternalReferenceID: &extRef,
	})
	require.NoError(t, err)

	assert.Equal(t, "ev-001-id", ev.ID)
	require.NotNil(t, ev.ExternalReferenceID)
	assert.Equal(t, "ev-ext-ref", *ev.ExternalReferenceID)

	assertRequestJSON(t, *capturedBodyPtr, "emergency_verifications/update_request.json")
}

func TestEmergencyVerificationsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/emergency_verifications": {status: http.StatusCreated, fixture: "emergency_verifications/create.json"},
	})

	ev, err := server.client.EmergencyVerifications().Create(context.Background(), &resource.EmergencyVerification{
		AddressID:                 "d3414687-40f4-4346-a267-c2c65117d28c",
		EmergencyCallingServiceID: "ecs-001-id",
	})
	require.NoError(t, err)

	assert.Equal(t, "ev-new-id", ev.ID)
	assert.Equal(t, "pending", ev.Status)

	assertRequestJSON(t, *capturedBodyPtr, "emergency_verifications/create_request.json")
}
