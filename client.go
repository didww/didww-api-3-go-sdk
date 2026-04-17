package didww

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/didww/didww-api-3-go-sdk/resource"
)

// Client is the DIDWW API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithEnvironment sets the API environment.
func WithEnvironment(env Environment) ClientOption {
	return func(c *Client) {
		c.baseURL = string(env)
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout sets the HTTP client timeout in milliseconds.
func WithTimeout(ms int) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = time.Duration(ms) * time.Millisecond
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new DIDWW API client.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    string(Sandbox),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// BaseURL returns the base URL of the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// APIKey returns the API key of the client.
func (c *Client) APIKey() string {
	return c.apiKey
}

// DownloadExport downloads an export file from the given URL and writes it to the destination writer.
func (c *Client) DownloadExport(ctx context.Context, downloadURL string, dest io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to create download request: %v", err)}
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("X-DIDWW-API-Version", apiVersion)
	req.Header.Set("User-Agent", "didww-go-sdk/"+sdkVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("download request failed: %v", err)}
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &ClientError{Message: fmt.Sprintf("download failed: HTTP %d %s", resp.StatusCode, string(body))}
	}

	if _, err := io.Copy(dest, resp.Body); err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to write export data: %v", err)}
	}
	return nil
}

// DownloadAndDecompressExport downloads a gzip-compressed export file (.csv.gz) and writes the decompressed CSV to dest.
func (c *Client) DownloadAndDecompressExport(ctx context.Context, downloadURL string, dest io.Writer) error {
	var buf bytes.Buffer
	if err := c.DownloadExport(ctx, downloadURL, &buf); err != nil {
		return err
	}

	gz, err := gzip.NewReader(&buf)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to create gzip reader: %v", err)}
	}
	defer gz.Close() //nolint:errcheck // best-effort close

	const maxDecompressedSize = 1 << 30 // 1 GB
	if _, err := io.Copy(dest, io.LimitReader(gz, maxDecompressedSize)); err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to decompress export data: %v", err)}
	}
	return nil
}

// --- Repository Accessors ---

func (c *Client) Balance() *SingletonRepository[resource.Balance] {
	return NewSingletonRepository[resource.Balance](c)
}

func (c *Client) Countries() *Repository[resource.Country] { return NewRepository[resource.Country](c) }
func (c *Client) Regions() *Repository[resource.Region]    { return NewRepository[resource.Region](c) }
func (c *Client) Cities() *Repository[resource.City]       { return NewRepository[resource.City](c) }
func (c *Client) Areas() *Repository[resource.Area]        { return NewRepository[resource.Area](c) }
func (c *Client) Pops() *Repository[resource.Pop]          { return NewRepository[resource.Pop](c) }
func (c *Client) VoiceInTrunks() *Repository[resource.VoiceInTrunk] {
	return NewRepository[resource.VoiceInTrunk](c)
}
func (c *Client) VoiceInTrunkGroups() *Repository[resource.VoiceInTrunkGroup] {
	return NewRepository[resource.VoiceInTrunkGroup](c)
}
func (c *Client) VoiceOutTrunks() *Repository[resource.VoiceOutTrunk] {
	return NewRepository[resource.VoiceOutTrunk](c)
}
func (c *Client) DIDs() *Repository[resource.DID] { return NewRepository[resource.DID](c) }
func (c *Client) DIDGroups() *Repository[resource.DIDGroup] {
	return NewRepository[resource.DIDGroup](c)
}
func (c *Client) DIDGroupTypes() *Repository[resource.DIDGroupType] {
	return NewRepository[resource.DIDGroupType](c)
}
func (c *Client) DIDReservations() *Repository[resource.DIDReservation] {
	return NewRepository[resource.DIDReservation](c)
}
func (c *Client) AvailableDIDs() *Repository[resource.AvailableDID] {
	return NewRepository[resource.AvailableDID](c)
}
func (c *Client) Orders() *Repository[resource.Order] { return NewRepository[resource.Order](c) }
func (c *Client) Identities() *Repository[resource.Identity] {
	return NewRepository[resource.Identity](c)
}
func (c *Client) Addresses() *Repository[resource.Address] { return NewRepository[resource.Address](c) }
func (c *Client) AddressVerifications() *Repository[resource.AddressVerification] {
	return NewRepository[resource.AddressVerification](c)
}
func (c *Client) Proofs() *Repository[resource.Proof] { return NewRepository[resource.Proof](c) }
func (c *Client) ProofTypes() *Repository[resource.ProofType] {
	return NewRepository[resource.ProofType](c)
}
func (c *Client) AddressRequirements() *Repository[resource.AddressRequirement] {
	return NewRepository[resource.AddressRequirement](c)
}
func (c *Client) AddressRequirementValidations() *Repository[resource.AddressRequirementValidation] {
	return NewRepository[resource.AddressRequirementValidation](c)
}
func (c *Client) Exports() *Repository[resource.Export] { return NewRepository[resource.Export](c) }
func (c *Client) CapacityPools() *Repository[resource.CapacityPool] {
	return NewRepository[resource.CapacityPool](c)
}
func (c *Client) SharedCapacityGroups() *Repository[resource.SharedCapacityGroup] {
	return NewRepository[resource.SharedCapacityGroup](c)
}
func (c *Client) PublicKeys() *Repository[resource.PublicKey] {
	return NewRepository[resource.PublicKey](c)
}
func (c *Client) EncryptedFiles() *Repository[resource.EncryptedFile] {
	return NewRepository[resource.EncryptedFile](c)
}
func (c *Client) SupportingDocumentTemplates() *Repository[resource.SupportingDocumentTemplate] {
	return NewRepository[resource.SupportingDocumentTemplate](c)
}
func (c *Client) PermanentSupportingDocuments() *Repository[resource.PermanentSupportingDocument] {
	return NewRepository[resource.PermanentSupportingDocument](c)
}
func (c *Client) NanpaPrefixes() *Repository[resource.NanpaPrefix] {
	return NewRepository[resource.NanpaPrefix](c)
}
func (c *Client) EmergencyCallingServices() *Repository[resource.EmergencyCallingService] {
	return NewRepository[resource.EmergencyCallingService](c)
}
func (c *Client) EmergencyVerifications() *Repository[resource.EmergencyVerification] {
	return NewRepository[resource.EmergencyVerification](c)
}
func (c *Client) EmergencyRequirementValidations() *Repository[resource.EmergencyRequirementValidation] {
	return NewRepository[resource.EmergencyRequirementValidation](c)
}
func (c *Client) EmergencyRequirements() *Repository[resource.EmergencyRequirement] {
	return NewRepository[resource.EmergencyRequirement](c)
}
func (c *Client) DIDHistory() *Repository[resource.DIDHistory] {
	return NewRepository[resource.DIDHistory](c)
}
func (c *Client) VoiceOutTrunkRegenerateCredentials() *Repository[resource.VoiceOutTrunkRegenerateCredential] {
	return NewRepository[resource.VoiceOutTrunkRegenerateCredential](c)
}
