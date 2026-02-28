package didww

import (
	"encoding/json"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

// Balance represents a DIDWW account balance.
type Balance struct {
	ID           string `json:"-"`
	TotalBalance string `json:"total_balance"`
	Credit       string `json:"credit"`
	Balance      string `json:"balance"`
}

// Country represents a country resource.
type Country struct {
	ID     string `json:"-"`
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
	ISO    string `json:"iso"`
	// Resolved relationships
	Regions []*Region `json:"-"`
}

// ResolveRelationships resolves included relationships for Country.
func (c *Country) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if regions, err := ResolveToMany[Region](included, rels, "regions"); err != nil {
		return err
	} else if regions != nil {
		c.Regions = regions
	}
	return nil
}

// Region represents a geographic region.
type Region struct {
	ID   string  `json:"-"`
	Name string  `json:"name"`
	ISO  *string `json:"iso"`
	// Resolved relationships
	Country *Country `json:"-"`
}

// ResolveRelationships resolves included relationships for Region.
func (r *Region) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if country, err := ResolveToOne[Country](included, rels, "country"); err != nil {
		return err
	} else if country != nil {
		r.Country = country
	}
	return nil
}

// City represents a city.
type City struct {
	ID   string `json:"-"`
	Name string `json:"name"`
	// Resolved relationships
	Country *Country `json:"-"`
	Region  *Region  `json:"-"`
	Area    *Area    `json:"-"`
}

// ResolveRelationships resolves included relationships for City.
func (c *City) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if country, err := ResolveToOne[Country](included, rels, "country"); err != nil {
		return err
	} else if country != nil {
		c.Country = country
	}
	if region, err := ResolveToOne[Region](included, rels, "region"); err != nil {
		return err
	} else if region != nil {
		c.Region = region
	}
	if area, err := ResolveToOne[Area](included, rels, "area"); err != nil {
		return err
	} else if area != nil {
		c.Area = area
	}
	return nil
}

// Area represents a geographic area.
type Area struct {
	ID   string `json:"-"`
	Name string `json:"name"`
	// Resolved relationships
	Country *Country `json:"-"`
}

// ResolveRelationships resolves included relationships for Area.
func (a *Area) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if country, err := ResolveToOne[Country](included, rels, "country"); err != nil {
		return err
	} else if country != nil {
		a.Country = country
	}
	return nil
}

// Pop represents a Point of Presence.
type Pop struct {
	ID   string `json:"-"`
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
	ID             string             `json:"-"`
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
	Pop               *Pop               `json:"-"`
	VoiceInTrunkGroup *VoiceInTrunkGroup `json:"-"`
}

// UnmarshalJSON implements custom unmarshalling for VoiceInTrunk.
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

// MarshalJSON implements custom marshalling for VoiceInTrunk.
func (v VoiceInTrunk) MarshalJSON() ([]byte, error) {
	type Alias VoiceInTrunk
	aux := &struct {
		Alias
		RawConfig json.RawMessage `json:"configuration,omitempty"`
	}{
		Alias: (Alias)(v),
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

// ResolveRelationships resolves included relationships for VoiceInTrunk.
func (v *VoiceInTrunk) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if pop, err := ResolveToOne[Pop](included, rels, "pop"); err != nil {
		return err
	} else if pop != nil {
		v.Pop = pop
	}
	if group, err := ResolveToOne[VoiceInTrunkGroup](included, rels, "voice_in_trunk_group"); err != nil {
		return err
	} else if group != nil {
		v.VoiceInTrunkGroup = group
	}
	return nil
}

// VoiceInTrunkGroup represents a group of voice inbound trunks.
type VoiceInTrunkGroup struct {
	ID            string `json:"-"`
	Name          string `json:"name,omitempty"`
	CapacityLimit *int   `json:"capacity_limit,omitempty"`
	CreatedAt     string `json:"created_at" api:"readonly"`
	// Relationship IDs for create/update
	VoiceInTrunkIDs []string `json:"-"`
	// Resolved relationships
	VoiceInTrunks []*VoiceInTrunk `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for VoiceInTrunkGroup.
func (g *VoiceInTrunkGroup) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if len(g.VoiceInTrunkIDs) > 0 {
		refs := make([]RelationshipRef, len(g.VoiceInTrunkIDs))
		for i, id := range g.VoiceInTrunkIDs {
			refs[i] = RelationshipRef{Type: "voice_in_trunks", ID: id}
		}
		rels["voice_in_trunks"] = ToManyRelationship(refs)
	}
	return rels, nil
}

// ResolveRelationships resolves included relationships for VoiceInTrunkGroup.
func (g *VoiceInTrunkGroup) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if trunks, err := ResolveToMany[VoiceInTrunk](included, rels, "voice_in_trunks"); err != nil {
		return err
	} else if trunks != nil {
		g.VoiceInTrunks = trunks
	}
	return nil
}

// VoiceOutTrunk represents a voice outbound trunk.
type VoiceOutTrunk struct {
	ID                  string                    `json:"-"`
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
	DefaultDIDID string   `json:"-"`
	DIDIDs       []string `json:"-"`
	// Resolved relationships
	DefaultDID *DID   `json:"-"`
	DIDs       []*DID `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for VoiceOutTrunk.
func (v *VoiceOutTrunk) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if v.DefaultDIDID != "" {
		rels["default_did"] = ToOneRelationship(RelationshipRef{Type: "dids", ID: v.DefaultDIDID})
	}
	if len(v.DIDIDs) > 0 {
		refs := make([]RelationshipRef, len(v.DIDIDs))
		for i, id := range v.DIDIDs {
			refs[i] = RelationshipRef{Type: "dids", ID: id}
		}
		rels["dids"] = ToManyRelationship(refs)
	}
	return rels, nil
}

// ResolveRelationships resolves included relationships for VoiceOutTrunk.
func (v *VoiceOutTrunk) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if defaultDid, err := ResolveToOne[DID](included, rels, "default_did"); err != nil {
		return err
	} else if defaultDid != nil {
		v.DefaultDID = defaultDid
	}
	if dids, err := ResolveToMany[DID](included, rels, "dids"); err != nil {
		return err
	} else if dids != nil {
		v.DIDs = dids
	}
	return nil
}

// DID represents a DID (phone number) resource.
type DID struct {
	ID                     string  `json:"-"`
	Blocked                bool    `json:"blocked" api:"readonly"`
	CapacityLimit          *int    `json:"capacity_limit"`
	Description            *string `json:"description"`
	Terminated             bool    `json:"terminated" api:"readonly"`
	AwaitingRegistration   bool    `json:"awaiting_registration" api:"readonly"`
	CreatedAt              string  `json:"created_at" api:"readonly"`
	BillingCyclesCount     *int    `json:"billing_cycles_count" api:"readonly"`
	Number                 string  `json:"number" api:"readonly"`
	ExpiresAt              string  `json:"expires_at" api:"readonly"`
	ChannelsIncludedCount  int     `json:"channels_included_count" api:"readonly"`
	DedicatedChannelsCount int     `json:"dedicated_channels_count"`
	// Resolved relationships
	Order               *Order               `json:"-"`
	AddressVerification *AddressVerification `json:"-"`
	DIDGroup            *DIDGroup            `json:"-"`
}

// ResolveRelationships resolves included relationships for DID.
func (d *DID) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if order, err := ResolveToOne[Order](included, rels, "order"); err != nil {
		return err
	} else if order != nil {
		d.Order = order
	}
	if av, err := ResolveToOne[AddressVerification](included, rels, "address_verification"); err != nil {
		return err
	} else if av != nil {
		d.AddressVerification = av
	}
	if dg, err := ResolveToOne[DIDGroup](included, rels, "did_group"); err != nil {
		return err
	} else if dg != nil {
		d.DIDGroup = dg
	}
	return nil
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

// MarshalJSON implements custom marshalling for OrderItemAttributes to exclude read-only fields.
func (a OrderItemAttributes) MarshalJSON() ([]byte, error) {
	type Alias OrderItemAttributes
	return marshalWritableAttrs(Alias(a))
}

// OrderItem represents an item within an order.
type OrderItem struct {
	Type       string              `json:"type"`
	Attributes OrderItemAttributes `json:"attributes"`
}

// Order represents a DIDWW order.
type Order struct {
	ID                string            `json:"-"`
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
	ID                  string             `json:"-"`
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
	CountryID string `json:"-"`
	// Resolved relationships
	Country *Country `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for Identity.
func (i *Identity) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if i.CountryID != "" {
		rels["country"] = ToOneRelationship(RelationshipRef{Type: "countries", ID: i.CountryID})
	}
	return rels, nil
}

// ResolveRelationships resolves included relationships for Identity.
func (i *Identity) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if country, err := ResolveToOne[Country](included, rels, "country"); err != nil {
		return err
	} else if country != nil {
		i.Country = country
	}
	return nil
}

// Export represents a CDR export.
type Export struct {
	ID             string                 `json:"-"`
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
	ID                      string          `json:"-"`
	Prefix                  string          `json:"prefix"`
	LocalPrefix             string          `json:"local_prefix"`
	Features                []enums.Feature `json:"features"`
	IsMetered               bool            `json:"is_metered"`
	AreaName                string          `json:"area_name"`
	AllowAdditionalChannels bool            `json:"allow_additional_channels"`
	// Resolved relationships
	Country           *Country            `json:"-"`
	City              *City               `json:"-"`
	Region            *Region             `json:"-"`
	DIDGroupType      *DIDGroupType       `json:"-"`
	StockKeepingUnits []*StockKeepingUnit `json:"-"`
}

// ResolveRelationships resolves included relationships for DIDGroup.
func (dg *DIDGroup) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if country, err := ResolveToOne[Country](included, rels, "country"); err != nil {
		return err
	} else if country != nil {
		dg.Country = country
	}
	if city, err := ResolveToOne[City](included, rels, "city"); err != nil {
		return err
	} else if city != nil {
		dg.City = city
	}
	if region, err := ResolveToOne[Region](included, rels, "region"); err != nil {
		return err
	} else if region != nil {
		dg.Region = region
	}
	if dgt, err := ResolveToOne[DIDGroupType](included, rels, "did_group_type"); err != nil {
		return err
	} else if dgt != nil {
		dg.DIDGroupType = dgt
	}
	if skus, err := ResolveToMany[StockKeepingUnit](included, rels, "stock_keeping_units"); err != nil {
		return err
	} else if skus != nil {
		dg.StockKeepingUnits = skus
	}
	return nil
}

// DIDGroupType represents a type of DID group.
type DIDGroupType struct {
	ID   string `json:"-"`
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
	ID                    string `json:"-"`
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
	Countries            []*Country             `json:"-"`
	SharedCapacityGroups []*SharedCapacityGroup `json:"-"`
	QtyBasedPricings     []*QtyBasedPricing     `json:"-"`
}

// ResolveRelationships resolves included relationships for CapacityPool.
func (cp *CapacityPool) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if countries, err := ResolveToMany[Country](included, rels, "countries"); err != nil {
		return err
	} else if countries != nil {
		cp.Countries = countries
	}
	if groups, err := ResolveToMany[SharedCapacityGroup](included, rels, "shared_capacity_groups"); err != nil {
		return err
	} else if groups != nil {
		cp.SharedCapacityGroups = groups
	}
	if pricings, err := ResolveToMany[QtyBasedPricing](included, rels, "qty_based_pricings"); err != nil {
		return err
	} else if pricings != nil {
		cp.QtyBasedPricings = pricings
	}
	return nil
}

// SharedCapacityGroup represents a shared capacity group.
type SharedCapacityGroup struct {
	ID                   string `json:"-"`
	Name                 string `json:"name"`
	SharedChannelsCount  int    `json:"shared_channels_count"`
	CreatedAt            string `json:"created_at" api:"readonly"`
	MeteredChannelsCount int    `json:"metered_channels_count"`
	// Relationship IDs for create/update
	CapacityPoolID string `json:"-"`
	// Resolved relationships
	CapacityPool *CapacityPool `json:"-"`
	DIDs         []*DID        `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for SharedCapacityGroup.
func (s *SharedCapacityGroup) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if s.CapacityPoolID != "" {
		rels["capacity_pool"] = ToOneRelationship(RelationshipRef{Type: "capacity_pools", ID: s.CapacityPoolID})
	}
	return rels, nil
}

// ResolveRelationships resolves included relationships for SharedCapacityGroup.
func (s *SharedCapacityGroup) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if pool, err := ResolveToOne[CapacityPool](included, rels, "capacity_pool"); err != nil {
		return err
	} else if pool != nil {
		s.CapacityPool = pool
	}
	if dids, err := ResolveToMany[DID](included, rels, "dids"); err != nil {
		return err
	} else if dids != nil {
		s.DIDs = dids
	}
	return nil
}

// Address represents a customer address.
type Address struct {
	ID          string `json:"-"`
	CityName    string `json:"city_name"`
	PostalCode  string `json:"postal_code"`
	Address     string `json:"address"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at" api:"readonly"`
	Verified    bool   `json:"verified" api:"readonly"`
	// Relationship IDs for create/update
	IdentityID string `json:"-"`
	CountryID  string `json:"-"`
	// Resolved relationships
	Country  *Country  `json:"-"`
	Identity *Identity `json:"-"`
	Proofs   []*Proof  `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for Address.
func (a *Address) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if a.CountryID != "" {
		rels["country"] = ToOneRelationship(RelationshipRef{Type: "countries", ID: a.CountryID})
	}
	if a.IdentityID != "" {
		rels["identity"] = ToOneRelationship(RelationshipRef{Type: "identities", ID: a.IdentityID})
	}
	return rels, nil
}

// ResolveRelationships resolves included relationships for Address.
func (a *Address) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if country, err := ResolveToOne[Country](included, rels, "country"); err != nil {
		return err
	} else if country != nil {
		a.Country = country
	}
	if identity, err := ResolveToOne[Identity](included, rels, "identity"); err != nil {
		return err
	} else if identity != nil {
		a.Identity = identity
	}
	if proofs, err := ResolveToMany[Proof](included, rels, "proofs"); err != nil {
		return err
	} else if proofs != nil {
		a.Proofs = proofs
	}
	return nil
}

// AvailableDID represents a DID available for purchase.
type AvailableDID struct {
	ID     string `json:"-"`
	Number string `json:"number"`
	// Resolved relationships
	DIDGroup    *DIDGroup    `json:"-"`
	NanpaPrefix *NanpaPrefix `json:"-"`
}

// ResolveRelationships resolves included relationships for AvailableDID.
func (a *AvailableDID) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if dg, err := ResolveToOne[DIDGroup](included, rels, "did_group"); err != nil {
		return err
	} else if dg != nil {
		a.DIDGroup = dg
	}
	if np, err := ResolveToOne[NanpaPrefix](included, rels, "nanpa_prefix"); err != nil {
		return err
	} else if np != nil {
		a.NanpaPrefix = np
	}
	return nil
}

// DIDReservation represents a reserved DID.
type DIDReservation struct {
	ID          string `json:"-"`
	ExpireAt    string `json:"expire_at" api:"readonly"`
	CreatedAt   string `json:"created_at" api:"readonly"`
	Description string `json:"description"`
	// Relationship IDs for create/update
	AvailableDIDID string `json:"-"`
	// Resolved relationships
	AvailableDID *AvailableDID `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for DIDReservation.
func (r *DIDReservation) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if r.AvailableDIDID != "" {
		rels["available_did"] = ToOneRelationship(RelationshipRef{Type: "available_dids", ID: r.AvailableDIDID})
	}
	return rels, nil
}

// ResolveRelationships resolves included relationships for DIDReservation.
func (r *DIDReservation) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if ad, err := ResolveToOne[AvailableDID](included, rels, "available_did"); err != nil {
		return err
	} else if ad != nil {
		r.AvailableDID = ad
	}
	return nil
}

// Proof represents a proof document.
type Proof struct {
	ID        string  `json:"-"`
	CreatedAt string  `json:"created_at" api:"readonly"`
	ExpiresAt *string `json:"expires_at" api:"readonly"`
	// Polymorphic entity relationship (type: "identities" or "addresses")
	EntityID   string `json:"-"`
	EntityType string `json:"-"`
	// Other relationship IDs
	ProofTypeID string   `json:"-"`
	FileIDs     []string `json:"-"`
	// Resolved relationships
	ProofType *ProofType `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for Proof.
func (p *Proof) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)

	if p.EntityID != "" && p.EntityType != "" {
		rels["entity"] = ToOneRelationship(RelationshipRef{Type: p.EntityType, ID: p.EntityID})
	}

	if p.ProofTypeID != "" {
		rels["proof_type"] = ToOneRelationship(RelationshipRef{Type: "proof_types", ID: p.ProofTypeID})
	}

	if len(p.FileIDs) > 0 {
		refs := make([]RelationshipRef, len(p.FileIDs))
		for i, id := range p.FileIDs {
			refs[i] = RelationshipRef{Type: "encrypted_files", ID: id}
		}
		rels["files"] = ToManyRelationship(refs)
	}

	return rels, nil
}

// UnmarshalRelationships implements RelationshipUnmarshaler for Proof.
func (p *Proof) UnmarshalRelationships(rels map[string]json.RawMessage) error {
	if raw, ok := rels["entity"]; ok {
		ref, err := ParseToOneRelationship(raw)
		if err != nil {
			return err
		}
		if ref != nil {
			p.EntityID = ref.ID
			p.EntityType = ref.Type
		}
	}

	if raw, ok := rels["proof_type"]; ok {
		ref, err := ParseToOneRelationship(raw)
		if err != nil {
			return err
		}
		if ref != nil {
			p.ProofTypeID = ref.ID
		}
	}

	return nil
}

// ResolveRelationships resolves included relationships for Proof.
func (p *Proof) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if pt, err := ResolveToOne[ProofType](included, rels, "proof_type"); err != nil {
		return err
	} else if pt != nil {
		p.ProofType = pt
	}
	return nil
}

// PublicKey represents a DIDWW public key for encryption.
type PublicKey struct {
	ID  string `json:"-"`
	Key string `json:"key"`
}

// AddressVerification represents an address verification request.
type AddressVerification struct {
	ID                 string                          `json:"-"`
	ServiceDescription *string                         `json:"service_description,omitempty"`
	CallbackURL        *string                         `json:"callback_url,omitempty"`
	CallbackMethod     *string                         `json:"callback_method,omitempty"`
	Status             enums.AddressVerificationStatus `json:"status" api:"readonly"`
	RejectReasons      *string                         `json:"reject_reasons" api:"readonly"`
	CreatedAt          string                          `json:"created_at" api:"readonly"`
	Reference          string                          `json:"reference" api:"readonly"`
	// Relationship IDs for create/update
	AddressID string   `json:"-"`
	DIDIDs    []string `json:"-"`
	// Resolved relationships
	AddressRel *Address `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for AddressVerification.
func (av *AddressVerification) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if av.AddressID != "" {
		rels["address"] = ToOneRelationship(RelationshipRef{Type: "addresses", ID: av.AddressID})
	}
	if len(av.DIDIDs) > 0 {
		refs := make([]RelationshipRef, len(av.DIDIDs))
		for i, id := range av.DIDIDs {
			refs[i] = RelationshipRef{Type: "dids", ID: id}
		}
		rels["dids"] = ToManyRelationship(refs)
	}
	return rels, nil
}

// ResolveRelationships resolves included relationships for AddressVerification.
func (av *AddressVerification) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if addr, err := ResolveToOne[Address](included, rels, "address"); err != nil {
		return err
	} else if addr != nil {
		av.AddressRel = addr
	}
	return nil
}

// EncryptedFile represents an encrypted file upload.
type EncryptedFile struct {
	ID          string  `json:"-"`
	Description string  `json:"description"`
	ExpireAt    *string `json:"expire_at" api:"readonly"`
}

// NanpaPrefix represents an NANPA prefix.
type NanpaPrefix struct {
	ID  string `json:"-"`
	NPA string `json:"npa"`
	NXX string `json:"nxx"`
	// Resolved relationships
	Country *Country `json:"-"`
}

// ResolveRelationships resolves included relationships for NanpaPrefix.
func (n *NanpaPrefix) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if country, err := ResolveToOne[Country](included, rels, "country"); err != nil {
		return err
	} else if country != nil {
		n.Country = country
	}
	return nil
}

// Requirement represents a regulatory requirement.
type Requirement struct {
	ID                         string   `json:"-"`
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
	Country                   *Country                    `json:"-"`
	DIDGroupType              *DIDGroupType               `json:"-"`
	PersonalPermanentDocument *SupportingDocumentTemplate `json:"-"`
	BusinessPermanentDocument *SupportingDocumentTemplate `json:"-"`
	PersonalOnetimeDocument   *SupportingDocumentTemplate `json:"-"`
	BusinessOnetimeDocument   *SupportingDocumentTemplate `json:"-"`
	PersonalProofTypes        []*ProofType                `json:"-"`
	BusinessProofTypes        []*ProofType                `json:"-"`
	AddressProofTypes         []*ProofType                `json:"-"`
}

// ResolveRelationships resolves included relationships for Requirement.
func (r *Requirement) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if country, err := ResolveToOne[Country](included, rels, "country"); err != nil {
		return err
	} else if country != nil {
		r.Country = country
	}
	if dgt, err := ResolveToOne[DIDGroupType](included, rels, "did_group_type"); err != nil {
		return err
	} else if dgt != nil {
		r.DIDGroupType = dgt
	}
	if doc, err := ResolveToOne[SupportingDocumentTemplate](included, rels, "personal_permanent_document"); err != nil {
		return err
	} else if doc != nil {
		r.PersonalPermanentDocument = doc
	}
	if doc, err := ResolveToOne[SupportingDocumentTemplate](included, rels, "business_permanent_document"); err != nil {
		return err
	} else if doc != nil {
		r.BusinessPermanentDocument = doc
	}
	if doc, err := ResolveToOne[SupportingDocumentTemplate](included, rels, "personal_onetime_document"); err != nil {
		return err
	} else if doc != nil {
		r.PersonalOnetimeDocument = doc
	}
	if doc, err := ResolveToOne[SupportingDocumentTemplate](included, rels, "business_onetime_document"); err != nil {
		return err
	} else if doc != nil {
		r.BusinessOnetimeDocument = doc
	}
	if pts, err := ResolveToMany[ProofType](included, rels, "personal_proof_types"); err != nil {
		return err
	} else if pts != nil {
		r.PersonalProofTypes = pts
	}
	if pts, err := ResolveToMany[ProofType](included, rels, "business_proof_types"); err != nil {
		return err
	} else if pts != nil {
		r.BusinessProofTypes = pts
	}
	if pts, err := ResolveToMany[ProofType](included, rels, "address_proof_types"); err != nil {
		return err
	} else if pts != nil {
		r.AddressProofTypes = pts
	}
	return nil
}

// RequirementValidation represents a requirement validation result.
type RequirementValidation struct {
	ID string `json:"-"`
	// Relationship IDs for create
	AddressID     string `json:"-"`
	IdentityID    string `json:"-"`
	RequirementID string `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for RequirementValidation.
func (rv *RequirementValidation) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if rv.AddressID != "" {
		rels["address"] = ToOneRelationship(RelationshipRef{Type: "addresses", ID: rv.AddressID})
	}
	if rv.IdentityID != "" {
		rels["identity"] = ToOneRelationship(RelationshipRef{Type: "identities", ID: rv.IdentityID})
	}
	if rv.RequirementID != "" {
		rels["requirement"] = ToOneRelationship(RelationshipRef{Type: "requirements", ID: rv.RequirementID})
	}
	return rels, nil
}

// ProofType represents a type of proof document.
type ProofType struct {
	ID         string `json:"-"`
	Name       string `json:"name"`
	EntityType string `json:"entity_type"`
}

// SupportingDocumentTemplate represents a supporting document template.
type SupportingDocumentTemplate struct {
	ID        string `json:"-"`
	Name      string `json:"name"`
	Permanent bool   `json:"permanent"`
	URL       string `json:"url"`
}

// PermanentSupportingDocument represents a permanent supporting document.
type PermanentSupportingDocument struct {
	ID        string `json:"-"`
	CreatedAt string `json:"created_at" api:"readonly"`
	// Relationship IDs for create/update
	TemplateID string   `json:"-"`
	IdentityID string   `json:"-"`
	FileIDs    []string `json:"-"`
	// Resolved relationships
	Template *SupportingDocumentTemplate `json:"-"`
}

// MarshalRelationships implements RelationshipMarshaler for PermanentSupportingDocument.
func (d *PermanentSupportingDocument) MarshalRelationships() (map[string]any, error) {
	rels := make(map[string]any)
	if d.TemplateID != "" {
		rels["template"] = ToOneRelationship(RelationshipRef{Type: "supporting_document_templates", ID: d.TemplateID})
	}
	if d.IdentityID != "" {
		rels["identity"] = ToOneRelationship(RelationshipRef{Type: "identities", ID: d.IdentityID})
	}
	if len(d.FileIDs) > 0 {
		refs := make([]RelationshipRef, len(d.FileIDs))
		for i, id := range d.FileIDs {
			refs[i] = RelationshipRef{Type: "encrypted_files", ID: id}
		}
		rels["files"] = ToManyRelationship(refs)
	}
	return rels, nil
}

// ResolveRelationships resolves included relationships for PermanentSupportingDocument.
func (d *PermanentSupportingDocument) ResolveRelationships(included IncludedResources, rels map[string]json.RawMessage) error {
	if tmpl, err := ResolveToOne[SupportingDocumentTemplate](included, rels, "template"); err != nil {
		return err
	} else if tmpl != nil {
		d.Template = tmpl
	}
	return nil
}

// VoiceOutTrunkRegenerateCredential represents a credential regeneration for voice out trunks.
type VoiceOutTrunkRegenerateCredential struct {
	ID string `json:"-"`
}
