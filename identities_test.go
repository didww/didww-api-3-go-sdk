package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func TestIdentitiesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/identities": {status: http.StatusOK, fixture: "identities/index.json"},
	})

	identities, err := client.Identities().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(identities) == 0 {
		t.Fatal("expected non-empty identities list")
	}

	first := identities[0]
	if first.FirstName != "John" {
		t.Errorf("expected FirstName 'John', got %q", first.FirstName)
	}
	if first.LastName != "Doe" {
		t.Errorf("expected LastName 'Doe', got %q", first.LastName)
	}
	if first.IdentityType != enums.IdentityTypePersonal {
		t.Errorf("expected IdentityType 'Personal', got %q", first.IdentityType)
	}
}

func TestIdentitiesCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/identities": {status: http.StatusCreated, fixture: "identities/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	idNumber := "ABC1234"
	companyName := "Test Company Limited"
	companyRegNumber := "543221"
	vatID := "GB1234"
	description := "test identity"
	personalTaxID := "987654321"
	externalRefID := "111"
	identity, err := server.client.Identities().Create(context.Background(), &Identity{
		FirstName:           "John",
		LastName:            "Doe",
		PhoneNumber:         "123456789",
		IDNumber:            &idNumber,
		BirthDate:           "1970-01-01",
		CompanyName:         &companyName,
		CompanyRegNumber:    &companyRegNumber,
		VatID:               &vatID,
		Description:         &description,
		PersonalTaxID:       &personalTaxID,
		IdentityType:        enums.IdentityTypeBusiness,
		ExternalReferenceID: &externalRefID,
		CountryID:           "1f6fc2bd-f081-4202-9b1a-d9cb88d942b9",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identity.ID != "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae" {
		t.Errorf("expected ID 'e96ae7d1-11d5-42bc-a5c5-211f3c3788ae', got %q", identity.ID)
	}
	if identity.FirstName != "John" {
		t.Errorf("expected FirstName 'John', got %q", identity.FirstName)
	}
	if identity.IdentityType != enums.IdentityTypeBusiness {
		t.Errorf("expected IdentityType 'Business', got %q", identity.IdentityType)
	}

	// Verify included country
	if identity.Country == nil {
		t.Fatal("expected non-nil Country")
	}
	if identity.Country.Name != "United States" {
		t.Errorf("expected country name 'United States', got %q", identity.Country.Name)
	}

	assertRequestJSON(t, capturedBody, "identities/create_request.json")
}

func TestIdentitiesCreatePersonal(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"POST /v3/identities": {status: http.StatusCreated, fixture: "identities/create_personal.json"},
	})

	identity, err := client.Identities().Create(context.Background(), &Identity{
		FirstName:    "John",
		LastName:     "Doe",
		PhoneNumber:  "123456789",
		IdentityType: enums.IdentityTypePersonal,
		BirthDate:    "1970-01-01",
		CountryID:    "1f6fc2bd-f081-4202-9b1a-d9cb88d942b9",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identity.ID != "9728ea13-cb5d-41fb-8a7f-796a005b0a13" {
		t.Errorf("expected ID '9728ea13-cb5d-41fb-8a7f-796a005b0a13', got %q", identity.ID)
	}
	if identity.IdentityType != enums.IdentityTypePersonal {
		t.Errorf("expected IdentityType 'Personal', got %q", identity.IdentityType)
	}
	if identity.FirstName != "John" {
		t.Errorf("expected FirstName 'John', got %q", identity.FirstName)
	}
	if identity.PersonalTaxID == nil || *identity.PersonalTaxID != "987654321" {
		t.Errorf("expected PersonalTaxID '987654321', got %v", identity.PersonalTaxID)
	}
}

func TestIdentitiesUpdate(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"PATCH /v3/identities/e96ae7d1-11d5-42bc-a5c5-211f3c3788ae": {status: http.StatusOK, fixture: "identities/update.json"},
	})

	identity, err := client.Identities().Update(context.Background(), &Identity{
		ID:        "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae",
		FirstName: "Jake",
		LastName:  "Johnson",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identity.ID != "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae" {
		t.Errorf("expected ID 'e96ae7d1-11d5-42bc-a5c5-211f3c3788ae', got %q", identity.ID)
	}
	if identity.FirstName != "Jake" {
		t.Errorf("expected FirstName 'Jake', got %q", identity.FirstName)
	}
	if identity.LastName != "Johnson" {
		t.Errorf("expected LastName 'Johnson', got %q", identity.LastName)
	}
	if identity.CompanyName == nil || *identity.CompanyName != "Some Company Limited" {
		t.Errorf("expected CompanyName 'Some Company Limited', got %v", identity.CompanyName)
	}
}

func TestIdentitiesDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/identities/e96ae7d1-11d5-42bc-a5c5-211f3c3788ae": {status: http.StatusNoContent},
	})

	err := client.Identities().Delete(context.Background(), "e96ae7d1-11d5-42bc-a5c5-211f3c3788ae")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
