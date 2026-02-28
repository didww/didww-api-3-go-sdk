// Purchases capacity by creating a capacity order item.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders_capacity/
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

	// Get capacity pools
	pools, err := client.CapacityPools().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	if len(pools) == 0 {
		panic("No capacity pools found")
	}
	pool := pools[0]
	fmt.Println("Capacity pool:", pool.Name)

	// Purchase capacity
	order := &didww.Order{
		Items: []didww.OrderItem{
			{
				Type: "capacity_order_items",
				Attributes: didww.OrderItemAttributes{
					CapacityPoolID: pool.ID,
					Qty:            1,
				},
			},
		},
	}

	created, err := client.Orders().Create(ctx, order)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Order %s status=%s items=%d\n", created.ID, created.Status, len(created.Items))
}
