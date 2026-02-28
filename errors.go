package didww

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ErrorSource contains the JSON pointer to the source of an error.
type ErrorSource struct {
	Pointer string `json:"pointer,omitempty"`
}

// ErrorDetail represents a single JSON:API error object.
type ErrorDetail struct {
	Title  string      `json:"title,omitempty"`
	Detail string      `json:"detail,omitempty"`
	Code   string      `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	Source ErrorSource `json:"source,omitempty"`
}

// APIError represents an error response from the DIDWW API.
type APIError struct {
	HTTPStatus int
	Errors     []ErrorDetail
}

func (e *APIError) Error() string {
	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = err.Detail
		if messages[i] == "" {
			messages[i] = err.Title
		}
	}
	return fmt.Sprintf("DIDWW API error (HTTP %d): %s", e.HTTPStatus, strings.Join(messages, "; "))
}

// ClientError represents a client-side error.
type ClientError struct {
	Message string
}

func (e *ClientError) Error() string {
	return e.Message
}

// ParseAPIErrors parses a JSON:API error response body.
func ParseAPIErrors(body []byte, httpStatus int) (*APIError, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response body")
	}

	var errResp struct {
		Errors []ErrorDetail `json:"errors"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil, err
	}

	return &APIError{
		HTTPStatus: httpStatus,
		Errors:     errResp.Errors,
	}, nil
}
