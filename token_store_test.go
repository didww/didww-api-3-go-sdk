package didww

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMemoryTokenStore(clock *fakeClock) *memoryTokenStore {
	store := NewMemoryTokenStore().(*memoryTokenStore)
	store.now = clock.Now
	return store
}

func TestMemoryTokenStoreGetMissing(t *testing.T) {
	store := newTestMemoryTokenStore(newFakeClock())

	_, ok, err := store.Get(context.Background(), "key")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestMemoryTokenStoreExpiry(t *testing.T) {
	ctx := context.Background()
	clock := newFakeClock()
	store := newTestMemoryTokenStore(clock)
	require.NoError(t, store.Set(ctx, "key", testToken(), 60*time.Second))

	clock.Advance(59 * time.Second)
	tok, ok, err := store.Get(ctx, "key")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, testToken(), tok)

	clock.Advance(time.Second)
	_, ok, err = store.Get(ctx, "key")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, store.entries)
}

func TestMemoryTokenStoreWithoutTTL(t *testing.T) {
	ctx := context.Background()
	clock := newFakeClock()
	store := newTestMemoryTokenStore(clock)
	require.NoError(t, store.Set(ctx, "key", testToken(), 0))

	clock.Advance(1000 * time.Hour)
	_, ok, err := store.Get(ctx, "key")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestMemoryTokenStoreReplace(t *testing.T) {
	ctx := context.Background()
	clock := newFakeClock()
	store := newTestMemoryTokenStore(clock)
	old := testToken()
	replacement := testToken()
	replacement.AccessToken = "new-token"

	require.NoError(t, store.Set(ctx, "key", old, 60*time.Second))
	clock.Advance(30 * time.Second)
	require.NoError(t, store.Set(ctx, "key", replacement, 60*time.Second))
	clock.Advance(59 * time.Second)

	tok, ok, err := store.Get(ctx, "key")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "new-token", tok.AccessToken)
}

func TestMemoryTokenStoreDelete(t *testing.T) {
	ctx := context.Background()
	store := newTestMemoryTokenStore(newFakeClock())
	require.NoError(t, store.Set(ctx, "key", testToken(), time.Minute))

	require.NoError(t, store.Delete(ctx, "key"))
	require.NoError(t, store.Delete(ctx, "missing"))

	_, ok, err := store.Get(ctx, "key")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestMemoryTokenStoreStringHidesTokens(t *testing.T) {
	store := newTestMemoryTokenStore(newFakeClock())
	require.NoError(t, store.Set(context.Background(), "didww:token:v1:abc", testToken(), time.Minute))

	assert.Equal(t, "MemoryTokenStore{entries: 1}", fmt.Sprintf("%v", store))
	assert.NotContains(t, fmt.Sprintf("%#v", store), "secret-access-token")
}
