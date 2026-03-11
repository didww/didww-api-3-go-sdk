package didww

import (
	"encoding/json"
	"strings"

	"github.com/didww/didww-api-3-go-sdk/jsonapi"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// Balance represents a DIDWW account balance.
type Balance struct {
	ID           string `json:"-" jsonapi:"balance"`
	TotalBalance string `json:"total_balance"`
	Credit       string `json:"credit"`
	Balance      string `json:"balance"`
}

// Country represents a country resource.
type Country struct {
	ID     string `json:"-" jsonapi:"countries"`
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
	ISO    string `json:"iso"`
	// Resolved relationships
	Regions []*Region `json:"-" rel:"regions"`
}

// Region represents a geographic region.
type Region struct {
	ID   string  `json:"-" jsonapi:"regions"`
	Name string  `json:"name"`
	ISO  *string `json:"iso"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
}

// City represents a city.
type City struct {
	ID   string `json:"-" jsonapi:"cities"`
	Name string `json:"name"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
	Region  *Region  `json:"-" rel:"region"`
	Area    *Area    `json:"-" rel:"area"`
}

// Area represents a geographic area.
type Area struct {
	ID   string `json:"-" jsonapi:"areas"`
	Name string `json:"name"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
}

// Pop represents a Point of Presence.
type Pop struct {
	ID   string `json:"-" jsonapi:"pops"`
	Name string `json:"name"`
}

// TrunkConfiguration is an interface for voice in trunk configurations.
type TrunkConfiguration interface {
	configurationType() string
}

// PSTNConfiguration represents a PSTN trunk configuration.
type PSTNConfiguration struct {
	Dst string `json:"dst"`
}

func (c *PSTNConfiguration) configurationType() string { return "pstn_configurations" }

// SIPConfiguration represents a SIP trunk configuration.
type SIPConfiguration struct {
	Username                   string                          `json:"username,omitempty"`
	Host                       string                          `json:"host,omitempty"`
	Port                       int                             `json:"port,omitempty"`
	CodecIDs                   []enums.Codec                   `json:"codec_ids,omitempty"`
	RxDtmfFormatID             enums.RxDtmfFormat              `json:"rx_dtmf_format_id,omitempty"`
	TxDtmfFormatID             enums.TxDtmfFormat              `json:"tx_dtmf_format_id,omitempty"`
	ResolveRuri                bool                            `json:"resolve_ruri,omitempty"`
	AuthEnabled                bool                            `json:"auth_enabled,omitempty"`
	AuthUser                   string                          `json:"auth_user,omitempty"`
	AuthPassword               string                          `json:"auth_password,omitempty"`
	AuthFromUser               string                          `json:"auth_from_user,omitempty"`
	AuthFromDomain             string                          `json:"auth_from_domain,omitempty"`
	SstEnabled                 bool                            `json:"sst_enabled,omitempty"`
	SstMinTimer                int                             `json:"sst_min_timer,omitempty"`
	SstMaxTimer                int                             `json:"sst_max_timer,omitempty"`
	SstAccept501               bool                            `json:"sst_accept_501,omitempty"`
	SipTimerB                  int                             `json:"sip_timer_b,omitempty"`
	DnsSrvFailoverTimer        int                             `json:"dns_srv_failover_timer,omitempty"`
	RtpPing                    bool                            `json:"rtp_ping,omitempty"`
	RtpTimeout                 int                             `json:"rtp_timeout,omitempty"`
	ForceSymmetricRtp          bool                            `json:"force_symmetric_rtp,omitempty"`
	SymmetricRtpIgnoreRtcp     bool                            `json:"symmetric_rtp_ignore_rtcp,omitempty"`
	ReroutingDisconnectCodeIDs []enums.ReroutingDisconnectCode `json:"rerouting_disconnect_code_ids,omitempty"`
	SstSessionExpires          *int                            `json:"sst_session_expires,omitempty"`
	SstRefreshMethodID         enums.SstRefreshMethod          `json:"sst_refresh_method_id,omitempty"`
	TransportProtocolID        enums.TransportProtocol         `json:"transport_protocol_id,omitempty"`
	MaxTransfers               int                             `json:"max_transfers,omitempty"`
	Max30xRedirects            int                             `json:"max_30x_redirects,omitempty"`
	MediaEncryptionMode        enums.MediaEncryptionMode       `json:"media_encryption_mode,omitempty"`
	StirShakenMode             enums.StirShakenMode            `json:"stir_shaken_mode,omitempty"`
	AllowedRtpIPs              []string                        `json:"allowed_rtp_ips,omitempty"`
}

func (c *SIPConfiguration) configurationType() string { return "sip_configurations" }

// parseTrunkConfiguration deserializes a nested configuration object.
func parseTrunkConfiguration(data []byte) (TrunkConfiguration, error) {
	var env struct {
		Type       string          `json:"type"`
		Attributes json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	switch env.Type {
	case "pstn_configurations":
		var cfg PSTNConfiguration
		if err := json.Unmarshal(env.Attributes, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	case "sip_configurations":
		var cfg SIPConfiguration
		if err := json.Unmarshal(env.Attributes, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	default:
		// Skip unsupported legacy configuration types (iax2, h323, etc.)
		return nil, nil
	}
}

// VoiceInTrunk represents a voice inbound trunk.
type VoiceInTrunk struct {
	ID             string             `json:"-" jsonapi:"voice_in_trunks"`
	Priority       int                `json:"priority,omitempty"`
	CapacityLimit  *int               `json:"capacity_limit,omitempty"`
	Weight         int                `json:"weight,omitempty"`
	Name           string             `json:"name,omitempty"`
	CliFormat      enums.CliFormat    `json:"cli_format,omitempty"`
	CliPrefix      *string            `json:"cli_prefix,omitempty"`
	Description    *string            `json:"description,omitempty"`
	RingingTimeout *int               `json:"ringing_timeout,omitempty"`
	Configuration  TrunkConfiguration `json:"-"`
	CreatedAt      string             `json:"created_at" api:"readonly"`
	// Resolved relationships
	Pop               *Pop               `json:"-" rel:"pop"`
	VoiceInTrunkGroup *VoiceInTrunkGroup `json:"-" rel:"voice_in_trunk_group"`
}

// UnmarshalJSON implements custom unmarshaling for VoiceInTrunk.
func (v *VoiceInTrunk) UnmarshalJSON(data []byte) error {
	type Alias VoiceInTrunk
	aux := &struct {
		*Alias
		RawConfig json.RawMessage `json:"configuration"`
	}{
		Alias: (*Alias)(v),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if len(aux.RawConfig) > 0 && string(aux.RawConfig) != "null" {
		config, err := parseTrunkConfiguration(aux.RawConfig)
		if err != nil {
			return err
		}
		v.Configuration = config
	}
	return nil
}

// MarshalJSON implements custom marshaling for VoiceInTrunk.
func (v VoiceInTrunk) MarshalJSON() ([]byte, error) { //nolint:gocritic // value receiver required for json.Marshal
	type Alias VoiceInTrunk
	aux := &struct {
		Alias
		RawConfig json.RawMessage `json:"configuration,omitempty"`
	}{
		Alias: Alias(v),
	}
	if v.Configuration != nil {
		configData := map[string]any{
			"type": v.Configuration.configurationType(),
		}
		attrs, err := json.Marshal(v.Configuration)
		if err != nil {
			return nil, err
		}
		configData["attributes"] = json.RawMessage(attrs)
		raw, err := json.Marshal(configData)
		if err != nil {
			return nil, err
		}
		aux.RawConfig = raw
	}
	return json.Marshal(aux)
}

// VoiceInTrunkGroup represents a group of voice inbound trunks.
type VoiceInTrunkGroup struct {
	ID            string `json:"-" jsonapi:"voice_in_trunk_groups"`
	Name          string `json:"name,omitempty"`
	CapacityLimit *int   `json:"capacity_limit,omitempty"`
	CreatedAt     string `json:"created_at" api:"readonly"`
	// Relationship IDs for create/update
	VoiceInTrunkIDs []string `json:"-" rel:"voice_in_trunks,voice_in_trunks"`
	// Resolved relationships
	VoiceInTrunks []*VoiceInTrunk `json:"-" rel:"voice_in_trunks"`
}

// VoiceOutTrunk represents a voice outbound trunk.
type VoiceOutTrunk struct {
	ID                  string                    `json:"-" jsonapi:"voice_out_trunks"`
	AllowedSipIPs       []string                  `json:"allowed_sip_ips,omitempty"`
	AllowedRtpIPs       []string                  `json:"allowed_rtp_ips,omitempty"`
	AllowAnyDidAsCli    bool                      `json:"allow_any_did_as_cli,omitempty"`
	Status              enums.VoiceOutTrunkStatus `json:"status" api:"readonly"`
	OnCliMismatchAction enums.OnCliMismatchAction `json:"on_cli_mismatch_action,omitempty"`
	Name                string                    `json:"name,omitempty"`
	CapacityLimit       *int                      `json:"capacity_limit,omitempty"`
	Username            string                    `json:"username" api:"readonly"`
	Password            string                    `json:"password" api:"readonly"`
	CreatedAt           string                    `json:"created_at" api:"readonly"`
	ThresholdReached    bool                      `json:"threshold_reached" api:"readonly"`
	ThresholdAmount     *string                   `json:"threshold_amount,omitempty"`
	MediaEncryptionMode enums.MediaEncryptionMode `json:"media_encryption_mode,omitempty"`
	DefaultDstAction    enums.DefaultDstAction    `json:"default_dst_action,omitempty"`
	DstPrefixes         []string                  `json:"dst_prefixes,omitempty"`
	ForceSymmetricRtp   bool                      `json:"force_symmetric_rtp,omitempty"`
	RtpPing             bool                      `json:"rtp_ping,omitempty"`
	CallbackURL         *string                   `json:"callback_url,omitempty"`
	// Relationship IDs for create/update
	DefaultDIDID string   `json:"-" rel:"default_did,dids"`
	DIDIDs       []string `json:"-" rel:"dids,dids"`
	// Resolved relationships
	DefaultDID *DID   `json:"-" rel:"default_did"`
	DIDs       []*DID `json:"-" rel:"dids"`
}

// DID represents a DID (phone number) resource.
type DID struct {
	ID                     string  `json:"-" jsonapi:"dids"`
	Blocked                bool    `json:"blocked" api:"readonly"`
	CapacityLimit          *int    `json:"capacity_limit"`
	Description            *string `json:"description"`
	Terminated             bool    `json:"terminated"`
	AwaitingRegistration   bool    `json:"awaiting_registration" api:"readonly"`
	CreatedAt              string  `json:"created_at" api:"readonly"`
	BillingCyclesCount     *int    `json:"billing_cycles_count" api:"readonly"`
	Number                 string  `json:"number" api:"readonly"`
	ExpiresAt              string  `json:"expires_at" api:"readonly"`
	ChannelsIncludedCount  int     `json:"channels_included_count" api:"readonly"`
	DedicatedChannelsCount int     `json:"dedicated_channels_count"`
	// Relationship IDs for create/update
	VoiceInTrunkID        string `json:"-" rel:"voice_in_trunk,voice_in_trunks"`
	VoiceInTrunkGroupID   string `json:"-" rel:"voice_in_trunk_group,voice_in_trunk_groups"`
	CapacityPoolID        string `json:"-" rel:"capacity_pool,capacity_pools"`
	SharedCapacityGroupID string `json:"-" rel:"shared_capacity_group,shared_capacity_groups"`
	// Resolved relationships
	Order               *Order               `json:"-" rel:"order"`
	AddressVerification *AddressVerification `json:"-" rel:"address_verification"`
	DIDGroup            *DIDGroup            `json:"-" rel:"did_group"`
	VoiceInTrunk        *VoiceInTrunk        `json:"-" rel:"voice_in_trunk"`
	VoiceInTrunkGroup   *VoiceInTrunkGroup   `json:"-" rel:"voice_in_trunk_group"`
	CapacityPool        *CapacityPool        `json:"-" rel:"capacity_pool"`
	SharedCapacityGroup *SharedCapacityGroup `json:"-" rel:"shared_capacity_group"`
}

// MarshalRelationships implements RelationshipMarshaler for DID.
// Ensures mutual exclusivity: setting a trunk nullifies the trunk group and vice versa.
func (d *DID) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if d.VoiceInTrunkID != "" {
		rels["voice_in_trunk_group"] = jsonapi.NullRelationship()
	}
	if d.VoiceInTrunkGroupID != "" {
		rels["voice_in_trunk"] = jsonapi.NullRelationship()
	}
	if d.CapacityPoolID != "" {
		rels["shared_capacity_group"] = jsonapi.NullRelationship()
	}
	if d.SharedCapacityGroupID != "" {
		rels["capacity_pool"] = jsonapi.NullRelationship()
	}
	return rels, nil
}

// OrderItemAttributes contains the attributes of an order item.
type OrderItemAttributes struct {
	Qty                int     `json:"qty,omitempty"`
	Nrc                string  `json:"nrc,omitempty" api:"readonly"`
	Mrc                string  `json:"mrc,omitempty" api:"readonly"`
	ProratedMrc        bool    `json:"prorated_mrc" api:"readonly"`
	BilledFrom         *string `json:"billed_from" api:"readonly"`
	BilledTo           *string `json:"billed_to" api:"readonly"`
	SetupPrice         string  `json:"setup_price,omitempty" api:"readonly"`
	MonthlyPrice       string  `json:"monthly_price,omitempty" api:"readonly"`
	DIDGroupID         string  `json:"did_group_id,omitempty"`
	SkuID              string  `json:"sku_id,omitempty"`
	AvailableDidID     string  `json:"available_did_id,omitempty"`
	DidReservationID   string  `json:"did_reservation_id,omitempty"`
	CapacityPoolID     string  `json:"capacity_pool_id,omitempty"`
	BillingCyclesCount *int    `json:"billing_cycles_count,omitempty"`
	NanpaPrefixID      string  `json:"nanpa_prefix_id,omitempty"`
}

// MarshalJSON implements custom marshaling for OrderItemAttributes to exclude read-only fields.
func (a OrderItemAttributes) MarshalJSON() ([]byte, error) { //nolint:gocritic // value receiver required for json.Marshal
	type Alias OrderItemAttributes
	return jsonapi.MarshalWritableAttrs(Alias(a))
}

// OrderItem represents an item within an order.
type OrderItem struct {
	Type       string              `json:"type"`
	Attributes OrderItemAttributes `json:"attributes"`
}

// Order represents a DIDWW order.
type Order struct {
	ID                string            `json:"-" jsonapi:"orders"`
	Amount            string            `json:"amount" api:"readonly"`
	Status            enums.OrderStatus `json:"status" api:"readonly"`
	CreatedAt         string            `json:"created_at" api:"readonly"`
	Description       string            `json:"description" api:"readonly"`
	Reference         string            `json:"reference" api:"readonly"`
	Items             []OrderItem       `json:"items"`
	AllowBackOrdering bool              `json:"allow_back_ordering,omitempty"`
	CallbackURL       *string           `json:"callback_url,omitempty"`
	CallbackMethod    *string           `json:"callback_method,omitempty"`
}

// Identity represents a customer identity.
type Identity struct {
	ID                  string             `json:"-" jsonapi:"identities"`
	FirstName           string             `json:"first_name"`
	LastName            string             `json:"last_name"`
	PhoneNumber         string             `json:"phone_number"`
	IDNumber            *string            `json:"id_number"`
	BirthDate           string             `json:"birth_date"`
	CompanyName         *string            `json:"company_name"`
	CompanyRegNumber    *string            `json:"company_reg_number"`
	VatID               *string            `json:"vat_id"`
	Description         *string            `json:"description"`
	PersonalTaxID       *string            `json:"personal_tax_id"`
	IdentityType        enums.IdentityType `json:"identity_type"`
	CreatedAt           string             `json:"created_at" api:"readonly"`
	ExternalReferenceID *string            `json:"external_reference_id"`
	Verified            bool               `json:"verified" api:"readonly"`
	// Relationship IDs for create/update
	CountryID string `json:"-" rel:"country,countries"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
}

// Export represents a CDR export.
type Export struct {
	ID             string                 `json:"-" jsonapi:"exports"`
	Status         enums.ExportStatus     `json:"status" api:"readonly"`
	CreatedAt      string                 `json:"created_at" api:"readonly"`
	URL            *string                `json:"url" api:"readonly"`
	CallbackURL    *string                `json:"callback_url,omitempty"`
	CallbackMethod *string                `json:"callback_method,omitempty"`
	ExportType     enums.ExportType       `json:"export_type"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
}

// DIDGroup represents a DID group.
type DIDGroup struct {
	ID                      string          `json:"-" jsonapi:"did_groups"`
	Prefix                  string          `json:"prefix"`
	Features                []enums.Feature `json:"features"`
	IsMetered               bool            `json:"is_metered"`
	AreaName                string          `json:"area_name"`
	AllowAdditionalChannels bool            `json:"allow_additional_channels"`
	// Resolved relationships
	Country           *Country            `json:"-" rel:"country"`
	City              *City               `json:"-" rel:"city"`
	Region            *Region             `json:"-" rel:"region"`
	DIDGroupType      *DIDGroupType       `json:"-" rel:"did_group_type"`
	StockKeepingUnits []*StockKeepingUnit `json:"-" rel:"stock_keeping_units"`
	Requirement       *Requirement        `json:"-" rel:"requirement"`
}

// DIDGroupType represents a type of DID group.
type DIDGroupType struct {
	ID   string `json:"-" jsonapi:"did_group_types"`
	Name string `json:"name"`
}

// StockKeepingUnit represents an SKU for DID pricing.
type StockKeepingUnit struct {
	ID                    string `json:"-"`
	SetupPrice            string `json:"setup_price"`
	MonthlyPrice          string `json:"monthly_price"`
	ChannelsIncludedCount int    `json:"channels_included_count"`
}

// QtyBasedPricing represents quantity-based pricing for capacity pools.
type QtyBasedPricing struct {
	ID           string `json:"-"`
	SetupPrice   string `json:"setup_price"`
	MonthlyPrice string `json:"monthly_price"`
	Qty          int    `json:"qty"`
}

// CapacityPool represents a capacity pool.
type CapacityPool struct {
	ID                    string `json:"-" jsonapi:"capacity_pools"`
	Name                  string `json:"name,omitempty"`
	RenewDate             string `json:"renew_date" api:"readonly"`
	TotalChannelsCount    int    `json:"total_channels_count"`
	AssignedChannelsCount int    `json:"assigned_channels_count" api:"readonly"`
	MinimumLimit          int    `json:"minimum_limit" api:"readonly"`
	MinimumQtyPerOrder    int    `json:"minimum_qty_per_order" api:"readonly"`
	SetupPrice            string `json:"setup_price" api:"readonly"`
	MonthlyPrice          string `json:"monthly_price" api:"readonly"`
	MeteredRate           string `json:"metered_rate" api:"readonly"`
	// Resolved relationships
	Countries            []*Country             `json:"-" rel:"countries"`
	SharedCapacityGroups []*SharedCapacityGroup `json:"-" rel:"shared_capacity_groups"`
	QtyBasedPricings     []*QtyBasedPricing     `json:"-" rel:"qty_based_pricings"`
}

// SharedCapacityGroup represents a shared capacity group.
type SharedCapacityGroup struct {
	ID                   string `json:"-" jsonapi:"shared_capacity_groups"`
	Name                 string `json:"name"`
	SharedChannelsCount  int    `json:"shared_channels_count"`
	CreatedAt            string `json:"created_at" api:"readonly"`
	MeteredChannelsCount int    `json:"metered_channels_count"`
	// Relationship IDs for create/update
	CapacityPoolID string `json:"-" rel:"capacity_pool,capacity_pools"`
	// Resolved relationships
	CapacityPool *CapacityPool `json:"-" rel:"capacity_pool"`
	DIDs         []*DID        `json:"-" rel:"dids"`
}

// Address represents a customer address.
type Address struct {
	ID          string `json:"-" jsonapi:"addresses"`
	CityName    string `json:"city_name"`
	PostalCode  string `json:"postal_code"`
	Address     string `json:"address"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at" api:"readonly"`
	Verified    bool   `json:"verified" api:"readonly"`
	// Relationship IDs for create/update
	IdentityID string `json:"-" rel:"identity,identities"`
	CountryID  string `json:"-" rel:"country,countries"`
	// Resolved relationships
	Country  *Country  `json:"-" rel:"country"`
	Identity *Identity `json:"-" rel:"identity"`
	Proofs   []*Proof  `json:"-" rel:"proofs"`
}

// AvailableDID represents a DID available for purchase.
type AvailableDID struct {
	ID     string `json:"-" jsonapi:"available_dids"`
	Number string `json:"number"`
	// Resolved relationships
	DIDGroup    *DIDGroup    `json:"-" rel:"did_group"`
	NanpaPrefix *NanpaPrefix `json:"-" rel:"nanpa_prefix"`
}

// DIDReservation represents a reserved DID.
type DIDReservation struct {
	ID          string `json:"-" jsonapi:"did_reservations"`
	ExpireAt    string `json:"expire_at" api:"readonly"`
	CreatedAt   string `json:"created_at" api:"readonly"`
	Description string `json:"description"`
	// Relationship IDs for create/update
	AvailableDIDID string `json:"-" rel:"available_did,available_dids"`
	// Resolved relationships
	AvailableDID *AvailableDID `json:"-" rel:"available_did"`
}

// Proof represents a proof document.
type Proof struct {
	ID        string  `json:"-" jsonapi:"proofs"`
	CreatedAt string  `json:"created_at" api:"readonly"`
	ExpiresAt *string `json:"expires_at" api:"readonly"`
	// Polymorphic entity relationship (type: "identities" or "addresses")
	EntityID   string `json:"-"`
	EntityType string `json:"-"`
	// Other relationship IDs
	ProofTypeID string   `json:"-" rel:"proof_type,proof_types"`
	FileIDs     []string `json:"-" rel:"files,encrypted_files"`
	// Resolved relationships
	ProofType *ProofType `json:"-" rel:"proof_type"`
}

// MarshalRelationships implements RelationshipMarshaler for Proof (polymorphic entity only).
func (p *Proof) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if p.EntityID != "" && p.EntityType != "" {
		rels["entity"] = jsonapi.ToOneRelationship(jsonapi.RelationshipRef{Type: p.EntityType, ID: p.EntityID})
	}
	return rels, nil
}

// UnmarshalRelationships implements RelationshipUnmarshaler for Proof.
// Handles polymorphic entity and proof_type ID extraction from response.
func (p *Proof) UnmarshalRelationships(rels map[string]json.RawMessage) error {
	if raw, ok := rels["entity"]; ok {
		ref, err := jsonapi.ParseToOneRelationship(raw)
		if err != nil {
			return err
		}
		if ref != nil {
			p.EntityID = ref.ID
			p.EntityType = ref.Type
		}
	}
	if raw, ok := rels["proof_type"]; ok {
		ref, err := jsonapi.ParseToOneRelationship(raw)
		if err != nil {
			return err
		}
		if ref != nil {
			p.ProofTypeID = ref.ID
		}
	}
	return nil
}

// PublicKey represents a DIDWW public key for encryption.
type PublicKey struct {
	ID  string `json:"-" jsonapi:"public_keys"`
	Key string `json:"key"`
}

// AddressVerification represents an address verification request.
type AddressVerification struct {
	ID                 string                          `json:"-" jsonapi:"address_verifications"`
	ServiceDescription *string                         `json:"service_description,omitempty"`
	CallbackURL        *string                         `json:"callback_url,omitempty"`
	CallbackMethod     *string                         `json:"callback_method,omitempty"`
	Status             enums.AddressVerificationStatus `json:"status" api:"readonly"`
	RejectReasons      []string                        `json:"reject_reasons" api:"readonly"`
	CreatedAt          string                          `json:"created_at" api:"readonly"`
	Reference          string                          `json:"reference" api:"readonly"`
	// Relationship IDs for create/update
	AddressID string   `json:"-" rel:"address,addresses"`
	DIDIDs    []string `json:"-" rel:"dids,dids"`
	// Resolved relationships
	AddressRel *Address `json:"-" rel:"address"`
}

// UnmarshalJSON splits the semicolon-separated reject_reasons string into a slice.
func (a *AddressVerification) UnmarshalJSON(data []byte) error {
	type Alias AddressVerification
	aux := &struct {
		RejectReasons *string `json:"reject_reasons"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.RejectReasons != nil {
		rawItems := strings.Split(*aux.RejectReasons, "; ")
		a.RejectReasons = make([]string, 0, len(rawItems))
		for _, item := range rawItems {
			if item != "" {
				a.RejectReasons = append(a.RejectReasons, item)
			}
		}
	}
	return nil
}

// EncryptedFile represents an encrypted file upload.
type EncryptedFile struct {
	ID          string  `json:"-" jsonapi:"encrypted_files"`
	Description string  `json:"description"`
	ExpireAt    *string `json:"expire_at" api:"readonly"`
}

// NanpaPrefix represents an NANPA prefix.
type NanpaPrefix struct {
	ID  string `json:"-" jsonapi:"nanpa_prefixes"`
	NPA string `json:"npa"`
	NXX string `json:"nxx"`
	// Resolved relationships
	Country *Country `json:"-" rel:"country"`
	Region  *Region  `json:"-" rel:"region"`
}

// Requirement represents a regulatory requirement.
type Requirement struct {
	ID                         string   `json:"-" jsonapi:"requirements"`
	IdentityType               string   `json:"identity_type"`
	PersonalAreaLevel          string   `json:"personal_area_level"`
	BusinessAreaLevel          string   `json:"business_area_level"`
	AddressAreaLevel           string   `json:"address_area_level"`
	PersonalProofQty           int      `json:"personal_proof_qty"`
	BusinessProofQty           int      `json:"business_proof_qty"`
	AddressProofQty            int      `json:"address_proof_qty"`
	PersonalMandatoryFields    []string `json:"personal_mandatory_fields"`
	BusinessMandatoryFields    []string `json:"business_mandatory_fields"`
	ServiceDescriptionRequired bool     `json:"service_description_required"`
	RestrictionMessage         string   `json:"restriction_message"`
	// Resolved relationships
	Country                   *Country                    `json:"-" rel:"country"`
	DIDGroupType              *DIDGroupType               `json:"-" rel:"did_group_type"`
	PersonalPermanentDocument *SupportingDocumentTemplate `json:"-" rel:"personal_permanent_document"`
	BusinessPermanentDocument *SupportingDocumentTemplate `json:"-" rel:"business_permanent_document"`
	PersonalOnetimeDocument   *SupportingDocumentTemplate `json:"-" rel:"personal_onetime_document"`
	BusinessOnetimeDocument   *SupportingDocumentTemplate `json:"-" rel:"business_onetime_document"`
	PersonalProofTypes        []*ProofType                `json:"-" rel:"personal_proof_types"`
	BusinessProofTypes        []*ProofType                `json:"-" rel:"business_proof_types"`
	AddressProofTypes         []*ProofType                `json:"-" rel:"address_proof_types"`
}

// RequirementValidation represents a requirement validation result.
type RequirementValidation struct {
	ID string `json:"-" jsonapi:"requirement_validations"`
	// Relationship IDs for create
	AddressID     string `json:"-" rel:"address,addresses"`
	IdentityID    string `json:"-" rel:"identity,identities"`
	RequirementID string `json:"-" rel:"requirement,requirements"`
}

// ProofType represents a type of proof document.
type ProofType struct {
	ID         string `json:"-" jsonapi:"proof_types"`
	Name       string `json:"name"`
	EntityType string `json:"entity_type"`
}

// SupportingDocumentTemplate represents a supporting document template.
type SupportingDocumentTemplate struct {
	ID        string `json:"-" jsonapi:"supporting_document_templates"`
	Name      string `json:"name"`
	Permanent bool   `json:"permanent"`
	URL       string `json:"url"`
}

// PermanentSupportingDocument represents a permanent supporting document.
type PermanentSupportingDocument struct {
	ID        string `json:"-" jsonapi:"permanent_supporting_documents"`
	CreatedAt string `json:"created_at" api:"readonly"`
	// Relationship IDs for create/update
	TemplateID string   `json:"-" rel:"template,supporting_document_templates"`
	IdentityID string   `json:"-" rel:"identity,identities"`
	FileIDs    []string `json:"-" rel:"files,encrypted_files"`
	// Resolved relationships
	Template *SupportingDocumentTemplate `json:"-" rel:"template"`
}

// VoiceOutTrunkRegenerateCredential represents a credential regeneration for voice out trunks.
type VoiceOutTrunkRegenerateCredential struct {
	ID string `json:"-" jsonapi:"voice_out_trunk_regenerate_credentials"`
}
