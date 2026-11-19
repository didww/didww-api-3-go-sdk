// CRUD for trunk groups with trunk relationships: creates two trunks, creates
// a trunk group with them, lists groups, renames the group and deletes the
// group and the trunks.
//
// A trunk in a trunk group cannot be deleted, and deleting a group keeps its
// trunks, so the group is deleted first and the trunks after it, also when a
// step fails.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/voice_in_trunk_groups/
package main

import (
	"context"
	"fmt"
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

	// Helper for SIP config with all required fields
	sipConfig := func(host string) *trunkconfiguration.SIPConfiguration {
		return &trunkconfiguration.SIPConfiguration{
			Host:                host,
			Port:                5060,
			Codecs:              []enums.Codec{enums.CodecPCMU, enums.CodecPCMA},
			TransportProtocolID: enums.TransportProtocolUDP,
			RxDtmfFormatID:      enums.RxDtmfFormatRFC2833,
			TxDtmfFormatID:      enums.TxDtmfFormatRFC2833,
			SstRefreshMethodID:  enums.SstRefreshMethodInvite,
			SstMinTimer:         600,
			SstMaxTimer:         900,
			SstSessionExpires:   didww.Ptr(900),
			SipTimerB:           8000,
			DnsSrvFailoverTimer: 2000,
			RtpTimeout:          30,
			MediaEncryptionMode: enums.MediaEncryptionModeDisabled,
			StirShakenMode:      enums.StirShakenModeDisabled,
		}
	}

	ts := time.Now().UnixMilli()

	// Create two trunks
	trunkA, err := client.VoiceInTrunks().Create(ctx, &resource.VoiceInTrunk{
		Name:          fmt.Sprintf("Group Trunk A %d", ts),
		Configuration: sipConfig("sip-a.example.com"),
	})
	if err != nil {
		panic(err)
	}
	defer deleteTrunk(ctx, client, trunkA.ID)
	fmt.Println("Created trunk A:", trunkA.ID)

	trunkB, err := client.VoiceInTrunks().Create(ctx, &resource.VoiceInTrunk{
		Name:          fmt.Sprintf("Group Trunk B %d", ts),
		Configuration: sipConfig("sip-b.example.com"),
	})
	if err != nil {
		panic(err)
	}
	defer deleteTrunk(ctx, client, trunkB.ID)
	fmt.Println("Created trunk B:", trunkB.ID)

	// Create a trunk group with both trunks
	extRef := fmt.Sprintf("go-vitg-%d", ts)
	group, err := client.VoiceInTrunkGroups().Create(ctx, &resource.VoiceInTrunkGroup{
		Name:                fmt.Sprintf("SDK Trunk Group %d", ts),
		CapacityLimit:       didww.Ptr(10),
		VoiceInTrunkIDs:     []string{trunkA.ID, trunkB.ID},
		ExternalReferenceID: &extRef,
	})
	if err != nil {
		panic(err)
	}
	// Deferred last, so it runs first: the trunks can be deleted only outside a group
	defer deleteTrunkGroup(ctx, client, group.ID)
	fmt.Printf("Created trunk group: %s - %s\n", group.ID, group.Name)

	// List trunk groups with included trunks
	params := didww.NewQueryParams().Include("voice_in_trunks")
	groups, err := client.VoiceInTrunkGroups().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nAll trunk groups (%d):\n", len(groups))
	for _, g := range groups {
		trunkCount := len(g.VoiceInTrunks)
		fmt.Printf("  %s (%d trunks)", g.Name, trunkCount)
		if g.ExternalReferenceID != nil {
			fmt.Printf(" [ref: %s]", *g.ExternalReferenceID)
		}
		fmt.Println()
	}

	// Rename the group: only the changed name is sent
	group.Name = fmt.Sprintf("Updated Group %d", ts)
	updated, err := client.VoiceInTrunkGroups().Update(ctx, group)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nUpdated name:", updated.Name)
}

func deleteTrunkGroup(ctx context.Context, client *didww.Client, id string) {
	if err := client.VoiceInTrunkGroups().Delete(ctx, id); err != nil {
		fmt.Println("Could not delete trunk group:", err)
		return
	}
	fmt.Println("Deleted trunk group")
}

func deleteTrunk(ctx context.Context, client *didww.Client, id string) {
	if err := client.VoiceInTrunks().Delete(ctx, id); err != nil {
		fmt.Println("Could not delete trunk:", err)
		return
	}
	fmt.Println("Deleted trunk", id)
}
