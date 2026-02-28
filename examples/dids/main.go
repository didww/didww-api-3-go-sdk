// Updates DID routing/capacity by assigning trunk and capacity pool.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/dids/
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

	// Get last ordered DID
	didParams := didww.NewQueryParams().
		Sort("-created_at").
		Page(1, 1)
	dids, err := client.DIDs().List(ctx, didParams)
	if err != nil {
		panic(err)
	}
	if len(dids) == 0 {
		panic("No DIDs found. Order a DID first.")
	}
	did := dids[0]

	// Get last SIP trunk
	trunkParams := didww.NewQueryParams().
		Sort("-created_at").
		Filter("configuration.type", "sip_configurations")
	trunks, err := client.VoiceInTrunks().List(ctx, trunkParams)
	if err != nil {
		panic(err)
	}
	if len(trunks) == 0 {
		panic("No SIP trunks found. Create a trunk first.")
	}

	// Update DID with capacity settings
	did.Description = examples.Ptr("Hi")
	did.CapacityLimit = examples.Ptr(5)
	did.DedicatedChannelsCount = 1

	savedDid, err := client.DIDs().Update(ctx, did)
	if err != nil {
		panic(err)
	}
	fmt.Printf("DID %s description=%s capacityLimit=%d dedicatedChannels=%d\n",
		savedDid.ID, *savedDid.Description, *savedDid.CapacityLimit, savedDid.DedicatedChannelsCount)
}
