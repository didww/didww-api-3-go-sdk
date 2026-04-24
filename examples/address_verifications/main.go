// Lists Address Verifications (with 2026-04-16 reject_comment / external_reference_id).
//
// AddressVerification ties an address to one or more DIDs and a set of
// supporting documents so DIDWW compliance can approve or reject the
// declaration. 2026-04-16 adds:
//   - reject_comment:        free-form comment accompanying a rejection
//   - external_reference_id: customer-supplied reference (max 100 chars)
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/address_verifications/
package main

import (
	"context"
	"fmt"
	"strings"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Address Verifications ===")
	params := didww.NewQueryParams().Include("address,dids")
	verifications, err := client.AddressVerifications().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d address verifications\n", len(verifications))

	limit := 5
	if len(verifications) < limit {
		limit = len(verifications)
	}
	for _, av := range verifications[:limit] {
		fmt.Printf("\nVerification: %s\n", av.ID)
		fmt.Printf("  Reference: %s\n", av.Reference)
		fmt.Printf("  Status: %s\n", av.Status)
		if av.ExternalReferenceID != nil {
			fmt.Printf("  External Reference: %s\n", *av.ExternalReferenceID)
		}
		if av.ServiceDescription != nil {
			fmt.Printf("  Service description: %s\n", *av.ServiceDescription)
		}
		if len(av.RejectReasons) > 0 {
			fmt.Printf("  Reject reasons: %s\n", strings.Join(av.RejectReasons, ", "))
		}
		if av.RejectComment != "" {
			fmt.Printf("  Reject comment: %s\n", av.RejectComment)
		}
	}

	// Filter: only rejected verifications
	fmt.Println("\n=== Rejected verifications ===")
	params = didww.NewQueryParams().Filter("status", "rejected")
	rejected, err := client.AddressVerifications().List(ctx, params)
	if err != nil {
		_ = rejected // silence unused
		panic(err)
	}
	fmt.Printf("Found %d rejected verifications\n", len(rejected))
	limit = 3
	if len(rejected) < limit {
		limit = len(rejected)
	}
	for _, av := range rejected[:limit] {
		comment := av.RejectComment
		if comment == "" && len(av.RejectReasons) > 0 {
			comment = strings.Join(av.RejectReasons, ", ")
		}
		fmt.Printf("  %s: %s\n", av.Reference, comment)
	}
}
