// Demonstrates exclusive trunk/trunk group assignment on DIDs.
// Assigning a trunk auto-nullifies the trunk group and vice versa.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/did_trunk_assignment/
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

	// Get a DID to work with
	fmt.Println("=== Finding DID ===")
	didParams := didww.NewQueryParams().
		Include("voice_in_trunk,voice_in_trunk_group").
		Page(1, 1)
	dids, err := client.DIDs().List(ctx, didParams)
	if err != nil {
		panic(err)
	}
	if len(dids) == 0 {
		panic("No DIDs found. Please order a DID first.")
	}
	did := dids[0]
	fmt.Printf("Using DID: %s (%s)\n", did.Number, did.ID)

	// Get a trunk
	fmt.Println("\n=== Finding Trunk ===")
	trunkParams := didww.NewQueryParams().Page(1, 1)
	trunks, err := client.VoiceInTrunks().List(ctx, trunkParams)
	if err != nil {
		panic(err)
	}
	if len(trunks) == 0 {
		panic("No trunks found. Please create a trunk first.")
	}
	trunk := trunks[0]
	fmt.Printf("Selected trunk: %s (%s)\n", trunk.Name, trunk.ID)

	// Get a trunk group
	fmt.Println("\n=== Finding Trunk Group ===")
	groupParams := didww.NewQueryParams().Page(1, 1)
	groups, err := client.VoiceInTrunkGroups().List(ctx, groupParams)
	if err != nil {
		panic(err)
	}
	if len(groups) == 0 {
		panic("No trunk groups found. Please create a trunk group first.")
	}
	trunkGroup := groups[0]
	fmt.Printf("Selected trunk group: %s (%s)\n", trunkGroup.Name, trunkGroup.ID)

	printDIDAssignment := func(didID string) {
		p := didww.NewQueryParams().Include("voice_in_trunk,voice_in_trunk_group")
		result, err := client.DIDs().Find(ctx, didID, p)
		if err != nil {
			panic(err)
		}
		trunkStr := "null"
		if result.VoiceInTrunk != nil {
			trunkStr = result.VoiceInTrunk.ID
		}
		groupStr := "null"
		if result.VoiceInTrunkGroup != nil {
			groupStr = result.VoiceInTrunkGroup.ID
		}
		fmt.Printf("   trunk = %s\n", trunkStr)
		fmt.Printf("   group = %s\n", groupStr)
	}

	// 1. Assign trunk to DID (auto-nullifies trunk group)
	fmt.Println("\n=== 1. Assigning trunk to DID ===")
	did.VoiceInTrunkID = trunk.ID
	did.VoiceInTrunkGroupID = ""
	_, err = client.DIDs().Update(ctx, did)
	if err != nil {
		panic(fmt.Sprintf("Error assigning trunk: %v", err))
	}
	printDIDAssignment(did.ID)

	// 2. Assign trunk group to DID (auto-nullifies trunk)
	fmt.Println("\n=== 2. Assigning trunk group to DID ===")
	freshDID, err := client.DIDs().Find(ctx, did.ID, nil)
	if err != nil {
		panic(err)
	}
	freshDID.VoiceInTrunkGroupID = trunkGroup.ID
	freshDID.VoiceInTrunkID = ""
	_, err = client.DIDs().Update(ctx, freshDID)
	if err != nil {
		panic(fmt.Sprintf("Error assigning trunk group: %v", err))
	}
	printDIDAssignment(did.ID)

	// 3. Re-assign trunk (auto-nullifies trunk group again)
	fmt.Println("\n=== 3. Re-assigning trunk ===")
	freshDID, err = client.DIDs().Find(ctx, did.ID, nil)
	if err != nil {
		panic(err)
	}
	freshDID.VoiceInTrunkID = trunk.ID
	freshDID.VoiceInTrunkGroupID = ""
	_, err = client.DIDs().Update(ctx, freshDID)
	if err != nil {
		panic(fmt.Sprintf("Error re-assigning trunk: %v", err))
	}
	printDIDAssignment(did.ID)

	// 4. Update description only (trunk stays assigned)
	fmt.Println("\n=== 4. Updating description only (trunk stays) ===")
	freshDID, err = client.DIDs().Find(ctx, did.ID, nil)
	if err != nil {
		panic(err)
	}
	freshDID.Description = examples.Ptr("DID with trunk assigned")
	_, err = client.DIDs().Update(ctx, freshDID)
	if err != nil {
		panic(fmt.Sprintf("Error updating description: %v", err))
	}
	printDIDAssignment(did.ID)

	fmt.Println("\nDemonstration complete!")
}
