package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDIDHistoryList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_history": {status: http.StatusOK, fixture: "did_history/index.json"},
	})

	records, err := client.DIDHistory().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, records, 2)
	assert.Equal(t, "a1b2c3d4-e5f6-7890-abcd-ef1234567890", records[0].ID)
	assert.Equal(t, "12025551234", records[0].DIDNumber)
	assert.Equal(t, "assigned", records[0].Action)
	assert.Equal(t, "api3", records[0].Method)
	assert.Equal(t, "renewed", records[1].Action)
	assert.Equal(t, "system", records[1].Method)
}

func TestDIDHistoryFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_history/a1b2c3d4-e5f6-7890-abcd-ef1234567890": {status: http.StatusOK, fixture: "did_history/show.json"},
	})

	record, err := client.DIDHistory().Find(context.Background(), "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	require.NoError(t, err)

	assert.Equal(t, "a1b2c3d4-e5f6-7890-abcd-ef1234567890", record.ID)
	assert.Equal(t, "12025551234", record.DIDNumber)
	assert.Equal(t, "assigned", record.Action)
	assert.Equal(t, "api3", record.Method)
	assert.False(t, record.CreatedAt.IsZero())
	// No meta for non-billing_cycles_count_changed actions
	assert.Nil(t, record.Meta)
}

func TestDIDHistoryFindBillingCyclesCountChanged(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/did_history/c3d4e5f6-a7b8-9012-cdef-123456789012": {status: http.StatusOK, fixture: "did_history/show_billing_cycles_count_changed.json"},
	})

	record, err := client.DIDHistory().Find(context.Background(), "c3d4e5f6-a7b8-9012-cdef-123456789012")
	require.NoError(t, err)

	assert.Equal(t, "c3d4e5f6-a7b8-9012-cdef-123456789012", record.ID)
	assert.Equal(t, "12025551234", record.DIDNumber)
	assert.Equal(t, "billing_cycles_count_changed", record.Action)
	assert.Equal(t, "system", record.Method)
	assert.False(t, record.CreatedAt.IsZero())
	// Meta fields present for billing_cycles_count_changed
	require.NotNil(t, record.Meta)
	assert.Equal(t, "2", record.Meta["from"])
	assert.Equal(t, "1", record.Meta["to"])
}
