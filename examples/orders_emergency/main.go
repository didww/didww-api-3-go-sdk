// Creates an Order with an EmergencyOrderItem (2026-04-16).
//
// This example demonstrates ordering an emergency calling service
// by submitting an EmergencyOrderItem with an emergency_calling_service_id
// obtained from the emergency requirements workflow.
//
// Usage: DIDWW_API_KEY=your_api_key EMERGENCY_CALLING_SERVICE_ID=xxx \
//
//	go run ./examples/orders_emergency/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/didww/didww-api-3-go-sdk/examples"
	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/orderitem"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	ecsID := os.Getenv("EMERGENCY_CALLING_SERVICE_ID")
	if ecsID == "" {
		fmt.Fprintln(os.Stderr, "EMERGENCY_CALLING_SERVICE_ID is required")
		os.Exit(1)
	}

	fmt.Println("=== Creating Emergency Order ===")
	extRef := "go-emergency-order"
	order, err := client.Orders().Create(ctx, &resource.Order{
		ExternalReferenceID: &extRef,
		Items: []orderitem.OrderItem{
			&orderitem.EmergencyOrderItem{
				EmergencyCallingServiceID: ecsID,
				Qty:                       1,
			},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating order: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Order created: %s\n", order.ID)
	fmt.Printf("  Reference: %s\n", order.Reference)
	fmt.Printf("  Status: %s\n", order.Status)
	fmt.Printf("  Amount: %s\n", order.Amount)
	fmt.Printf("  Description: %s\n", order.Description)
	if order.ExternalReferenceID != nil {
		fmt.Printf("  External Reference: %s\n", *order.ExternalReferenceID)
	}

	for i, item := range order.Items {
		fmt.Printf("\n  Item %d:\n", i+1)
		if emItem, ok := item.(*orderitem.EmergencyOrderItem); ok {
			fmt.Printf("    Type: emergency_order_items\n")
			fmt.Printf("    Qty: %d\n", emItem.Qty)
			fmt.Printf("    NRC: %s\n", emItem.Nrc)
			fmt.Printf("    MRC: %s\n", emItem.Mrc)
		}
	}
}
