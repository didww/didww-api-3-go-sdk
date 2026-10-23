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
	var capturedHeader http.Header

	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/encrypted_files": {status: http.StatusCreated, fixture: "encrypted_files/create.json"},
	}, func(r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		capturedHeader = r.Header.Clone()
		capturedBody, _ = io.ReadAll(r.Body)
	})

	id, err := server.client.UploadEncryptedFile(
		context.Background(),
		[]byte("encrypted-content"),
		"sample.pdf.enc",
		"fingerprint-123",
		"sample.pdf",
	)
	require.NoError(t, err)

	assert.Equal(t, "6eed102c-66a9-4a9b-a95f-4312d70ec12a", id)

	// Verify multipart content type
	assert.Contains(t, capturedContentType, "multipart/form-data")

	// Verify the bearer token is sent instead of the API key
	assert.Equal(t, "Bearer "+testAccessToken, capturedHeader.Get("Authorization"))
	assert.NotContains(t, capturedHeader, "Api-Key")
	assert.NotContains(t, capturedHeader, "X-Didww-Api-Version")
	assert.Equal(t, "didww-go-sdk/"+sdkVersion, capturedHeader.Get("User-Agent"))

	// Verify form fields are present in the body
	bodyStr := string(capturedBody)
	assert.Contains(t, bodyStr, "fingerprint-123")
	assert.Contains(t, bodyStr, "sample.pdf")
	assert.Contains(t, bodyStr, "sample.pdf.enc")
}

func TestUploadEncryptedFileRetriedAfter401(t *testing.T) {
	var log requestLog
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		if r.Header.Get("Authorization") == "Bearer token-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusCreated)
		w.Write(loadFixture(t, "encrypted_files/create.json"))
	})
	client := newTestClient(t, srv.URL)

	id, err := client.UploadEncryptedFile(context.Background(), []byte("encrypted-content"), "sample.pdf.enc", "fingerprint-123", "sample.pdf")
	require.NoError(t, err)

	assert.Equal(t, "6eed102c-66a9-4a9b-a95f-4312d70ec12a", id)
	requests := log.all()
	require.Len(t, requests, 2)
	assert.Contains(t, requests[0].body, "encrypted-content")
	assert.Equal(t, requests[0].body, requests[1].body)
	assert.Equal(t, "Bearer token-1", requests[0].authorization)
	assert.Equal(t, "Bearer token-2", requests[1].authorization)
	assert.Equal(t, 2, srv.mints())
	assert.JSONEq(t, `{"aud":["api.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(0)))
	assert.JSONEq(t, `{"aud":["api.didww.com"],"api_version":"2026-10-22"}`, string(srv.mintBody(1)))
}

func TestUploadEncryptedFileSecond401(t *testing.T) {
	var log requestLog
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		log.record(r)
		w.WriteHeader(http.StatusUnauthorized)
	})
	client := newTestClient(t, srv.URL)

	_, err := client.UploadEncryptedFile(context.Background(), []byte("encrypted-content"), "sample.pdf.enc", "fingerprint-123", "sample.pdf")

	var clientErr *ClientError
	require.ErrorAs(t, err, &clientErr)
	assert.Contains(t, clientErr.Message, "HTTP 401")
	assert.Len(t, log.all(), 2)
	assert.Equal(t, 2, srv.mints())
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

func TestEncryptEncryptMethod(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/public_keys": {status: http.StatusOK, fixture: "public_keys/index.json"},
	})

	enc, err := NewEncrypt(context.Background(), client)
	require.NoError(t, err)

	plaintext := []byte("test data for encryption")
	encrypted, err := enc.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// RSA-OAEP with 2048-bit key produces 256-byte output per key
	// Encrypted data should be larger than 2 * 256 bytes (two RSA blocks + AES data)
	assert.Greater(t, len(encrypted), 512)
}
