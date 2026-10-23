package didww

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
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

// Sentinel errors of the token service, matched by *AuthError with errors.Is.
var (
	// ErrInvalidClient means the token service refused the API key (401). It is not retried.
	ErrInvalidClient = errors.New("didww: token service refused the API key")
	// ErrInvalidRequest means the token service rejected the request (400), e.g. an unknown scope.
	ErrInvalidRequest = errors.New("didww: token service rejected the request")
	// ErrRateLimited means the token service rate-limited the API key (429).
	// Further mints fail without a request until AuthError.RetryAfter has passed.
	ErrRateLimited = errors.New("didww: token service rate limit exceeded")
	// ErrAuthServer means the last attempt to reach the token service was
	// answered with a 5xx.
	ErrAuthServer = errors.New("didww: token service failed")
	// ErrAuthConnection means the last attempt could not reach the token
	// service, or TLS failed.
	ErrAuthConnection = errors.New("didww: token service unreachable")
	// ErrInvalidGrant means the granted token is empty or does not cover the requested audience.
	ErrInvalidGrant = errors.New("didww: token service granted an unusable token")
)

const authErrPrefix = "didww: token service: "

// AuthError is returned when an access token cannot be obtained. It never
// contains the API key or a token. Use errors.Is with the sentinels above to
// tell the cases apart.
type AuthError struct {
	StatusCode  int           // HTTP status of the token service; 0 on a connection failure
	Code        string        // "error" field of the answer, e.g. "invalid_client"
	Description string        // "error_description" field of the answer
	RetryAfter  time.Duration // set with ErrRateLimited
	Err         error         // underlying transport error, if any
	kind        error
}

func (e *AuthError) Error() string {
	switch {
	case e.Code != "" && e.Description != "":
		return authErrPrefix + e.Code + ": " + e.Description
	case e.Code != "":
		return authErrPrefix + e.Code
	case e.Description != "":
		return authErrPrefix + e.Description
	case e.Err != nil:
		return authErrPrefix + e.Err.Error()
	default:
		return fmt.Sprintf("didww: token service responded with status %d", e.StatusCode)
	}
}

// Is reports whether target is the sentinel matching this error.
func (e *AuthError) Is(target error) bool {
	return e.kind != nil && target == e.kind
}

// Unwrap returns the underlying transport error, if any.
func (e *AuthError) Unwrap() error {
	return e.Err
}

// wrapSendError returns token and client errors as is and wraps transport errors.
func wrapSendError(msg string, err error) error {
	var authErr *AuthError
	var clientErr *ClientError
	if errors.As(err, &authErr) || errors.As(err, &clientErr) {
		return err
	}
	return &ClientError{Message: fmt.Sprintf("%s: %v", msg, err)}
}

// unexpectedStatusError builds the *APIError for a response with a status the
// caller does not accept.
func unexpectedStatusError(status int, body []byte) *APIError {
	if apiErr, err := ParseAPIErrors(body, status); err == nil && len(apiErr.Errors) > 0 {
		return apiErr
	}
	return &APIError{
		HTTPStatus: status,
		Errors:     []ErrorDetail{{Title: "Unexpected status", Detail: string(body)}},
	}
}

// Sentinel errors of the prices services, matched by *PricesError with errors.Is.
var (
	// ErrPricesBadRequest means the request was refused (400), e.g. a filter value of the wrong format.
	ErrPricesBadRequest = errors.New("didww: prices service rejected the request")
	// ErrPricesNotAuthorized means the token was refused again after it was replaced once (401).
	ErrPricesNotAuthorized = errors.New("didww: prices service refused the token")
	// ErrPricesAccessDenied means the token does not allow the request (403).
	ErrPricesAccessDenied = errors.New("didww: prices service denied access")
	// ErrPricePlanNotAssigned means the account has no price plan for the service
	// (404 price_plan_not_assigned, and any other 404 but price_not_found).
	ErrPricePlanNotAssigned = errors.New("didww: price plan not assigned")
	// ErrPriceNotFound means no price matches the lookup (404 price_not_found).
	ErrPriceNotFound = errors.New("didww: price not found")
	// ErrPricesNotAcceptable means the service cannot answer in the requested format (406).
	ErrPricesNotAcceptable = errors.New("didww: prices service cannot answer in the requested format")
	// ErrPricesTooManyRequests means the request budget is spent (429). It is not retried;
	// PricesError.RetryAfter says when to try again.
	ErrPricesTooManyRequests = errors.New("didww: prices service rate limit exceeded")
	// ErrPricesServiceUnavailable means the service cannot answer yet (503), e.g. a price
	// file that is still being generated. PricesError.RetryAfter says when to try again.
	ErrPricesServiceUnavailable = errors.New("didww: prices service unavailable")
	// ErrPricesUnexpectedStatus means a status the call does not expect, such as a redirect,
	// 202 or 204.
	ErrPricesUnexpectedStatus = errors.New("didww: prices service responded with an unexpected status")
	// ErrPricesConnection means the request failed in transport: connection, timeout, TLS or
	// a broken response body. It is not retried.
	ErrPricesConnection = errors.New("didww: prices service unreachable")
)

// PricesError is returned by the prices services. Use errors.Is with the
// sentinels above to tell the cases apart.
type PricesError struct {
	StatusCode int           // HTTP status; 0 when no response was received
	Code       string        // code of the first error object, e.g. "price_not_found"
	Errors     []ErrorDetail // error objects of the answer, if any
	RetryAfter time.Duration // Retry-After of the answer, when given in seconds
	Err        error         // underlying transport error, if any
	kind       error
}

func (e *PricesError) Error() string {
	messages := make([]string, 0, len(e.Errors))
	for _, d := range e.Errors {
		switch {
		case d.Detail != "":
			messages = append(messages, d.Detail)
		case d.Title != "":
			messages = append(messages, d.Title)
		}
	}
	switch {
	case len(messages) > 0:
		return "didww: prices service: " + strings.Join(messages, "; ")
	case e.Err != nil:
		return "didww: prices service: " + e.Err.Error()
	default:
		return fmt.Sprintf("didww: prices service responded with status %d", e.StatusCode)
	}
}

// Is reports whether target is the sentinel matching this error.
func (e *PricesError) Is(target error) bool {
	return e.kind != nil && target == e.kind
}

// Unwrap returns the underlying transport error, if any.
func (e *PricesError) Unwrap() error {
	return e.Err
}
