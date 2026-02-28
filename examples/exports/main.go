// Creates and lists CDR exports.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/exports/
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

	// Create an export
	export := &didww.Export{
		ExportType: "cdr_in",
		Filters:    map[string]interface{}{"year": 2025, "month": 1},
	}
	created, err := client.Exports().Create(ctx, export)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created export:", created.ID)
	fmt.Println("  type:", created.ExportType)
	fmt.Println("  status:", created.Status)

	// List exports
	exports, err := client.Exports().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nAll exports (%d):\n", len(exports))
	for _, e := range exports {
		fmt.Printf("  %s %s [%s]\n", e.ID, e.ExportType, e.Status)
	}
}
