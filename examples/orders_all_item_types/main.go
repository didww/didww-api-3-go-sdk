// Places one order with the three DID order item variants:
//   - DidOrderItem: DIDs by SKU and quantity
//   - AvailableDidOrderItem: a specific available DID
//   - ReservationDidOrderItem: a DID reserved beforehand
//
// All three are did_order_items on the wire, so they fit one order. Capacity
// items (CapacityOrderItem) cannot share an order with DID items: the API
// answers "items - should have same type". Buy capacity in its own order, as
// in examples/orders_capacity.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders_all_item_types/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/orderitem"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// --- Find two available DIDs with their DID group and SKUs ---
	params := didww.NewQueryParams().
		Include("did_group.stock_keeping_units").
		Page(1, 10)
	available, err := client.AvailableDIDs().List(ctx, params)
	if err != nil {
		panic(err)
	}
	if len(available) < 2 {
		panic("Need at least 2 available DIDs for this example")
	}
	ad1, ad2 := available[0], available[1]
	skuID1 := firstSKU(ad1)
	skuID2 := firstSKU(ad2)
	fmt.Printf("Available DID 1: %s (%s) - for AvailableDidOrderItem\n", ad1.ID, ad1.Number)
	fmt.Printf("Available DID 2: %s (%s) - for ReservationDidOrderItem\n", ad2.ID, ad2.Number)

	// --- Find a SKU for the DidOrderItem (by quantity) ---
	groups, err := client.DIDGroups().List(ctx, didww.NewQueryParams().Include("stock_keeping_units").Page(1, 10))
	if err != nil {
		panic(err)
	}
	var skuForQty string
	for _, g := range groups {
		if len(g.StockKeepingUnits) > 0 {
			skuForQty = g.StockKeepingUnits[0].ID
			fmt.Printf("SKU for qty order: %s (group=%s)\n", skuForQty, g.ID)
			break
		}
	}
	if skuForQty == "" {
		panic("No DID group with stock_keeping_units found")
	}

	// --- Reserve the second available DID ---
	reservation, err := client.DIDReservations().Create(ctx, &resource.DIDReservation{
		Description:    "Reserved for order example",
		AvailableDIDID: ad2.ID,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Reservation:", reservation.ID)

	// --- One order with all three DID item variants ---
	order, err := client.Orders().Create(ctx, &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.DidOrderItem{
				SkuID: skuForQty,
				Qty:   1,
			},
			&orderitem.AvailableDidOrderItem{
				DidOrderItem:   orderitem.DidOrderItem{SkuID: skuID1},
				AvailableDidID: ad1.ID,
			},
			&orderitem.ReservationDidOrderItem{
				DidOrderItem:     orderitem.DidOrderItem{SkuID: skuID2},
				DidReservationID: reservation.ID,
			},
		},
	})
	if err != nil {
		// The reservation is not used by an order, so release it
		if delErr := client.DIDReservations().Delete(ctx, reservation.ID); delErr != nil {
			fmt.Println("Could not delete the reservation:", delErr)
		}
		panic(err)
	}

	fmt.Printf("\nOrder ID: %s\n", order.ID)
	fmt.Printf("Amount: %s\n", order.Amount)
	fmt.Printf("Status: %s\n", order.Status)
	fmt.Printf("Created at: %s\n", order.CreatedAt)
	fmt.Printf("Reference: %s\n", order.Reference)
	fmt.Printf("Items count: %d\n", len(order.Items))
	for i, item := range order.Items {
		fmt.Printf("  Item %d: %T\n", i+1, item)
	}

	// --- DIDs that belong to this order (none yet while the order is pending) ---
	dids, err := client.DIDs().List(ctx, didww.NewQueryParams().Filter("order.id", order.ID))
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nDIDs in order (%d):\n", len(dids))
	for _, did := range dids {
		capacityLimit := "none"
		if did.CapacityLimit != nil {
			capacityLimit = fmt.Sprint(*did.CapacityLimit)
		}
		fmt.Printf("  %s | %s | capacity limit=%s\n", did.ID, did.Number, capacityLimit)
	}
}

// firstSKU returns the first SKU of the available DID's included DID group.
func firstSKU(ad *resource.AvailableDID) string {
	if ad.DIDGroup == nil || len(ad.DIDGroup.StockKeepingUnits) == 0 {
		panic(fmt.Sprintf("No stock_keeping_units for available DID %s", ad.ID))
	}
	return ad.DIDGroup.StockKeepingUnits[0].ID
}
