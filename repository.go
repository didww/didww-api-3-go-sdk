package didww

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Repository provides CRUD operations for a JSON:API resource.
type Repository[T any] struct {
	client       *Client
	resourcePath string
	resourceType string
}

// List retrieves a collection of resources.
func (r *Repository[T]) List(ctx context.Context, params *QueryParams) ([]*T, error) {
	body, err := r.client.doRequest(ctx, http.MethodGet, r.resourcePath, params, nil)
	if err != nil {
		return nil, err
	}
	return unmarshalMany[T](body)
}

// Find retrieves a single resource by ID.
func (r *Repository[T]) Find(ctx context.Context, id string, params ...*QueryParams) (*T, error) {
	path := r.resourcePath + "/" + id
	var qp *QueryParams
	if len(params) > 0 {
		qp = params[0]
	}
	body, err := r.client.doRequest(ctx, http.MethodGet, path, qp, nil)
	if err != nil {
		return nil, err
	}
	return unmarshalOne[T](body)
}

// Create creates a new resource.
func (r *Repository[T]) Create(ctx context.Context, resource *T) (*T, error) {
	reqBody, err := marshalResource(resource, r.resourceType)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to serialize resource: %v", err)}
	}
	body, err := r.client.doRequest(ctx, http.MethodPost, r.resourcePath, nil, reqBody)
	if err != nil {
		return nil, err
	}
	return unmarshalOne[T](body)
}

// Update updates an existing resource. The resource must have a non-empty ID.
func (r *Repository[T]) Update(ctx context.Context, resource *T) (*T, error) {
	id := getID(resource)
	if id == "" {
		return nil, &ClientError{Message: "resource ID is required for update"}
	}
	path := r.resourcePath + "/" + id
	reqBody, err := marshalResource(resource, r.resourceType)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to serialize resource: %v", err)}
	}
	body, err := r.client.doRequest(ctx, http.MethodPatch, path, nil, reqBody)
	if err != nil {
		return nil, err
	}
	return unmarshalOne[T](body)
}

// Delete removes a resource by ID.
func (r *Repository[T]) Delete(ctx context.Context, id string) error {
	path := r.resourcePath + "/" + id
	_, err := r.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
	return err
}

// SingletonRepository provides read access to a singleton resource.
type SingletonRepository[T any] struct {
	client       *Client
	resourcePath string
}

// Find retrieves the singleton resource.
func (r *SingletonRepository[T]) Find(ctx context.Context) (*T, error) {
	body, err := r.client.doRequest(ctx, http.MethodGet, r.resourcePath, nil, nil)
	if err != nil {
		return nil, err
	}
	return unmarshalOne[T](body)
}

const jsonapiMediaType = "application/vnd.api+json"

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

	// Set API key header for non-public endpoints
	if !strings.Contains(path, "public_keys") {
		req.Header.Set("Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

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

// buildURL constructs the full URL for a resource path.
func (c *Client) buildURL(resource string) string {
	base := c.baseURL
	if !strings.HasSuffix(base, "/v3") {
		base += "/v3"
	}
	return base + "/" + resource
}
