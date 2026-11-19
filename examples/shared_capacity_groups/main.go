// Creates a shared capacity group in a capacity pool, lists the groups,
// updates the new group and deletes it, also when a step fails.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/shared_capacity_groups/
package main

import (
	"context"
	"fmt"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
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
	fmt.Printf("Using capacity pool: %s (%s)\n", pool.Name, pool.ID)

	// Create a shared capacity group
	ts := time.Now().UnixMilli()
	extRef := fmt.Sprintf("go-scg-%d", ts)
	created, err := client.SharedCapacityGroups().Create(ctx, &resource.SharedCapacityGroup{
		Name:                 fmt.Sprintf("SDK Channel Group %d", ts),
		MeteredChannelsCount: 10,
		SharedChannelsCount:  1,
		CapacityPoolID:       pool.ID,
		ExternalReferenceID:  &extRef,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if delErr := client.SharedCapacityGroups().Delete(ctx, created.ID); delErr != nil {
			fmt.Println("Could not delete shared capacity group:", delErr)
			return
		}
		fmt.Println("\nDeleted shared capacity group")
	}()
	fmt.Printf("Created: %s name=%s metered=%d shared=%d\n",
		created.ID, created.Name, created.MeteredChannelsCount, created.SharedChannelsCount)
	if created.ExternalReferenceID != nil {
		fmt.Printf("  External Reference: %s\n", *created.ExternalReferenceID)
	}

	// List the shared capacity groups of the pool with their pool and DIDs
	params := didww.NewQueryParams().
		Filter("capacity_pool.id", pool.ID).
		Include("capacity_pool,dids")
	groups, err := client.SharedCapacityGroups().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nShared capacity groups in %s (%d):\n", pool.Name, len(groups))
	for _, g := range groups {
		fmt.Printf("  %s shared=%d metered=%d dids=%d\n",
			g.Name, g.SharedChannelsCount, g.MeteredChannelsCount, len(g.DIDs))
	}

	// Update the group: only the changed fields are sent
	created.Name = fmt.Sprintf("SDK Channel Group %d (updated)", ts)
	created.MeteredChannelsCount = 5
	updated, err := client.SharedCapacityGroups().Update(ctx, created)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nUpdated: %s name=%s metered=%d shared=%d\n",
		updated.ID, updated.Name, updated.MeteredChannelsCount, updated.SharedChannelsCount)
}
