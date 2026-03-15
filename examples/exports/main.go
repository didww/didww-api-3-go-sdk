// Creates and lists CDR exports.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/exports/
package main

import (
	"context"
	"fmt"

	"github.com/didww/didww-api-3-go-sdk/examples"
	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// Create an export
	export := &resource.Export{
		ExportType: enums.ExportTypeCdrIn,
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
