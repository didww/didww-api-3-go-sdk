// Reserves a DID and then places an order from that reservation.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/orders_reservation_dids/
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
	params := didww.NewQueryParams().
		Include("did_group.stock_keeping_units").
		Page(1, 1)
	available, err := client.AvailableDIDs().List(ctx, params)
	if err != nil {
		panic(err)
	}
	if len(available) == 0 {
		panic("No available DIDs found")
	}
	ad := available[0]
	fmt.Println("Reserving DID:", ad.Number)

	if ad.DIDGroup == nil || len(ad.DIDGroup.StockKeepingUnits) == 0 {
		panic("No stock_keeping_units found in included did_group")
	}
	skuID := ad.DIDGroup.StockKeepingUnits[0].ID

	// Create reservation
	reservation := &resource.DIDReservation{
		Description:    "SDK example reservation",
		AvailableDIDID: ad.ID,
	}
	created, err := client.DIDReservations().Create(ctx, reservation)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created reservation: %s (expires: %s)\n", created.ID, created.ExpiresAt)

	// Order reserved DID
	order := &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.ReservationDidOrderItem{
				DidOrderItem: orderitem.DidOrderItem{
					SkuID: skuID,
				},
				DidReservationID: created.ID,
			},
		},
	}
	orderedOrder, err := client.Orders().Create(ctx, order)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Order %s status=%s\n", orderedOrder.ID, orderedOrder.Status)
}
