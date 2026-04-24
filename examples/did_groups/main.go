// Fetches DID groups with included SKUs and shows group details.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/did_groups/
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

	// Fetch DID groups filtered by country with included stock_keeping_units
	params := didww.NewQueryParams().
		Filter("country.id", "1f6fc2bd-f081-4202-9b1a-d9cb88d942b9").
		Filter("area_name", "Beverly Hills").
		Include("stock_keeping_units")
	didGroups, err := client.DIDGroups().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d DID groups\n", len(didGroups))

	for _, dg := range didGroups {
		fmt.Printf("%s - %s prefix=%s features=%v metered=%v allow_additional_channels=%v\n",
			dg.ID, dg.AreaName, dg.Prefix, dg.Features, dg.IsMetered, dg.AllowAdditionalChannels)
		if dg.ServiceRestrictions != nil {
			fmt.Printf("  Service restrictions: %s\n", *dg.ServiceRestrictions)
		}
	}

	// Fetch a specific DID group
	if len(didGroups) > 0 {
		includeParams := didww.NewQueryParams().Include("stock_keeping_units")
		dg, err := client.DIDGroups().Find(ctx, didGroups[0].ID, includeParams)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Found: %s prefix=%s\n", dg.AreaName, dg.Prefix)
	}
}
