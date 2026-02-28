// CRUD for voice out trunks (requires account config).
//
// Note: Voice Out Trunks and some OnCliMismatchAction values (e.g. replace_cli, randomize_cli)
// require additional account configuration. Contact DIDWW support to enable.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/voice_out_trunks/
package main

import (
	"context"
	"fmt"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk"
	"github.com/didww/didww-api-3-go-sdk/examples"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// Create a voice out trunk
	trunk := &didww.VoiceOutTrunk{
		Name:                fmt.Sprintf("SDK Outbound Trunk %d", time.Now().UnixMilli()),
		AllowedSipIPs:       []string{"192.168.1.1"},
		AllowedRtpIPs:       []string{"192.168.1.1"},
		DstPrefixes:         []string{},
		DefaultDstAction:    enums.DefaultDstActionAllowAll,
		OnCliMismatchAction: enums.OnCliMismatchActionRejectCall,
		MediaEncryptionMode: enums.MediaEncryptionModeDisabled,
		ThresholdAmount:     examples.Ptr("100.00"),
	}
	created, err := client.VoiceOutTrunks().Create(ctx, trunk)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created voice out trunk:", created.ID)
	fmt.Println("  name:", created.Name)
	fmt.Println("  username:", created.Username)
	fmt.Println("  password:", created.Password)
	fmt.Println("  status:", created.Status)

	// List voice out trunks
	trunks, err := client.VoiceOutTrunks().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nAll voice out trunks (%d):\n", len(trunks))
	for _, t := range trunks {
		fmt.Printf("  %s (%s)\n", t.Name, t.Status)
	}

	// Update
	created.Name = "Updated Outbound Trunk"
	created.AllowedSipIPs = []string{"10.0.0.0/8"}
	updated, err := client.VoiceOutTrunks().Update(ctx, created)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nUpdated name:", updated.Name)

	// Delete
	if err := client.VoiceOutTrunks().Delete(ctx, created.ID); err != nil {
		panic(err)
	}
	fmt.Println("Deleted voice out trunk")
}
