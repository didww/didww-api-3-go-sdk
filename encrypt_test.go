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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestKeyPair generates a 2048-bit RSA key pair for testing.
func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return key, string(pubPEM)
}

func TestEncryptWithKeysRoundTrip(t *testing.T) {
	privA, pemA := generateTestKeyPair(t)
	privB, pemB := generateTestKeyPair(t)

	plaintext := []byte("Hello, DIDWW encryption!")
	encrypted, err := EncryptWithKeys(plaintext, pemA, pemB)
	require.NoError(t, err)

	// RSA-OAEP with 2048-bit key produces 256-byte output per key
	rsaBlockSize := 256
	require.Greater(t, len(encrypted), rsaBlockSize*2, "encrypted data too short")

	// Extract the three parts
	encRSAa := encrypted[:rsaBlockSize]
	encRSAb := encrypted[rsaBlockSize : rsaBlockSize*2]
	encryptedAES := encrypted[rsaBlockSize*2:]

	// Decrypt AES credentials with private key A
	aesCredentials, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privA, encRSAa, nil)
	require.NoError(t, err)

	// AES credentials = 32-byte key + 16-byte IV = 48 bytes
	require.Len(t, aesCredentials, 48)

	aesKey := aesCredentials[:32]
	aesIV := aesCredentials[32:]

	// Decrypt AES data
	block, err := aes.NewCipher(aesKey)
	require.NoError(t, err)
	decrypted := make([]byte, len(encryptedAES))
	cipher.NewCBCDecrypter(block, aesIV).CryptBlocks(decrypted, encryptedAES)

	// Remove PKCS7 padding
	padding := int(decrypted[len(decrypted)-1])
	decrypted = decrypted[:len(decrypted)-padding]

	assert.Equal(t, string(plaintext), string(decrypted))

	// Verify key B can also decrypt the same AES credentials
	aesCredentialsB, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privB, encRSAb, nil)
	require.NoError(t, err)
	assert.Equal(t, string(aesCredentials), string(aesCredentialsB))
}

func TestCalculateFingerprint(t *testing.T) {
	_, pemA := generateTestKeyPair(t)
	_, pemB := generateTestKeyPair(t)

	fingerprint := CalculateFingerprint(pemA, pemB)

	// Format: hex_sha1_a:::hex_sha1_b
	assert.Contains(t, fingerprint, ":::")
	parts := strings.Split(fingerprint, ":::")
	require.Len(t, parts, 2)
	// Each SHA-1 hex digest is 40 characters
	for i, part := range parts {
		assert.Len(t, part, 40, "part[%d]", i)
	}
	// Two different keys should have different fingerprints
	assert.NotEqual(t, parts[0], parts[1], "expected different fingerprints for different keys")
}

func TestFingerprintIsConsistent(t *testing.T) {
	_, pemA := generateTestKeyPair(t)
	_, pemB := generateTestKeyPair(t)

	fp1 := CalculateFingerprint(pemA, pemB)
	fp2 := CalculateFingerprint(pemA, pemB)

	assert.Equal(t, fp1, fp2)
}

func TestEncryptWithKeysProducesUniqueOutput(t *testing.T) {
	_, pemA := generateTestKeyPair(t)
	_, pemB := generateTestKeyPair(t)

	plaintext := []byte("Same input")
	enc1, err := EncryptWithKeys(plaintext, pemA, pemB)
	require.NoError(t, err)
	enc2, err := EncryptWithKeys(plaintext, pemA, pemB)
	require.NoError(t, err)

	// Each encryption uses random AES key + IV, so outputs differ
	assert.NotEqual(t, string(enc2), string(enc1))
}

func TestUploadEncryptedFile(t *testing.T) {
	var capturedContentType string
	var capturedBody []byte
	var capturedAuth string
	var capturedAPIVersion string

	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/encrypted_files": {status: http.StatusCreated, fixture: "encrypted_files/create.json"},
	}, func(r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		capturedAuth = r.Header.Get("Api-Key")
		capturedAPIVersion = r.Header.Get("X-DIDWW-API-Version")
		capturedBody, _ = io.ReadAll(r.Body)
	})

	ids, err := server.client.UploadEncryptedFile(
		context.Background(),
		[]byte("encrypted-content"),
		"sample.pdf.enc",
		"fingerprint-123",
		"sample.pdf",
	)
	require.NoError(t, err)

	require.Len(t, ids, 2)
	assert.Equal(t, "6eed102c-66a9-4a9b-a95f-4312d70ec12a", ids[0])
	assert.Equal(t, "371eafbd-ac6a-485c-aadf-9e3c5da37eb4", ids[1])

	// Verify multipart content type
	assert.Contains(t, capturedContentType, "multipart/form-data")

	// Verify API key is sent
	assert.Equal(t, "test-api-key", capturedAuth)
	assert.Equal(t, apiVersion, capturedAPIVersion)

	// Verify form fields are present in the body
	bodyStr := string(capturedBody)
	assert.Contains(t, bodyStr, "fingerprint-123")
	assert.Contains(t, bodyStr, "sample.pdf")
	assert.Contains(t, bodyStr, "sample.pdf.enc")
}

func TestNewEncrypt(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/public_keys": {status: http.StatusOK, fixture: "public_keys/index.json"},
	})

	enc, err := NewEncrypt(context.Background(), client)
	require.NoError(t, err)

	fp := enc.Fingerprint()
	assert.Contains(t, fp, ":::")
}
