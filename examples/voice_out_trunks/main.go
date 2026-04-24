// CRUD for voice out trunks using 2026-04-16 polymorphic authentication_method.
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

	"github.com/didww/didww-api-3-go-sdk/v3/examples"
	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/authenticationmethod"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// List voice out trunks
	trunks, err := client.VoiceOutTrunks().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d voice out trunks\n", len(trunks))
	for _, t := range trunks {
		fmt.Printf("  %s (%s)\n", t.Name, t.Status)
		fmt.Printf("    ID: %s\n", t.ID)
		if t.AuthenticationMethod != nil {
			fmt.Printf("    Auth type: %s\n", t.AuthenticationMethod.AuthenticationType())
			switch am := t.AuthenticationMethod.(type) {
			case *authenticationmethod.CredentialsAndIp:
				fmt.Printf("    Username: %s\n", am.Username)
			case *authenticationmethod.IpOnly:
				fmt.Printf("    Allowed SIP IPs: %v\n", am.AllowedSipIPs)
			case *authenticationmethod.Twilio:
				fmt.Printf("    Twilio Account SID: %s\n", am.TwilioAccountSid)
			}
		}
		if t.ExternalReferenceID != nil {
			fmt.Printf("    External Reference ID: %s\n", *t.ExternalReferenceID)
		}
		fmt.Printf("    Emergency Enable All: %v\n", t.EmergencyEnableAll)
		if t.RtpTimeout != nil {
			fmt.Printf("    RTP Timeout: %d\n", *t.RtpTimeout)
		}
	}

	// Create a voice out trunk with credentials_and_ip authentication
	// NOTE: 203.0.113.0/24 is RFC 5737 TEST-NET-3 documentation space.
	// Replace with the real CIDR of your SIP infrastructure.
	suffix := fmt.Sprintf("%d", time.Now().UnixMilli())
	extRef := fmt.Sprintf("go-example-%s", suffix[:8])
	rtpTimeout := 60
	trunk := &resource.VoiceOutTrunk{
		Name: fmt.Sprintf("SDK Outbound Trunk %s", suffix),
		AuthenticationMethod: &authenticationmethod.CredentialsAndIp{
			AllowedSipIPs: []string{"203.0.113.0/24"},
		},
		AllowedRtpIPs:       []string{"203.0.113.1"},
		DstPrefixes:         []string{},
		DefaultDstAction:    enums.DefaultDstActionAllowAll,
		OnCliMismatchAction: enums.OnCliMismatchActionRejectCall,
		MediaEncryptionMode: enums.MediaEncryptionModeDisabled,
		ThresholdAmount:     examples.Ptr("100.00"),
		ExternalReferenceID: &extRef,
		RtpTimeout:          &rtpTimeout,
	}
	created, err := client.VoiceOutTrunks().Create(ctx, trunk)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nCreated voice out trunk:", created.ID)
	fmt.Println("  name:", created.Name)
	fmt.Println("  auth type:", created.AuthenticationMethod.AuthenticationType())
	if cam, ok := created.AuthenticationMethod.(*authenticationmethod.CredentialsAndIp); ok {
		fmt.Println("  username:", cam.Username)
	}
	fmt.Println("  status:", created.Status)
	if created.ExternalReferenceID != nil {
		fmt.Println("  external reference:", *created.ExternalReferenceID)
	}

	// Update - change name and tech_prefix
	fmt.Println("\n=== Updating Voice Out Trunk ===")
	created.Name = "Updated Outbound Trunk"
	created.AuthenticationMethod = &authenticationmethod.CredentialsAndIp{
		AllowedSipIPs: []string{"203.0.113.0/24"},
		TechPrefix:    "9",
	}
	updated, err := client.VoiceOutTrunks().Update(ctx, created)
	if err != nil {
		panic(err)
	}
	fmt.Println("Updated name:", updated.Name)
	fmt.Println("  New auth type:", updated.AuthenticationMethod.AuthenticationType())

	// Delete
	if err := client.VoiceOutTrunks().Delete(ctx, created.ID); err != nil {
		panic(err)
	}
	fmt.Println("\nDeleted voice out trunk")
}
