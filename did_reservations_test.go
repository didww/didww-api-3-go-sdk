package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/resource"
)

func TestDIDReservationsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_reservations": {status: http.StatusOK, fixture: "did_reservations/index.json"},
	})

	reservations, err := client.DIDReservations().List(context.Background(), nil)
	require.NoError(t, err)

	require.NotEmpty(t, reservations)
}

func TestDIDReservationsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/did_reservations": {status: http.StatusCreated, fixture: "did_reservations/create.json"},
	})

	reservation, err := server.client.DIDReservations().Create(context.Background(), &resource.DIDReservation{
		Description:    "DIDWW",
		AvailableDIDID: "857d1462-5f43-4238-b007-ff05f282e41b",
	})
	require.NoError(t, err)

	assert.Equal(t, "fd38d3ff-80cf-4e67-a605-609a2884a5c4", reservation.ID)

	assertRequestJSON(t, *capturedBodyPtr, "did_reservations/create_request.json")
}

func TestDIDReservationsFindWithIncludedAvailableDID(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_reservations/fd38d3ff-80cf-4e67-a605-609a2884a5c4": {status: http.StatusOK, fixture: "did_reservations/show.json"},
	})

	params := NewQueryParams().Include("available_did")
	reservation, err := client.DIDReservations().Find(context.Background(), "fd38d3ff-80cf-4e67-a605-609a2884a5c4", params)
	require.NoError(t, err)

	assert.Equal(t, "fd38d3ff-80cf-4e67-a605-609a2884a5c4", reservation.ID)
	assert.Equal(t, "DIDWW", reservation.Description)

	require.NotNil(t, reservation.AvailableDID)
	assert.Equal(t, "857d1462-5f43-4238-b007-ff05f282e41b", reservation.AvailableDID.ID)
	assert.Equal(t, "19492033398", reservation.AvailableDID.Number)
}

func TestDIDReservationsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/did_reservations/fd38d3ff-80cf-4e67-a605-609a2884a5c4": {status: http.StatusNoContent},
	})

	err := client.DIDReservations().Delete(context.Background(), "fd38d3ff-80cf-4e67-a605-609a2884a5c4")
	require.NoError(t, err)
}
