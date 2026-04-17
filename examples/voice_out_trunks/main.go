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

	"github.com/didww/didww-api-3-go-sdk/examples"
	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/authenticationmethod"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// Create a voice out trunk with ip_only authentication
	trunk := &resource.VoiceOutTrunk{
		Name: fmt.Sprintf("SDK Outbound Trunk %d", time.Now().UnixMilli()),
		AuthenticationMethod: &authenticationmethod.IpOnly{
			AllowedSipIPs: []string{"203.0.113.1"},
		},
		AllowedRtpIPs:       []string{"203.0.113.1"},
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
	fmt.Println("  auth type:", created.AuthenticationMethod.AuthenticationType())
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
