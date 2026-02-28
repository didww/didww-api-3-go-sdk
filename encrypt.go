package didww

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" //nolint:gosec // SHA-1 required for DIDWW fingerprint protocol
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// Encrypt provides file encryption using DIDWW public keys.
// It uses a hybrid encryption scheme: AES-256-CBC for data encryption
// and RSA-OAEP (SHA-256) for encrypting the AES credentials.
type Encrypt struct {
	client      *Client
	publicKeys  [2]string
	fingerprint string
}

// NewEncrypt creates a new Encrypt instance by fetching public keys from the API.
// The public_keys endpoint does not require an API key.
func NewEncrypt(ctx context.Context, client *Client) (*Encrypt, error) {
	e := &Encrypt{client: client}
	if err := e.Reset(ctx); err != nil {
		return nil, err
	}
	return e, nil
}

// Reset fetches fresh public keys from the API.
func (e *Encrypt) Reset(ctx context.Context) error {
	body, err := e.client.doRequest(ctx, http.MethodGet, "public_keys", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch public keys: %w", err)
	}

	var envelope struct {
		Data []struct {
			Attributes struct {
				Key string `json:"key"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("failed to parse public keys response: %w", err)
	}
	if len(envelope.Data) < 2 {
		return fmt.Errorf("expected at least 2 public keys, got %d", len(envelope.Data))
	}

	e.publicKeys = [2]string{envelope.Data[0].Attributes.Key, envelope.Data[1].Attributes.Key}
	e.fingerprint = CalculateFingerprint(e.publicKeys[0], e.publicKeys[1])
	return nil
}

// Fingerprint returns the fingerprint of the current public keys.
func (e *Encrypt) Fingerprint() string {
	return e.fingerprint
}

// Encrypt encrypts data using the current public keys.
func (e *Encrypt) Encrypt(data []byte) ([]byte, error) {
	return EncryptWithKeys(data, e.publicKeys[0], e.publicKeys[1])
}

// EncryptWithKeys encrypts data using a hybrid RSA-OAEP + AES-256-CBC scheme.
// The output format is: RSA_encrypted_credentials_A | RSA_encrypted_credentials_B | AES_encrypted_data
func EncryptWithKeys(data []byte, publicKeyA, publicKeyB string) ([]byte, error) {
	// Generate random AES-256-CBC key (32 bytes) and IV (16 bytes)
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}
	aesIV := make([]byte, 16)
	if _, err := rand.Read(aesIV); err != nil {
		return nil, fmt.Errorf("failed to generate AES IV: %w", err)
	}

	// Encrypt data with AES-256-CBC + PKCS7 padding
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	paddedData := pkcs7Pad(data, aes.BlockSize)
	encryptedAES := make([]byte, len(paddedData))
	cipher.NewCBCEncrypter(block, aesIV).CryptBlocks(encryptedAES, paddedData)

	// Concatenate AES key + IV as credentials
	aesCredentials := make([]byte, 48)
	copy(aesCredentials, aesKey)
	copy(aesCredentials[32:], aesIV)

	// RSA-OAEP encrypt credentials with each public key
	encRSAa, err := encryptRSAOAEP(publicKeyA, aesCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with key A: %w", err)
	}
	encRSAb, err := encryptRSAOAEP(publicKeyB, aesCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with key B: %w", err)
	}

	// Concatenate: RSA_A + RSA_B + AES_encrypted
	result := make([]byte, 0, len(encRSAa)+len(encRSAb)+len(encryptedAES))
	result = append(result, encRSAa...)
	result = append(result, encRSAb...)
	result = append(result, encryptedAES...)
	return result, nil
}

// CalculateFingerprint calculates the encryption fingerprint from two PEM-encoded public keys.
// Format: SHA1(keyA_der)_hex:::SHA1(keyB_der)_hex
func CalculateFingerprint(publicKeyA, publicKeyB string) string {
	return fingerprintFor(publicKeyA) + ":::" + fingerprintFor(publicKeyB)
}

// UploadEncryptedFile uploads an encrypted file via multipart/form-data POST.
// Returns the list of encrypted file IDs created by the API.
func (c *Client) UploadEncryptedFile(ctx context.Context, encryptedData []byte, fileName, fingerprint, description string) ([]string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := w.WriteField("encrypted_files[encryption_fingerprint]", fingerprint); err != nil {
		return nil, fmt.Errorf("failed to write fingerprint field: %w", err)
	}
	if err := w.WriteField("encrypted_files[items][][description]", description); err != nil {
		return nil, fmt.Errorf("failed to write description field: %w", err)
	}
	part, err := w.CreateFormFile("encrypted_files[items][][file]", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create file part: %w", err)
	}
	if _, writeErr := part.Write(encryptedData); writeErr != nil {
		return nil, fmt.Errorf("failed to write file data: %w", writeErr)
	}
	if closeErr := w.Close(); closeErr != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", closeErr)
	}

	u := c.buildURL("encrypted_files")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &buf)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to create request: %v", err)}
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	if resp.StatusCode >= 400 {
		return nil, &ClientError{Message: fmt.Sprintf("upload failed: HTTP %d %s", resp.StatusCode, string(body))}
	}

	var result struct {
		IDs []string `json:"ids"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("unexpected upload response: %s", string(body))}
	}
	if result.IDs == nil {
		return nil, &ClientError{Message: fmt.Sprintf("unexpected upload response: %s", string(body))}
	}
	return result.IDs, nil
}

func encryptRSAOAEP(publicKeyPEM string, data []byte) ([]byte, error) {
	block, _ := pem.Decode([]byte(normalizePublicKey(publicKeyPEM)))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, data, nil)
}

func fingerprintFor(publicKeyPEM string) string {
	block, _ := pem.Decode([]byte(normalizePublicKey(publicKeyPEM)))
	if block == nil {
		return ""
	}
	digest := sha1.Sum(block.Bytes) //nolint:gosec // SHA-1 required for DIDWW fingerprint protocol
	return hex.EncodeToString(digest[:])
}

func normalizePublicKey(publicKeyPEM string) string {
	s := strings.TrimSpace(publicKeyPEM)
	if !strings.HasPrefix(s, "-----BEGIN") {
		s = "-----BEGIN PUBLIC KEY-----\n" + s + "\n-----END PUBLIC KEY-----"
	}
	return s
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// Ensure Encrypt uses the correct RSA-OAEP hash.
var _ crypto.Hash = crypto.SHA256
