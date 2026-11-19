package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
)

// README: Usage
func (s *suite) usage() {
	s.section("Usage")
	s.check("Check balance and list DID groups with stock keeping units", func() error {
		client, err := didww.NewClient(s.apiKey)
		if err != nil {
			return err
		}

		// Check balance
		balance, err := client.Balance().Find(context.Background())
		if err != nil {
			return err
		}
		s.logf("Balance: %s", balance.TotalBalance)

		// List DID groups with stock keeping units
		params := didww.NewQueryParams().
			Include("stock_keeping_units").
			Filter("area_name", "Acapulco")
		didGroups, err := client.DIDGroups().List(context.Background(), params)
		if err != nil {
			return err
		}
		s.logf("DID groups: %d", len(didGroups))
		return expect(balance.TotalBalance != "", "empty total balance")
	})
}

// README: Configuration, Custom HTTP Client
func (s *suite) configuration() {
	s.section("Configuration")
	s.check("Production client (read-only)", s.productionReadOnly)

	s.check("Custom HTTP client with a proxy (sandbox)", func() error {
		proxyEnv := os.Getenv("DIDWW_README_PROXY_URL")
		if proxyEnv == "" {
			return skip("set DIDWW_README_PROXY_URL to an HTTP proxy to run it")
		}
		proxyURL, err := url.Parse(proxyEnv)
		if err != nil {
			return err
		}
		var proxied atomic.Int32 // counts the requests sent through the proxy
		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: func(r *http.Request) (*url.URL, error) {
					proxied.Add(1)
					return http.ProxyURL(proxyURL)(r)
				},
			},
		}

		// The sandbox instead of the README's production environment; a store
		// of its own, so that the token is minted through the proxy as well
		client, err := didww.NewClient(s.apiKey,
			didww.WithEnvironment(didww.Sandbox),
			didww.WithHTTPClient(httpClient),
			didww.WithTokenStore(didww.NewMemoryTokenStore()),
		)
		if err != nil {
			return err
		}
		balance, err := client.Balance().Find(s.ctx)
		if err != nil {
			return err
		}
		s.logf("balance %s, %d requests through the proxy", balance.TotalBalance, proxied.Load())
		return expect(proxied.Load() >= 2, "%d requests went through the proxy, want the token and the API request", proxied.Load())
	})
}

// productionReadOnly runs the README's production client snippet with its own
// production key, read-only: the read scope makes the token service issue a
// token that the API accepts for GET and HEAD only, so a write would be
// refused with 403. It mints the token, reads the balance and the countries
// and looks up one price, and prints no key, token, amount or account data.
// This is the only check that talks to production.
func (s *suite) productionReadOnly() error {
	ctx := s.ctx
	apiKey := os.Getenv("DIDWW_README_PRODUCTION_API_KEY")
	if apiKey == "" {
		return skip("set DIDWW_README_PRODUCTION_API_KEY to run it read-only against production")
	}

	tokens := &scopeRecorder{TokenStore: didww.NewMemoryTokenStore()}
	client, err := didww.NewClient(apiKey,
		didww.WithEnvironment(didww.Production),
		didww.WithScope(didww.ScopeRead), // not in the README: read-only, whatever is called
		didww.WithTokenStore(tokens),
	)
	if err != nil {
		return err
	}
	if client.BaseURL() != string(didww.Production) {
		return fmt.Errorf("base URL %s", client.BaseURL())
	}

	if _, err = client.Balance().Find(ctx); err != nil {
		return fmt.Errorf("balance: %w", err)
	}
	countries, err := client.Countries().List(ctx, nil)
	if err != nil {
		return fmt.Errorf("countries: %w", err)
	}
	priceFact := "price lookup ok"
	_, err = client.VoiceInPrices().Lookup(ctx, map[string]string{"number": "12125551234"})
	switch {
	case errors.Is(err, didww.ErrPriceNotFound):
		priceFact = "price lookup answered: no price for the number"
	case errors.Is(err, didww.ErrPricePlanNotAssigned):
		priceFact = "price lookup answered: no Voice IN price plan"
	case err != nil:
		return fmt.Errorf("price lookup: %w", err)
	}

	scopes := tokens.grantedScopes()
	for _, scope := range scopes {
		if scope != string(didww.ScopeRead) {
			return fmt.Errorf("a token was granted scope %q, want read", scope)
		}
	}
	s.logf("production: %d tokens minted (scope read), balance read, %d countries, %s", len(scopes), len(countries), priceFact)
	return expect(len(scopes) >= 2, "%d tokens minted, want 2 (API and Voice IN prices)", len(scopes))
}

// scopeRecorder is a token store that records the scope of the tokens it
// stores, never the tokens themselves.
type scopeRecorder struct {
	didww.TokenStore
	mu     sync.Mutex
	scopes []string
}

func (r *scopeRecorder) Set(ctx context.Context, key string, t didww.Token, ttl time.Duration) error { //nolint:gocritic // signature of didww.TokenStore
	r.mu.Lock()
	r.scopes = append(r.scopes, t.Scope)
	r.mu.Unlock()
	return r.TokenStore.Set(ctx, key, t, ttl)
}

func (r *scopeRecorder) grantedScopes() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.scopes...)
}

// README: Authentication (Scope, Token cache)
func (s *suite) authentication() {
	ctx := s.ctx
	s.section("Authentication")
	s.check("Scope", func() error {
		readClient, err := didww.NewClient(s.apiKey, didww.WithScope(didww.ScopeRead)) // GET and HEAD only
		if err != nil {
			return err
		}
		if _, err = didww.NewClient(s.apiKey, didww.WithScope(didww.ScopeFull)); err != nil {
			return err
		}
		if _, err = didww.NewClient(s.apiKey); err != nil { // the default: scope is not sent, the token service grants full
			return err
		}

		balance, err := readClient.Balance().Find(ctx)
		if err != nil {
			return err
		}
		s.logf("read scope: balance %s", balance.TotalBalance)
		return nil
	})

	s.check("Token cache: a custom TokenStore", func() error {
		myRedisStore := newMapTokenStore()
		client, err := didww.NewClient(s.apiKey, didww.WithTokenStore(myRedisStore))
		if err != nil {
			return err
		}
		if _, err := client.Balance().Find(ctx); err != nil {
			return err
		}
		stored := myRedisStore.len()
		s.logf("tokens in the store: %d", stored)
		return expect(stored == 1, "the store holds %d tokens, want 1", stored)
	})

	s.check("Token cache: a TokenStore backed by Redis, shared by two clients", s.redisTokenCache)
}

// redisTokenCache mints a token through a Redis store and checks that a second
// client, with its own connection to Redis, reuses it instead of minting one.
func (s *suite) redisTokenCache() (err error) {
	ctx := s.ctx
	redisURL := os.Getenv("DIDWW_README_REDIS_URL")
	if redisURL == "" {
		return skip("set DIDWW_README_REDIS_URL to a Redis server to run it")
	}
	prefix := "didww-readme:" + s.suffix + ":" // keys of this run only

	first, err := dialRedis(ctx, redisURL, prefix)
	if err != nil {
		return err
	}
	defer s.undo(&err, "delete the tokens from Redis", func() error { return first.close(ctx) })
	second, err := dialRedis(ctx, redisURL, prefix)
	if err != nil {
		return err
	}
	defer s.undo(&err, "close the second Redis connection", func() error { return second.close(ctx) })

	firstClient, err := didww.NewClient(s.apiKey, didww.WithTokenStore(first))
	if err != nil {
		return err
	}
	if _, err = firstClient.Balance().Find(ctx); err != nil {
		return err
	}
	secondClient, err := didww.NewClient(s.apiKey, didww.WithTokenStore(second))
	if err != nil {
		return err
	}
	if _, err = secondClient.Balance().Find(ctx); err != nil {
		return err
	}

	_, firstSets := first.counts()
	secondHits, secondSets := second.counts()
	s.logf("first client stored %d token(s); second client read %d and stored %d", firstSets, secondHits, secondSets)
	return expect(firstSets == 1 && secondHits >= 1 && secondSets == 0,
		"the second client did not reuse the token of the first one")
}

// mapTokenStore is a didww.TokenStore kept in a map, in place of a store
// shared between processes such as Redis.
type mapTokenStore struct {
	mu     sync.Mutex
	tokens map[string]didww.Token
}

var _ didww.TokenStore = (*mapTokenStore)(nil)

func newMapTokenStore() *mapTokenStore {
	return &mapTokenStore{tokens: map[string]didww.Token{}}
}

func (m *mapTokenStore) Get(_ context.Context, key string) (t didww.Token, ok bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok = m.tokens[key]
	return t, ok, nil
}

func (m *mapTokenStore) Set(_ context.Context, key string, t didww.Token, _ time.Duration) error { //nolint:gocritic // signature of didww.TokenStore
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[key] = t
	return nil
}

func (m *mapTokenStore) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, key)
	return nil
}

func (m *mapTokenStore) len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.tokens)
}

// README: Error Handling
func (s *suite) errorHandling() {
	ctx, client := s.ctx, s.client
	s.section("Error Handling")
	s.check("APIError and ClientError", func() error {
		trunk, err := client.VoiceInTrunks().Find(ctx, "nonexistent")
		if err == nil {
			return fmt.Errorf("found trunk %s, want an error", trunk.ID)
		}
		var apiErr *didww.APIError
		if errors.As(err, &apiErr) {
			s.logf("HTTP Status: %d", apiErr.HTTPStatus)
			for _, e := range apiErr.Errors {
				s.logf("Error: %s", e.Detail)
			}
		}

		var clientErr *didww.ClientError
		if errors.As(err, &clientErr) {
			s.logf("Client error: %s", clientErr.Message)
		}

		if errors.As(err, &apiErr) && apiErr.HTTPStatus == http.StatusTooManyRequests {
			s.logf("API rate limit, retry after %s", apiErr.RetryAfter)
		}
		return expect(apiErr != nil && apiErr.HTTPStatus == http.StatusNotFound, "got %v, want an *APIError with status 404", err)
	})

	s.pricesErrors()
}

// README: Webhook Signature Validation
func (s *suite) webhook() {
	s.section("Webhook Signature Validation")
	s.check("RequestValidator", func() error {
		// The example of the DIDWW API documentation (callbacks details)
		const (
			callbackSecret = "szrdgh6547umt7tht7xbqhj6g9gdbyp7" //nolint:gosec // the published documentation example
			requestURL     = "https://mycompany.com/didww_callbacks?opaque=123"
		)
		form := url.Values{
			"id":     {"bf2cee72-6caa-4ae2-917e-bea01945691e"},
			"status": {"completed"},
			"type":   {"orders"},
		}
		r := httptest.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set(didww.SignatureHeaderName, "30f66e9d72eb5e193051fd02952f70d8e934b4ff")
		if err := r.ParseForm(); err != nil {
			return err
		}

		validator := didww.NewRequestValidator(callbackSecret)

		// In your webhook handler:
		signature := r.Header.Get(didww.SignatureHeaderName) // "X-DIDWW-Signature"
		payload := map[string]string{}                       // parsed form/query payload
		for key := range r.PostForm {
			payload[key] = r.PostForm.Get(key)
		}
		valid := validator.Validate(requestURL, payload, signature)

		payload["status"] = "canceled"
		tampered := validator.Validate(requestURL, payload, signature)
		s.logf("valid: %t, tampered payload valid: %t", valid, tampered)
		return expect(valid && !tampered, "valid=%t tampered=%t, want true and false", valid, tampered)
	})
}
