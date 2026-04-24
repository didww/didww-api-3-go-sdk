// Lists capacity pools with included shared capacity groups and qty-based pricings.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/capacity_pools/
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

	// List capacity pools with includes
	params := didww.NewQueryParams().Include("shared_capacity_groups", "qty_based_pricings")
	pools, err := client.CapacityPools().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Capacity pools (%d):\n", len(pools))

	for _, pool := range pools {
		fmt.Printf("\n  %s\n", pool.Name)
		fmt.Printf("    total channels: %d\n", pool.TotalChannelsCount)
		fmt.Printf("    assigned channels: %d\n", pool.AssignedChannelsCount)
		fmt.Printf("    renew date: %s\n", pool.RenewDate)

		// Shared capacity groups (included)
		if len(pool.SharedCapacityGroups) > 0 {
			fmt.Printf("    shared capacity groups (%d):\n", len(pool.SharedCapacityGroups))
			for _, g := range pool.SharedCapacityGroups {
				fmt.Printf("      %s shared=%d metered=%d\n",
					g.Name, g.SharedChannelsCount, g.MeteredChannelsCount)
			}
		}

		// Qty-based pricings (included)
		if len(pool.QtyBasedPricings) > 0 {
			fmt.Printf("    qty-based pricings (%d):\n", len(pool.QtyBasedPricings))
			for _, p := range pool.QtyBasedPricings {
				fmt.Printf("      qty=%d setup=%s monthly=%s\n",
					p.Qty, p.SetupPrice, p.MonthlyPrice)
			}
		}
	}
}
