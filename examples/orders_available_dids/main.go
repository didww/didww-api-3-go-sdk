// Orders an available DID using included DID group SKU.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders_available_dids/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/orderitem"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// Get available DIDs with included DID group and SKUs
	params := didww.NewQueryParams().Include("did_group.stock_keeping_units")
	available, err := client.AvailableDIDs().List(ctx, params)
	if err != nil {
		panic(err)
	}
	if len(available) == 0 {
		panic("No available DIDs found")
	}
	ad := available[0]
	fmt.Println("Available DID:", ad.Number)

	if ad.DIDGroup == nil || len(ad.DIDGroup.StockKeepingUnits) == 0 {
		panic("No stock_keeping_units found in included did_group")
	}

	// Create order with available DID
	order := &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.AvailableDidOrderItem{
				DidOrderItem: orderitem.DidOrderItem{
					SkuID: ad.DIDGroup.StockKeepingUnits[0].ID,
				},
				AvailableDidID: ad.ID,
			},
		},
	}

	created, err := client.Orders().Create(ctx, order)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Order %s status=%s items=%d\n", created.ID, created.Status, len(created.Items))
}
