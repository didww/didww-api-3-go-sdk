package didww

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptedFilesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/encrypted_files": {status: http.StatusOK, fixture: "encrypted_files/index.json"},
	})

	files, err := client.EncryptedFiles().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, files, 1)
	assert.Equal(t, "7f2fbdca-8008-44ce-bcb6-3537ea5efaac", files[0].ID)
	assert.Equal(t, "file.enc", files[0].Description)
}

func TestEncryptedFilesFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/encrypted_files/6eed102c-66a9-4a9b-a95f-4312d70ec12a": {status: http.StatusOK, fixture: "encrypted_files/show.json"},
	})

	file, err := client.EncryptedFiles().Find(context.Background(), "6eed102c-66a9-4a9b-a95f-4312d70ec12a")
	require.NoError(t, err)

	assert.Equal(t, "6eed102c-66a9-4a9b-a95f-4312d70ec12a", file.ID)
	assert.Equal(t, "some description", file.Description)
	assert.Equal(t, time.Date(2021, 4, 1, 10, 0, 0, 0, time.UTC), file.CreatedAt)
}

func TestEncryptedFilesFindWithExpiration(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/encrypted_files/371eafbd-ac6a-485c-aadf-9e3c5da37eb4": {status: http.StatusOK, fixture: "encrypted_files/show_with_expiration.json"},
	})

	file, err := client.EncryptedFiles().Find(context.Background(), "371eafbd-ac6a-485c-aadf-9e3c5da37eb4")
	require.NoError(t, err)

	assert.Equal(t, "371eafbd-ac6a-485c-aadf-9e3c5da37eb4", file.ID)
	assert.Equal(t, time.Date(2021, 4, 1, 12, 0, 0, 0, time.UTC), file.CreatedAt)
	require.NotNil(t, file.ExpireAt)
	assert.Equal(t, time.Date(2021, 4, 6, 16, 38, 34, 437000000, time.UTC), *file.ExpireAt)
}

func TestEncryptedFilesDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/encrypted_files/7f2fbdca-8008-44ce-bcb6-3537ea5efaac": {status: http.StatusNoContent},
	})

	err := client.EncryptedFiles().Delete(context.Background(), "7f2fbdca-8008-44ce-bcb6-3537ea5efaac")
	require.NoError(t, err)
}
