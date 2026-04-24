// CRUD for trunk groups with trunk relationships.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/voice_in_trunk_groups/
package main

import (
	"context"
	"fmt"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk"
	"github.com/didww/didww-api-3-go-sdk/examples"
	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/resource/trunkconfiguration"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// Helper for SIP config with all required fields
	sipConfig := func(host string) *trunkconfiguration.SIPConfiguration {
		return &trunkconfiguration.SIPConfiguration{
			Host:                host,
			Port:                5060,
			CodecIDs:            []enums.Codec{enums.CodecPCMU, enums.CodecPCMA},
			TransportProtocolID: enums.TransportProtocolUDP,
			RxDtmfFormatID:      enums.RxDtmfFormatRFC2833,
			TxDtmfFormatID:      enums.TxDtmfFormatRFC2833,
			SstRefreshMethodID:  enums.SstRefreshMethodInvite,
			SstMinTimer:         600,
			SstMaxTimer:         900,
			SstSessionExpires:   examples.Ptr(900),
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
	fmt.Println("Created trunk A:", trunkA.ID)

	trunkB, err := client.VoiceInTrunks().Create(ctx, &resource.VoiceInTrunk{
		Name:          fmt.Sprintf("Group Trunk B %d", ts),
		Configuration: sipConfig("sip-b.example.com"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created trunk B:", trunkB.ID)

	// Create a trunk group with both trunks
	extRef := fmt.Sprintf("go-vitg-%d", ts)
	group, err := client.VoiceInTrunkGroups().Create(ctx, &resource.VoiceInTrunkGroup{
		Name:                fmt.Sprintf("SDK Trunk Group %d", ts),
		CapacityLimit:       examples.Ptr(10),
		VoiceInTrunkIDs:     []string{trunkA.ID, trunkB.ID},
		ExternalReferenceID: &extRef,
	})
	if err != nil {
		panic(err)
	}
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

	// Update group name
	group.Name = "Updated Group"
	updated, err := client.VoiceInTrunkGroups().Update(ctx, group)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nUpdated name:", updated.Name)

	// Cleanup: deleting the group cascades to delete assigned trunks
	if err := client.VoiceInTrunkGroups().Delete(ctx, group.ID); err != nil {
		panic(err)
	}
	fmt.Println("Deleted trunk group (cascaded to trunks)")
}
