// CRUD for SMS trunk groups: creates two inbound SMS trunks, puts them in a
// group, lists groups, and deletes the group.
//
// A group spreads a DID's inbound messages over its trunks by their priority
// (higher first). http_out trunks cannot join a group. Deleting a group also
// deletes the trunks in it.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/sms_trunk_groups/
package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/smstrunkconfiguration"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()
	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)

	fmt.Println("=== Creating SMS Trunks ===")
	primary, err := client.SmsTrunks().Create(ctx, &resource.SmsTrunk{
		Name:     "Go primary " + suffix,
		Priority: didww.Ptr(200),
		Configuration: &smstrunkconfiguration.HTTPIn{
			URL:        "https://example.com/sms",
			HTTPMethod: smstrunkconfiguration.HTTPMethodGet,
			QueryParameters: map[string]any{
				"text": "{SMS_TEXT}",
				"from": "{SMS_SRC_ADDR}",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created SMS trunk: %s - %s\n", primary.ID, primary.Name)

	fallback, err := client.SmsTrunks().Create(ctx, &resource.SmsTrunk{
		Name:     "Go fallback " + suffix,
		Priority: didww.Ptr(100),
		Configuration: &smstrunkconfiguration.SmsToEmail{
			SendTo:  "sms@example.com",
			Subject: "Inbound SMS",
			Message: "You have a new message.",
		},
	})
	if err != nil {
		_ = client.SmsTrunks().Delete(ctx, primary.ID)
		panic(err)
	}
	fmt.Printf("Created SMS trunk: %s - %s\n", fallback.ID, fallback.Name)

	fmt.Println("\n=== Creating SMS Trunk Group ===")
	group, err := client.SmsTrunkGroups().Create(ctx, &resource.SmsTrunkGroup{
		Name:                "Go SMS Group " + suffix,
		ExternalReferenceID: didww.Ptr("go-sms-group-" + suffix),
		SmsTrunkIDs:         []string{primary.ID, fallback.ID},
	})
	if err != nil {
		_ = client.SmsTrunks().Delete(ctx, primary.ID)
		_ = client.SmsTrunks().Delete(ctx, fallback.ID)
		panic(err)
	}
	fmt.Printf("Created SMS trunk group: %s - %s (%s trunks)\n", group.ID, group.Name, group.Meta["sms_trunks_count"])

	fmt.Println("\n=== Existing SMS Trunk Groups ===")
	params := didww.NewQueryParams().
		Include("sms_trunks").
		Page(1, 5)
	groups, err := client.SmsTrunkGroups().List(ctx, params)
	if err != nil {
		panic(err)
	}
	for _, g := range groups {
		names := make([]string, len(g.SmsTrunks))
		for i, t := range g.SmsTrunks {
			names[i] = t.Name
		}
		fmt.Printf("%s (%s trunks) %s\n", g.Name, g.Meta["sms_trunks_count"], strings.Join(names, ", "))
	}

	fmt.Println("\n=== Deleting SMS Trunk Group (and its trunks) ===")
	if err := client.SmsTrunkGroups().Delete(ctx, group.ID); err != nil {
		panic(err)
	}
	fmt.Println("SMS trunk group deleted")
}
