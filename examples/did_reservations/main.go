// Creates, lists, finds and deletes DID reservations.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/did_reservations/
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

	// Get an available DID to reserve
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
	fmt.Println("Reserving DID:", available[0].Number)

	// Create a reservation
	reservation := &didww.DIDReservation{
		Description:    "SDK example reservation",
		AvailableDIDID: available[0].ID,
	}
	created, err := client.DIDReservations().Create(ctx, reservation)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created reservation:", created.ID)
	fmt.Println("  description:", created.Description)
	fmt.Println("  expires at:", created.ExpireAt)

	// List reservations with includes
	listParams := didww.NewQueryParams().Include("available_did")
	reservations, err := client.DIDReservations().List(ctx, listParams)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nAll reservations (%d):\n", len(reservations))
	for _, r := range reservations {
		number := "unknown"
		if r.AvailableDID != nil {
			number = r.AvailableDID.Number
		}
		fmt.Printf("  %s - %s\n", r.ID, number)
	}

	// Find by ID
	found, err := client.DIDReservations().Find(ctx, created.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nFound reservation:", found.ID)

	// Delete reservation
	if err := client.DIDReservations().Delete(ctx, created.ID); err != nil {
		panic(err)
	}
	fmt.Println("Deleted reservation")
}
