// Lists trunks, creates SIP and PSTN trunks, updates and deletes them.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/trunks/
package main

import (
	"context"
	"fmt"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/trunkconfiguration"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// List voice in trunks with included relationships
	params := didww.NewQueryParams().
		Include("pop", "voice_in_trunk_group").
		Page(1, 10)
	trunks, err := client.VoiceInTrunks().List(ctx, params)
	if err != nil {
		panic(err)
	}
	for _, trunk := range trunks {
		fmt.Printf("%s [%T]\n", trunk.Name, trunk.Configuration)
	}

	// Create a SIP trunk
	newTrunk := &resource.VoiceInTrunk{
		Name:           "My SIP Trunk",
		Priority:       1,
		Weight:         100,
		CliFormat:      enums.CliFormatE164,
		RingingTimeout: examples.Ptr(30),
		CapacityLimit:  examples.Ptr(10),
		Configuration: &trunkconfiguration.SIPConfiguration{
			Host:                "sip.example.com",
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
			ReroutingDisconnectCodeIDs: []enums.ReroutingDisconnectCode{
				enums.DCSIP408RequestTimeout,
				enums.DCSIP503ServiceUnavailable,
			},
		},
	}
	created, err := client.VoiceInTrunks().Create(ctx, newTrunk)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created trunk: %s - %s\n", created.ID, created.Name)

	// Update trunk
	created.Description = examples.Ptr("Updated description")
	updated, err := client.VoiceInTrunks().Update(ctx, created)
	if err != nil {
		panic(err)
	}
	fmt.Println("Updated trunk description:", *updated.Description)

	// Delete trunk
	if err := client.VoiceInTrunks().Delete(ctx, created.ID); err != nil {
		panic(err)
	}
	fmt.Println("SIP trunk deleted")

	// --- Create a PSTN trunk ---
	pstnTrunk := &resource.VoiceInTrunk{
		Name:           "My PSTN Trunk",
		RingingTimeout: examples.Ptr(30),
		Configuration: &trunkconfiguration.PSTNConfiguration{
			Dst: "12125551234",
		},
	}
	createdPstn, err := client.VoiceInTrunks().Create(ctx, pstnTrunk)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created PSTN trunk: %s - %s\n", createdPstn.ID, createdPstn.Name)

	if pstn, ok := createdPstn.Configuration.(*trunkconfiguration.PSTNConfiguration); ok {
		fmt.Println("  DST:", pstn.Dst)
	}

	// Delete PSTN trunk
	if err := client.VoiceInTrunks().Delete(ctx, createdPstn.ID); err != nil {
		panic(err)
	}
	fmt.Println("PSTN trunk deleted")
}
