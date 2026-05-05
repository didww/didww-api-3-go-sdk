// End-to-end SIP registration flow on /voice_in_trunks (API 2026-04-16):
// create with sip_registration enabled → rename → disable by setting Host
// → re-enable by toggling the flag. Demonstrates how the SDK keeps the
// dependent fields (Host, Port, UseDIDInRuri) aligned with the server's
// validation rules. The sandbox trunk is left in place after the script
// completes.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/voice_in_trunk_sip_registration/
package main

import (
	"context"
	"fmt"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/trunkconfiguration"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Go SDK — SIP registration flow ===")

	// 1) Create with sip_registration enabled.
	fmt.Println("\n[1/4] Create with sip_registration enabled...")
	ringingTimeout := 30
	created, err := client.VoiceInTrunks().Create(ctx, &resource.VoiceInTrunk{
		Name:           fmt.Sprintf("go-sip-registration-%d", time.Now().Unix()),
		Priority:       1,
		Weight:         100,
		CliFormat:      enums.CliFormatE164,
		RingingTimeout: &ringingTimeout,
		Configuration: &trunkconfiguration.SIPConfiguration{
			EnabledSipRegistration: didww.Ptr(true),
			UseDIDInRuri:           didww.Ptr(true),
			CnamLookup:             didww.Ptr(false),
			CodecIDs:               []enums.Codec{enums.CodecPCMU, enums.CodecPCMA},
			TransportProtocolID:    enums.TransportProtocolUDP,
		},
	})
	if err != nil {
		fmt.Printf("  ✗ create failed: %v\n", err)
		return
	}
	trunkID := created.ID
	cfg1 := created.Configuration.(*trunkconfiguration.SIPConfiguration)
	fmt.Printf("  id=%s\n", trunkID)
	fmt.Printf("  IncomingAuthUsername=%q\n", cfg1.IncomingAuthUsername)
	fmt.Printf("  IncomingAuthPassword=%q\n", cfg1.IncomingAuthPassword)

	// 2) Rename — single-field PATCH.
	fmt.Println("\n[2/4] Rename trunk...")
	newName := fmt.Sprintf("go-renamed-%d", time.Now().Unix())
	if _, err := client.VoiceInTrunks().Update(ctx, &resource.VoiceInTrunk{
		ID:   trunkID,
		Name: newName,
	}); err != nil {
		fmt.Printf("  ✗ rename failed: %v\n", err)
		return
	}
	fmt.Printf("  name=%s\n", newName)

	// 3) Disable sip_registration by setting Host.
	fmt.Println("\n[3/4] Disable by setting Host...")
	if _, err := client.VoiceInTrunks().Update(ctx, &resource.VoiceInTrunk{
		ID: trunkID,
		Configuration: &trunkconfiguration.SIPConfiguration{
			Host: "203.0.113.10",
		},
	}); err != nil {
		fmt.Printf("  ✗ disable failed: %v\n", err)
		return
	}
	fresh3, _ := client.VoiceInTrunks().Find(ctx, trunkID)
	cfg3 := fresh3.Configuration.(*trunkconfiguration.SIPConfiguration)
	enabled3 := "<nil>"
	if cfg3.EnabledSipRegistration != nil {
		enabled3 = fmt.Sprintf("%v", *cfg3.EnabledSipRegistration)
	}
	useDid3 := "<nil>"
	if cfg3.UseDIDInRuri != nil {
		useDid3 = fmt.Sprintf("%v", *cfg3.UseDIDInRuri)
	}
	fmt.Printf("  EnabledSipRegistration=%s\n", enabled3)
	fmt.Printf("  UseDIDInRuri=%s\n", useDid3)
	fmt.Printf("  Host=%q\n", cfg3.Host)
	fmt.Printf("  IncomingAuthUsername=%q\n", cfg3.IncomingAuthUsername)

	// 4) Re-enable sip_registration. The SDK should send host=null / port=null
	//    on the wire so the server clears the values it had persisted.
	fmt.Println("\n[4/4] Re-enable by toggling EnabledSipRegistration...")
	if _, err := client.VoiceInTrunks().Update(ctx, &resource.VoiceInTrunk{
		ID: trunkID,
		Configuration: &trunkconfiguration.SIPConfiguration{
			EnabledSipRegistration: didww.Ptr(true),
			UseDIDInRuri:           didww.Ptr(true),
		},
	}); err != nil {
		fmt.Printf("  ✗ FAIL: %v\n", err)
		fmt.Printf("\n=== FAIL at re-enable — trunk %s left in sandbox ===\n", trunkID)
		return
	}
	fresh4, _ := client.VoiceInTrunks().Find(ctx, trunkID)
	cfg4 := fresh4.Configuration.(*trunkconfiguration.SIPConfiguration)
	enabled4 := "<nil>"
	if cfg4.EnabledSipRegistration != nil {
		enabled4 = fmt.Sprintf("%v", *cfg4.EnabledSipRegistration)
	}
	fmt.Printf("  EnabledSipRegistration=%s\n", enabled4)
	fmt.Printf("  Host=%q\n", cfg4.Host)
	fmt.Printf("  IncomingAuthUsername=%q\n", cfg4.IncomingAuthUsername)
	fmt.Printf("\n=== PASS — trunk %s left in sandbox ===\n", trunkID)
}
