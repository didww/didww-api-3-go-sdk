package didww

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestDIDReservationsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_reservations": {status: http.StatusOK, fixture: "did_reservations/index.json"},
	})

	reservations, err := client.DIDReservations().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reservations) == 0 {
		t.Fatal("expected non-empty reservations list")
	}
}

func TestDIDReservationsCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/did_reservations": {status: http.StatusCreated, fixture: "did_reservations/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	reservation, err := server.client.DIDReservations().Create(context.Background(), &DIDReservation{
		Description:    "DIDWW",
		AvailableDIDID: "857d1462-5f43-4238-b007-ff05f282e41b",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reservation.ID != "fd38d3ff-80cf-4e67-a605-609a2884a5c4" {
		t.Errorf("expected ID 'fd38d3ff-80cf-4e67-a605-609a2884a5c4', got %q", reservation.ID)
	}

	assertRequestJSON(t, capturedBody, "did_reservations/create_request.json")
}

func TestDIDReservationsFindWithIncludedAvailableDID(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_reservations/fd38d3ff-80cf-4e67-a605-609a2884a5c4": {status: http.StatusOK, fixture: "did_reservations/show.json"},
	})

	params := NewQueryParams().Include("available_did")
	reservation, err := client.DIDReservations().Find(context.Background(), "fd38d3ff-80cf-4e67-a605-609a2884a5c4", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reservation.ID != "fd38d3ff-80cf-4e67-a605-609a2884a5c4" {
		t.Errorf("expected ID 'fd38d3ff-80cf-4e67-a605-609a2884a5c4', got %q", reservation.ID)
	}
	if reservation.Description != "DIDWW" {
		t.Errorf("expected Description 'DIDWW', got %q", reservation.Description)
	}

	if reservation.AvailableDID == nil {
		t.Fatal("expected non-nil AvailableDID")
	}
	if reservation.AvailableDID.ID != "857d1462-5f43-4238-b007-ff05f282e41b" {
		t.Errorf("expected AvailableDID ID '857d1462-5f43-4238-b007-ff05f282e41b', got %q", reservation.AvailableDID.ID)
	}
	if reservation.AvailableDID.Number != "19492033398" {
		t.Errorf("expected AvailableDID Number '19492033398', got %q", reservation.AvailableDID.Number)
	}
}

func TestDIDReservationsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/did_reservations/fd38d3ff-80cf-4e67-a605-609a2884a5c4": {status: http.StatusNoContent},
	})

	err := client.DIDReservations().Delete(context.Background(), "fd38d3ff-80cf-4e67-a605-609a2884a5c4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
