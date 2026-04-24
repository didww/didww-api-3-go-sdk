// Lists and creates emergency verifications (2026-04-16).
//
// Emergency verifications link an address to an emergency calling service,
// validating that the end-user's location meets regulatory requirements.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/emergency_verifications/
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/didww/didww-api-3-go-sdk/v3/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Emergency Verifications ===")
	verifications, err := client.EmergencyVerifications().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d emergency verifications\n", len(verifications))

	for _, ev := range verifications {
		fmt.Printf("\nVerification: %s\n", ev.ID)
		fmt.Printf("  Reference: %s\n", ev.Reference)
		fmt.Printf("  Status: %s\n", ev.Status)
		fmt.Printf("  Created: %s\n", ev.CreatedAt.Format(time.RFC3339))
		if ev.ExternalReferenceID != nil {
			fmt.Printf("  External Reference: %s\n", *ev.ExternalReferenceID)
		}
		if len(ev.RejectReasons) > 0 {
			fmt.Printf("  Reject Reasons: %v\n", ev.RejectReasons)
		}
	}
}
