// Lists identities with country and birth_country (2026-04-16).
//
// 2026-04-16 adds:
//   - birth_country has_one relationship on Identity
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/identities/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Identities ===")
	params := didww.NewQueryParams().Include("country,birth_country")
	identities, err := client.Identities().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d identities\n", len(identities))

	limit := 10
	if len(identities) < limit {
		limit = len(identities)
	}
	for _, id := range identities[:limit] {
		fmt.Printf("\nIdentity: %s\n", id.ID)
		fmt.Printf("  Name: %s %s\n", id.FirstName, id.LastName)
		fmt.Printf("  Phone: %s\n", id.PhoneNumber)
		fmt.Printf("  Type: %s\n", id.IdentityType)
		if id.Country != nil {
			fmt.Printf("  Country: %s\n", id.Country.Name)
		}
		if id.BirthCountry != nil {
			fmt.Printf("  Birth Country: %s\n", id.BirthCountry.Name)
		}
		fmt.Printf("  Birth Date: %s\n", id.BirthDate)
		fmt.Printf("  Verified: %v\n", id.Verified)
	}
}
