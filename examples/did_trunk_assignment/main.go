// Demonstrates exclusive trunk/trunk group assignment on DIDs.
// Assigning a trunk auto-nullifies the trunk group and vice versa.
//
// The example creates its own trunk and trunk group. At the end the DID's
// original voice routing and description are restored and the created trunk
// and trunk group are deleted, also when a step fails. The routing is restored
// by assigning the original trunk or trunk group again, so the example uses a
// DID that is routed to one of them.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/did_trunk_assignment/
package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/trunkconfiguration"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()
	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)

	// Get a DID and remember its routing and description to put them back at the end
	did := findRoutedDID(ctx, client)
	fmt.Printf("Using DID: %s (%s)\n", did.Number, did.ID)
	printDIDAssignment(ctx, client, did.ID)

	// Create a trunk
	trunk, err := client.VoiceInTrunks().Create(ctx, &resource.VoiceInTrunk{
		Name: "Assignment Trunk " + suffix,
		Pop:  didww.Ptr(enums.PopFRA), // optional; leave it nil to let DIDWW choose
		Configuration: &trunkconfiguration.SIPConfiguration{
			Host:                "sip.example.com",
			Port:                5060,
			Codecs:              []enums.Codec{enums.CodecPCMU, enums.CodecPCMA},
			TransportProtocolID: enums.TransportProtocolUDP,
		},
	})
	if err != nil {
		panic(err)
	}
	defer deleteTrunk(ctx, client, trunk.ID)
	fmt.Println("\nCreated trunk:", trunk.ID)

	// Create a trunk group
	group, err := client.VoiceInTrunkGroups().Create(ctx, &resource.VoiceInTrunkGroup{
		Name:          "Assignment Group " + suffix,
		CapacityLimit: didww.Ptr(10),
	})
	if err != nil {
		panic(err)
	}
	defer deleteTrunkGroup(ctx, client, group.ID)
	fmt.Println("Created trunk group:", group.ID)

	// Restore the DID first, so the created trunk and group are no longer in use
	defer restoreDID(ctx, client, did)

	// 1. Assign trunk to DID (sends voice_in_trunk_group as null)
	updateDID(ctx, client, &resource.DID{ID: did.ID, VoiceInTrunkID: trunk.ID})
	fmt.Println("\n1. Assigned trunk:")
	printDIDAssignment(ctx, client, did.ID)

	// 2. Assign trunk group to DID (sends voice_in_trunk as null)
	updateDID(ctx, client, &resource.DID{ID: did.ID, VoiceInTrunkGroupID: group.ID})
	fmt.Println("\n2. Assigned trunk group (the trunk is cleared):")
	printDIDAssignment(ctx, client, did.ID)

	// 3. Re-assign trunk (sends voice_in_trunk_group as null again)
	updateDID(ctx, client, &resource.DID{ID: did.ID, VoiceInTrunkID: trunk.ID})
	fmt.Println("\n3. Re-assigned trunk:")
	printDIDAssignment(ctx, client, did.ID)

	// 4. Update the description only: only changed fields are sent, so the trunk stays
	updateDID(ctx, client, &resource.DID{ID: did.ID, Description: didww.Ptr("DID with trunk assigned")})
	fmt.Println("\n4. Updated description only (trunk stays):")
	printDIDAssignment(ctx, client, did.ID)
}

// findRoutedDID returns the first DID that is routed to a voice IN trunk or trunk group.
func findRoutedDID(ctx context.Context, client *didww.Client) *resource.DID {
	params := didww.NewQueryParams().
		Filter("terminated", "false").
		Include("voice_in_trunk,voice_in_trunk_group").
		Page(1, 50)
	dids, err := client.DIDs().List(ctx, params)
	if err != nil {
		panic(err)
	}
	for _, did := range dids {
		if did.VoiceInTrunk != nil || did.VoiceInTrunkGroup != nil {
			return did
		}
	}
	panic("No DID routed to a voice IN trunk or trunk group found. Route a DID first.")
}

func updateDID(ctx context.Context, client *didww.Client, did *resource.DID) {
	if _, err := client.DIDs().Update(ctx, did); err != nil {
		panic(err)
	}
}

func printDIDAssignment(ctx context.Context, client *didww.Client, didID string) {
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

// restoreDID assigns the original trunk or trunk group again and puts back the
// original description, which is sent as null when it was empty.
func restoreDID(ctx context.Context, client *didww.Client, original *resource.DID) {
	current, err := client.DIDs().Find(ctx, original.ID)
	if err != nil {
		fmt.Println("\nCould not restore the DID:", err)
		return
	}
	if original.VoiceInTrunk != nil {
		current.VoiceInTrunkID = original.VoiceInTrunk.ID
	} else {
		current.VoiceInTrunkGroupID = original.VoiceInTrunkGroup.ID
	}
	current.Description = original.Description
	if _, err := client.DIDs().Update(ctx, current); err != nil {
		fmt.Println("\nCould not restore the DID:", err)
		return
	}
	fmt.Println("\nRestored the DID's original routing:")
	printDIDAssignment(ctx, client, original.ID)
}

func deleteTrunk(ctx context.Context, client *didww.Client, id string) {
	if err := client.VoiceInTrunks().Delete(ctx, id); err != nil {
		fmt.Println("Could not delete trunk:", err)
		return
	}
	fmt.Println("Deleted trunk")
}

func deleteTrunkGroup(ctx context.Context, client *didww.Client, id string) {
	if err := client.VoiceInTrunkGroups().Delete(ctx, id); err != nil {
		fmt.Println("Could not delete trunk group:", err)
		return
	}
	fmt.Println("Deleted trunk group")
}
