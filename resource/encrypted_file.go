package resource

import "time"

// EncryptedFile represents an encrypted file upload.
type EncryptedFile struct {
	ID          string     `json:"-" jsonapi:"encrypted_files"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expires_at" api:"readonly"`
}
