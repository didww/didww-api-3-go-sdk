// Creates and lists CDR exports (cdr_in / cdr_out).
//
// 2026-04-16 additions:
//   - external_reference_id: customer-supplied reference (max 100 chars)
//
// Filter semantics on CDR exports:
//   - filters.from: lower bound, INCLUSIVE (server: time_start >= from)
//   - filters.to:   upper bound, EXCLUSIVE (server: time_start < to)
//
// To cover a whole day, pass from: "2026-04-15 00:00:00", to: "2026-04-16 00:00:00".
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/exports/
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/didww/didww-api-3-go-sdk/examples"
	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// List existing exports
	fmt.Println("=== Existing Exports ===")
	exports, err := client.Exports().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d exports\n", len(exports))

	limit := 5
	if len(exports) < limit {
		limit = len(exports)
	}
	for _, e := range exports[:limit] {
		fmt.Printf("Export: %s\n", e.ID)
		fmt.Printf("  Type: %s\n", e.ExportType)
		fmt.Printf("  Status: %s\n", e.Status)
		fmt.Printf("  Created: %s\n", e.CreatedAt.Format(time.RFC3339))
		if e.URL != nil {
			fmt.Printf("  URL: %s\n", *e.URL)
		}
		if e.ExternalReferenceID != nil {
			fmt.Printf("  External Reference: %s\n", *e.ExternalReferenceID)
		}
		fmt.Println()
	}

	// Create a CDR-In export for yesterday (from is inclusive, to is exclusive)
	fmt.Println("\n=== Creating CDR-In Export (yesterday) ===")
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	suffix := fmt.Sprintf("%d", now.UnixMilli())[:8]
	extRef := fmt.Sprintf("go-cdr-in-%s", suffix)

	export := &resource.Export{
		ExportType: enums.ExportTypeCdrIn,
		Filters: map[string]interface{}{
			"from": yesterday.Format("2006-01-02") + " 00:00:00", // inclusive
			"to":   now.Format("2006-01-02") + " 00:00:00",       // exclusive
		},
		ExternalReferenceID: &extRef,
	}
	created, err := client.Exports().Create(ctx, export)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created CDR-In export: %s\n", created.ID)
	fmt.Printf("  Status: %s\n", created.Status)
	if created.ExternalReferenceID != nil {
		fmt.Printf("  External Reference: %s\n", *created.ExternalReferenceID)
	}

	// Find and inspect the specific export
	if len(exports) > 0 {
		fmt.Println("\n=== Specific Export Details ===")
		specific, err := client.Exports().Find(ctx, exports[0].ID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Export: %s\n", specific.ID)
		fmt.Printf("  Type: %s\n", specific.ExportType)
		fmt.Printf("  Status: %s\n", specific.Status)
		if specific.ExternalReferenceID != nil {
			fmt.Printf("  External Reference: %s\n", *specific.ExternalReferenceID)
		}
	}
}
