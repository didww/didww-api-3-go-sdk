package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestRequirementsList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/requirements": {status: http.StatusOK, fixture: "requirements/index.json"},
	})

	requirements, err := client.Requirements().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(requirements) == 0 {
		t.Fatal("expected non-empty requirements list")
	}
}

func TestRequirementsFindWithIncludes(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/requirements/25d12afe-1ec6-4fe3-9621-b250dd1fb959": {status: http.StatusOK, fixture: "requirements/show.json"},
	})

	params := NewQueryParams().Include("country,did_group_type,personal_permanent_document,business_permanent_document,personal_onetime_document,business_onetime_document,personal_proof_types,business_proof_types,address_proof_types")
	req, err := client.Requirements().Find(context.Background(), "25d12afe-1ec6-4fe3-9621-b250dd1fb959", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.ID != "25d12afe-1ec6-4fe3-9621-b250dd1fb959" {
		t.Errorf("expected ID '25d12afe-1ec6-4fe3-9621-b250dd1fb959', got %q", req.ID)
	}
	if req.IdentityType != "Any" {
		t.Errorf("expected IdentityType 'Any', got %q", req.IdentityType)
	}
	if req.PersonalProofQty != 1 {
		t.Errorf("expected PersonalProofQty 1, got %d", req.PersonalProofQty)
	}
	if !req.ServiceDescriptionRequired {
		t.Error("expected ServiceDescriptionRequired to be true")
	}

	// Verify resolved country
	if req.Country == nil {
		t.Fatal("expected non-nil Country")
	}
	if req.Country.Name != "Spain" {
		t.Errorf("expected country name 'Spain', got %q", req.Country.Name)
	}

	// Verify resolved DIDGroupType
	if req.DIDGroupType == nil {
		t.Fatal("expected non-nil DIDGroupType")
	}
	if req.DIDGroupType.Name != "Local" {
		t.Errorf("expected DIDGroupType name 'Local', got %q", req.DIDGroupType.Name)
	}

	// Verify resolved document templates
	if req.PersonalPermanentDocument == nil {
		t.Fatal("expected non-nil PersonalPermanentDocument")
	}
	if req.PersonalPermanentDocument.Name != "Belgium Registration Form" {
		t.Errorf("expected PersonalPermanentDocument name 'Belgium Registration Form', got %q", req.PersonalPermanentDocument.Name)
	}
	if req.PersonalOnetimeDocument == nil {
		t.Fatal("expected non-nil PersonalOnetimeDocument")
	}
	if req.PersonalOnetimeDocument.Name != "Generic LOI" {
		t.Errorf("expected PersonalOnetimeDocument name 'Generic LOI', got %q", req.PersonalOnetimeDocument.Name)
	}

	// Verify resolved proof types
	if len(req.PersonalProofTypes) != 1 {
		t.Fatalf("expected 1 personal proof type, got %d", len(req.PersonalProofTypes))
	}
	if req.PersonalProofTypes[0].Name != "Drivers License" {
		t.Errorf("expected personal proof type 'Drivers License', got %q", req.PersonalProofTypes[0].Name)
	}
	if len(req.BusinessProofTypes) != 7 {
		t.Fatalf("expected 7 business proof types, got %d", len(req.BusinessProofTypes))
	}
	if len(req.AddressProofTypes) != 1 {
		t.Fatalf("expected 1 address proof type, got %d", len(req.AddressProofTypes))
	}
	if req.AddressProofTypes[0].Name != "Copy of Phone Bill" {
		t.Errorf("expected address proof type 'Copy of Phone Bill', got %q", req.AddressProofTypes[0].Name)
	}
}
