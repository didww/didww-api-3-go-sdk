// Routes a DID's inbound SMS to an SMS trunk, then to an SMS trunk group.
// The two are exclusive: assigning one sends the other as null in the same
// PATCH, just like the voice in trunk and voice in trunk group. At the end
// the DID is unrouted, as an SMS trunk or group cannot be deleted while DIDs
// are routed to it.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/did_sms_routing/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	did := findDID(ctx, client)
	trunk := findInboundTrunk(ctx, client)
	group := findTrunkGroup(ctx, client)

	// 1. Assign the SMS trunk (sends sms_trunk_group as null)
	fmt.Println("\n=== 1. Assigning SMS trunk to DID ===")
	did.SmsTrunkID = trunk.ID
	if _, err := client.DIDs().Update(ctx, did); err != nil {
		panic(fmt.Sprintf("Error assigning SMS trunk: %v", err))
	}
	printSMSRouting(ctx, client, did.ID)
	printRoutedDIDs(ctx, client, trunk)

	// 2. Assign the SMS trunk group (sends sms_trunk as null)
	if group != nil {
		assignTrunkGroup(ctx, client, did.ID, group)
	}

	// 3. Unroute the DID (sends sms_trunk and sms_trunk_group as null)
	fmt.Println("\n=== 3. Unrouting the DID's inbound SMS ===")
	if _, err := client.DIDs().Update(ctx, &resource.DID{
		ID:                   did.ID,
		NullifySmsTrunk:      true,
		NullifySmsTrunkGroup: true,
	}); err != nil {
		panic(fmt.Sprintf("Error unrouting DID: %v", err))
	}
	printSMSRouting(ctx, client, did.ID)
}

func findDID(ctx context.Context, client *didww.Client) *resource.DID {
	fmt.Println("=== Finding DID ===")
	dids, err := client.DIDs().List(ctx, didww.NewQueryParams().Page(1, 1))
	if err != nil {
		panic(err)
	}
	if len(dids) == 0 {
		panic("No DIDs found. Please order a DID first.")
	}
	did := dids[0]
	fmt.Printf("Using DID: %s (%s)\n", did.Number, did.ID)
	return did
}

func findInboundTrunk(ctx context.Context, client *didww.Client) *resource.SmsTrunk {
	fmt.Println("\n=== Finding inbound SMS trunk ===")
	trunks, err := client.SmsTrunks().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	for _, t := range trunks {
		if t.AllowIn {
			fmt.Printf("Selected SMS trunk: %s (%s)\n", t.Name, t.ID)
			return t
		}
	}
	panic("No inbound SMS trunk found. Run examples/sms_trunks or create one first.")
}

// findTrunkGroup returns the first SMS trunk group, or nil when there is none.
func findTrunkGroup(ctx context.Context, client *didww.Client) *resource.SmsTrunkGroup {
	fmt.Println("\n=== Finding SMS trunk group ===")
	groups, err := client.SmsTrunkGroups().List(ctx, didww.NewQueryParams().Page(1, 1))
	if err != nil {
		panic(err)
	}
	if len(groups) == 0 {
		fmt.Println("No SMS trunk group found; step 2 is skipped.")
		return nil
	}
	group := groups[0]
	fmt.Printf("Selected SMS trunk group: %s (%s)\n", group.Name, group.ID)
	return group
}

func printSMSRouting(ctx context.Context, client *didww.Client, didID string) {
	p := didww.NewQueryParams().Include("sms_trunk,sms_trunk_group")
	result, err := client.DIDs().Find(ctx, didID, p)
	if err != nil {
		panic(err)
	}
	trunkStr := "null"
	if result.SmsTrunk != nil {
		trunkStr = result.SmsTrunk.ID
	}
	groupStr := "null"
	if result.SmsTrunkGroup != nil {
		groupStr = result.SmsTrunkGroup.ID
	}
	fmt.Printf("   sms_trunk       = %s\n", trunkStr)
	fmt.Printf("   sms_trunk_group = %s\n", groupStr)
}

func printRoutedDIDs(ctx context.Context, client *didww.Client, trunk *resource.SmsTrunk) {
	fmt.Printf("\n=== DIDs routed to %s ===\n", trunk.Name)
	routed, err := client.DIDs().List(ctx, didww.NewQueryParams().Filter("sms_trunk.id", trunk.ID))
	if err != nil {
		panic(err)
	}
	for _, d := range routed {
		fmt.Printf("   %s\n", d.Number)
	}
}

func assignTrunkGroup(ctx context.Context, client *didww.Client, didID string, group *resource.SmsTrunkGroup) {
	fmt.Println("\n=== 2. Assigning SMS trunk group to DID (clears the SMS trunk) ===")
	freshDID, err := client.DIDs().Find(ctx, didID)
	if err != nil {
		panic(err)
	}
	freshDID.SmsTrunkGroupID = group.ID
	if _, err := client.DIDs().Update(ctx, freshDID); err != nil {
		panic(fmt.Sprintf("Error assigning SMS trunk group: %v", err))
	}
	printSMSRouting(ctx, client, didID)
}
