package didww

import (
	"fmt"
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

// --- Repository Accessors ---

func (c *Client) Balance() *SingletonRepository[Balance] {
	return &SingletonRepository[Balance]{client: c, resourcePath: "balance"}
}

func (c *Client) Countries() *Repository[Country] {
	return &Repository[Country]{client: c, resourcePath: "countries", resourceType: "countries"}
}

func (c *Client) Regions() *Repository[Region] {
	return &Repository[Region]{client: c, resourcePath: "regions", resourceType: "regions"}
}

func (c *Client) Cities() *Repository[City] {
	return &Repository[City]{client: c, resourcePath: "cities", resourceType: "cities"}
}

func (c *Client) Areas() *Repository[Area] {
	return &Repository[Area]{client: c, resourcePath: "areas", resourceType: "areas"}
}

func (c *Client) Pops() *Repository[Pop] {
	return &Repository[Pop]{client: c, resourcePath: "pops", resourceType: "pops"}
}

func (c *Client) VoiceInTrunks() *Repository[VoiceInTrunk] {
	return &Repository[VoiceInTrunk]{client: c, resourcePath: "voice_in_trunks", resourceType: "voice_in_trunks"}
}

func (c *Client) VoiceInTrunkGroups() *Repository[VoiceInTrunkGroup] {
	return &Repository[VoiceInTrunkGroup]{client: c, resourcePath: "voice_in_trunk_groups", resourceType: "voice_in_trunk_groups"}
}

func (c *Client) VoiceOutTrunks() *Repository[VoiceOutTrunk] {
	return &Repository[VoiceOutTrunk]{client: c, resourcePath: "voice_out_trunks", resourceType: "voice_out_trunks"}
}

func (c *Client) DIDs() *Repository[DID] {
	return &Repository[DID]{client: c, resourcePath: "dids", resourceType: "dids"}
}

func (c *Client) DIDGroups() *Repository[DIDGroup] {
	return &Repository[DIDGroup]{client: c, resourcePath: "did_groups", resourceType: "did_groups"}
}

func (c *Client) DIDGroupTypes() *Repository[DIDGroupType] {
	return &Repository[DIDGroupType]{client: c, resourcePath: "did_group_types", resourceType: "did_group_types"}
}

func (c *Client) DIDReservations() *Repository[DIDReservation] {
	return &Repository[DIDReservation]{client: c, resourcePath: "did_reservations", resourceType: "did_reservations"}
}

func (c *Client) AvailableDIDs() *Repository[AvailableDID] {
	return &Repository[AvailableDID]{client: c, resourcePath: "available_dids", resourceType: "available_dids"}
}

func (c *Client) Orders() *Repository[Order] {
	return &Repository[Order]{client: c, resourcePath: "orders", resourceType: "orders"}
}

func (c *Client) Identities() *Repository[Identity] {
	return &Repository[Identity]{client: c, resourcePath: "identities", resourceType: "identities"}
}

func (c *Client) Addresses() *Repository[Address] {
	return &Repository[Address]{client: c, resourcePath: "addresses", resourceType: "addresses"}
}

func (c *Client) AddressVerifications() *Repository[AddressVerification] {
	return &Repository[AddressVerification]{client: c, resourcePath: "address_verifications", resourceType: "address_verifications"}
}

func (c *Client) Proofs() *Repository[Proof] {
	return &Repository[Proof]{client: c, resourcePath: "proofs", resourceType: "proofs"}
}

func (c *Client) ProofTypes() *Repository[ProofType] {
	return &Repository[ProofType]{client: c, resourcePath: "proof_types", resourceType: "proof_types"}
}

func (c *Client) Requirements() *Repository[Requirement] {
	return &Repository[Requirement]{client: c, resourcePath: "requirements", resourceType: "requirements"}
}

func (c *Client) RequirementValidations() *Repository[RequirementValidation] {
	return &Repository[RequirementValidation]{client: c, resourcePath: "requirement_validations", resourceType: "requirement_validations"}
}

func (c *Client) Exports() *Repository[Export] {
	return &Repository[Export]{client: c, resourcePath: "exports", resourceType: "exports"}
}

func (c *Client) CapacityPools() *Repository[CapacityPool] {
	return &Repository[CapacityPool]{client: c, resourcePath: "capacity_pools", resourceType: "capacity_pools"}
}

func (c *Client) SharedCapacityGroups() *Repository[SharedCapacityGroup] {
	return &Repository[SharedCapacityGroup]{client: c, resourcePath: "shared_capacity_groups", resourceType: "shared_capacity_groups"}
}

func (c *Client) PublicKeys() *SingletonRepository[PublicKey] {
	return &SingletonRepository[PublicKey]{client: c, resourcePath: "public_keys"}
}

func (c *Client) EncryptedFiles() *Repository[EncryptedFile] {
	return &Repository[EncryptedFile]{client: c, resourcePath: "encrypted_files", resourceType: "encrypted_files"}
}

func (c *Client) SupportingDocumentTemplates() *Repository[SupportingDocumentTemplate] {
	return &Repository[SupportingDocumentTemplate]{client: c, resourcePath: "supporting_document_templates", resourceType: "supporting_document_templates"}
}

func (c *Client) PermanentSupportingDocuments() *Repository[PermanentSupportingDocument] {
	return &Repository[PermanentSupportingDocument]{client: c, resourcePath: "permanent_supporting_documents", resourceType: "permanent_supporting_documents"}
}

func (c *Client) NanpaPrefixes() *Repository[NanpaPrefix] {
	return &Repository[NanpaPrefix]{client: c, resourcePath: "nanpa_prefixes", resourceType: "nanpa_prefixes"}
}

func (c *Client) VoiceOutTrunkRegenerateCredentials() *Repository[VoiceOutTrunkRegenerateCredential] {
	return &Repository[VoiceOutTrunkRegenerateCredential]{client: c, resourcePath: "voice_out_trunk_regenerate_credentials", resourceType: "voice_out_trunk_regenerate_credentials"}
}
