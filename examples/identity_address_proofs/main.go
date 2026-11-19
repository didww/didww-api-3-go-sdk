// Creates an identity and an address, uploads an encrypted file and attaches
// proofs to both (2026-04-16).
//
// Steps:
//  1. Create an identity with a country and a birth_country (2026-04-16)
//  2. Create an address linked to the identity
//  3. Pick a proof type for the identity and one for the address
//  4. Encrypt the bundled example.pdf and upload it, one file per request
//  5. Create a proof for the identity and one for the address
//  6. Read the identity and the address again and print their verified flags
//  7. Delete the address, the identity and the uploaded files, also when a
//     step fails
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/identity_address_proofs/
package main

import (
	"context"
	_ "embed"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
)

//go:embed example.pdf
var examplePDF []byte

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	var identityID, addressID string
	var fileIDs []string
	defer func() {
		fmt.Println("\nCleaning up...")
		if addressID != "" {
			report(client.Addresses().Delete(ctx, addressID), "address", addressID)
		}
		if identityID != "" {
			report(client.Identities().Delete(ctx, identityID), "identity", identityID)
		}
		for _, id := range fileIDs {
			report(client.EncryptedFiles().Delete(ctx, id), "encrypted file", id)
		}
	}()

	// --- Step 1: Create an identity ---
	countries, err := client.Countries().List(ctx, didww.NewQueryParams().Page(1, 1))
	if err != nil {
		panic(err)
	}
	if len(countries) == 0 {
		panic("No countries found")
	}
	country := countries[0]
	fmt.Printf("Using country: %s (%s)\n", country.Name, country.ID)

	identity, err := client.Identities().Create(ctx, &resource.Identity{
		FirstName:      "John",
		LastName:       "Doe",
		PhoneNumber:    "12125551234",
		IdentityType:   enums.IdentityTypePersonal,
		CountryID:      country.ID,
		BirthCountryID: country.ID, // 2026-04-16: birth_country relationship
	})
	if err != nil {
		panic(err)
	}
	identityID = identity.ID
	fmt.Printf("Created identity: %s (%s %s)\n", identity.ID, identity.FirstName, identity.LastName)

	// --- Step 2: Create an address linked to the identity ---
	address, err := client.Addresses().Create(ctx, &resource.Address{
		CityName:   "New York",
		PostalCode: "10001",
		Address:    "123 Main St",
		IdentityID: identity.ID,
		CountryID:  country.ID,
	})
	if err != nil {
		panic(err)
	}
	addressID = address.ID
	fmt.Printf("Created address: %s (%s)\n", address.ID, address.Address)

	// --- Step 3: Pick proof types (entity_type is personal, business or address) ---
	proofTypes, err := client.ProofTypes().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	identityProofType := findProofType(proofTypes, string(enums.IdentityTypePersonal))
	addressProofType := findProofType(proofTypes, "address")
	if identityProofType == nil || addressProofType == nil {
		panic("No proof type found for a personal identity or an address")
	}
	fmt.Printf("Identity proof type: %s (%s)\n", identityProofType.Name, identityProofType.ID)
	fmt.Printf("Address proof type: %s (%s)\n", addressProofType.Name, addressProofType.ID)

	// --- Step 4: Encrypt and upload the file, once for each proof ---
	enc, err := didww.NewEncrypt(ctx, client)
	if err != nil {
		panic(err)
	}
	for _, label := range []string{"identity_proof", "address_proof"} {
		fileIDs = append(fileIDs, uploadFile(ctx, client, enc, label))
	}
	fmt.Printf("Uploaded encrypted files: %v\n", fileIDs)

	// --- Step 5: Create the proofs ---
	identityProof, err := client.Proofs().Create(ctx, &resource.Proof{
		EntityID:    identity.ID,
		EntityType:  "identities",
		ProofTypeID: identityProofType.ID,
		FileIDs:     []string{fileIDs[0]},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created identity proof: %s (type=%s)\n", identityProof.ID, identityProofType.Name)

	addressProof, err := client.Proofs().Create(ctx, &resource.Proof{
		EntityID:    address.ID,
		EntityType:  "addresses",
		ProofTypeID: addressProofType.ID,
		FileIDs:     []string{fileIDs[1]},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created address proof: %s (type=%s)\n", addressProof.ID, addressProofType.Name)

	// --- Step 6: Read the verified flags ---
	identity, err = client.Identities().Find(ctx, identity.ID, didww.NewQueryParams().Include("country,birth_country"))
	if err != nil {
		panic(err)
	}
	address, err = client.Addresses().Find(ctx, address.ID, didww.NewQueryParams().Include("proofs"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nIdentity %s: verified=%t\n", identity.ID, identity.Verified)
	if identity.BirthCountry != nil {
		fmt.Printf("  Birth Country: %s\n", identity.BirthCountry.Name)
	}
	fmt.Printf("Address %s: verified=%t, proofs=%d\n", address.ID, address.Verified, len(address.Proofs))
}

// uploadFile encrypts the bundled PDF and uploads it; one file per request.
func uploadFile(ctx context.Context, client *didww.Client, enc *didww.Encrypt, label string) string {
	encrypted, err := enc.Encrypt(examplePDF)
	if err != nil {
		panic(err)
	}
	id, err := client.UploadEncryptedFile(ctx, encrypted, label+".pdf.enc", enc.Fingerprint(), label+".pdf")
	if err != nil {
		panic(err)
	}
	return id
}

func findProofType(proofTypes []*resource.ProofType, entityType string) *resource.ProofType {
	for _, pt := range proofTypes {
		if pt.EntityType == entityType {
			return pt
		}
	}
	return nil
}

func report(err error, what, id string) {
	if err != nil {
		fmt.Printf("  Could not delete %s %s: %v\n", what, id, err)
		return
	}
	fmt.Printf("  Deleted %s: %s\n", what, id)
}
