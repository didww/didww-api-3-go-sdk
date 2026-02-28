// Lists countries, demonstrates filtering, and fetches one country by ID.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/countries/
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

	// List all countries
	countries, err := client.Countries().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	for _, country := range countries {
		fmt.Printf("%s (+%s) [%s]\n", country.Name, country.Prefix, country.ISO)
	}

	// Filter countries by name
	params := didww.NewQueryParams().
		Filter("name", "United States").
		Page(1, 10)
	filtered, err := client.Countries().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nFiltered: %d countries\n", len(filtered))

	// Find a specific country
	if len(filtered) > 0 {
		country, err := client.Countries().Find(ctx, filtered[0].ID)
		if err != nil {
			panic(err)
		}
		fmt.Println("Found:", country.Name)
	}
}
