// Inspects Emergency orders (2026-04-16).
//
// Emergency orders are created server-side when an EmergencyCallingService
// is activated or renewed — customers cannot POST them directly.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders_emergency/
package main

import (
	"context"
	"fmt"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/orderitem"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Recent Orders (filtering for Emergency) ===")
	params := didww.NewQueryParams().Sort("-created_at").Page(1, 50)
	orders, err := client.Orders().List(ctx, params)
	if err != nil {
		panic(err)
	}

	var emergencyOrders []*resource.Order
	for _, order := range orders {
		for _, item := range order.Items {
			if _, ok := item.(*orderitem.EmergencyOrderItem); ok {
				emergencyOrders = append(emergencyOrders, order)
				break
			}
		}
	}
	fmt.Printf("Found %d emergency orders out of %d total\n", len(emergencyOrders), len(orders))

	for i, order := range emergencyOrders {
		if i == 5 {
			break
		}
		fmt.Printf("\nOrder: %s\n", order.ID)
		fmt.Printf("  Reference: %s\n", order.Reference)
		fmt.Printf("  Status: %s\n", order.Status)
		fmt.Printf("  Amount: %s\n", order.Amount)
		fmt.Printf("  Created: %s\n", order.CreatedAt.Format(time.RFC3339))
		for _, item := range order.Items {
			if emItem, ok := item.(*orderitem.EmergencyOrderItem); ok {
				fmt.Printf("  Emergency Calling Service: %s\n", emItem.EmergencyCallingServiceID)
				fmt.Printf("  Qty: %d, NRC: %s, MRC: %s\n", emItem.Qty, emItem.Nrc, emItem.Mrc)
			}
		}
	}

	// Follow the link from an EmergencyCallingService to its order
	fmt.Println("\n=== Emergency Calling Service -> Order ===")
	ecsParams := didww.NewQueryParams().Include("order").Page(1, 1)
	services, err := client.EmergencyCallingServices().List(ctx, ecsParams)
	if err != nil {
		panic(err)
	}
	if len(services) == 0 {
		fmt.Println("No emergency_calling_services on this account")
		return
	}
	svc := services[0]
	fmt.Printf("ECS %s (%s)\n", svc.ID, svc.Name)
	if svc.Order != nil {
		fmt.Printf("  -> Order %s — status: %s, amount: %s\n", svc.Order.ID, svc.Order.Status, svc.Order.Amount)
	} else {
		fmt.Println("  -> No order linked yet")
	}
}
