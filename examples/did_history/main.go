// Lists DID ownership history (2026-04-16).
// Records are retained for the last 90 days only.
//
// Server-side filters supported:
//
//	did_number (eq), action (eq), method (eq),
//	created_at_gteq, created_at_lteq
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/did_history/
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

	// List most recent DID history events
	fmt.Println("=== Recent DID History ===")
	events, err := client.DIDHistory().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d events in the last 90 days\n", len(events))

	limit := 10
	if len(events) < limit {
		limit = len(events)
	}
	for _, event := range events[:limit] {
		fmt.Printf("  %s  %-16s  %-28s  via %s\n",
			event.CreatedAt.Format("2006-01-02T15:04:05Z"),
			event.DIDNumber,
			event.Action,
			event.Method,
		)
	}

	// Filter by action
	fmt.Println("\n=== Only 'assigned' events ===")
	params := didww.NewQueryParams().Filter("action", "assigned")
	assigned, err := client.DIDHistory().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d assignments\n", len(assigned))

	// Filter by a specific DID number
	if len(events) > 0 {
		number := events[0].DIDNumber
		fmt.Printf("\n=== History for DID %s ===\n", number)
		params = didww.NewQueryParams().Filter("did_number", number)
		perNumber, err := client.DIDHistory().List(ctx, params)
		if err != nil {
			panic(err)
		}
		for _, event := range perNumber {
			fmt.Printf("  %s  %s  via %s\n",
				event.CreatedAt.Format("2006-01-02T15:04:05Z"),
				event.Action,
				event.Method,
			)
		}
	}
}
