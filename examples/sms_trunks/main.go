// CRUD for SMS trunks: lists trunks, creates an HTTP IN trunk, updates its
// configuration, regenerates the credentials of an HTTP OUT trunk, and deletes
// both trunks.
//
// Configuration types (package smstrunkconfiguration):
//   - sms_to_email, http_in: DIDWW delivers inbound messages by email or HTTP
//   - http_out, smpp_esme: you connect to DIDWW; DIDWW generates the
//     credentials, which are read-only
//   - smpp_smsc: DIDWW binds to your SMSC
//
// The read-only AllowIn / AllowOut tell which directions a trunk handles.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/sms_trunks/
package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
	"github.com/didww/didww-api-3-go-sdk/v3/examples"
	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/smstrunkconfiguration"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	fmt.Println("=== Existing SMS Trunks ===")
	params := didww.NewQueryParams().
		Include("sms_trunk_group").
		Page(1, 10)
	trunks, err := client.SmsTrunks().List(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d SMS trunks\n", len(trunks))
	for _, trunk := range trunks {
		configType := "none"
		if trunk.Configuration != nil {
			configType = trunk.Configuration.ConfigurationType()
		}
		fmt.Printf("%s [%s]\n", trunk.Name, configType)
		fmt.Printf("  ID: %s\n", trunk.ID)
		fmt.Printf("  Inbound: %t, outbound: %t\n", trunk.AllowIn, trunk.AllowOut)
		if trunk.SmsTrunkGroup != nil {
			fmt.Printf("  Group: %s\n", trunk.SmsTrunkGroup.Name)
		}
		if host := trunk.Meta["host"]; host != "" {
			fmt.Printf("  Submit URL: %s\n", host)
		}
	}

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)

	fmt.Println("\n=== Creating HTTP IN Trunk ===")
	created, err := client.SmsTrunks().Create(ctx, &resource.SmsTrunk{
		Name:                "Go HTTP IN " + suffix,
		ExternalReferenceID: didww.Ptr("go-sms-" + suffix),
		Configuration: &smstrunkconfiguration.HTTPIn{
			URL:        "https://example.com/sms",
			HTTPMethod: smstrunkconfiguration.HTTPMethodPost,
			BodyType:   smstrunkconfiguration.BodyTypeJSON,
			// Placeholders: {SMS_TEXT}, {SMS_TEXT_BASE64_ENCODED}, {SMS_SRC_ADDR}, {SMS_DST_ADDR}, {SMS_TIME}
			Body:                     `{"text":"{SMS_TEXT}","from":"{SMS_SRC_ADDR}","to":"{SMS_DST_ADDR}"}`,
			ValidateHTTPSCertificate: didww.Ptr(true),
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created SMS trunk: %s - %s\n", created.ID, created.Name)
	if cfg, ok := created.Configuration.(*smstrunkconfiguration.HTTPIn); ok {
		fmt.Printf("  Configuration: %s -> %s\n", cfg.ConfigurationType(), cfg.URL)
	}

	// The configuration type cannot change; attributes left out keep their values.
	fmt.Println("\n=== Updating HTTP IN Trunk ===")
	created.Name = "Go HTTP IN " + suffix + " (v2)"
	created.Configuration = &smstrunkconfiguration.HTTPIn{URL: "https://example.com/sms/v2"}
	updated, err := client.SmsTrunks().Update(ctx, created)
	if err != nil {
		fmt.Println("Error updating SMS trunk:", err)
	} else if cfg, ok := updated.Configuration.(*smstrunkconfiguration.HTTPIn); ok {
		fmt.Printf("Updated: %s -> %s\n", updated.Name, cfg.URL)
	}

	// The generated credentials of http_out and smpp_esme trunks are rotated
	// with an SmsTrunkRegenerateCredential and read back from the trunk.
	fmt.Println("\n=== Creating HTTP OUT Trunk and Regenerating Its Credentials ===")
	httpOut, err := client.SmsTrunks().Create(ctx, &resource.SmsTrunk{
		Name: "Go HTTP OUT " + suffix,
		Configuration: &smstrunkconfiguration.HTTPOut{
			CallbackURL: "https://example.com/sms/status",
			AllowedIPs:  []string{"203.0.113.10"},
		},
	})
	if err != nil {
		fmt.Println("Error creating HTTP OUT trunk:", err)
	} else {
		fmt.Printf("Created SMS trunk: %s - %s\n", httpOut.ID, httpOut.Name)
		regenerateCredentials(ctx, client, httpOut.ID)

		if err := client.SmsTrunks().Delete(ctx, httpOut.ID); err != nil {
			fmt.Println("Error deleting HTTP OUT trunk:", err)
		} else {
			fmt.Println("HTTP OUT trunk deleted")
		}
	}

	fmt.Println("\n=== Deleting HTTP IN Trunk ===")
	if err := client.SmsTrunks().Delete(ctx, created.ID); err != nil {
		panic(err)
	}
	fmt.Println("SMS trunk deleted")
}

func regenerateCredentials(ctx context.Context, client *didww.Client, trunkID string) {
	_, err := client.SmsTrunkRegenerateCredentials().Create(ctx, &resource.SmsTrunkRegenerateCredential{
		SmsTrunkID: trunkID,
	})
	if err != nil {
		fmt.Println("Error regenerating credentials:", err)
		return
	}

	trunk, err := client.SmsTrunks().Find(ctx, trunkID)
	if err != nil {
		fmt.Println("Error reloading SMS trunk:", err)
		return
	}
	if cfg, ok := trunk.Configuration.(*smstrunkconfiguration.HTTPOut); ok {
		// Use cfg.AuthName and cfg.AuthPassword to submit messages to the
		// URL in trunk.Meta["host"]. Printing cfg itself masks them.
		fmt.Printf("Credentials regenerated: %v\n", cfg)
		fmt.Printf("  Submit URL: %s\n", trunk.Meta["host"])
	}
}
