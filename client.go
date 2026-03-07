package didww

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to create download request: %v", err)}
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("X-DIDWW-API-Version", apiVersion)

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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to create download request: %v", err)}
	}
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("X-DIDWW-API-Version", apiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("download request failed: %v", err)}
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &ClientError{Message: fmt.Sprintf("download failed: HTTP %d %s", resp.StatusCode, string(body))}
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to create gzip reader: %v", err)}
	}
	defer gz.Close() //nolint:errcheck // best-effort close

	if _, err := io.Copy(dest, gz); err != nil {
		return &ClientError{Message: fmt.Sprintf("failed to decompress export data: %v", err)}
	}
	return nil
}

// --- Repository Accessors ---

func (c *Client) Balance() *SingletonRepository[Balance] { return NewSingletonRepository[Balance](c) }

func (c *Client) Countries() *Repository[Country]          { return NewRepository[Country](c) }
func (c *Client) Regions() *Repository[Region]             { return NewRepository[Region](c) }
func (c *Client) Cities() *Repository[City]                { return NewRepository[City](c) }
func (c *Client) Areas() *Repository[Area]                 { return NewRepository[Area](c) }
func (c *Client) Pops() *Repository[Pop]                   { return NewRepository[Pop](c) }
func (c *Client) VoiceInTrunks() *Repository[VoiceInTrunk] { return NewRepository[VoiceInTrunk](c) }
func (c *Client) VoiceInTrunkGroups() *Repository[VoiceInTrunkGroup] {
	return NewRepository[VoiceInTrunkGroup](c)
}
func (c *Client) VoiceOutTrunks() *Repository[VoiceOutTrunk] { return NewRepository[VoiceOutTrunk](c) }
func (c *Client) DIDs() *Repository[DID]                     { return NewRepository[DID](c) }
func (c *Client) DIDGroups() *Repository[DIDGroup]           { return NewRepository[DIDGroup](c) }
func (c *Client) DIDGroupTypes() *Repository[DIDGroupType]   { return NewRepository[DIDGroupType](c) }
func (c *Client) DIDReservations() *Repository[DIDReservation] {
	return NewRepository[DIDReservation](c)
}
func (c *Client) AvailableDIDs() *Repository[AvailableDID] { return NewRepository[AvailableDID](c) }
func (c *Client) Orders() *Repository[Order]               { return NewRepository[Order](c) }
func (c *Client) Identities() *Repository[Identity]        { return NewRepository[Identity](c) }
func (c *Client) Addresses() *Repository[Address]          { return NewRepository[Address](c) }
func (c *Client) AddressVerifications() *Repository[AddressVerification] {
	return NewRepository[AddressVerification](c)
}
func (c *Client) Proofs() *Repository[Proof]             { return NewRepository[Proof](c) }
func (c *Client) ProofTypes() *Repository[ProofType]     { return NewRepository[ProofType](c) }
func (c *Client) Requirements() *Repository[Requirement] { return NewRepository[Requirement](c) }
func (c *Client) RequirementValidations() *Repository[RequirementValidation] {
	return NewRepository[RequirementValidation](c)
}
func (c *Client) Exports() *Repository[Export]             { return NewRepository[Export](c) }
func (c *Client) CapacityPools() *Repository[CapacityPool] { return NewRepository[CapacityPool](c) }
func (c *Client) SharedCapacityGroups() *Repository[SharedCapacityGroup] {
	return NewRepository[SharedCapacityGroup](c)
}
func (c *Client) PublicKeys() *Repository[PublicKey]         { return NewRepository[PublicKey](c) }
func (c *Client) EncryptedFiles() *Repository[EncryptedFile] { return NewRepository[EncryptedFile](c) }
func (c *Client) SupportingDocumentTemplates() *Repository[SupportingDocumentTemplate] {
	return NewRepository[SupportingDocumentTemplate](c)
}
func (c *Client) PermanentSupportingDocuments() *Repository[PermanentSupportingDocument] {
	return NewRepository[PermanentSupportingDocument](c)
}
func (c *Client) NanpaPrefixes() *Repository[NanpaPrefix] { return NewRepository[NanpaPrefix](c) }
func (c *Client) VoiceOutTrunkRegenerateCredentials() *Repository[VoiceOutTrunkRegenerateCredential] {
	return NewRepository[VoiceOutTrunkRegenerateCredential](c)
}
