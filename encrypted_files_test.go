package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestEncryptedFilesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/encrypted_files": {status: http.StatusOK, fixture: "encrypted_files/index.json"},
	})

	files, err := client.EncryptedFiles().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 encrypted file, got %d", len(files))
	}
	if files[0].ID != "7f2fbdca-8008-44ce-bcb6-3537ea5efaac" {
		t.Errorf("expected ID '7f2fbdca-8008-44ce-bcb6-3537ea5efaac', got %q", files[0].ID)
	}
	if files[0].Description != "file.enc" {
		t.Errorf("expected Description 'file.enc', got %q", files[0].Description)
	}
}

func TestEncryptedFilesFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/encrypted_files/6eed102c-66a9-4a9b-a95f-4312d70ec12a": {status: http.StatusOK, fixture: "encrypted_files/show.json"},
	})

	file, err := client.EncryptedFiles().Find(context.Background(), "6eed102c-66a9-4a9b-a95f-4312d70ec12a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if file.ID != "6eed102c-66a9-4a9b-a95f-4312d70ec12a" {
		t.Errorf("expected ID '6eed102c-66a9-4a9b-a95f-4312d70ec12a', got %q", file.ID)
	}
	if file.Description != "some description" {
		t.Errorf("expected Description 'some description', got %q", file.Description)
	}
}

func TestEncryptedFilesFindWithExpiration(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/encrypted_files/371eafbd-ac6a-485c-aadf-9e3c5da37eb4": {status: http.StatusOK, fixture: "encrypted_files/show_with_expiration.json"},
	})

	file, err := client.EncryptedFiles().Find(context.Background(), "371eafbd-ac6a-485c-aadf-9e3c5da37eb4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if file.ID != "371eafbd-ac6a-485c-aadf-9e3c5da37eb4" {
		t.Errorf("expected ID '371eafbd-ac6a-485c-aadf-9e3c5da37eb4', got %q", file.ID)
	}
	if file.ExpireAt == nil || *file.ExpireAt != "2021-04-06T16:38:34.437Z" {
		t.Errorf("expected ExpireAt '2021-04-06T16:38:34.437Z', got %v", file.ExpireAt)
	}
}

func TestEncryptedFilesDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/encrypted_files/7f2fbdca-8008-44ce-bcb6-3537ea5efaac": {status: http.StatusNoContent},
	})

	err := client.EncryptedFiles().Delete(context.Background(), "7f2fbdca-8008-44ce-bcb6-3537ea5efaac")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
