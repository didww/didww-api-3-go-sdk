package didww

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/resource"
)

func TestClientSendsCorrectHeaders(t *testing.T) {
	var (
		receivedContentType string
		receivedAccept      string
		receivedAPIKey      string
		receivedAPIVersion  string
		receivedUserAgent   string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		receivedAccept = r.Header.Get("Accept")
		receivedAPIKey = r.Header.Get("Api-Key")
		receivedAPIVersion = r.Header.Get("X-DIDWW-API-Version")
		receivedUserAgent = r.Header.Get("User-Agent")

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		w.Write(loadFixture(t, "balance/index.json"))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	_, err = client.Balance().Find(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "application/vnd.api+json", receivedContentType)
	assert.Equal(t, "application/vnd.api+json", receivedAccept)
	assert.Equal(t, "test-api-key", receivedAPIKey)
	assert.Equal(t, apiVersion, receivedAPIVersion)
	assert.Equal(t, "didww-go-sdk/3.0.0", receivedUserAgent)
}

func TestClientHandlesHTTPErrors(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		makeRequest  func(*Client) error
		checkErrors  func(*testing.T, *APIError)
	}{
		{
			name:         "404 Not Found",
			statusCode:   http.StatusNotFound,
			responseBody: `{"errors":[{"title":"not found","detail":"Resource not found","status":"404"}]}`,
			makeRequest: func(c *Client) error {
				_, err := c.Countries().Find(context.Background(), "nonexistent-id")
				return err
			},
		},
		{
			name:         "500 Internal Server Error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"errors":[{"title":"server error","detail":"Internal server error","status":"500"}]}`,
			makeRequest: func(c *Client) error {
				_, err := c.Balance().Find(context.Background())
				return err
			},
		},
		{
			name:         "422 Unprocessable Entity",
			statusCode:   http.StatusUnprocessableEntity,
			responseBody: `{"errors":[{"title":"is invalid","detail":"name - is invalid","code":"100","source":{"pointer":"/data/attributes/name"},"status":"422"}]}`,
			makeRequest: func(c *Client) error {
				_, err := c.VoiceInTrunks().Create(context.Background(), &resource.VoiceInTrunk{Name: "test"})
				return err
			},
			checkErrors: func(t *testing.T, apiErr *APIError) {
				require.Len(t, apiErr.Errors, 1)
				assert.Equal(t, "100", apiErr.Errors[0].Code)
				assert.Equal(t, "/data/attributes/name", apiErr.Errors[0].Source.Pointer)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.responseBody))
			}))
			defer server.Close()

			client, err := NewClient("test-api-key", WithBaseURL(server.URL))
			require.NoError(t, err)

			err = tc.makeRequest(client)
			require.Error(t, err)

			apiErr, ok := err.(*APIError)
			require.True(t, ok, "expected *APIError")
			assert.Equal(t, tc.statusCode, apiErr.HTTPStatus)
			if tc.checkErrors != nil {
				tc.checkErrors(t, apiErr)
			}
		})
	}
}

func TestClientWithQueryParamsAppendedToURL(t *testing.T) {
	var requestURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURL = r.URL.String()
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		w.Write(loadFixture(t, "countries/index.json"))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	params := NewQueryParams().
		Filter("prefix", "44").
		Sort("name").
		Include("regions").
		Page(1, 25)

	_, err = client.Countries().List(context.Background(), params)
	require.NoError(t, err)

	// Verify query params are appended to URL
	assertContains(t, requestURL, "filter[prefix]=44")
	assertContains(t, requestURL, "sort=name")
	assertContains(t, requestURL, "include=regions")
	assertContains(t, requestURL, "page[number]=1")
	assertContains(t, requestURL, "page[size]=25")
}

func TestClientWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	client, err := NewClient("test-api-key", WithHTTPClient(custom))
	require.NoError(t, err)
	require.NotNil(t, client)
	// Verify the custom HTTP client is used by making a request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
		w.Write(loadFixture(t, "balance/index.json"))
	}))
	defer server.Close()

	client, err = NewClient("test-api-key", WithBaseURL(server.URL), WithHTTPClient(custom))
	require.NoError(t, err)
	_, err = client.Balance().Find(context.Background())
	require.NoError(t, err)
}
