package didww

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
)

func TestSmsTrunkRegenerateCredentialsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/sms_trunk_regenerate_credentials": {status: http.StatusCreated, fixture: "sms_trunk_regenerate_credentials/create.json"},
	})

	cred, err := server.client.SmsTrunkRegenerateCredentials().Create(context.Background(), &resource.SmsTrunkRegenerateCredential{
		SmsTrunkID: testSmsTrunkHTTPOutID,
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunk_regenerate_credentials/create_request.json")
	assert.Equal(t, "3c4d5e6f-0001-4c7d-8e9f-0a1b2c3d4e01", cred.ID)
}

func TestSmsTrunkRegenerateCredentialsCreateError(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/sms_trunk_regenerate_credentials": {status: http.StatusUnprocessableEntity, fixture: "sms_trunk_regenerate_credentials/create_error.json"},
	})

	_, err := client.SmsTrunkRegenerateCredentials().Create(context.Background(), &resource.SmsTrunkRegenerateCredential{
		SmsTrunkID: testSmsTrunkHTTPInID,
	})
	require.Error(t, err)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr), "expected *APIError")
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "/data/relationships/sms_trunk", apiErr.Errors[0].Source.Pointer)
}
