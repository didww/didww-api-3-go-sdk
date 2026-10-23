// Creates SMS exports (sms_in / sms_out) for yesterday and reads the columns
// of a completed SMS-Out export.
//
// Filters on SMS exports (Export.Filters):
//   - "from", "to" (required): "YYYY-MM-DD HH:MM:SS" UTC; from inclusive, to exclusive
//   - "source_address_eq" / "source_address_cont", "destination_address_eq" /
//     "destination_address_cont", "status", "sms_id"
//   - "sms_trunk.id": one id or a slice of ids
//   - sms_in only: "trunk_type", one or more of the enums.ExportTrunkType* values
//   - sms_out only: "service_type" ("a2p" or "p2p"), "sender_id_verification.id"
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/exports_sms/
package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	now := time.Now().UTC()
	from := now.AddDate(0, 0, -1).Format("2006-01-02") + " 00:00:00" // inclusive
	to := now.Format("2006-01-02") + " 00:00:00"                     // exclusive
	suffix := strconv.FormatInt(now.UnixNano(), 36)

	fmt.Println("=== Creating SMS-Out Export (A2P, yesterday) ===")
	smsOut, err := client.Exports().Create(ctx, &resource.Export{
		ExportType: enums.ExportTypeSmsOut,
		Filters: map[string]interface{}{
			"from":         from,
			"to":           to,
			"service_type": "a2p",
		},
		ExternalReferenceID: didww.Ptr("go-sms-out-" + suffix),
	})
	if err != nil {
		fmt.Println("Error creating SMS-Out export:", err)
	} else {
		fmt.Printf("Created SMS-Out export: %s (status: %s)\n", smsOut.ID, smsOut.Status)
	}

	fmt.Println("\n=== Creating SMS-In Export (HTTP IN and email trunks, yesterday) ===")
	smsIn, err := client.Exports().Create(ctx, &resource.Export{
		ExportType: enums.ExportTypeSmsIn,
		Filters: map[string]interface{}{
			"from": from,
			"to":   to,
			"trunk_type": []string{
				enums.ExportTrunkTypeHTTPIn,
				enums.ExportTrunkTypeEmail,
			},
		},
		ExternalReferenceID: didww.Ptr("go-sms-in-" + suffix),
	})
	if err != nil {
		fmt.Println("Error creating SMS-In export:", err)
	} else {
		fmt.Printf("Created SMS-In export: %s (status: %s)\n", smsIn.ID, smsIn.Status)
	}

	fmt.Println("\n=== Completed SMS-Out exports ===")
	params := didww.NewQueryParams().
		Filter("export_type", string(enums.ExportTypeSmsOut)).
		Filter("status", string(enums.ExportStatusCompleted))
	completed, err := client.Exports().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d completed SMS-Out exports\n", len(completed))
	if len(completed) == 0 || completed[0].URL == nil {
		return
	}

	export := completed[0]
	var csv bytes.Buffer
	if err := client.DownloadAndDecompressExport(ctx, *export.URL, &csv); err != nil {
		panic(err)
	}
	header, _, _ := strings.Cut(csv.String(), "\n")
	fmt.Printf("Export %s (%s) columns: %s\n", export.ID, export.ExportType, header)
}
