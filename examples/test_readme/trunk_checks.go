package main

import (
	"errors"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/authenticationmethod"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/smstrunkconfiguration"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/trunkconfiguration"
)

const sipHost = "sip.example.com"

// createSIPTrunk creates a SIP trunk for a check.
func (s *suite) createSIPTrunk(name string) (*resource.VoiceInTrunk, error) {
	return s.client.VoiceInTrunks().Create(s.ctx, &resource.VoiceInTrunk{
		Name: name + " " + s.suffix,
		Configuration: &trunkconfiguration.SIPConfiguration{
			Host:                sipHost,
			Port:                5060,
			Codecs:              []enums.Codec{enums.CodecPCMU, enums.CodecPCMA},
			TransportProtocolID: enums.TransportProtocolUDP,
		},
	})
}

// README: Voice In Trunks, SIP Registration
func (s *suite) voiceInTrunks() {
	ctx, client := s.ctx, s.client
	s.section("Voice In Trunks")

	s.check("Create a SIP trunk, update and delete it", func() (err error) {
		// Create SIP trunk
		ringingTimeout := 30
		trunk := &resource.VoiceInTrunk{
			Name:           "My SIP Trunk " + s.suffix,
			Priority:       1,
			Weight:         100,
			CliFormat:      enums.CliFormatE164,
			RingingTimeout: &ringingTimeout,
			Configuration: &trunkconfiguration.SIPConfiguration{
				Host:                sipHost,
				Port:                5060,
				Codecs:              []enums.Codec{enums.CodecPCMA, enums.CodecPCMU, enums.CodecTelephoneEvent},
				IceMode:             enums.IceModeAccept,
				TransportProtocolID: enums.TransportProtocolUDP,
			},
		}
		created, err := client.VoiceInTrunks().Create(ctx, trunk)
		if err != nil {
			return err
		}
		deleted := false
		defer func() {
			if !deleted {
				s.undo(&err, "delete the trunk", s.deleteVoiceInTrunk(created.ID))
			}
		}()

		// Update trunk
		desc := "Updated"
		created.Description = &desc
		updated, err := client.VoiceInTrunks().Update(ctx, created)
		if err != nil {
			return err
		}
		if updated.Description == nil || *updated.Description != desc {
			return errors.New("description not updated")
		}

		// Delete trunk
		err = client.VoiceInTrunks().Delete(ctx, created.ID)
		deleted = err == nil
		return err
	})

	s.check("Pin a trunk to a point of presence", func() (err error) {
		created, err := s.createSIPTrunk("README POP trunk")
		if err != nil {
			return err
		}
		defer s.undo(&err, "delete the trunk", s.deleteVoiceInTrunk(created.ID))

		trunk, err := client.VoiceInTrunks().Find(ctx, created.ID)
		if err != nil {
			return err
		}
		trunk.Pop = didww.Ptr(enums.PopAMS)
		updated, err := client.VoiceInTrunks().Update(ctx, trunk)
		if err != nil {
			return err
		}
		return expect(updated.Pop != nil && *updated.Pop == enums.PopAMS, "pop %v, want %s", updated.Pop, enums.PopAMS)
	})

	s.check("SIP registration: enable on create, disable by setting Host", func() (err error) {
		trunk := &resource.VoiceInTrunk{
			Name: "Office (registered) " + s.suffix,
			Configuration: &trunkconfiguration.SIPConfiguration{
				EnabledSipRegistration: didww.Ptr(true),
				UseDIDInRuri:           didww.Ptr(true),
				CnamLookup:             didww.Ptr(true),
			},
		}
		created, err := client.VoiceInTrunks().Create(ctx, trunk)
		if err != nil {
			return err
		}
		defer s.undo(&err, "delete the trunk", s.deleteVoiceInTrunk(created.ID))
		sip, ok := created.Configuration.(*trunkconfiguration.SIPConfiguration)
		if !ok {
			return fmt.Errorf("configuration %T, want SIP", created.Configuration)
		}
		// sip.IncomingAuthUsername and sip.IncomingAuthPassword are server-generated
		if sip.IncomingAuthUsername == "" || sip.IncomingAuthPassword == "" {
			return errors.New("no SIP registration credentials were generated")
		}

		update := &resource.VoiceInTrunk{
			ID: created.ID,
			Configuration: &trunkconfiguration.SIPConfiguration{
				Host: sipHost,
			},
		}
		updated, err := client.VoiceInTrunks().Update(ctx, update)
		if err != nil {
			return err
		}
		cfg, ok := updated.Configuration.(*trunkconfiguration.SIPConfiguration)
		return expect(ok && cfg.Host == sipHost && (cfg.EnabledSipRegistration == nil || !*cfg.EnabledSipRegistration),
			"SIP registration is still enabled")
	})
}

// README: Voice In Trunk Groups
func (s *suite) voiceInTrunkGroups() {
	ctx, client := s.ctx, s.client
	s.section("Voice In Trunk Groups")
	s.check("Create a trunk group with two trunks", func() (err error) {
		trunkA, err := s.createSIPTrunk("README group trunk A")
		if err != nil {
			return err
		}
		defer s.undo(&err, "delete trunk A", s.deleteVoiceInTrunk(trunkA.ID))
		trunkB, err := s.createSIPTrunk("README group trunk B")
		if err != nil {
			return err
		}
		defer s.undo(&err, "delete trunk B", s.deleteVoiceInTrunk(trunkB.ID))

		capacityLimit := 50
		group := &resource.VoiceInTrunkGroup{
			Name:            "Primary Group " + s.suffix,
			CapacityLimit:   &capacityLimit,
			VoiceInTrunkIDs: []string{trunkA.ID, trunkB.ID},
		}
		created, err := client.VoiceInTrunkGroups().Create(ctx, group)
		if err != nil {
			return err
		}
		// Deleted before the trunks: a trunk in a group cannot be deleted
		defer s.undo(&err, "delete the trunk group", func() error { return client.VoiceInTrunkGroups().Delete(ctx, created.ID) })

		found, err := client.VoiceInTrunkGroups().Find(ctx, created.ID, didww.NewQueryParams().Include("voice_in_trunks"))
		if err != nil {
			return err
		}
		return expect(len(found.VoiceInTrunks) == 2, "the group has %d trunks, want 2", len(found.VoiceInTrunks))
	})
}

// README: Voice Out Trunks
func (s *suite) voiceOutTrunks() {
	ctx, client := s.ctx, s.client
	s.section("Voice Out Trunks")
	s.check("Create a voice OUT trunk with credentials and IP authentication", func() (err error) {
		// NOTE: 203.0.113.0/24 is RFC 5737 TEST-NET-3 documentation space.
		trunk := &resource.VoiceOutTrunk{
			Name: "My Outbound Trunk " + s.suffix,
			AuthenticationMethod: &authenticationmethod.CredentialsAndIp{
				AllowedSipIPs: []string{"203.0.113.0/24"},
			},
			DefaultDstAction:    enums.DefaultDstActionAllowAll,
			OnCliMismatchAction: enums.OnCliMismatchActionRejectCall,
			MediaEncryptionMode: enums.MediaEncryptionModeDisabled,
			Codecs: []enums.VoiceOutCodec{
				enums.VoiceOutCodecOpus,
				enums.VoiceOutCodecPCMA,
				enums.VoiceOutCodecTelephoneEvent,
			},
		}
		created, err := client.VoiceOutTrunks().Create(ctx, trunk)
		if err != nil {
			return err
		}
		defer s.undo(&err, "delete the voice OUT trunk", func() error { return client.VoiceOutTrunks().Delete(ctx, created.ID) })

		auth, ok := created.AuthenticationMethod.(*authenticationmethod.CredentialsAndIp)
		if !ok {
			return fmt.Errorf("authentication method %T, want credentials_and_ip", created.AuthenticationMethod)
		}
		// auth.Username and auth.Password are server-generated
		return expect(auth.Username != "" && auth.Password != "" && len(created.Codecs) == 3,
			"credentials generated: %t, codecs %v", auth.Username != "", created.Codecs)
	})
}

// README: SMS Trunks
func (s *suite) smsTrunks() {
	ctx, client := s.ctx, s.client
	s.section("SMS Trunks")

	s.check("Create an HTTP IN trunk, update it, clear attributes and list by type", func() (err error) {
		trunk, err := client.SmsTrunks().Create(ctx, &resource.SmsTrunk{
			Name: "Inbound webhook " + s.suffix,
			Configuration: &smstrunkconfiguration.HTTPIn{
				URL:        "https://example.com/sms",
				HTTPMethod: smstrunkconfiguration.HTTPMethodPost,
				BodyType:   smstrunkconfiguration.BodyTypeJSON,
				Body:       `{"text":"{SMS_TEXT}","from":"{SMS_SRC_ADDR}","to":"{SMS_DST_ADDR}"}`,
			},
		})
		if err != nil {
			return err
		}
		defer s.undo(&err, "delete the SMS trunk", s.deleteSmsTrunk(trunk.ID))

		// On update the configuration type must stay the same; attributes left out keep their values.
		trunk.Configuration = &smstrunkconfiguration.HTTPIn{URL: "https://example.com/sms/v2"}
		trunk, err = client.SmsTrunks().Update(ctx, trunk)
		if err != nil {
			return err
		}
		cfg, ok := trunk.Configuration.(*smstrunkconfiguration.HTTPIn)
		if !ok || cfg.URL != "https://example.com/sms/v2" || cfg.Body == nil {
			return fmt.Errorf("configuration after update: %v", trunk.Configuration)
		}

		// Nullable attributes are cleared with a Generic configuration of the same type: nil is sent as null
		trunk.Configuration = &smstrunkconfiguration.Generic{
			Type:       smstrunkconfiguration.TypeHTTPIn,
			Attributes: map[string]any{"http_method": "get", "body_type": nil, "body": nil},
		}
		trunk, err = client.SmsTrunks().Update(ctx, trunk)
		if err != nil {
			return err
		}
		cfg, ok = trunk.Configuration.(*smstrunkconfiguration.HTTPIn)
		if !ok || cfg.HTTPMethod != smstrunkconfiguration.HTTPMethodGet || cfg.Body != nil {
			return fmt.Errorf("configuration after clearing: %v", trunk.Configuration)
		}

		// List by configuration type, with the group
		params := didww.NewQueryParams().
			Filter("configuration.type", "http_in,sms_to_email").
			Include("sms_trunk_group")
		trunks, err := client.SmsTrunks().List(ctx, params)
		return s.count("http_in and sms_to_email trunks", len(trunks), err)
	})

	s.check("Regenerate the credentials of an HTTP OUT trunk", func() (err error) {
		created, err := s.createOutboundSmsTrunk("README HTTP OUT")
		if err != nil {
			return err
		}
		outTrunkID := created.ID
		defer s.undo(&err, "delete the SMS trunk", s.deleteSmsTrunk(outTrunkID))
		before := ""
		if cfg, ok := created.Configuration.(*smstrunkconfiguration.HTTPOut); ok {
			before = cfg.AuthPassword
		}

		_, err = client.SmsTrunkRegenerateCredentials().Create(ctx, &resource.SmsTrunkRegenerateCredential{
			SmsTrunkID: outTrunkID,
		})
		if err != nil {
			return err
		}
		outTrunk, err := client.SmsTrunks().Find(ctx, outTrunkID)
		if err != nil {
			return err
		}
		switch cfg := outTrunk.Configuration.(type) {
		case *smstrunkconfiguration.HTTPOut:
			// cfg.AuthName and cfg.AuthPassword are the new credentials
			return expect(cfg.AuthPassword != "" && cfg.AuthPassword != before, "the password did not change")
		case *smstrunkconfiguration.SmppEsme:
			// cfg.SystemID and cfg.Password are the new credentials
			return errors.New("got an smpp_esme configuration, want http_out")
		}
		return fmt.Errorf("configuration %T, want http_out", outTrunk.Configuration)
	})
}

// README: SMS Trunk Groups
func (s *suite) smsTrunkGroups() {
	ctx, client := s.ctx, s.client
	s.section("SMS Trunk Groups")
	s.check("Create a group and add a trunk to it", func() (err error) {
		first, err := s.createInboundSmsTrunk("README group trunk 1")
		if err != nil {
			return err
		}
		group, err := client.SmsTrunkGroups().Create(ctx, &resource.SmsTrunkGroup{
			Name:        "Inbound SMS " + s.suffix,
			SmsTrunkIDs: []string{first.ID},
		})
		if err != nil {
			s.undo(&err, "delete the SMS trunk", s.deleteSmsTrunk(first.ID))
			return err
		}
		groupID := group.ID
		// Deleting a group also deletes the trunks in it
		defer s.undo(&err, "delete the SMS trunk group", func() error { return client.SmsTrunkGroups().Delete(ctx, groupID) })
		s.logf("sms_trunks_count: %s", group.Meta["sms_trunks_count"]) // 1
		if group.Meta["sms_trunks_count"] != "1" {
			return fmt.Errorf("sms_trunks_count %q, want 1", group.Meta["sms_trunks_count"])
		}

		another, err := s.createInboundSmsTrunk("README group trunk 2")
		if err != nil {
			return err
		}
		added := false
		defer func() {
			if !added {
				s.undo(&err, "delete the SMS trunk", s.deleteSmsTrunk(another.ID))
			}
		}()

		group, err = client.SmsTrunkGroups().Find(ctx, group.ID, didww.NewQueryParams().Include("sms_trunks"))
		if err != nil {
			return err
		}
		ids := []string{another.ID}
		for _, t := range group.SmsTrunks {
			ids = append(ids, t.ID)
		}
		if _, err = client.SmsTrunkGroups().Update(ctx, &resource.SmsTrunkGroup{ID: group.ID, SmsTrunkIDs: ids}); err != nil {
			return err
		}
		added = true

		group, err = client.SmsTrunkGroups().Find(ctx, groupID, didww.NewQueryParams().Include("sms_trunks"))
		if err != nil {
			return err
		}
		return expect(len(group.SmsTrunks) == 2, "the group has %d trunks, want 2", len(group.SmsTrunks))
	})
}

// README: Sender ID Verifications
func (s *suite) senderIDVerifications() {
	ctx, client := s.ctx, s.client
	s.section("Sender ID Verifications")

	var verifications []*resource.SenderIDVerification
	s.check("List active verifications with their SMS trunk", func() (err error) {
		params := didww.NewQueryParams().
			Filter("status", string(enums.SenderIDVerificationStatusActive)).
			Include("sms_trunk")
		verifications, err = client.SenderIDVerifications().List(ctx, params)
		return s.count("active Sender ID verifications", len(verifications), err)
	})

	s.check("Route a verification through an SMS trunk, then unassign it", func() (err error) {
		if len(verifications) == 0 {
			return skip("the account has no active Sender ID verification; verifications are created outside the API")
		}
		original := verifications[0]
		trunk, err := s.createOutboundSmsTrunk("README sender ID")
		if err != nil {
			return err
		}
		defer s.undo(&err, "delete the SMS trunk", s.deleteSmsTrunk(trunk.ID))
		defer s.undo(&err, "restore the verification's SMS trunk", func() error {
			restore := &resource.SenderIDVerification{ID: original.ID, NullifySmsTrunk: true}
			if original.SmsTrunk != nil {
				restore = &resource.SenderIDVerification{ID: original.ID, SmsTrunkID: original.SmsTrunk.ID}
			}
			_, restoreErr := client.SenderIDVerifications().Update(ctx, restore)
			return restoreErr
		})

		// http_out, smpp_esme or smpp_smsc trunks only
		if _, err = client.SenderIDVerifications().Update(ctx, &resource.SenderIDVerification{
			ID:         verifications[0].ID,
			SmsTrunkID: trunk.ID,
		}); err != nil {
			return err
		}

		// Outbound messages stop until a trunk is assigned again
		_, err = client.SenderIDVerifications().Update(ctx, &resource.SenderIDVerification{
			ID:              verifications[0].ID,
			NullifySmsTrunk: true,
		})
		return err
	})
}

// README: Dirty PATCH Serialization
func (s *suite) dirtyPatch() {
	ctx, client := s.ctx, s.client
	s.section("Dirty PATCH Serialization")

	s.check("Update a fetched DID (dedicated channels)", func() (err error) {
		target, err := s.firstDID(
			didww.NewQueryParams().Include("did_group"),
			func(d *resource.DID) bool { return d.DIDGroup != nil && d.DIDGroup.AllowAdditionalChannels },
			"active DID whose DID group allows additional channels")
		if err != nil {
			return err
		}
		original := target.DedicatedChannelsCount
		defer s.undo(&err, "restore the dedicated channels", func() error {
			current, findErr := client.DIDs().Find(ctx, target.ID)
			if findErr != nil {
				return findErr
			}
			current.DedicatedChannelsCount = original
			_, updateErr := client.DIDs().Update(ctx, current)
			return updateErr
		})

		did, err := client.DIDs().Find(ctx, target.ID)
		if err != nil {
			return err
		}
		did.DedicatedChannelsCount = 5
		// PATCH payload includes only "dedicated_channels_count", not all attributes
		updated, err := client.DIDs().Update(ctx, did)
		if err != nil {
			return err
		}
		return expect(updated.DedicatedChannelsCount == 5, "dedicated channels %d, want 5", updated.DedicatedChannelsCount)
	})

	s.check("Build a resource for update", func() (err error) {
		target, err := s.anyDID()
		if err != nil {
			return err
		}
		defer s.undo(&err, "restore the description", func() error { return s.restoreDescription(target.ID, target.Description) })

		desc := "New description"
		updated, err := client.DIDs().Update(ctx, &resource.DID{
			ID:          target.ID,
			Description: &desc,
		})
		if err != nil {
			return err
		}
		// PATCH payload includes only "description"
		return expect(updated.Description != nil && *updated.Description == desc, "description not updated")
	})

	s.check("Clear a field with explicit null", func() (err error) {
		target, err := s.anyDID()
		if err != nil {
			return err
		}
		defer s.undo(&err, "restore the description", func() error { return s.restoreDescription(target.ID, target.Description) })
		if target.Description == nil {
			// Give the DID a description to clear
			if _, err = client.DIDs().Update(ctx, &resource.DID{ID: target.ID, Description: didww.Ptr("To be cleared")}); err != nil {
				return err
			}
		}

		did, err := client.DIDs().Find(ctx, target.ID)
		if err != nil {
			return err
		}
		did.Description = nil
		// PATCH payload: { "description": null }
		updated, err := client.DIDs().Update(ctx, did)
		if err != nil {
			return err
		}
		return expect(updated.Description == nil, "description %q, want null", derefString(updated.Description))
	})

	s.dirtyRelationships()
}

// dirtyRelationships assigns a check trunk to a routed DID, updates that trunk
// through the DID's include and puts the DID's routing back.
func (s *suite) dirtyRelationships() {
	ctx, client := s.ctx, s.client

	original, setupErr := s.routedDID()
	var checkTrunk *resource.VoiceInTrunk
	if setupErr == nil {
		checkTrunk, setupErr = s.createSIPTrunk("README dirty PATCH")
	}
	if setupErr == nil {
		defer s.check("Restore the DID's routing and delete the check trunk", func() (err error) {
			defer s.undo(&err, "delete the trunk", s.deleteVoiceInTrunk(checkTrunk.ID))
			return s.restoreVoiceRouting(original)
		})
	}

	s.check("Clear a relationship: assigning a trunk clears the trunk group", func() error {
		if setupErr != nil {
			return setupErr
		}
		did, err := client.DIDs().Find(ctx, original.ID)
		if err != nil {
			return err
		}
		did.VoiceInTrunkID = checkTrunk.ID
		// PATCH payload includes: "relationships": { "voice_in_trunk": { "data": { ... } }, "voice_in_trunk_group": { "data": null } }
		if _, err = client.DIDs().Update(ctx, did); err != nil {
			return err
		}

		routed, err := client.DIDs().Find(ctx, original.ID, didww.NewQueryParams().Include("voice_in_trunk,voice_in_trunk_group"))
		if err != nil {
			return err
		}
		return expect(routed.VoiceInTrunk != nil && routed.VoiceInTrunk.ID == checkTrunk.ID && routed.VoiceInTrunkGroup == nil,
			"the DID is not routed to the check trunk only")
	})

	s.check("Update an included resource", func() error {
		if setupErr != nil {
			return setupErr
		}
		params := didww.NewQueryParams().Include("voice_in_trunk")
		did, err := client.DIDs().Find(ctx, original.ID, params)
		if err != nil {
			return err
		}
		trunk := did.VoiceInTrunk
		// Only the check's own trunk is changed
		if trunk == nil || trunk.ID != checkTrunk.ID {
			return errors.New("the DID is not routed to the check trunk")
		}
		desc := "Updated via include"
		trunk.Description = &desc
		updated, err := client.VoiceInTrunks().Update(ctx, trunk)
		if err != nil {
			return err
		}
		return expect(updated.Description != nil && *updated.Description == desc, "description not updated")
	})
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
