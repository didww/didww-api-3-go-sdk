package didww

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"strings"
	"testing"
)

// generateTestKeyPair generates a 2048-bit RSA key pair for testing.
func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return key, string(pubPEM)
}

func TestEncryptWithKeysRoundTrip(t *testing.T) {
	privA, pemA := generateTestKeyPair(t)
	privB, pemB := generateTestKeyPair(t)

	plaintext := []byte("Hello, DIDWW encryption!")
	encrypted, err := EncryptWithKeys(plaintext, pemA, pemB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// RSA-OAEP with 2048-bit key produces 256-byte output per key
	rsaBlockSize := 256
	if len(encrypted) <= rsaBlockSize*2 {
		t.Fatalf("encrypted data too short: %d bytes", len(encrypted))
	}

	// Extract the three parts
	encRSAa := encrypted[:rsaBlockSize]
	encRSAb := encrypted[rsaBlockSize : rsaBlockSize*2]
	encryptedAES := encrypted[rsaBlockSize*2:]

	// Decrypt AES credentials with private key A
	aesCredentials, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privA, encRSAa, nil)
	if err != nil {
		t.Fatalf("failed to decrypt RSA A: %v", err)
	}

	// AES credentials = 32-byte key + 16-byte IV = 48 bytes
	if len(aesCredentials) != 48 {
		t.Fatalf("expected 48-byte AES credentials, got %d", len(aesCredentials))
	}

	aesKey := aesCredentials[:32]
	aesIV := aesCredentials[32:]

	// Decrypt AES data
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		t.Fatalf("failed to create AES cipher: %v", err)
	}
	decrypted := make([]byte, len(encryptedAES))
	cipher.NewCBCDecrypter(block, aesIV).CryptBlocks(decrypted, encryptedAES)

	// Remove PKCS7 padding
	padding := int(decrypted[len(decrypted)-1])
	decrypted = decrypted[:len(decrypted)-padding]

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted %q, want %q", decrypted, plaintext)
	}

	// Verify key B can also decrypt the same AES credentials
	aesCredentialsB, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privB, encRSAb, nil)
	if err != nil {
		t.Fatalf("failed to decrypt RSA B: %v", err)
	}
	if string(aesCredentialsB) != string(aesCredentials) {
		t.Error("key B produced different AES credentials than key A")
	}
}

func TestCalculateFingerprint(t *testing.T) {
	_, pemA := generateTestKeyPair(t)
	_, pemB := generateTestKeyPair(t)

	fingerprint := CalculateFingerprint(pemA, pemB)

	// Format: hex_sha1_a:::hex_sha1_b
	if !strings.Contains(fingerprint, ":::") {
		t.Fatalf("expected fingerprint to contain ':::', got %q", fingerprint)
	}
	parts := strings.Split(fingerprint, ":::")
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(parts))
	}
	// Each SHA-1 hex digest is 40 characters
	for i, part := range parts {
		if len(part) != 40 {
			t.Errorf("part[%d] length = %d, want 40", i, len(part))
		}
	}
	// Two different keys should have different fingerprints
	if parts[0] == parts[1] {
		t.Error("expected different fingerprints for different keys")
	}
}

func TestFingerprintIsConsistent(t *testing.T) {
	_, pemA := generateTestKeyPair(t)
	_, pemB := generateTestKeyPair(t)

	fp1 := CalculateFingerprint(pemA, pemB)
	fp2 := CalculateFingerprint(pemA, pemB)

	if fp1 != fp2 {
		t.Errorf("fingerprints differ: %q vs %q", fp1, fp2)
	}
}

func TestEncryptWithKeysProducesUniqueOutput(t *testing.T) {
	_, pemA := generateTestKeyPair(t)
	_, pemB := generateTestKeyPair(t)

	plaintext := []byte("Same input")
	enc1, err := EncryptWithKeys(plaintext, pemA, pemB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	enc2, err := EncryptWithKeys(plaintext, pemA, pemB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Each encryption uses random AES key + IV, so outputs differ
	if string(enc1) == string(enc2) {
		t.Error("expected different encrypted outputs for same plaintext")
	}
}

func TestUploadEncryptedFile(t *testing.T) {
	var capturedContentType string
	var capturedBody []byte
	var capturedAuth string

	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/encrypted_files": {status: http.StatusCreated, fixture: "encrypted_files/create.json"},
	}, func(r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		capturedAuth = r.Header.Get("Api-Key")
		capturedBody, _ = io.ReadAll(r.Body)
	})

	ids, err := server.client.UploadEncryptedFile(
		context.Background(),
		[]byte("encrypted-content"),
		"sample.pdf.enc",
		"fingerprint-123",
		"sample.pdf",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(ids))
	}
	if ids[0] != "6eed102c-66a9-4a9b-a95f-4312d70ec12a" {
		t.Errorf("unexpected first ID: %q", ids[0])
	}
	if ids[1] != "371eafbd-ac6a-485c-aadf-9e3c5da37eb4" {
		t.Errorf("unexpected second ID: %q", ids[1])
	}

	// Verify multipart content type
	if !strings.HasPrefix(capturedContentType, "multipart/form-data") {
		t.Errorf("expected multipart/form-data content type, got %q", capturedContentType)
	}

	// Verify API key is sent
	if capturedAuth != "test-api-key" {
		t.Errorf("expected Api-Key 'test-api-key', got %q", capturedAuth)
	}

	// Verify form fields are present in the body
	bodyStr := string(capturedBody)
	if !strings.Contains(bodyStr, "fingerprint-123") {
		t.Error("expected body to contain fingerprint")
	}
	if !strings.Contains(bodyStr, "sample.pdf") {
		t.Error("expected body to contain description")
	}
	if !strings.Contains(bodyStr, "sample.pdf.enc") {
		t.Error("expected body to contain filename")
	}
}

func TestNewEncrypt(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/public_keys": {status: http.StatusOK, fixture: "public_keys/index.json"},
	})

	enc, err := NewEncrypt(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fp := enc.Fingerprint()
	if !strings.Contains(fp, ":::") {
		t.Errorf("expected fingerprint to contain ':::', got %q", fp)
	}
}
