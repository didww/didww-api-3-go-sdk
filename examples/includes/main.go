// Reads resources with included relationships (the JSON:API include parameter).
//
// The SDK sets the pointer (or slice) field of a relationship from the
// resources in the response's included data. Request the relationships you
// read with Include: a relationship whose resource is not in the response
// stays nil, the same as a relationship without a related resource, and its
// ID is not kept.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/includes/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// --- Without and with Include ---
	fmt.Println("=== Regions without include ===")
	regions, err := client.Regions().List(ctx, didww.NewQueryParams().Page(1, 3))
	if err != nil {
		panic(err)
	}
	for _, region := range regions {
		fmt.Printf("  %s - country resolved: %t\n", region.Name, region.Country != nil)
	}

	fmt.Println("\n=== Regions with country ===")
	regions, err = client.Regions().List(ctx, didww.NewQueryParams().Include("country").Page(1, 3))
	if err != nil {
		panic(err)
	}
	for _, region := range regions {
		country := "(not included)"
		if region.Country != nil {
			country = region.Country.Name
		}
		fmt.Printf("  %s - country: %s\n", region.Name, country)
	}

	// --- DID groups with included stock keeping units (to-many) ---
	fmt.Println("\n=== DID groups with SKUs ===")
	groups, err := client.DIDGroups().List(ctx, didww.NewQueryParams().Include("stock_keeping_units").Page(1, 3))
	if err != nil {
		panic(err)
	}
	for _, group := range groups {
		fmt.Printf("  %s %s - %d SKU(s)\n", group.Prefix, group.AreaName, len(group.StockKeepingUnits))
		for _, sku := range group.StockKeepingUnits {
			fmt.Printf("    $%s/mo, %d channels\n", sku.MonthlyPrice, sku.ChannelsIncludedCount)
		}
	}

	// --- Available DIDs with a nested include: did_group, then its SKUs ---
	fmt.Println("\n=== Available DIDs with did_group.stock_keeping_units ===")
	available, err := client.AvailableDIDs().List(ctx, didww.NewQueryParams().Include("did_group.stock_keeping_units").Page(1, 3))
	if err != nil {
		panic(err)
	}
	for _, ad := range available {
		if ad.DIDGroup == nil {
			fmt.Printf("  %s - DID group not included\n", ad.Number)
			continue
		}
		fmt.Printf("  %s - group: %s %s, %d SKU(s)\n",
			ad.Number, ad.DIDGroup.Prefix, ad.DIDGroup.AreaName, len(ad.DIDGroup.StockKeepingUnits))
	}

	// --- DIDs with their voice IN trunk ---
	fmt.Println("\n=== DIDs with voice_in_trunk ===")
	dids, err := client.DIDs().List(ctx, didww.NewQueryParams().Include("voice_in_trunk").Page(1, 3))
	if err != nil {
		panic(err)
	}
	for _, did := range dids {
		// Requested with Include, so nil here means the DID has no trunk
		trunk := "(none)"
		if did.VoiceInTrunk != nil {
			trunk = did.VoiceInTrunk.Name
		}
		// Not requested, so nil here says nothing about the DID's trunk group
		fmt.Printf("  %s - trunk: %s, trunk group resolved: %t\n", did.Number, trunk, did.VoiceInTrunkGroup != nil)
	}
}
