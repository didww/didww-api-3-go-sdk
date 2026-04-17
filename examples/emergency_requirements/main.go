// Lists emergency service requirements for a country/did_group_type (2026-04-16).
//
// Emergency requirements describe what address precision, identity type,
// and supporting fields an end-customer must provide to enable 911/112
// on a DID.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/emergency_requirements/
package main

import (
	"context"
	"fmt"
	"strings"

	didww "github.com/didww/didww-api-3-go-sdk"
	"github.com/didww/didww-api-3-go-sdk/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Emergency Requirements ===")
	params := didww.NewQueryParams().Include("country,did_group_type")
	requirements, err := client.EmergencyRequirements().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d emergency requirements\n", len(requirements))

	limit := 5
	if len(requirements) < limit {
		limit = len(requirements)
	}
	for _, req := range requirements[:limit] {
		fmt.Printf("\nRequirement: %s\n", req.ID)
		if req.Country != nil {
			fmt.Printf("  Country: %s\n", req.Country.Name)
		}
		if req.DIDGroupType != nil {
			fmt.Printf("  DID Group Type: %s\n", req.DIDGroupType.Name)
		}
		fmt.Printf("  Identity type required: %s\n", req.IdentityType)
		fmt.Printf("  Address area level: %s\n", req.AddressAreaLevel)
		if len(req.AddressMandatoryFields) > 0 {
			fmt.Printf("  Address mandatory fields: %s\n", strings.Join(req.AddressMandatoryFields, ", "))
		}
	}
}
