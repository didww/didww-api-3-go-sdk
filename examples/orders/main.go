// Lists orders and creates/cancels a DID order using live SKU lookup.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk"
	"github.com/didww/didww-api-3-go-sdk/examples"
	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/orderitem"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// List orders
	orders, err := client.Orders().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	for _, order := range orders {
		fmt.Printf("Order %s: %s ($%s)", order.ID, order.Status, order.Amount)
		if order.ExternalReferenceID != nil {
			fmt.Printf(" [ref: %s]", *order.ExternalReferenceID)
		}
		fmt.Println()
		for _, item := range order.Items {
			fmt.Printf("  - %T\n", item)
		}
	}

	// Create an order with DID order items
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
	skuID := didGroups[0].StockKeepingUnits[0].ID

	extRef := "go-order-example"
	newOrder := &resource.Order{
		ExternalReferenceID: &extRef,
		Items: []orderitem.OrderItem{
			&orderitem.DidOrderItem{
				SkuID: skuID,
				Qty:   1,
			},
		},
	}
	created, err := client.Orders().Create(ctx, newOrder)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created order: %s - %s\n", created.ID, created.Status)

	// Delete order (cancel)
	if err := client.Orders().Delete(ctx, created.ID); err != nil {
		panic(err)
	}
	fmt.Println("Order cancelled")
}
