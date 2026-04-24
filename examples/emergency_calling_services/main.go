// Lists emergency calling services (2026-04-16).
//
// Emergency calling services are customer-owned subscriptions that enable 911/112
// on DIDs. They are created via the Orders API with an EmergencyOrderItem.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/emergency_calling_services/
package main

import (
	"context"
	"fmt"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk"
	"github.com/didww/didww-api-3-go-sdk/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Emergency Calling Services ===")
	params := didww.NewQueryParams().Include("country,did_group_type")
	services, err := client.EmergencyCallingServices().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d emergency calling services\n", len(services))

	for _, svc := range services {
		fmt.Printf("\nService: %s\n", svc.ID)
		fmt.Printf("  Name: %s\n", svc.Name)
		fmt.Printf("  Reference: %s\n", svc.Reference)
		fmt.Printf("  Status: %s\n", svc.Status)
		fmt.Printf("  Created: %s\n", svc.CreatedAt.Format(time.RFC3339))
		if svc.ActivatedAt != nil {
			fmt.Printf("  Activated: %s\n", svc.ActivatedAt.Format(time.RFC3339))
		}
		if svc.Country != nil {
			fmt.Printf("  Country: %s\n", svc.Country.Name)
		}
		if svc.DIDGroupType != nil {
			fmt.Printf("  DID Group Type: %s\n", svc.DIDGroupType.Name)
		}
		if svc.Meta != nil {
			fmt.Printf("  Setup Price: %s\n", svc.Meta["setup_price"])
			fmt.Printf("  Monthly Price: %s\n", svc.Meta["monthly_price"])
		}
	}

	// Find a specific service
	if len(services) > 0 {
		fmt.Printf("\n=== Details for %s ===\n", services[0].ID)
		svc, err := client.EmergencyCallingServices().Find(ctx, services[0].ID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  Name: %s\n", svc.Name)
		fmt.Printf("  Status: %s\n", svc.Status)
	}
}
