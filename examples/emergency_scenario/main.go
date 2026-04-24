// End-to-end Emergency Calling Service scenario (2026-04-16).
//
// This example walks through the full flow of purchasing an Emergency
// Calling Service:
//
//  0. Find an address with a country, find an available DID with emergency
//     feature in that country, order via available DID, wait for completion.
//  1. Find a DID with the emergency feature that is not yet emergency-enabled.
//  2. Look up emergency requirements for that DID's country + did_group_type.
//  3. Find an existing identity on the account.
//  4. Find an existing address for that identity.
//  5. Validate the (emergency_requirement, address, identity) triple.
//  6. Create an emergency verification with callback_method="post".
//  7. Fetch the created verification to confirm its status.
//  8. Fetch the auto-created emergency_calling_service via the verification.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/emergency_scenario/
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk"
	"github.com/didww/didww-api-3-go-sdk/examples"
	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/orderitem"
)

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// === Step 0: Order an available DID with emergency feature ===
	fmt.Println("=== Step 0: Order an available DID with emergency feature ===")

	// Find an address first so we know what country to order in
	addrParams := didww.NewQueryParams().Include("country")
	addresses, err := client.Addresses().List(ctx, addrParams)
	if err != nil {
		log.Fatal("Failed to list addresses: ", err)
	}
	if len(addresses) == 0 {
		log.Fatal("No addresses on this account. Please create an address first.")
	}
	addressForOrder := addresses[0]
	addressCountry := addressForOrder.Country
	if addressCountry == nil {
		log.Fatal("Address has no included country")
	}
	fmt.Printf("  Using address country: %s (%s)\n", addressCountry.Name, addressCountry.ID)

	// Find an available DID with emergency feature in that country
	availParams := didww.NewQueryParams().
		Filter("did_group.features", "emergency").
		Filter("country.id", addressCountry.ID).
		Include("did_group,did_group.stock_keeping_units").
		Page(1, 1)
	availableDIDs, err := client.AvailableDIDs().List(ctx, availParams)
	if err != nil {
		log.Fatal("Failed to list available DIDs: ", err)
	}
	if len(availableDIDs) == 0 {
		log.Fatal("No available DIDs with emergency feature in this country.")
	}

	availableDID := availableDIDs[0]
	didGroup := availableDID.DIDGroup
	if didGroup == nil {
		log.Fatal("Available DID has no included did_group")
	}
	if len(didGroup.StockKeepingUnits) == 0 {
		log.Fatal("No SKU found for this DID group.")
	}
	sku := didGroup.StockKeepingUnits[0]

	fmt.Printf("  Available DID: %s\n", availableDID.Number)
	fmt.Printf("  DID Group: %s\n", didGroup.AreaName)

	order, err := client.Orders().Create(ctx, &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.AvailableDidOrderItem{
				DidOrderItem: orderitem.DidOrderItem{
					SkuID: sku.ID,
				},
				AvailableDidID: availableDID.ID,
			},
		},
	})
	if err != nil {
		log.Fatal("Failed to create order: ", err)
	}
	fmt.Printf("  Order: %s -- %s\n", order.ID, order.Status)

	// Wait for order to complete
	for i := 0; i < 10; i++ {
		if order.Status == "completed" {
			break
		}
		time.Sleep(5 * time.Second)
		order, err = client.Orders().Find(ctx, order.ID)
		if err != nil {
			log.Fatal("Failed to fetch order: ", err)
		}
	}
	if order.Status != "completed" {
		log.Fatalf("  Order did not complete (status: %s).", order.Status)
	}
	fmt.Println("  Order completed")

	// === Step 1: Find the newly ordered DID ===
	fmt.Println("\n=== Step 1: Find the newly ordered DID ===")
	didParams := didww.NewQueryParams().
		Filter("did_group.features", "emergency").
		Filter("emergency_enabled", "false").
		Include("did_group,did_group.country,did_group.did_group_type,emergency_calling_service").
		Sort("-created_at").
		Page(1, 10)
	dids, err := client.DIDs().List(ctx, didParams)
	if err != nil {
		log.Fatal("Failed to list DIDs: ", err)
	}

	// Pick a DID that is not yet assigned to an ECS
	var did *resource.DID
	for _, d := range dids {
		if d.EmergencyCallingService == nil {
			did = d
			break
		}
	}
	if did == nil {
		log.Fatal("No available DID without an existing Emergency Calling Service.")
	}

	didGroup = did.DIDGroup
	if didGroup == nil {
		log.Fatal("DID has no included did_group")
	}
	country := didGroup.Country
	dgt := didGroup.DIDGroupType

	fmt.Printf("  DID:            %s (%s)\n", did.Number, did.ID)
	fmt.Printf("  DID Group:      %s\n", didGroup.ID)
	if country != nil {
		fmt.Printf("  Country:        %s (%s)\n", country.Name, country.ID)
	}
	if dgt != nil {
		fmt.Printf("  DID Group Type: %s (%s)\n", dgt.Name, dgt.ID)
	}

	// === Step 2: Get emergency requirements ===
	fmt.Println("\n=== Step 2: Get emergency requirements for country + did_group_type ===")
	reqParams := didww.NewQueryParams()
	if country != nil {
		reqParams = reqParams.Filter("country.id", country.ID)
	}
	if dgt != nil {
		reqParams = reqParams.Filter("did_group_type.id", dgt.ID)
	}
	requirements, err := client.EmergencyRequirements().List(ctx, reqParams)
	if err != nil {
		log.Fatal("Failed to list emergency requirements: ", err)
	}
	if len(requirements) == 0 {
		log.Fatal("No emergency requirements found for this DID group")
	}
	req := requirements[0]
	fmt.Printf("  Emergency Requirement: %s\n", req.ID)
	fmt.Printf("  Identity type:         %s\n", req.IdentityType)
	fmt.Printf("  Address area level:    %s\n", req.AddressAreaLevel)

	// === Step 3: Find an existing identity ===
	fmt.Println("\n=== Step 3: Find identity ===")
	identityParams := didww.NewQueryParams().Page(1, 1)
	identities, err := client.Identities().List(ctx, identityParams)
	if err != nil {
		log.Fatal("Failed to list identities: ", err)
	}
	if len(identities) == 0 {
		log.Fatal("No identities found. Create an identity first.")
	}
	identity := identities[0]
	fmt.Printf("  Identity: %s\n", identity.ID)
	fmt.Printf("  Type:     %s\n", identity.IdentityType)

	// === Step 4: Find an existing address ===
	fmt.Println("\n=== Step 4: Find address ===")
	addrListParams := didww.NewQueryParams().Page(1, 1)
	addrList, err := client.Addresses().List(ctx, addrListParams)
	if err != nil {
		log.Fatal("Failed to list addresses: ", err)
	}
	if len(addrList) == 0 {
		log.Fatal("No addresses found. Create an address first.")
	}
	addr := addrList[0]
	fmt.Printf("  Address: %s\n", addr.ID)

	// === Step 5: Validate emergency requirement (dry-run) ===
	fmt.Println("\n=== Step 5: Validate emergency requirement (requirement + address + identity) ===")
	_, err = client.EmergencyRequirementValidations().Create(ctx, &resource.EmergencyRequirementValidation{
		EmergencyRequirementID: req.ID,
		AddressID:              addr.ID,
		IdentityID:             identity.ID,
	})
	if err != nil {
		log.Fatal("Validation failed: ", err)
	}
	fmt.Println("  Validation passed -- this combination can be used for emergency calling.")

	// === Step 6: Create an emergency verification ===
	fmt.Println("\n=== Step 6: Create emergency verification ===")
	suffix := randomHex(4)
	callbackMethod := "post"
	externalRef := fmt.Sprintf("go-scenario-%s", suffix)
	verification, err := client.EmergencyVerifications().Create(ctx, &resource.EmergencyVerification{
		CallbackURL:         examples.Ptr("https://example.com/webhooks/emergency"),
		CallbackMethod:      &callbackMethod,
		ExternalReferenceID: &externalRef,
		AddressID:           addr.ID,
		DIDIDs:              []string{did.ID},
	})
	if err != nil {
		log.Fatal("Failed to create emergency verification: ", err)
	}
	fmt.Printf("  Created verification: %s\n", verification.ID)
	fmt.Printf("  Reference:            %s\n", verification.Reference)
	fmt.Printf("  Status:               %s\n", verification.Status)
	if verification.ExternalReferenceID != nil {
		fmt.Printf("  External Reference:   %s\n", *verification.ExternalReferenceID)
	}

	// === Step 7: Fetch the verification to confirm status ===
	fmt.Println("\n=== Step 7: Fetch the created verification ===")
	fetchParams := didww.NewQueryParams().Include("address,emergency_calling_service,dids")
	fetched, err := client.EmergencyVerifications().Find(ctx, verification.ID, fetchParams)
	if err != nil {
		log.Fatal("Failed to fetch verification: ", err)
	}
	fmt.Printf("  Verification: %s\n", fetched.ID)
	fmt.Printf("  Status:       %s\n", fetched.Status)
	if fetched.DIDs != nil {
		numbers := make([]string, len(fetched.DIDs))
		for i, d := range fetched.DIDs {
			numbers[i] = d.Number
		}
		fmt.Printf("  DIDs:         %s\n", strings.Join(numbers, ", "))
	}
	if fetched.AddressRel != nil {
		fmt.Printf("  Address:      %s\n", fetched.AddressRel.ID)
	}

	// === Step 8: Fetch the auto-created emergency_calling_service ===
	fmt.Println("\n=== Step 8: Fetch emergency calling service ===")
	ecs := fetched.EmergencyCallingService
	if ecs != nil {
		// Re-fetch with includes for full details
		svcParams := didww.NewQueryParams().Include("country,did_group_type,dids")
		service, err := client.EmergencyCallingServices().Find(ctx, ecs.ID, svcParams)
		if err != nil {
			log.Fatal("Failed to fetch emergency calling service: ", err)
		}
		fmt.Printf("  Service:        %s\n", service.ID)
		fmt.Printf("  Name:           %s\n", service.Name)
		fmt.Printf("  Reference:      %s\n", service.Reference)
		fmt.Printf("  Status:         %s\n", service.Status)
		if service.Country != nil {
			fmt.Printf("  Country:        %s\n", service.Country.Name)
		}
		if service.DIDGroupType != nil {
			fmt.Printf("  DID Group Type: %s\n", service.DIDGroupType.Name)
		}
		if service.DIDs != nil {
			numbers := make([]string, len(service.DIDs))
			for i, d := range service.DIDs {
				numbers[i] = d.Number
			}
			fmt.Printf("  Attached DIDs:  %s\n", strings.Join(numbers, ", "))
		}
	} else {
		fmt.Println("  No emergency_calling_service linked yet (may be created asynchronously).")
	}

	fmt.Println("\nDone! Emergency calling service flow completed.")
}
