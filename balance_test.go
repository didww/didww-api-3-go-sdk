package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/balance": {status: http.StatusOK, fixture: "balance/index.json"},
	})

	balance, err := client.Balance().Find(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "4c39e0bf-683b-4697-9322-5abaf4011883", balance.ID)
	assert.Equal(t, "60.00", balance.TotalBalance)
	assert.Equal(t, "10.00", balance.Credit)
	assert.Equal(t, "50.00", balance.Balance)
}
