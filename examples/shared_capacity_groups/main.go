// Creates a shared capacity group in a capacity pool.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/shared_capacity_groups/
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

	// Get a capacity pool
	pools, err := client.CapacityPools().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	if len(pools) == 0 {
		panic("No capacity pools found")
	}
	pool := pools[0]

	// Create a shared capacity group
	group := &didww.SharedCapacityGroup{
		Name:                 fmt.Sprintf("SDK Channel Group %d", time.Now().UnixMilli()),
		MeteredChannelsCount: 10,
		SharedChannelsCount:  1,
		CapacityPoolID:       pool.ID,
	}
	created, err := client.SharedCapacityGroups().Create(ctx, group)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created: %s name=%s metered=%d shared=%d\n",
		created.ID, created.Name, created.MeteredChannelsCount, created.SharedChannelsCount)
}
