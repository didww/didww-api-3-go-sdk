package didww

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentSandboxURL(t *testing.T) {
	assert.Equal(t, Environment("https://sandbox-api.didww.com/v3"), Sandbox)
}

func TestEnvironmentProductionURL(t *testing.T) {
	assert.Equal(t, Environment("https://api.didww.com/v3"), Production)
}

func TestNewClientRequiresAPIKey(t *testing.T) {
	_, err := NewClient("")
	require.Error(t, err)
}

func TestNewClientWithValidAPIKey(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClientDefaultsToSandbox(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)
	assert.Equal(t, string(Sandbox), client.BaseURL())
}

func TestNewClientWithProductionEnvironment(t *testing.T) {
	client, err := NewClient("test-api-key", WithEnvironment(Production))
	require.NoError(t, err)
	assert.Equal(t, string(Production), client.BaseURL())
}

func TestNewClientWithCustomBaseURL(t *testing.T) {
	customURL := "http://localhost:3000/v3"
	client, err := NewClient("test-api-key", WithBaseURL(customURL))
	require.NoError(t, err)
	assert.Equal(t, customURL, client.BaseURL())
}

func TestNewClientWithTimeout(t *testing.T) {
	client, err := NewClient("test-api-key", WithTimeout(5000))
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClientWithMultipleOptions(t *testing.T) {
	client, err := NewClient("test-api-key",
		WithEnvironment(Production),
		WithTimeout(10000),
	)
	require.NoError(t, err)
	assert.Equal(t, string(Production), client.BaseURL())
}

func TestClientAPIKey(t *testing.T) {
	client, err := NewClient("my-secret-key")
	require.NoError(t, err)
	assert.Equal(t, "my-secret-key", client.APIKey())
}
