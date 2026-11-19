package main

import (
	_ "embed"
	"errors"
	"fmt"
	"sort"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/orderitem"
)

// examplePDF is uploaded as the proof documents.
//
//go:embed example.pdf
var examplePDF []byte

// The address and emergency verifications are submitted to DIDWW for a review
// and cannot be deleted. Each check prints a line starting with REVIEW with the
// ID and the reference of what it submitted.

// addressVerification runs the README's address verification snippet for a
// DID of a Ukrainian local DID group, whose DIDs need an address. It orders
// such a DID when the account has none without a verification, and creates an
// identity and an address that meet the group's address requirement.
func (s *suite) addressVerification() (err error) {
	ctx, client := s.ctx, s.client

	group, requirement, err := s.groupWithAddressRequirement()
	if err != nil {
		return err
	}
	s.logf("DID group %s %s, address requirement %s (identity type %s)",
		group.Prefix, group.AreaName, requirement.ID, requirement.IdentityType)

	did, err := s.didWithoutAddressVerification(group)
	if err != nil {
		return err
	}
	s.logf("DID %s", did.Number)

	reg := &registration{}
	defer s.cleanupRegistration(&err, reg)
	err = s.registerCompliantAddress(reg, group, requirement)
	if err != nil {
		return err
	}

	_, err = client.AddressRequirementValidations().Create(ctx, &resource.AddressRequirementValidation{
		AddressRequirementID: requirement.ID,
		AddressID:            reg.addressID,
		IdentityID:           reg.identityID,
	})
	if err != nil {
		return fmt.Errorf("address requirement validation: %w", err)
	}
	s.logf("the identity and the address meet the requirement")

	cbURL := "http://example.com/callback"
	cbMethod := "GET"
	verification := &resource.AddressVerification{
		CallbackURL:    &cbURL,
		CallbackMethod: &cbMethod,
		AddressID:      reg.addressID,
		DIDIDs:         []string{did.ID},
	}
	if requirement.ServiceDescriptionRequired {
		verification.ServiceDescription = didww.Ptr("Inbound calls of an SDK README check")
	}
	created, err := client.AddressVerifications().Create(ctx, verification)
	if err != nil {
		return err
	}
	reg.submitted = true
	fmt.Printf("REVIEW address_verification %s %s\n", created.ID, created.Reference)
	return nil
}

// registration tracks what a check created for an address verification.
type registration struct {
	identityID string
	addressID  string
	fileIDs    []string
	submitted  bool
}

// cleanupRegistration deletes the identity, address and files when no
// verification refers to them. After a verification was submitted, the
// address cannot be deleted; it is left in place with its identity and files.
func (s *suite) cleanupRegistration(err *error, reg *registration) { //nolint:gocritic // sets the caller's named result
	ctx, client := s.ctx, s.client
	if reg.submitted {
		if reg.addressID == "" {
			return
		}
		if deleteErr := client.Addresses().Delete(ctx, reg.addressID); deleteErr != nil {
			s.logf("address %s, identity %s and the uploaded files are left in place for the review: %v",
				reg.addressID, reg.identityID, deleteErr)
			return
		}
		s.undo(err, "delete the identity", func() error { return client.Identities().Delete(ctx, reg.identityID) })
		return
	}
	if reg.addressID != "" {
		s.undo(err, "delete the address", func() error { return client.Addresses().Delete(ctx, reg.addressID) })
	}
	if reg.identityID != "" {
		s.undo(err, "delete the identity", func() error { return client.Identities().Delete(ctx, reg.identityID) })
	}
	for _, id := range reg.fileIDs {
		fileID := id
		s.undo(err, "delete the encrypted file", func() error { return client.EncryptedFiles().Delete(ctx, fileID) })
	}
}

// groupWithAddressRequirement returns a Ukrainian local DID group that needs
// an address, Kiev when it does, with its address requirement. The
// address_requirement relationship is only set when it is included.
func (s *suite) groupWithAddressRequirement() (*resource.DIDGroup, *resource.AddressRequirement, error) {
	ctx, client := s.ctx, s.client
	ukraineID, err := s.countryID("UA")
	if err != nil {
		return nil, nil, err
	}
	ukrainianGroups := func(areaName string) (*resource.DIDGroup, error) {
		params := didww.NewQueryParams().
			Filter("country.id", ukraineID).
			Filter("number_type", string(enums.NumberTypeLocal)).
			Include("address_requirement,city,stock_keeping_units").
			Page(1, 500)
		if areaName != "" {
			params = params.Filter("area_name", areaName)
		}
		groups, listErr := client.DIDGroups().List(ctx, params)
		if listErr != nil {
			return nil, listErr
		}
		for _, g := range groups {
			if g.AddressRequirement != nil && len(g.StockKeepingUnits) > 0 {
				return g, nil
			}
		}
		return nil, nil
	}

	group, err := ukrainianGroups("Kiev")
	if err == nil && group == nil {
		group, err = ukrainianGroups("") // any other Ukrainian local group with a requirement
	}
	if err != nil {
		return nil, nil, err
	}
	if group == nil {
		return nil, nil, errors.New("no Ukrainian local DID group with an address requirement")
	}

	requirement, err := client.AddressRequirements().Find(ctx, group.AddressRequirement.ID, didww.NewQueryParams().Include(
		"personal_proof_types,business_proof_types,address_proof_types,personal_permanent_document,business_permanent_document"))
	if err != nil {
		return nil, nil, err
	}
	return group, requirement, nil
}

// didWithoutAddressVerification returns an active DID of the group without an
// address verification, or orders one.
func (s *suite) didWithoutAddressVerification(group *resource.DIDGroup) (*resource.DID, error) {
	params := didww.NewQueryParams().
		Filter("did_group.id", group.ID).
		Filter("terminated", "false").
		Include("address_verification").
		Page(1, 50)
	dids, err := s.client.DIDs().List(s.ctx, params)
	if err != nil {
		return nil, err
	}
	for _, did := range dids {
		if did.AddressVerification == nil {
			return did, nil
		}
	}
	return s.orderAvailableDID(didww.NewQueryParams().Filter("did_group.id", group.ID))
}

// registerCompliantAddress creates an identity and an address in the group's
// country that meet the requirement: every mandatory identity field, the
// proofs and the permanent document it asks for.
func (s *suite) registerCompliantAddress(reg *registration, group *resource.DIDGroup, requirement *resource.AddressRequirement) error {
	ctx, client := s.ctx, s.client
	ukraineID, err := s.countryID("UA")
	if err != nil {
		return err
	}

	business := requirement.IdentityType == string(enums.IdentityTypeBusiness)
	identity, err := client.Identities().Create(ctx, compliantIdentity(business, ukraineID))
	if err != nil {
		return err
	}
	reg.identityID = identity.ID

	cityName := group.AreaName
	if group.City != nil {
		cityName = group.City.Name
	}
	address, err := client.Addresses().Create(ctx, &resource.Address{
		CityName:   cityName,
		PostalCode: "65000",
		Address:    "1 Derybasivska St",
		IdentityID: identity.ID,
		CountryID:  ukraineID,
	})
	if err != nil {
		return err
	}
	reg.addressID = address.ID
	s.logf("identity %s, address %s", identity.ID, address.ID)

	identityProofTypes, identityProofQty, template := requirement.PersonalProofTypes, requirement.PersonalProofQty, requirement.PersonalPermanentDocument
	if business {
		identityProofTypes, identityProofQty, template = requirement.BusinessProofTypes, requirement.BusinessProofQty, requirement.BusinessPermanentDocument
	}
	err = s.createProofs(reg, "identities", identity.ID, identityProofTypes, identityProofQty)
	if err != nil {
		return err
	}
	err = s.createProofs(reg, "addresses", address.ID, requirement.AddressProofTypes, requirement.AddressProofQty)
	if err != nil {
		return err
	}
	if template == nil {
		return nil
	}
	fileID, err := s.uploadExamplePDF(reg, "permanent_document")
	if err != nil {
		return err
	}
	_, err = client.PermanentSupportingDocuments().Create(ctx, &resource.PermanentSupportingDocument{
		TemplateID: template.ID,
		IdentityID: identity.ID,
		FileIDs:    []string{fileID},
	})
	return err
}

// compliantIdentity fills every field an address requirement can make mandatory.
func compliantIdentity(business bool, countryID string) *resource.Identity {
	identity := &resource.Identity{
		FirstName:     "John",
		LastName:      "Doe",
		PhoneNumber:   "380487000000",
		IdentityType:  enums.IdentityTypePersonal,
		IDNumber:      didww.Ptr("AB123456"),
		PersonalTaxID: didww.Ptr("1234567890"),
		ContactEmail:  didww.Ptr("john.doe@example.com"),
		CountryID:     countryID,
	}
	if business {
		identity.IdentityType = enums.IdentityTypeBusiness
		identity.CompanyName = didww.Ptr("Example LLC")
		identity.CompanyRegNumber = didww.Ptr("12345678")
		identity.VatID = didww.Ptr("UA123456789012")
		return identity
	}
	identity.BirthDate = "1980-01-01"
	identity.BirthCountryID = countryID
	return identity
}

// createProofs creates qty proofs of different types for an identity or an address.
func (s *suite) createProofs(reg *registration, entityType, entityID string, proofTypes []*resource.ProofType, qty int) error {
	if qty > len(proofTypes) {
		return fmt.Errorf("the requirement asks for %d %s proofs of %d types", qty, entityType, len(proofTypes))
	}
	for _, proofType := range proofTypes[:qty] {
		fileID, err := s.uploadExamplePDF(reg, proofType.Name)
		if err != nil {
			return err
		}
		proof, err := s.client.Proofs().Create(s.ctx, &resource.Proof{
			EntityID:    entityID,
			EntityType:  entityType,
			ProofTypeID: proofType.ID,
			FileIDs:     []string{fileID},
		})
		if err != nil {
			return err
		}
		s.logf("%s proof %s (%s)", entityType, proof.ID, proofType.Name)
	}
	return nil
}

// uploadExamplePDF encrypts the bundled PDF and uploads it.
func (s *suite) uploadExamplePDF(reg *registration, description string) (string, error) {
	enc, err := didww.NewEncrypt(s.ctx, s.client)
	if err != nil {
		return "", err
	}
	encrypted, err := enc.Encrypt(examplePDF)
	if err != nil {
		return "", err
	}
	id, err := s.client.UploadEncryptedFile(s.ctx, encrypted, "example.pdf.enc", enc.Fingerprint(), description)
	if err != nil {
		return "", err
	}
	reg.fileIDs = append(reg.fileIDs, id)
	return id, nil
}

// emergencyVerification runs the README's emergency verification snippet the
// way examples/emergency_scenario does: an emergency-capable DID in the
// country of an account address, Ukraine first, checked first against the
// emergency requirement.
func (s *suite) emergencyVerification() error {
	ctx, client := s.ctx, s.client

	address, err := s.emergencyAddress()
	if err != nil {
		return err
	}
	did, err := s.emergencyDID(address.Country.ID)
	if err != nil {
		return err
	}
	if did.DIDGroup == nil {
		return fmt.Errorf("DID %s: did_group not included", did.Number)
	}
	s.logf("address %s in %s, DID %s", address.ID, address.Country.Name, did.Number)

	reqParams := didww.NewQueryParams().
		Filter("country.id", address.Country.ID).
		Filter("number_type", string(did.DIDGroup.NumberType))
	requirements, err := client.EmergencyRequirements().List(ctx, reqParams)
	if err != nil {
		return err
	}
	if len(requirements) == 0 {
		return fmt.Errorf("no emergency requirement for %s %s DIDs", address.Country.Name, did.DIDGroup.NumberType)
	}
	_, err = client.EmergencyRequirementValidations().Create(ctx, &resource.EmergencyRequirementValidation{
		EmergencyRequirementID: requirements[0].ID,
		AddressID:              address.ID,
		IdentityID:             address.Identity.ID,
	})
	if err != nil {
		return fmt.Errorf("emergency requirement validation: %w", err)
	}

	verification := &resource.EmergencyVerification{
		AddressID: address.ID,
		DIDIDs:    []string{did.ID},
	}
	created, err := client.EmergencyVerifications().Create(ctx, verification)
	if err != nil {
		return err
	}
	fmt.Printf("REVIEW emergency_verification %s %s\n", created.ID, created.Reference)
	return nil
}

// emergencyAddress returns an account address with its country and identity:
// Ukrainian and verified ones first.
func (s *suite) emergencyAddress() (*resource.Address, error) {
	params := didww.NewQueryParams().Include("country,identity").Page(1, 100)
	addresses, err := s.client.Addresses().List(s.ctx, params)
	if err != nil {
		return nil, err
	}
	rank := func(a *resource.Address) int {
		r := 0
		if a.Country == nil || a.Country.ISO != "UA" {
			r += 2
		}
		if !a.Verified {
			r++
		}
		return r
	}
	sort.SliceStable(addresses, func(i, j int) bool { return rank(addresses[i]) < rank(addresses[j]) })
	for _, a := range addresses {
		if a.Country != nil && a.Identity != nil {
			return a, nil
		}
	}
	return nil, errors.New("the account has no address with a country and an identity")
}

// emergencyDID returns an active emergency-capable DID of the country without
// an emergency calling service, or orders one.
func (s *suite) emergencyDID(countryID string) (*resource.DID, error) {
	params := didww.NewQueryParams().
		Filter("did_group.features", string(enums.FeatureEmergency)).
		Filter("country.id", countryID).
		Filter("emergency_enabled", "false").
		Filter("awaiting_registration", "false").
		Filter("terminated", "false").
		Include("did_group,emergency_calling_service").
		Page(1, 50)
	dids, err := s.client.DIDs().List(s.ctx, params)
	if err != nil {
		return nil, err
	}
	for _, did := range dids {
		if did.EmergencyCallingService == nil && did.DIDGroup != nil {
			return did, nil
		}
	}
	return s.orderAvailableDID(didww.NewQueryParams().
		Filter("did_group.features", string(enums.FeatureEmergency)).
		Filter("country.id", countryID))
}

// orderAvailableDID orders the first available DID that matches params, waits
// for the order to complete and returns the DID with its DID group.
func (s *suite) orderAvailableDID(params *didww.QueryParams) (*resource.DID, error) {
	ctx, client := s.ctx, s.client
	available, err := client.AvailableDIDs().List(ctx, params.Include("did_group.stock_keeping_units").Page(1, 1))
	if err != nil {
		return nil, err
	}
	if len(available) == 0 || available[0].DIDGroup == nil || len(available[0].DIDGroup.StockKeepingUnits) == 0 {
		return nil, errors.New("no available DID to order")
	}
	ad := available[0]
	order, err := client.Orders().Create(ctx, &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.AvailableDidOrderItem{
				DidOrderItem:   orderitem.DidOrderItem{SkuID: ad.DIDGroup.StockKeepingUnits[0].ID},
				AvailableDidID: ad.ID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	for i := 0; i < 12 && !order.IsCompleted(); i++ {
		if order.IsCanceled() {
			return nil, fmt.Errorf("order %s was canceled", order.ID)
		}
		time.Sleep(5 * time.Second)
		if order, err = client.Orders().Find(ctx, order.ID); err != nil {
			return nil, err
		}
	}
	if !order.IsCompleted() {
		return nil, fmt.Errorf("order %s is still %s", order.ID, order.Status)
	}
	dids, err := client.DIDs().List(ctx, didww.NewQueryParams().
		Filter("order.id", order.ID).
		Include("did_group,emergency_calling_service"))
	if err != nil {
		return nil, err
	}
	if len(dids) == 0 {
		return nil, fmt.Errorf("order %s has no DID", order.ID)
	}
	s.logf("ordered DID %s (order %s)", dids[0].Number, order.ID)
	return dids[0], nil
}
