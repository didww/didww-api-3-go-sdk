package didww

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// TokenStore caches access tokens. Keys are opaque strings that already cover
// the environment, the API key (hashed), the audience, the API version and the
// scope, so one store may be shared by several clients or processes.
//
// A store holds live bearer tokens: protect it like the API key. A token is
// stored for the lifetime granted by the token service; it is refreshed when
// the API refuses it, not ahead of its expiry.
//
// Implementations should return when ctx is done: the SDK bounds the store
// calls it makes on its own behalf with a context deadline.
type TokenStore interface {
	// Get returns the token stored under key; ok is false when there is none
	// or it has expired.
	Get(ctx context.Context, key string) (t Token, ok bool, err error)
	// Set stores t under key for ttl; a ttl of zero or less means no expiry.
	Set(ctx context.Context, key string, t Token, ttl time.Duration) error
	// Delete removes the token stored under key, if any.
	Delete(ctx context.Context, key string) error
}

// NewMemoryTokenStore returns a TokenStore that keeps tokens in process
// memory. Expired entries are dropped when read. It is safe for concurrent use.
func NewMemoryTokenStore() TokenStore {
	return &memoryTokenStore{entries: map[string]memoryTokenEntry{}, now: time.Now}
}

type memoryTokenEntry struct {
	token     Token
	expiresAt time.Time
}

type memoryTokenStore struct {
	mu      sync.Mutex
	entries map[string]memoryTokenEntry
	now     func() time.Time
}

func (s *memoryTokenStore) Get(_ context.Context, key string) (Token, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	if !ok {
		return Token{}, false, nil
	}
	if !e.expiresAt.IsZero() && !s.now().Before(e.expiresAt) {
		delete(s.entries, key)
		return Token{}, false, nil
	}
	return e.token, true, nil
}

func (s *memoryTokenStore) Set(_ context.Context, key string, t Token, ttl time.Duration) error { //nolint:gocritic // TokenStore signature
	e := memoryTokenEntry{token: t}
	s.mu.Lock()
	defer s.mu.Unlock()
	if ttl > 0 {
		e.expiresAt = s.now().Add(ttl)
	}
	s.entries[key] = e
	return nil
}

func (s *memoryTokenStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
	return nil
}

// String describes the store without its tokens.
func (s *memoryTokenStore) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return "MemoryTokenStore{entries: " + strconv.Itoa(len(s.entries)) + "}"
}

// GoString describes the store without its tokens.
func (s *memoryTokenStore) GoString() string {
	return s.String()
}
