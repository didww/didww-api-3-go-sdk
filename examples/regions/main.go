// Lists regions with filters/includes and fetches a specific region.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/regions/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// Fetch regions filtered by country and name, with included country
	params := didww.NewQueryParams().
		Filter("country.id", "1f6fc2bd-f081-4202-9b1a-d9cb88d942b9").
		Filter("name", "Arizona").
		Include("country").
		Sort("-name")
	regions, err := client.Regions().List(ctx, params)
	if err != nil {
		panic(err)
	}
	for _, region := range regions {
		fmt.Printf("%s - %s\n", region.ID, region.Name)
	}

	// Fetch a specific region
	if len(regions) > 0 {
		includeParams := didww.NewQueryParams().Include("country")
		region, err := client.Regions().Find(ctx, regions[0].ID, includeParams)
		if err != nil {
			panic(err)
		}
		fmt.Println("Found:", region.Name)
	}
}
