// Lists recent orders, orders a DID by SKU and cancels (deletes) that order.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders/
package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/orderitem"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// List the five most recent orders
	fmt.Println("=== Recent Orders ===")
	listParams := didww.NewQueryParams().Sort("-created_at").Page(1, 5)
	orders, err := client.Orders().List(ctx, listParams)
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

	// Find a SKU to order
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

	// Order a DID by SKU
	fmt.Println("\n=== Creating DID Order ===")
	extRef := "go-order-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	created, err := client.Orders().Create(ctx, &resource.Order{
		ExternalReferenceID: &extRef,
		Items: []orderitem.OrderItem{
			&orderitem.DidOrderItem{
				SkuID: skuID,
				Qty:   1,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created order: %s - %s\n", created.ID, created.Status)
	fmt.Printf("  Amount: %s\n", created.Amount)
	fmt.Printf("  Reference: %s\n", created.Reference)
	fmt.Printf("  External reference: %s\n", extRef)

	// Cancel the order (delete it)
	fmt.Println("\n=== Canceling Order ===")
	if err := client.Orders().Delete(ctx, created.ID); err != nil {
		panic(err)
	}
	fmt.Println("Order canceled")
}
