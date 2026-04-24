// Validates emergency requirement data before ordering (2026-04-16).
//
// EmergencyRequirementValidation performs a dry-run check that the given
// address and identity satisfy an EmergencyRequirement. A successful POST
// returns 204 No Content, meaning the data is valid.
//
// Usage: DIDWW_API_KEY=your_api_key \
//
//	EMERGENCY_REQUIREMENT_ID=xxx ADDRESS_ID=yyy IDENTITY_ID=zzz \
//	go run ./examples/emergency_requirement_validations/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/didww/didww-api-3-go-sdk/v3/examples"
	"github.com/didww/didww-api-3-go-sdk/v3/resource"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	emergencyReqID := os.Getenv("EMERGENCY_REQUIREMENT_ID")
	addressID := os.Getenv("ADDRESS_ID")
	identityID := os.Getenv("IDENTITY_ID")

	if emergencyReqID == "" || addressID == "" || identityID == "" {
		fmt.Fprintln(os.Stderr, "EMERGENCY_REQUIREMENT_ID, ADDRESS_ID, and IDENTITY_ID are required")
		os.Exit(1)
	}

	fmt.Println("=== Validating Emergency Requirement ===")
	fmt.Printf("  Emergency Requirement: %s\n", emergencyReqID)
	fmt.Printf("  Address: %s\n", addressID)
	fmt.Printf("  Identity: %s\n", identityID)

	_, err := client.EmergencyRequirementValidations().Create(ctx, &resource.EmergencyRequirementValidation{
		EmergencyRequirementID: emergencyReqID,
		AddressID:              addressID,
		IdentityID:             identityID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nValidation passed (204 No Content)")
}
