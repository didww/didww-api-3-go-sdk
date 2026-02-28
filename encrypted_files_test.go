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
		"GET /v3/encrypted_files/7f2fbdca-8008-44ce-bcb6-3537ea5efaac": {status: http.StatusOK, fixture: "encrypted_files/index.json"},
	})

	file, err := client.EncryptedFiles().Find(context.Background(), "7f2fbdca-8008-44ce-bcb6-3537ea5efaac")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if file.ID != "7f2fbdca-8008-44ce-bcb6-3537ea5efaac" {
		t.Errorf("expected ID '7f2fbdca-8008-44ce-bcb6-3537ea5efaac', got %q", file.ID)
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
