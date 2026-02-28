// Creates a DID order by SKU resolved from DID groups.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders_sku/
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

	params := didww.NewQueryParams().
		Include("stock_keeping_units").
		Page(1, 1)
	didGroups, err := client.DIDGroups().List(ctx, params)
	if err != nil {
		panic(err)
	}
	if len(didGroups) == 0 || len(didGroups[0].StockKeepingUnits) == 0 {
		panic("No DID group with stock_keeping_units found")
	}

	order := &didww.Order{
		Items: []didww.OrderItem{
			{
				Type: "did_order_items",
				Attributes: didww.OrderItemAttributes{
					SkuID: didGroups[0].StockKeepingUnits[0].ID,
					Qty:   2,
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

	if len(created.Items) > 0 {
		fmt.Printf("Item type=%s\n", created.Items[0].Type)
	}
}
