// Orders a DID number by NPA/NXX prefix.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders_nanpa/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk"
	"github.com/didww/didww-api-3-go-sdk/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// Step 1: find the NANPA prefix by NPA-NXX (e.g. 201-221)
	npa := "201"
	nxx := "221"
	params := didww.NewQueryParams().Filter("npanxx", npa+nxx).Page(1, 1)
	nanpaPrefixes, err := client.NanpaPrefixes().List(ctx, params)
	if err != nil {
		panic(err)
	}
	if len(nanpaPrefixes) == 0 {
		panic(fmt.Sprintf("NANPA prefix %s-%s not found", npa, nxx))
	}
	nanpaPrefix := nanpaPrefixes[0]
	fmt.Printf("NANPA prefix: %s NPA=%s NXX=%s\n", nanpaPrefix.ID, nanpaPrefix.NPA, nanpaPrefix.NXX)

	// Step 2: find a DID group for this prefix and load its SKUs
	dgParams := didww.NewQueryParams().
		Filter("nanpa_prefix.id", nanpaPrefix.ID).
		Include("stock_keeping_units").
		Page(1, 1)
	didGroups, err := client.DIDGroups().List(ctx, dgParams)
	if err != nil {
		panic(err)
	}
	if len(didGroups) == 0 || len(didGroups[0].StockKeepingUnits) == 0 {
		panic("No DID group with SKUs found for this NANPA prefix")
	}
	sku := didGroups[0].StockKeepingUnits[0]
	fmt.Printf("DID group: %s SKU: %s (monthly=%s)\n", didGroups[0].ID, sku.ID, sku.MonthlyPrice)

	// Step 3: create the order
	order := &didww.Order{
		AllowBackOrdering: true,
		Items: []didww.OrderItem{
			{
				Type: "did_order_items",
				Attributes: didww.OrderItemAttributes{
					SkuID:         sku.ID,
					NanpaPrefixID: nanpaPrefix.ID,
					Qty:           1,
				},
			},
		},
	}

	created, err := client.Orders().Create(ctx, order)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Order %s amount=%s status=%s ref=%s\n",
		created.ID, created.Amount, created.Status, created.Reference)
}
