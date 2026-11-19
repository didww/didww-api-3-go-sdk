package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
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
	s.check("Production client", func() error {
		client, err := didww.NewClient(s.apiKey,
			didww.WithEnvironment(didww.Production),
		)
		if err != nil {
			return err
		}
		if client.BaseURL() != string(didww.Production) {
			return fmt.Errorf("base URL %s", client.BaseURL())
		}
		return skip("production environment: the client is built, no request is sent")
	})

	s.check("Custom HTTP client with a proxy", func() error {
		proxyURL, _ := url.Parse("http://proxy.example.com:8080")
		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}

		client, err := didww.NewClient(s.apiKey,
			didww.WithEnvironment(didww.Production),
			didww.WithHTTPClient(httpClient),
		)
		if err != nil {
			return err
		}
		_ = client
		return skip("production environment through a placeholder proxy: the client is built, no request is sent")
	})
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
