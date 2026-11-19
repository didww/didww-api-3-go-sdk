// Runs the Go code snippets of the SDK README against the sandbox, one check
// per snippet (or per README section when its snippets belong together).
//
// Each check runs the code of its snippet, with the placeholders such as
// "uuid" replaced by real IDs, and deletes or restores what it creates or
// changes, also when it fails. A failing check is reported and the next one
// runs.
//
// Every check uses the sandbox except the production client check, which runs
// only with DIDWW_README_PRODUCTION_API_KEY and is read-only: its client asks
// for read-scoped tokens and only mints them, reads the balance and the
// countries and looks up one price. Without the variable it is skipped.
//
// The address and emergency verification checks submit real verifications
// (ordering a DID when needed), which cannot be deleted; each prints a line
// "REVIEW address_verification <id> <reference>" or
// "REVIEW emergency_verification <id> <reference>" for the review.
//
// Optional environment variables (the check is skipped when one is missing):
//   - DIDWW_README_PROXY_URL: an HTTP proxy for the custom HTTP client snippet,
//     which then sends sandbox requests through it
//   - DIDWW_README_REDIS_URL: a redis:// URL for the token cache shared by two
//     clients through Redis
//
// Prints "Results: N passed, M failed" and exits with status 1 when a check
// fails.
//
// Usage: DIDWW_API_KEY=your_sandbox_api_key go run ./examples/test_readme/
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/smstrunkconfiguration"
)

func main() {
	// The client of every check but the read-only production one
	client := examples.ClientFromEnv()
	if client.BaseURL() != string(didww.Sandbox) {
		fmt.Fprintln(os.Stderr, "test_readme runs against the sandbox only")
		os.Exit(1)
	}

	s := &suite{
		ctx:    context.Background(),
		client: client,
		apiKey: client.APIKey(),
		suffix: strconv.FormatInt(time.Now().UnixNano(), 36),
	}

	// In the order of the README
	s.usage()
	s.configuration()
	s.authentication()
	s.readOnlyResources()
	s.dids()
	s.voiceInTrunks()
	s.voiceInTrunkGroups()
	s.voiceOutTrunks()
	s.smsTrunks()
	s.smsTrunkGroups()
	s.senderIDVerifications()
	s.orders()
	s.didReservations()
	s.sharedCapacityGroups()
	s.identitiesAndAddresses()
	s.emergencyServices()
	s.didHistory()
	s.exports()
	s.priceLookups()
	s.priceFiles()
	s.filtering()
	s.dirtyPatch()
	s.errorHandling()
	s.dateFields()
	s.webhook()

	os.Exit(s.summary())
}

// suite runs the checks and counts their results.
type suite struct {
	ctx    context.Context
	client *didww.Client
	apiKey string
	suffix string

	passed, failed int
	failures       []string
	skips          []string
}

// skipped is returned by a check whose snippet cannot run in the sandbox.
type skipped struct{ reason string }

func (e skipped) Error() string { return "skipped: " + e.reason }

func skip(format string, args ...any) error {
	return skipped{reason: fmt.Sprintf(format, args...)}
}

// expect returns an error with the message when ok is false.
func expect(ok bool, format string, args ...any) error {
	if ok {
		return nil
	}
	return fmt.Errorf(format, args...)
}

func (s *suite) section(title string) {
	fmt.Printf("\n== %s ==\n", title)
}

// check runs one check, records whether it passed, failed or was skipped,
// and goes on with the next one. A panic in a check is a failure.
func (s *suite) check(name string, fn func() error) {
	err := runCheck(fn)
	var sk skipped
	switch {
	case errors.As(err, &sk):
		s.skips = append(s.skips, name+": "+sk.reason)
		fmt.Printf("  SKIP %s: %s\n", name, sk.reason)
	case err != nil:
		s.failed++
		s.failures = append(s.failures, name+": "+err.Error())
		fmt.Printf("  FAIL %s: %v\n", name, err)
	default:
		s.passed++
		fmt.Printf("  PASS %s\n", name)
	}
}

func runCheck(fn func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("panic: %v", p)
		}
	}()
	return fn()
}

// logf prints a line of check output.
func (s *suite) logf(format string, args ...any) {
	fmt.Printf("    "+format+"\n", args...)
}

// undo runs a cleanup step; its error fails the check unless the check
// already failed. err points at the named error result of the check.
func (s *suite) undo(err *error, what string, fn func() error) { //nolint:gocritic // sets the caller's named result
	if cleanupErr := fn(); cleanupErr != nil {
		s.logf("cleanup failed: %s: %v", what, cleanupErr)
		if *err == nil {
			*err = fmt.Errorf("%s: %w", what, cleanupErr)
		}
	}
}

func (s *suite) summary() int {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("Results: %d passed, %d failed\n", s.passed, s.failed)
	if len(s.skips) > 0 {
		fmt.Printf("\nSkipped (%d):\n", len(s.skips))
		for _, line := range s.skips {
			fmt.Printf("  - %s\n", line)
		}
	}
	if len(s.failures) > 0 {
		fmt.Println("\nFailures:")
		for _, line := range s.failures {
			fmt.Printf("  - %s\n", line)
		}
	}
	fmt.Println(strings.Repeat("=", 50))
	if s.failed > 0 {
		return 1
	}
	return 0
}

// --- Fixtures shared by the checks ---

// usCountryID returns the ID of the United States.
func (s *suite) usCountryID() (string, error) {
	return s.countryID("US")
}

// countryID returns the ID of the country with the ISO code.
func (s *suite) countryID(iso string) (string, error) {
	countries, err := s.client.Countries().List(s.ctx, didww.NewQueryParams().Filter("iso", iso))
	if err != nil {
		return "", err
	}
	if len(countries) == 0 {
		return "", fmt.Errorf("country %s not found", iso)
	}
	return countries[0].ID, nil
}

// firstDID returns the first active DID that matches params, or a skip error
// when there is none.
func (s *suite) firstDID(params *didww.QueryParams, match func(*resource.DID) bool, what string) (*resource.DID, error) {
	dids, err := s.client.DIDs().List(s.ctx, params.Filter("terminated", "false").Page(1, 50))
	if err != nil {
		return nil, err
	}
	for _, did := range dids {
		if match == nil || match(did) {
			return did, nil
		}
	}
	return nil, skip("the account has no %s", what)
}

// anyDID returns the first active DID.
func (s *suite) anyDID() (*resource.DID, error) {
	return s.firstDID(didww.NewQueryParams(), nil, "active DID")
}

// routedDID returns the first active DID routed to a voice IN trunk or trunk
// group, with both relationships included, so that its routing can be put back.
func (s *suite) routedDID() (*resource.DID, error) {
	return s.firstDID(
		didww.NewQueryParams().Include("voice_in_trunk,voice_in_trunk_group"),
		func(d *resource.DID) bool { return d.VoiceInTrunk != nil || d.VoiceInTrunkGroup != nil },
		"active DID routed to a voice IN trunk or trunk group (its routing is put back by assigning it again)",
	)
}

// restoreVoiceRouting assigns the original trunk or trunk group of the DID again.
func (s *suite) restoreVoiceRouting(original *resource.DID) error {
	restore := &resource.DID{ID: original.ID}
	if original.VoiceInTrunk != nil {
		restore.VoiceInTrunkID = original.VoiceInTrunk.ID
	} else {
		restore.VoiceInTrunkGroupID = original.VoiceInTrunkGroup.ID
	}
	_, err := s.client.DIDs().Update(s.ctx, restore)
	return err
}

// restoreDescription puts back the description of a DID; nil is sent as null.
func (s *suite) restoreDescription(didID string, description *string) error {
	current, err := s.client.DIDs().Find(s.ctx, didID)
	if err != nil {
		return err
	}
	current.Description = description
	_, err = s.client.DIDs().Update(s.ctx, current)
	return err
}

// skuID returns the ID of the first SKU of the first DID group that has one.
func (s *suite) skuID() (string, error) {
	params := didww.NewQueryParams().Include("stock_keeping_units").Page(1, 10)
	groups, err := s.client.DIDGroups().List(s.ctx, params)
	if err != nil {
		return "", err
	}
	for _, g := range groups {
		if len(g.StockKeepingUnits) > 0 {
			return g.StockKeepingUnits[0].ID, nil
		}
	}
	return "", errors.New("no DID group with stock_keeping_units found")
}

// createInboundSmsTrunk creates an http_in SMS trunk for a check.
func (s *suite) createInboundSmsTrunk(name string) (*resource.SmsTrunk, error) {
	return s.client.SmsTrunks().Create(s.ctx, &resource.SmsTrunk{
		Name: name + " " + s.suffix,
		Configuration: &smstrunkconfiguration.HTTPIn{
			URL:        "https://example.com/sms",
			HTTPMethod: smstrunkconfiguration.HTTPMethodPost,
			BodyType:   smstrunkconfiguration.BodyTypeJSON,
			Body:       `{"text":"{SMS_TEXT}"}`,
		},
	})
}

// createOutboundSmsTrunk creates an http_out SMS trunk for a check.
func (s *suite) createOutboundSmsTrunk(name string) (*resource.SmsTrunk, error) {
	return s.client.SmsTrunks().Create(s.ctx, &resource.SmsTrunk{
		Name: name + " " + s.suffix,
		Configuration: &smstrunkconfiguration.HTTPOut{
			CallbackURL: "https://example.com/sms/status",
			AllowedIPs:  []string{"203.0.113.10"},
		},
	})
}

func (s *suite) deleteSmsTrunk(id string) func() error {
	return func() error { return s.client.SmsTrunks().Delete(s.ctx, id) }
}

func (s *suite) deleteVoiceInTrunk(id string) func() error {
	return func() error {
		err := s.client.VoiceInTrunks().Delete(s.ctx, id)
		var apiErr *didww.APIError
		if errors.As(err, &apiErr) && apiErr.HTTPStatus == http.StatusNotFound {
			return nil // deleting a trunk group deletes its trunks
		}
		return err
	}
}
