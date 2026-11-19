// Lists Voice IN trunks with their trunk group, creates a SIP trunk and a PSTN
// trunk, finds the SIP trunk by ID, updates both trunks and deletes them, also
// when a step fails.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/trunks/
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

	// List voice in trunks with included relationships
	params := didww.NewQueryParams().
		Include("voice_in_trunk_group").
		Page(1, 10)
	trunks, err := client.VoiceInTrunks().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("=== Voice IN trunks (%d) ===\n", len(trunks))
	for _, trunk := range trunks {
		printTrunk(trunk)
	}

	// Create a SIP trunk
	sipTrunk, err := client.VoiceInTrunks().Create(ctx, &resource.VoiceInTrunk{
		Name:           "My SIP Trunk " + suffix,
		Priority:       1,
		Weight:         100,
		CliFormat:      enums.CliFormatE164,
		RingingTimeout: didww.Ptr(30),
		CapacityLimit:  didww.Ptr(10),
		Pop:            didww.Ptr(enums.PopFRA), // optional; leave it nil to let DIDWW choose
		Configuration: &trunkconfiguration.SIPConfiguration{
			Host: "sip.example.com",
			Port: 5060,
			// Codecs by name, in priority order; left out, DIDWW applies its default list
			Codecs:              []enums.Codec{enums.CodecPCMU, enums.CodecPCMA, enums.CodecTelephoneEvent},
			IceMode:             enums.IceModeAccept,
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
			ReroutingDisconnectCodeIDs: []enums.ReroutingDisconnectCode{
				enums.DCSIP408RequestTimeout,
				enums.DCSIP503ServiceUnavailable,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	defer deleteTrunk(ctx, client, "SIP", sipTrunk.ID)
	fmt.Printf("\nCreated SIP trunk: %s - %s\n", sipTrunk.ID, sipTrunk.Name)

	// Create a PSTN trunk
	pstnTrunk, err := client.VoiceInTrunks().Create(ctx, &resource.VoiceInTrunk{
		Name:           "My PSTN Trunk " + suffix,
		RingingTimeout: didww.Ptr(30),
		Configuration: &trunkconfiguration.PSTNConfiguration{
			Dst: "12125551234",
		},
	})
	if err != nil {
		panic(err)
	}
	defer deleteTrunk(ctx, client, "PSTN", pstnTrunk.ID)
	fmt.Printf("Created PSTN trunk: %s - %s\n", pstnTrunk.ID, pstnTrunk.Name)

	// Find the SIP trunk by ID
	found, err := client.VoiceInTrunks().Find(ctx, sipTrunk.ID, didww.NewQueryParams().Include("voice_in_trunk_group"))
	if err != nil {
		panic(err)
	}
	fmt.Println("\nFound trunk:")
	printTrunk(found)

	// Update the SIP trunk: only the changed description is sent
	found.Description = didww.Ptr("Updated description")
	updatedSIP, err := client.VoiceInTrunks().Update(ctx, found)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nUpdated SIP trunk description:", *updatedSIP.Description)

	// Update the PSTN trunk: the configuration type must stay the same
	pstnTrunk.Name = "My PSTN Trunk " + suffix + " (v2)"
	pstnTrunk.Configuration = &trunkconfiguration.PSTNConfiguration{Dst: "12125556789"}
	updatedPSTN, err := client.VoiceInTrunks().Update(ctx, pstnTrunk)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Updated PSTN trunk: %s", updatedPSTN.Name)
	if pstn, ok := updatedPSTN.Configuration.(*trunkconfiguration.PSTNConfiguration); ok {
		fmt.Printf(" -> %s", pstn.Dst)
	}
	fmt.Print("\n\n")
}

func printTrunk(trunk *resource.VoiceInTrunk) {
	pop := "auto"
	if trunk.Pop != nil {
		pop = *trunk.Pop
	}
	group := "none"
	if trunk.VoiceInTrunkGroup != nil {
		group = trunk.VoiceInTrunkGroup.Name
	}
	fmt.Printf("%s [%T] pop=%s group=%s\n", trunk.Name, trunk.Configuration, pop, group)
	switch cfg := trunk.Configuration.(type) {
	case *trunkconfiguration.SIPConfiguration:
		fmt.Printf("  host=%s codecs=%v ice=%s rtcp_mux=%s rtcp_feedback=%s\n",
			cfg.Host, cfg.Codecs, cfg.IceMode, cfg.RtcpMuxMode, cfg.RtcpFeedbackMode)
	case *trunkconfiguration.PSTNConfiguration:
		fmt.Printf("  dst=%s\n", cfg.Dst)
	}
}

func deleteTrunk(ctx context.Context, client *didww.Client, kind, id string) {
	if err := client.VoiceInTrunks().Delete(ctx, id); err != nil {
		fmt.Printf("Could not delete %s trunk: %v\n", kind, err)
		return
	}
	fmt.Printf("%s trunk deleted\n", kind)
}
