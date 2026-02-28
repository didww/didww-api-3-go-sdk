package didww

import (
	"testing"
)

func TestEnvironmentSandboxURL(t *testing.T) {
	if Sandbox != "https://sandbox-api.didww.com/v3" {
		t.Errorf("expected Sandbox URL to be https://sandbox-api.didww.com/v3, got %s", Sandbox)
	}
}

func TestEnvironmentProductionURL(t *testing.T) {
	if Production != "https://api.didww.com/v3" {
		t.Errorf("expected Production URL to be https://api.didww.com/v3, got %s", Production)
	}
}

func TestNewClientRequiresAPIKey(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error when creating client with empty API key")
	}
}

func TestNewClientWithValidAPIKey(t *testing.T) {
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestNewClientDefaultsToSandbox(t *testing.T) {
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BaseURL() != string(Sandbox) {
		t.Errorf("expected default base URL to be %s, got %s", Sandbox, client.BaseURL())
	}
}

func TestNewClientWithProductionEnvironment(t *testing.T) {
	client, err := NewClient("test-api-key", WithEnvironment(Production))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BaseURL() != string(Production) {
		t.Errorf("expected base URL to be %s, got %s", Production, client.BaseURL())
	}
}

func TestNewClientWithCustomBaseURL(t *testing.T) {
	customURL := "http://localhost:3000/v3"
	client, err := NewClient("test-api-key", WithBaseURL(customURL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BaseURL() != customURL {
		t.Errorf("expected base URL to be %s, got %s", customURL, client.BaseURL())
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	client, err := NewClient("test-api-key", WithTimeout(5000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestNewClientWithMultipleOptions(t *testing.T) {
	client, err := NewClient("test-api-key",
		WithEnvironment(Production),
		WithTimeout(10000),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BaseURL() != string(Production) {
		t.Errorf("expected base URL to be %s, got %s", Production, client.BaseURL())
	}
}

func TestClientAPIKey(t *testing.T) {
	client, err := NewClient("my-secret-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.APIKey() != "my-secret-key" {
		t.Errorf("expected API key to be my-secret-key, got %s", client.APIKey())
	}
}
