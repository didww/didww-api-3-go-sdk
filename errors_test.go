package didww

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiErrorParsing(t *testing.T) {
	body := `{
		"errors": [
			{
				"title": "is invalid",
				"detail": "voice_in_trunk_group - is invalid",
				"code": "100",
				"source": {
					"pointer": "/data/attributes/voice_in_trunk_group_id"
				},
				"status": "422"
			}
		]
	}`

	apiErr, err := ParseAPIErrors([]byte(body), http.StatusUnprocessableEntity)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.HTTPStatus)

	require.Len(t, apiErr.Errors, 1)

	e := apiErr.Errors[0]
	assert.Equal(t, "is invalid", e.Title)
	assert.Equal(t, "voice_in_trunk_group - is invalid", e.Detail)
	assert.Equal(t, "100", e.Code)
	assert.Equal(t, "/data/attributes/voice_in_trunk_group_id", e.Source.Pointer)
	assert.Equal(t, "422", e.Status)
}

func TestApiErrorMultipleErrors(t *testing.T) {
	body := `{
		"errors": [
			{
				"title": "can't be blank",
				"detail": "name - can't be blank",
				"source": {"pointer": "/data/attributes/name"},
				"status": "422"
			},
			{
				"title": "is invalid",
				"detail": "configuration - is invalid",
				"source": {"pointer": "/data/attributes/configuration"},
				"status": "422"
			}
		]
	}`

	apiErr, err := ParseAPIErrors([]byte(body), http.StatusUnprocessableEntity)
	require.NoError(t, err)

	require.Len(t, apiErr.Errors, 2)

	assert.Equal(t, "name - can't be blank", apiErr.Errors[0].Detail)
	assert.Equal(t, "configuration - is invalid", apiErr.Errors[1].Detail)
}

func TestApiErrorWithoutCode(t *testing.T) {
	body := `{
		"errors": [
			{
				"title": "not found",
				"detail": "Resource not found",
				"status": "404"
			}
		]
	}`

	apiErr, err := ParseAPIErrors([]byte(body), http.StatusNotFound)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, apiErr.HTTPStatus)

	assert.Equal(t, "", apiErr.Errors[0].Code)
}

func TestApiErrorWithoutSourcePointer(t *testing.T) {
	body := `{
		"errors": [
			{
				"title": "server error",
				"detail": "Internal server error",
				"status": "500"
			}
		]
	}`

	apiErr, err := ParseAPIErrors([]byte(body), http.StatusInternalServerError)
	require.NoError(t, err)

	assert.Equal(t, "", apiErr.Errors[0].Source.Pointer)
}

func TestApiErrorEmptyBody(t *testing.T) {
	_, err := ParseAPIErrors([]byte(""), http.StatusInternalServerError)
	require.Error(t, err)
}

func TestApiErrorInvalidJSON(t *testing.T) {
	_, err := ParseAPIErrors([]byte("not json"), http.StatusInternalServerError)
	require.Error(t, err)
}

func TestApiErrorImplementsError(t *testing.T) {
	apiErr := &APIError{
		HTTPStatus: 422,
		Errors: []ErrorDetail{
			{Title: "is invalid", Detail: "name - is invalid"},
		},
	}

	errMsg := apiErr.Error()
	require.NotEmpty(t, errMsg)
}

func TestClientError(t *testing.T) {
	err := &ClientError{Message: "connection timeout"}
	assert.Equal(t, "connection timeout", err.Error())
}

func TestErrorDetailJSONRoundTrip(t *testing.T) {
	original := ErrorDetail{
		Title:  "is invalid",
		Detail: "name - is invalid",
		Code:   "100",
		Status: "422",
		Source: ErrorSource{Pointer: "/data/attributes/name"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ErrorDetail
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Title, decoded.Title)
	assert.Equal(t, original.Detail, decoded.Detail)
	assert.Equal(t, original.Code, decoded.Code)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.Source.Pointer, decoded.Source.Pointer)
}
