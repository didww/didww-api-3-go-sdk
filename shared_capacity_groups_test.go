package didww

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestSharedCapacityGroupsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/shared_capacity_groups": {status: http.StatusOK, fixture: "shared_capacity_groups/index.json"},
	})

	groups, err := client.SharedCapacityGroups().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(groups) != 4 {
		t.Fatalf("expected 4 shared capacity groups, got %d", len(groups))
	}
}

func TestSharedCapacityGroupsFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/shared_capacity_groups/89f987e2-0862-4bf4-a3f4-cdc89af0d875": {status: http.StatusOK, fixture: "shared_capacity_groups/show.json"},
	})

	params := NewQueryParams().Include("capacity_pool,dids")
	group, err := client.SharedCapacityGroups().Find(context.Background(), "89f987e2-0862-4bf4-a3f4-cdc89af0d875", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if group.ID != "89f987e2-0862-4bf4-a3f4-cdc89af0d875" {
		t.Errorf("expected ID '89f987e2-0862-4bf4-a3f4-cdc89af0d875', got %q", group.ID)
	}
	if group.Name != "didww" {
		t.Errorf("expected Name 'didww', got %q", group.Name)
	}
	if group.SharedChannelsCount != 19 {
		t.Errorf("expected SharedChannelsCount 19, got %d", group.SharedChannelsCount)
	}

	// Verify capacity pool is resolved
	if group.CapacityPool == nil {
		t.Fatal("expected non-nil CapacityPool")
	}
	if group.CapacityPool.ID != "f288d07c-e2fc-4ae6-9837-b18fb469c324" {
		t.Errorf("expected CapacityPool ID 'f288d07c-e2fc-4ae6-9837-b18fb469c324', got %q", group.CapacityPool.ID)
	}
}

func TestSharedCapacityGroupsCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/shared_capacity_groups": {status: http.StatusCreated, fixture: "shared_capacity_groups/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	group, err := server.client.SharedCapacityGroups().Create(context.Background(), &SharedCapacityGroup{
		Name:                 "java-sdk",
		SharedChannelsCount:  5,
		MeteredChannelsCount: 0,
		CapacityPoolID:       "f288d07c-e2fc-4ae6-9837-b18fb469c324",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if group.ID == "" {
		t.Error("expected non-empty ID")
	}

	assertRequestJSON(t, capturedBody, "shared_capacity_groups/create_request.json")
}

func TestSharedCapacityGroupsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/shared_capacity_groups/89f987e2-0862-4bf4-a3f4-cdc89af0d875": {status: http.StatusNoContent},
	})

	err := client.SharedCapacityGroups().Delete(context.Background(), "89f987e2-0862-4bf4-a3f4-cdc89af0d875")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
