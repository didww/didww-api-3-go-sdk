// Looks up Voice IN, Voice OUT, SMS OUT and SMS IN prices for one number.
//
// Prices are served by separate hosts (api-vin, api-vout and api-sms); the SDK
// mints a token for each of them on first use. A read-scoped client
// (didww.WithScope(didww.ScopeRead)) is enough for prices. The account needs a
// price plan of the service, otherwise the call fails with
// didww.ErrPricePlanNotAssigned.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/prices_lookup/ [number]
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	number := "18005550123"
	if len(os.Args) > 1 {
		number = os.Args[1]
	}

	show(ctx, "Voice IN", client.VoiceInPrices(), map[string]string{"number": number})
	show(ctx, "Voice IN, calls from the US", client.VoiceInPrices(), map[string]string{"number": number, "src_country_iso": "US"})
	show(ctx, "Voice OUT", client.VoiceOutPrices(), map[string]string{"number": number})
	show(ctx, "SMS OUT A2P", client.SMSOutA2PPrices(), map[string]string{"dst_number": number})
	show(ctx, "SMS OUT P2P, whole country", client.SMSOutP2PPrices(), map[string]string{"dst_country_iso": "US"})
	show(ctx, "SMS OUT OTP", client.SMSOutOTPPrices(), map[string]string{"dst_number": number, "dst_country_iso": "US"})
	show(ctx, "SMS IN", client.SMSInPrices(), map[string]string{"dst_number": number})

	// Filters are checked before any request is sent: Voice OUT takes number only.
	_, err := client.VoiceOutPrices().Lookup(ctx, map[string]string{"number": number, "src_country_iso": "US"})
	fmt.Printf("\nRejected locally: %v\n", err)
}

func show[T any](ctx context.Context, title string, prices *didww.PricesService[T], filters map[string]string) {
	fmt.Printf("\n=== %s ===\n", title)
	result, err := prices.Lookup(ctx, filters)
	switch {
	case errors.Is(err, didww.ErrPricePlanNotAssigned), errors.Is(err, didww.ErrPriceNotFound):
		fmt.Printf("  %v\n", err)
		return
	case err != nil:
		panic(err)
	}

	fmt.Printf("%d matching price(s)\n", result.TotalRecords)
	for _, row := range result.Rows {
		// Prices and fees are decimal strings in USD; nil fields are printed as null.
		attrs, err := json.Marshal(row)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  %s\n", attrs)
	}
}
