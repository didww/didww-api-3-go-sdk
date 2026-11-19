package main

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/orderitem"
)

// count logs how many resources a list returned.
func (s *suite) count(what string, n int, err error) error {
	if err != nil {
		return err
	}
	s.logf("got %d %s", n, what)
	return nil
}

// README: Read-Only Resources
func (s *suite) readOnlyResources() {
	ctx, client := s.ctx, s.client
	s.section("Read-Only Resources")

	s.check("Countries: list and find", func() error {
		countries, err := client.Countries().List(ctx, nil)
		if err != nil {
			return err
		}
		if len(countries) == 0 {
			return errors.New("no countries")
		}
		country, err := client.Countries().Find(ctx, countries[0].ID)
		if err != nil {
			return err
		}
		s.logf("got %d countries, found %s", len(countries), country.Name)
		return expect(country.ID == countries[0].ID, "found country %s, want %s", country.ID, countries[0].ID)
	})
	s.check("Regions", func() error {
		regions, err := client.Regions().List(ctx, nil)
		return s.count("regions", len(regions), err)
	})
	s.check("Cities", func() error {
		cities, err := client.Cities().List(ctx, nil)
		return s.count("cities", len(cities), err)
	})
	s.check("Areas", func() error {
		areas, err := client.Areas().List(ctx, nil)
		return s.count("areas", len(areas), err)
	})
	s.check("NANPA prefixes", func() error {
		prefixes, err := client.NanpaPrefixes().List(ctx, nil)
		return s.count("NANPA prefixes", len(prefixes), err)
	})
	s.check("DID groups with stock keeping units, filtered by number type", func() error {
		params := didww.NewQueryParams().
			Include("stock_keeping_units").
			Filter("number_type", string(enums.NumberTypeTollFree))
		groups, err := client.DIDGroups().List(ctx, params)
		if err != nil {
			return err
		}
		for _, g := range groups {
			if g.NumberType != enums.NumberTypeTollFree {
				return fmt.Errorf("DID group %s has number type %s", g.ID, g.NumberType)
			}
		}
		return s.count("toll-free DID groups", len(groups), nil)
	})
	s.check("Available DIDs with DID group and stock keeping units", func() error {
		params := didww.NewQueryParams().Include("did_group.stock_keeping_units")
		available, err := client.AvailableDIDs().List(ctx, params)
		if err != nil {
			return err
		}
		for _, ad := range available {
			if ad.DIDGroup == nil {
				return fmt.Errorf("available DID %s: did_group not included", ad.ID)
			}
		}
		return s.count("available DIDs", len(available), nil)
	})
	s.check("Proof types", func() error {
		proofTypes, err := client.ProofTypes().List(ctx, nil)
		return s.count("proof types", len(proofTypes), err)
	})
	s.check("Public keys", func() error {
		publicKeys, err := client.PublicKeys().List(ctx, nil)
		if err != nil {
			return err
		}
		return expect(len(publicKeys) >= 2, "got %d public keys, want 2", len(publicKeys))
	})
	s.check("Address requirements", func() error {
		requirements, err := client.AddressRequirements().List(ctx, nil)
		return s.count("address requirements", len(requirements), err)
	})
	s.check("Emergency requirements", func() error {
		emergReqs, err := client.EmergencyRequirements().List(ctx, nil)
		return s.count("emergency requirements", len(emergReqs), err)
	})
	s.check("DID history", func() error {
		history, err := client.DIDHistory().List(ctx, nil)
		return s.count("DID history entries", len(history), err)
	})
	s.check("Supporting document templates", func() error {
		templates, err := client.SupportingDocumentTemplates().List(ctx, nil)
		return s.count("supporting document templates", len(templates), err)
	})
	s.check("Balance", func() error {
		balance, err := client.Balance().Find(ctx)
		if err != nil {
			return err
		}
		s.logf("total balance %s", balance.TotalBalance)
		return expect(balance.TotalBalance != "", "empty total balance")
	})
}

// README: DIDs
func (s *suite) dids() {
	ctx, client := s.ctx, s.client
	s.section("DIDs")

	s.check("List DIDs and update one (dirty-only PATCH)", func() (err error) {
		dids, err := client.DIDs().List(ctx, nil)
		if err != nil {
			return err
		}
		target, err := s.anyDID()
		if err != nil {
			return err
		}
		s.logf("got %d DIDs, updating %s", len(dids), target.Number)

		did, err := client.DIDs().Find(ctx, target.ID)
		if err != nil {
			return err
		}
		original := did.Description
		defer s.undo(&err, "restore the description", func() error { return s.restoreDescription(target.ID, original) })

		desc := "Updated"
		did.Description = &desc
		updated, err := client.DIDs().Update(ctx, did)
		if err != nil {
			return err
		}
		return expect(updated.Description != nil && *updated.Description == desc, "description not updated")
	})

	s.check("SMS routing: SMS trunk, SMS trunk group, unroute", s.smsRouting)
}

// smsRouting routes a DID's inbound messages to an SMS trunk, then to an SMS
// trunk group, unroutes it and puts its original routing back.
func (s *suite) smsRouting() (err error) {
	ctx, client := s.ctx, s.client
	original, err := s.firstDID(
		didww.NewQueryParams().
			Filter("did_group.features", string(enums.FeatureSmsIn)).
			Include("sms_trunk,sms_trunk_group"),
		nil, "active DID with the sms_in feature")
	if err != nil {
		return err
	}

	trunk, err := s.createInboundSmsTrunk("README SMS routing")
	if err != nil {
		return err
	}
	defer s.undo(&err, "delete the SMS trunk", s.deleteSmsTrunk(trunk.ID))

	member, err := s.createInboundSmsTrunk("README SMS routing group member")
	if err != nil {
		return err
	}
	group, err := client.SmsTrunkGroups().Create(ctx, &resource.SmsTrunkGroup{
		Name:        "README SMS routing group " + s.suffix, // SMS trunks and groups share one name space
		SmsTrunkIDs: []string{member.ID},
	})
	if err != nil {
		s.undo(&err, "delete the SMS trunk", s.deleteSmsTrunk(member.ID))
		return err
	}
	// Deleting the group also deletes its trunk
	defer s.undo(&err, "delete the SMS trunk group", func() error { return client.SmsTrunkGroups().Delete(ctx, group.ID) })
	defer s.undo(&err, "restore the DID's SMS routing", func() error { return s.restoreSMSRouting(original) })

	// sends sms_trunk and sms_trunk_group: null
	if _, err = client.DIDs().Update(ctx, &resource.DID{ID: original.ID, SmsTrunkID: trunk.ID}); err != nil {
		return err
	}

	params := didww.NewQueryParams().
		Filter("sms_trunk.id", trunk.ID).
		Include("sms_trunk,sms_trunk_group")
	dids, err := client.DIDs().List(ctx, params) // did.SmsTrunk, did.SmsTrunkGroup
	if err != nil {
		return err
	}
	if len(dids) != 1 || dids[0].SmsTrunk == nil || dids[0].SmsTrunk.ID != trunk.ID {
		return errors.New("the DID is not listed as routed to the SMS trunk")
	}

	// sends sms_trunk_group and sms_trunk: null
	if _, err = client.DIDs().Update(ctx, &resource.DID{ID: original.ID, SmsTrunkGroupID: group.ID}); err != nil {
		return err
	}
	err = s.expectSMSRouting(original.ID, "", group.ID)
	if err != nil {
		return err
	}

	// unroute: sends sms_trunk: null and sms_trunk_group: null
	if _, err = client.DIDs().Update(ctx, &resource.DID{ID: original.ID, NullifySmsTrunk: true, NullifySmsTrunkGroup: true}); err != nil {
		return err
	}
	return s.expectSMSRouting(original.ID, "", "")
}

func (s *suite) expectSMSRouting(didID, trunkID, groupID string) error {
	did, err := s.client.DIDs().Find(s.ctx, didID, didww.NewQueryParams().Include("sms_trunk,sms_trunk_group"))
	if err != nil {
		return err
	}
	gotTrunk, gotGroup := "", ""
	if did.SmsTrunk != nil {
		gotTrunk = did.SmsTrunk.ID
	}
	if did.SmsTrunkGroup != nil {
		gotGroup = did.SmsTrunkGroup.ID
	}
	return expect(gotTrunk == trunkID && gotGroup == groupID,
		"sms_trunk=%q sms_trunk_group=%q, want %q and %q", gotTrunk, gotGroup, trunkID, groupID)
}

func (s *suite) restoreSMSRouting(original *resource.DID) error {
	restore := &resource.DID{ID: original.ID, NullifySmsTrunk: true, NullifySmsTrunkGroup: true}
	switch {
	case original.SmsTrunk != nil:
		restore = &resource.DID{ID: original.ID, SmsTrunkID: original.SmsTrunk.ID}
	case original.SmsTrunkGroup != nil:
		restore = &resource.DID{ID: original.ID, SmsTrunkGroupID: original.SmsTrunkGroup.ID}
	}
	_, err := s.client.DIDs().Update(s.ctx, restore)
	return err
}

// README: Orders
func (s *suite) orders() {
	ctx, client := s.ctx, s.client
	s.section("Orders")
	s.check("Create a DID order and cancel it", func() error {
		skuID, err := s.skuID()
		if err != nil {
			return err
		}

		// Create order with DID order item
		order := &resource.Order{
			Items: []orderitem.OrderItem{
				&orderitem.DidOrderItem{
					SkuID: skuID,
					Qty:   2,
				},
			},
		}
		created, err := client.Orders().Create(ctx, order)
		if err != nil {
			return err
		}
		s.logf("order %s: %s", created.ID, created.Status)

		// Delete order (cancel)
		return client.Orders().Delete(ctx, created.ID)
	})
}

// README: DID Reservations
func (s *suite) didReservations() {
	ctx, client := s.ctx, s.client
	s.section("DID Reservations")
	s.check("Create and delete a DID reservation", func() (err error) {
		available, err := client.AvailableDIDs().List(ctx, didww.NewQueryParams().Page(1, 1))
		if err != nil {
			return err
		}
		if len(available) == 0 {
			return errors.New("no available DIDs")
		}

		reservation := &resource.DIDReservation{
			Description:    "Reserved for client",
			AvailableDIDID: available[0].ID,
		}
		created, err := client.DIDReservations().Create(ctx, reservation)
		if err != nil {
			return err
		}
		s.logf("reservation %s expires at %s", created.ID, created.ExpiresAt)

		// Delete reservation
		return client.DIDReservations().Delete(ctx, created.ID)
	})
}

// README: Shared Capacity Groups
func (s *suite) sharedCapacityGroups() {
	ctx, client := s.ctx, s.client
	s.section("Shared Capacity Groups")
	s.check("Create a shared capacity group", func() (err error) {
		const sharedChannels = 20
		pools, err := client.CapacityPools().List(ctx, nil)
		if err != nil {
			return err
		}
		var pool *resource.CapacityPool
		for _, p := range pools {
			if p.TotalChannelsCount-p.AssignedChannelsCount >= sharedChannels {
				pool = p
				break
			}
		}
		if pool == nil {
			return skip("no capacity pool with %d unassigned channels", sharedChannels)
		}

		group := &resource.SharedCapacityGroup{
			Name:                 "Shared Group " + s.suffix,
			SharedChannelsCount:  sharedChannels,
			MeteredChannelsCount: 0,
			CapacityPoolID:       pool.ID,
		}
		created, err := client.SharedCapacityGroups().Create(ctx, group)
		if err != nil {
			return err
		}
		defer s.undo(&err, "delete the shared capacity group", func() error {
			return client.SharedCapacityGroups().Delete(ctx, created.ID)
		})
		s.logf("shared capacity group %s in %s", created.ID, pool.Name)
		return expect(created.SharedChannelsCount == sharedChannels, "shared channels %d", created.SharedChannelsCount)
	})
}

// README: Identities, Addresses, Address Verifications
func (s *suite) identitiesAndAddresses() {
	ctx, client := s.ctx, s.client
	s.section("Identities and Addresses")

	var identity *resource.Identity
	s.check("Create an identity", func() error {
		countryID, err := s.usCountryID()
		if err != nil {
			return err
		}

		identity = &resource.Identity{
			FirstName:    "John",
			LastName:     "Doe",
			PhoneNumber:  "12125551234",
			IdentityType: enums.IdentityTypePersonal,
			CountryID:    countryID,
		}
		created, err := client.Identities().Create(ctx, identity)
		if err != nil {
			identity = nil
			return err
		}
		identity = created
		s.logf("identity %s", created.ID)
		return nil
	})

	s.check("Create an address", func() (err error) {
		if identity == nil {
			return errors.New("no identity: its check failed")
		}
		countryID, err := s.usCountryID()
		if err != nil {
			return err
		}

		address := &resource.Address{
			CityName:   "New York",
			PostalCode: "10001",
			Address:    "123 Main St",
			IdentityID: identity.ID,
			CountryID:  countryID,
		}
		created, err := client.Addresses().Create(ctx, address)
		if err != nil {
			return err
		}
		s.logf("address %s", created.ID)
		return client.Addresses().Delete(ctx, created.ID)
	})

	if identity != nil {
		s.check("Delete the identity", func() error {
			return client.Identities().Delete(ctx, identity.ID)
		})
	}

	s.check("Create an address verification", func() error {
		cbURL := "http://example.com/callback"
		cbMethod := "GET"
		verification := &resource.AddressVerification{
			CallbackURL:    &cbURL,
			CallbackMethod: &cbMethod,
			AddressID:      "address-uuid",
			DIDIDs:         []string{"did-uuid"},
		}
		submit := func() (*resource.AddressVerification, error) {
			return client.AddressVerifications().Create(ctx, verification)
		}
		_ = submit
		return skip("submits DIDs for an address review, which cannot be deleted; needs a DID whose DID group requires an address")
	})
}

// README: Emergency Services
func (s *suite) emergencyServices() {
	ctx, client := s.ctx, s.client
	s.section("Emergency Services")

	s.check("Emergency requirements filtered by country", func() error {
		countryID, err := s.usCountryID()
		if err != nil {
			return err
		}
		params := didww.NewQueryParams().Filter("country.id", countryID)
		reqs, err := client.EmergencyRequirements().List(ctx, params)
		return s.count("emergency requirements in the US", len(reqs), err)
	})

	s.check("Create an emergency verification", func() error {
		verification := &resource.EmergencyVerification{
			AddressID: "address-uuid",
			DIDIDs:    []string{"did-uuid"},
		}
		submit := func() (*resource.EmergencyVerification, error) {
			return client.EmergencyVerifications().Create(ctx, verification)
		}
		_ = submit
		return skip("submits DIDs for an emergency review, which cannot be deleted; needs an emergency-capable DID " +
			"and a matching address (see examples/emergency_scenario)")
	})

	s.check("Emergency calling services", func() error {
		services, err := client.EmergencyCallingServices().List(ctx, nil)
		return s.count("emergency calling services", len(services), err)
	})
}

// README: DID History
func (s *suite) didHistory() {
	ctx, client := s.ctx, s.client
	s.section("DID History")
	s.check("DID history of one DID", func() error {
		did, err := s.anyDID()
		if err != nil {
			return err
		}

		// List DID history entries
		params := didww.NewQueryParams().Filter("did_number", did.Number)
		history, err := client.DIDHistory().List(ctx, params)
		return s.count("history entries of "+did.Number, len(history), err)
	})
}

// README: Exports
func (s *suite) exports() {
	ctx, client := s.ctx, s.client
	s.section("Exports")

	// from and to must fall within the current month and the two previous ones
	const layout = "2006-01-02 15:04:05"
	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -14)

	s.check("Create a CDR IN export", func() error {
		export := &resource.Export{
			ExportType: enums.ExportTypeCdrIn,
			Filters:    map[string]interface{}{"from": from.Format(layout), "to": to.Format(layout)},
		}
		created, err := client.Exports().Create(ctx, export)
		if err != nil {
			return err
		}
		s.logf("export %s: %s", created.ID, created.Status)
		return nil
	})

	s.check("Create an SMS OUT export, read it and download a completed export", func() error {
		filters := map[string]interface{}{
			"from":         from.Format(layout), // inclusive
			"to":           to.Format(layout),   // exclusive
			"service_type": enums.ExportServiceTypeA2P,
		}
		trunks, err := client.SmsTrunks().List(ctx, didww.NewQueryParams().Page(1, 1))
		if err != nil {
			return err
		}
		if len(trunks) > 0 {
			filters["sms_trunk.id"] = []string{trunks[0].ID}
		}
		smsOut, err := client.Exports().Create(ctx, &resource.Export{
			ExportType: enums.ExportTypeSmsOut,
			Filters:    filters,
		})
		if err != nil {
			return err
		}
		s.logf("export %s: %s", smsOut.ID, smsOut.Status)

		// The export is pending until its file is generated; read it again later
		done, err := client.Exports().Find(ctx, smsOut.ID)
		if err != nil {
			return err
		}
		if done.URL == nil {
			// Download an export that is already completed instead
			var completed []*resource.Export
			completed, err = client.Exports().List(ctx, didww.NewQueryParams().Filter("status", "completed").Page(1, 1))
			if err != nil {
				return err
			}
			if len(completed) == 0 || completed[0].URL == nil {
				s.logf("no completed export to download")
				return nil
			}
			done = completed[0]
		}
		return s.download(*done.URL)
	})
}

func (s *suite) download(exportURL string) error {
	var gz bytes.Buffer
	if err := s.client.DownloadExport(s.ctx, exportURL, &gz); err != nil { // gzipped CSV, as served
		return err
	}
	var csv bytes.Buffer
	if err := s.client.DownloadAndDecompressExport(s.ctx, exportURL, &csv); err != nil {
		return err
	}
	s.logf("downloaded %d bytes, %d bytes decompressed", gz.Len(), csv.Len())
	return nil
}

// README: Filtering, Sorting, and Pagination
func (s *suite) filtering() {
	ctx, client := s.ctx, s.client
	s.section("Filtering, Sorting, and Pagination")
	s.check("Regions with filter, include, sort and page", func() error {
		countryID, err := s.usCountryID()
		if err != nil {
			return err
		}
		params := didww.NewQueryParams().
			Filter("country.id", countryID).
			Filter("name", "Arizona").
			Include("country").
			Sort("name").
			Page(1, 25)

		regions, err := client.Regions().List(ctx, params)
		if err != nil {
			return err
		}
		for _, region := range regions {
			if region.Country == nil || region.Country.ID != countryID {
				return fmt.Errorf("region %s: country not included", region.Name)
			}
		}
		return s.count("regions", len(regions), nil)
	})
}

// README: Date and Datetime Fields
func (s *suite) dateFields() {
	ctx, client := s.ctx, s.client
	s.section("Date and Datetime Fields")
	s.check("DID datetimes and identity birth date", func() error {
		first, err := s.anyDID()
		if err != nil {
			return err
		}
		did, err := client.DIDs().Find(ctx, first.ID)
		if err != nil {
			return err
		}
		s.logf("created at %s, expires at %v", did.CreatedAt, did.ExpiresAt)

		identities, err := client.Identities().List(ctx, didww.NewQueryParams().Page(1, 1))
		if err != nil {
			return err
		}
		if len(identities) == 0 {
			s.logf("no identity to read a birth date from")
			return nil
		}
		identity, err := client.Identities().Find(ctx, identities[0].ID)
		if err != nil {
			return err
		}
		s.logf("birth date %q", identity.BirthDate)
		return expect(!did.CreatedAt.IsZero(), "zero created_at")
	})
}
