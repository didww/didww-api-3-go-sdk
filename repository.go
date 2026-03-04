package didww

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/didww/didww-api-3-go-sdk/jsonapi"
)

// Repository provides CRUD operations for a JSON:API resource.
type Repository[T any] struct {
	client       *Client
	resourceType string // JSON:API type, also used as URL path
}

// NewRepository creates a Repository for resource type T.
// The JSON:API type is read from T's struct tag.
func NewRepository[T any](client *Client) *Repository[T] {
	return &Repository[T]{
		client:       client,
		resourceType: jsonapi.ResourceType[T](),
	}
}

// List retrieves a collection of resources.
func (r *Repository[T]) List(ctx context.Context, params *QueryParams) ([]*T, error) {
	body, err := r.client.doRequest(ctx, http.MethodGet, r.resourceType, params, nil)
	if err != nil {
		return nil, err
	}
	return jsonapi.UnmarshalMany[T](body)
}

// Find retrieves a single resource by ID.
func (r *Repository[T]) Find(ctx context.Context, id string, params ...*QueryParams) (*T, error) {
	path := r.resourceType + "/" + id
	var qp *QueryParams
	if len(params) > 0 {
		qp = params[0]
	}
	body, err := r.client.doRequest(ctx, http.MethodGet, path, qp, nil)
	if err != nil {
		return nil, err
	}
	return jsonapi.UnmarshalOne[T](body)
}

// Create creates a new resource.
func (r *Repository[T]) Create(ctx context.Context, resource *T) (*T, error) {
	reqBody, err := jsonapi.Marshal(resource)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to serialize resource: %v", err)}
	}
	body, err := r.client.doRequest(ctx, http.MethodPost, r.resourceType, nil, reqBody)
	if err != nil {
		return nil, err
	}
	return jsonapi.UnmarshalOne[T](body)
}

// Update updates an existing resource. The resource must have a non-empty ID.
// Only fields modified since loading are included in the PATCH request.
// Pointer fields set to nil produce explicit JSON null.
//
// The returned resource has a fresh clean-state snapshot, so use it for any
// subsequent updates. The input resource's tracking state is consumed by the
// call and should not be reused.
func (r *Repository[T]) Update(ctx context.Context, resource *T) (*T, error) {
	id := jsonapi.GetID(resource)
	if id == "" {
		return nil, &ClientError{Message: "resource ID is required for update"}
	}
	path := r.resourceType + "/" + id
	reqBody, err := jsonapi.MarshalPatch(resource)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to serialize resource: %v", err)}
	}
	body, err := r.client.doRequest(ctx, http.MethodPatch, path, nil, reqBody)
	if err != nil {
		return nil, err
	}
	jsonapi.ForgetCleanState(resource)
	return jsonapi.UnmarshalOne[T](body)
}

// Delete removes a resource by ID.
func (r *Repository[T]) Delete(ctx context.Context, id string) error {
	path := r.resourceType + "/" + id
	_, err := r.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
	return err
}

// SingletonRepository provides read access to a singleton resource.
type SingletonRepository[T any] struct {
	client       *Client
	resourceType string // JSON:API type, also used as URL path
}

// NewSingletonRepository creates a SingletonRepository for resource type T.
// The JSON:API type is read from T's struct tag.
func NewSingletonRepository[T any](client *Client) *SingletonRepository[T] {
	return &SingletonRepository[T]{
		client:       client,
		resourceType: jsonapi.ResourceType[T](),
	}
}

// Find retrieves the singleton resource.
func (r *SingletonRepository[T]) Find(ctx context.Context) (*T, error) {
	body, err := r.client.doRequest(ctx, http.MethodGet, r.resourceType, nil, nil)
	if err != nil {
		return nil, err
	}
	return jsonapi.UnmarshalOne[T](body)
}

const (
	jsonapiMediaType = "application/vnd.api+json"
	apiVersion       = "2022-05-10"
)

// doRequest executes an HTTP request and returns the response body.
func (c *Client) doRequest(ctx context.Context, method, path string, params *QueryParams, reqBody []byte) ([]byte, error) {
	// Build URL
	u := c.buildURL(path)
	if params != nil {
		encoded := params.Encode()
		if encoded != "" {
			u += "?" + encoded
		}
	}

	var bodyReader io.Reader
	if reqBody != nil {
		bodyReader = bytes.NewReader(reqBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	// Set headers
	req.Header.Set("Content-Type", jsonapiMediaType)
	req.Header.Set("Accept", jsonapiMediaType)
	req.Header.Set("User-Agent", "didww-go-sdk/0.1.0")
	req.Header.Set("X-DIDWW-API-Version", apiVersion)

	// Set API key header for non-public endpoints
	if !isPublicEndpoint(path) {
		req.Header.Set("Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		apiErr, parseErr := ParseAPIErrors(body, resp.StatusCode)
		if parseErr != nil {
			return nil, &APIError{
				HTTPStatus: resp.StatusCode,
				Errors:     []ErrorDetail{{Title: "Unknown error", Detail: string(body)}},
			}
		}
		return nil, apiErr
	}

	// For DELETE (204 No Content), return nil body
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	return body, nil
}

// publicEndpoints lists resource paths that do not require API key authentication.
var publicEndpoints = map[string]bool{
	"public_keys": true,
}

// isPublicEndpoint returns true if the given path is a public (no auth) endpoint.
func isPublicEndpoint(path string) bool {
	// Extract the first path segment (e.g. "public_keys" from "public_keys/some-id")
	base := path
	if i := strings.Index(path, "/"); i >= 0 {
		base = path[:i]
	}
	return publicEndpoints[base]
}

// buildURL constructs the full URL for a resource path.
func (c *Client) buildURL(resource string) string {
	base := c.baseURL
	if !strings.HasSuffix(base, "/v3") {
		base += "/v3"
	}
	return base + "/" + resource
}
