// Lists Sender ID verifications with their SMS trunks, routes an active
// verification's outbound messages through an SMS trunk that sends them, and
// restores the verification's original trunk.
//
// A verification is created and reviewed outside the API; only its name and
// its SMS trunk can be changed, and a closed one cannot be changed at all.
// Only http_out, smpp_esme and smpp_smsc trunks are accepted.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/sender_id_verifications/
package main

import (
	"context"
	"fmt"
	"strings"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Sender ID Verifications ===")
	params := didww.NewQueryParams().Include("sms_trunk")
	verifications, err := client.SenderIDVerifications().List(ctx, params)
	if err != nil {
		panic(err)
	}
	var verification *resource.SenderIDVerification
	for _, v := range verifications {
		routedThrough := "no SMS trunk"
		if v.SmsTrunk != nil {
			routedThrough = v.SmsTrunk.Name
		}
		fmt.Printf("%s %s (%s, %s): %s via %s\n",
			v.Reference, v.Name, v.SenderType, v.Status, strings.Join(v.SenderIDs, ", "), routedThrough)
		if v.IsChangesRequired() {
			comment := ""
			if v.RejectComment != nil {
				comment = *v.RejectComment
			}
			fmt.Printf("  changes required: %s %s\n", strings.Join(v.RejectReasons, "; "), comment)
		}
		if verification == nil && v.IsActive() {
			verification = v
		}
	}
	if verification == nil {
		fmt.Println("No active Sender ID verification")
		return
	}

	trunks, err := client.SmsTrunks().List(ctx, nil)
	if err != nil {
		panic(err)
	}
	var trunk *resource.SmsTrunk
	for _, t := range trunks {
		if t.AllowOut {
			trunk = t
			break
		}
	}
	if trunk == nil {
		fmt.Println("No SMS trunk that sends messages")
		return
	}

	fmt.Printf("\n=== Routing %s through %s ===\n", verification.Reference, trunk.Name)
	routed, err := client.SenderIDVerifications().Update(ctx, &resource.SenderIDVerification{
		ID:         verification.ID,
		SmsTrunkID: trunk.ID,
	})
	if err != nil {
		panic(fmt.Sprintf("Error routing the verification: %v", err))
	}
	fmt.Printf("Routed %s through %s\n", routed.Reference, trunk.ID)

	fmt.Println("\nActive verifications routed through this trunk:")
	filtered := didww.NewQueryParams().
		Filter("status", string(enums.SenderIDVerificationStatusActive)).
		Filter("sms_trunk.id", trunk.ID)
	active, err := client.SenderIDVerifications().List(ctx, filtered)
	if err != nil {
		panic(err)
	}
	for _, v := range active {
		fmt.Printf("  %s %s\n", v.Reference, v.Name)
	}

	// Restore the original trunk; NullifySmsTrunk unassigns the trunk, which
	// stops outbound messages until a trunk is assigned again.
	fmt.Printf("\n=== Restoring the original SMS trunk of %s ===\n", verification.Reference)
	restore := &resource.SenderIDVerification{ID: verification.ID, NullifySmsTrunk: true}
	if verification.SmsTrunk != nil {
		restore = &resource.SenderIDVerification{ID: verification.ID, SmsTrunkID: verification.SmsTrunk.ID}
	}
	restored, err := client.SenderIDVerifications().Update(ctx, restore)
	if err != nil {
		panic(fmt.Sprintf("Error restoring the SMS trunk: %v", err))
	}
	if verification.SmsTrunk != nil {
		fmt.Printf("%s is routed through %s again\n", restored.Reference, verification.SmsTrunk.Name)
	} else {
		fmt.Printf("%s has no SMS trunk again\n", restored.Reference)
	}
}
