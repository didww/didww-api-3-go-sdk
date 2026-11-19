package didww

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
)

// pricesTestServer serves POST /token like newAuthTestServer and answers the
// n-th (1-based) other request with respond, recording those requests.
type pricesTestServer struct {
	*authTestServer
	mu       sync.Mutex
	requests []*http.Request
}

// newPricesTestServer starts a pricesTestServer and returns a client whose
// token service and price hosts are all served by it.
func newPricesTestServer(t *testing.T, respond func(n int, w http.ResponseWriter, r *http.Request)) (*pricesTestServer, *Client) {
	t.Helper()
	s := &pricesTestServer{}
	s.authTestServer = newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.requests = append(s.requests, r.Clone(context.Background()))
		n := len(s.requests)
		s.mu.Unlock()
		respond(n, w, r)
	})
	client := newTestClient(t, s.URL, WithVoiceInPricesURL(s.URL), WithVoiceOutPricesURL(s.URL), WithSMSPricesURL(s.URL))
	return s, client
}

// requestCount returns the number of requests other than token requests.
func (s *pricesTestServer) requestCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.requests)
}

// request returns the i-th (0-based) request other than a token request.
func (s *pricesTestServer) request(i int) *http.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.requests[i]
}

// mintAudiences returns the audiences of every token request, in order.
func (s *pricesTestServer) mintAudiences(t *testing.T) []string {
	t.Helper()
	audiences := []string{}
	for i := 0; i < s.mints(); i++ {
		var body struct {
			Audience []string `json:"aud"`
		}
		require.NoError(t, json.Unmarshal(s.mintBody(i), &body))
		audiences = append(audiences, body.Audience...)
	}
	return audiences
}

// writePricesJSON answers with a JSON:API body.
func writePricesJSON(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", jsonapiMediaType)
	w.WriteHeader(status)
	w.Write(body)
}

// closedServerURL returns the URL of a server that no longer accepts connections.
func closedServerURL() string {
	s := httptest.NewServer(http.NotFoundHandler())
	u := s.URL
	s.Close()
	return u
}

type lookupFunc func(c *Client, filters map[string]string) error

var (
	lookupVoiceIn = func(c *Client, f map[string]string) error {
		_, err := c.VoiceInPrices().Lookup(context.Background(), f)
		return err
	}
	lookupVoiceOut = func(c *Client, f map[string]string) error {
		_, err := c.VoiceOutPrices().Lookup(context.Background(), f)
		return err
	}
	lookupSMSOutA2P = func(c *Client, f map[string]string) error {
		_, err := c.SMSOutA2PPrices().Lookup(context.Background(), f)
		return err
	}
	lookupSMSOutP2P = func(c *Client, f map[string]string) error {
		_, err := c.SMSOutP2PPrices().Lookup(context.Background(), f)
		return err
	}
	lookupSMSOutOTP = func(c *Client, f map[string]string) error {
		_, err := c.SMSOutOTPPrices().Lookup(context.Background(), f)
		return err
	}
	lookupSMSIn = func(c *Client, f map[string]string) error {
		_, err := c.SMSInPrices().Lookup(context.Background(), f)
		return err
	}
)

func TestVoiceInPricesLookup(t *testing.T) {
	fixture := loadFixture(t, "prices/voice_in/lookup/200.json")
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusOK, fixture)
	})

	result, err := client.VoiceInPrices().Lookup(context.Background(), map[string]string{"number": "18005550123"})
	require.NoError(t, err)

	require.Equal(t, 1, s.requestCount())
	req := s.request(0)
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "/v3/lookup/prices", req.URL.Path)
	assert.Equal(t, url.Values{"filter[number]": {"18005550123"}}, req.URL.Query())
	assert.Equal(t, "application/vnd.api+json", req.Header.Get("Accept"))
	assert.Equal(t, "Bearer token-1", req.Header.Get("Authorization"))
	assert.Equal(t, userAgent, req.Header.Get("User-Agent"))
	assert.Empty(t, req.Header.Get("Api-Key"))
	assert.Equal(t, []string{"api-vin.didww.com"}, s.mintAudiences(t))

	assert.Equal(t, 2, result.TotalRecords)
	require.Len(t, result.Rows, 2)
	priced, rejected := result.Rows[0], result.Rows[1]
	assert.Equal(t, "df73d9e4-46ce-4af0-817e-f090dc8144f1", priced.ID)
	assert.Equal(t, "1800", priced.Destination)
	assert.Equal(t, "United States", *priced.Country)
	assert.Nil(t, priced.SrcCountryName)
	assert.Equal(t, "cellular", priced.SourceType)
	assert.Equal(t, "0.025", *priced.Rate)
	assert.Equal(t, "0.0", *priced.ConnectionFee)
	assert.Equal(t, 60, priced.InitialInterval)
	assert.Equal(t, 60, priced.NextInterval)
	assert.False(t, priced.RejectCalls)
	assert.Equal(t, "cb1335e1-6cfd-4921-8724-153d9ed67ac3", rejected.ID)
	assert.Equal(t, "United Kingdom", *rejected.SrcCountryName)
	assert.Nil(t, rejected.Rate)
	assert.Nil(t, rejected.ConnectionFee)
	assert.True(t, rejected.RejectCalls)
}

func TestVoiceInPricesLookupWithSourceCountry(t *testing.T) {
	fixture := loadFixture(t, "prices/voice_in/lookup/200.json")
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusOK, fixture)
	})

	_, err := client.VoiceInPrices().Lookup(context.Background(), map[string]string{"number": "448001234567", "src_country_iso": "GB"})
	require.NoError(t, err)

	assert.Equal(t, url.Values{"filter[number]": {"448001234567"}, "filter[src_country_iso]": {"GB"}}, s.request(0).URL.Query())
}

func TestVoiceOutPricesLookup(t *testing.T) {
	fixture := loadFixture(t, "prices/voice_out/lookup/200.json")
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusOK, fixture)
	})

	result, err := client.VoiceOutPrices().Lookup(context.Background(), map[string]string{"number": "493012345678"})
	require.NoError(t, err)

	req := s.request(0)
	assert.Equal(t, "/v3/lookup/prices", req.URL.Path)
	assert.Equal(t, url.Values{"filter[number]": {"493012345678"}}, req.URL.Query())
	assert.Equal(t, []string{"api-vout.didww.com"}, s.mintAudiences(t))

	assert.Equal(t, 2, result.TotalRecords)
	require.Len(t, result.Rows, 2)
	scheduled, plain := result.Rows[0], result.Rows[1]
	assert.Equal(t, "7b82cac2-c42a-4fa9-8da4-ccba2843068f", scheduled.ID)
	assert.Equal(t, enums.VoiceOutPriceCategoryInternational, scheduled.Category)
	assert.Equal(t, "4930", scheduled.Prefix)
	assert.Equal(t, "0.0045", *scheduled.Rate)
	assert.Equal(t, "0.005", *scheduled.NextRate)
	assert.Equal(t, 30, *scheduled.NextInitialIncrement)
	assert.Equal(t, 6, *scheduled.NextRecurringIncrement)
	require.NotNil(t, scheduled.EffectiveAt)
	assert.True(t, scheduled.EffectiveAt.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, "Increase and Billing increment change", *scheduled.Comment)
	assert.Equal(t, enums.VoiceOutPriceCategoryOriginBased, plain.Category)
	assert.Nil(t, plain.NextRate)
	assert.Nil(t, plain.NextInitialIncrement)
	assert.Nil(t, plain.EffectiveAt)
	assert.Nil(t, plain.Comment)
}

func TestSMSPricesLookup(t *testing.T) {
	fixture := loadFixture(t, "prices/sms_out/lookup/200.json")
	tests := []struct {
		name   string
		lookup lookupFunc
		path   string
	}{
		{"SMSOutA2P", lookupSMSOutA2P, "/v3/sms_out/a2p/lookup/prices"},
		{"SMSOutP2P", lookupSMSOutP2P, "/v3/sms_out/p2p/lookup/prices"},
		{"SMSOutOTP", lookupSMSOutOTP, "/v3/sms_out/otp/lookup/prices"},
		{"SMSIn", lookupSMSIn, "/v3/sms_in/lookup/prices"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
				writePricesJSON(w, http.StatusOK, fixture)
			})

			require.NoError(t, tt.lookup(client, map[string]string{"dst_number": "4915123456789"}))
			require.NoError(t, tt.lookup(client, map[string]string{"dst_country_iso": "FR"}))
			require.NoError(t, tt.lookup(client, map[string]string{"dst_number": "4915123456789", "dst_country_iso": "DE"}))

			require.Equal(t, 3, s.requestCount())
			for i := 0; i < 3; i++ {
				assert.Equal(t, tt.path, s.request(i).URL.Path)
				assert.Equal(t, "Bearer token-1", s.request(i).Header.Get("Authorization"))
			}
			assert.Equal(t, url.Values{"filter[dst_number]": {"4915123456789"}}, s.request(0).URL.Query())
			assert.Equal(t, url.Values{"filter[dst_country_iso]": {"FR"}}, s.request(1).URL.Query())
			assert.Equal(t, url.Values{"filter[dst_number]": {"4915123456789"}, "filter[dst_country_iso]": {"DE"}}, s.request(2).URL.Query())
			assert.Equal(t, []string{"api-sms.didww.com"}, s.mintAudiences(t))
		})
	}
}

func TestSMSOutPricesLookupRows(t *testing.T) {
	fixture := loadFixture(t, "prices/sms_out/lookup/200.json")
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusOK, fixture)
	})

	result, err := client.SMSOutA2PPrices().Lookup(context.Background(), map[string]string{"dst_number": "4915123456789"})
	require.NoError(t, err)

	assert.Equal(t, 2, result.TotalRecords)
	require.Len(t, result.Rows, 2)
	priced, rejected := result.Rows[0], result.Rows[1]
	assert.Equal(t, "d8e4d372-42a6-4bb6-80ec-6c29c49aa9fd", priced.ID)
	assert.Nil(t, priced.SourcePrefix)
	assert.Equal(t, "49151", *priced.DestinationPrefix)
	assert.Equal(t, 262, *priced.MCC)
	assert.Equal(t, 1, *priced.MNC)
	assert.Equal(t, "Telekom Deutschland", *priced.MNCName)
	assert.False(t, priced.RejectSMS)
	assert.Equal(t, "0.075000", *priced.Price)
	assert.True(t, rejected.RejectSMS)
	assert.Nil(t, rejected.Price)
}

func TestSMSInPricesLookupRows(t *testing.T) {
	fixture := loadFixture(t, "prices/sms_in/lookup/200.json")
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusOK, fixture)
	})

	result, err := client.SMSInPrices().Lookup(context.Background(), map[string]string{"dst_number": "12025550123"})
	require.NoError(t, err)

	assert.Equal(t, 1, result.TotalRecords)
	require.Len(t, result.Rows, 1)
	price := result.Rows[0]
	assert.Equal(t, "e93f925e-25ee-4060-bd50-a790b4945d3f", price.ID)
	assert.Nil(t, price.SourcePrefix)
	assert.Equal(t, 1, price.SourceNumberLengthMin)
	assert.Equal(t, 15, price.SourceNumberLengthMax)
	assert.Equal(t, "1", *price.DestinationPrefix)
	assert.Equal(t, 11, price.DestinationNumberLengthMin)
	assert.Equal(t, 11, price.DestinationNumberLengthMax)
	assert.Equal(t, "United States", *price.Country)
	assert.Equal(t, "0.005000", price.Price)
}

func TestPricesLookupFilterValidation(t *testing.T) {
	tests := []struct {
		name    string
		lookup  lookupFunc
		filters map[string]string
		message string
	}{
		{"VoiceIn without number", lookupVoiceIn, map[string]string{"src_country_iso": "GB"},
			"filter number is required on lookup/prices"},
		{"VoiceIn with a blank number", lookupVoiceIn, map[string]string{"number": ""},
			"filter number is required on lookup/prices"},
		{"VoiceIn with an unknown filter", lookupVoiceIn, map[string]string{"number": "18005550123", "dst_country_iso": "US"},
			"filter dst_country_iso is not allowed on lookup/prices (allowed: number, src_country_iso)"},
		{"VoiceOut with src_country_iso", lookupVoiceOut, map[string]string{"number": "493012345678", "src_country_iso": "DE"},
			"filter src_country_iso is not allowed on lookup/prices (allowed: number)"},
		{"VoiceOut with unknown filters", lookupVoiceOut, map[string]string{"number": "493012345678", "b": "1", "a": "1"},
			"filter a, b is not allowed on lookup/prices (allowed: number)"},
		{"SMSOutA2P without filters", lookupSMSOutA2P, map[string]string{},
			"one of filters dst_number, dst_country_iso is required on sms_out/a2p/lookup/prices"},
		{"SMSOutP2P with nil filters", lookupSMSOutP2P, nil,
			"one of filters dst_number, dst_country_iso is required on sms_out/p2p/lookup/prices"},
		{"SMSOutOTP with a blank dst_number", lookupSMSOutOTP, map[string]string{"dst_number": " "},
			"one of filters dst_number, dst_country_iso is required on sms_out/otp/lookup/prices"},
		{"SMSIn with number", lookupSMSIn, map[string]string{"number": "12025550123"},
			"filter number is not allowed on sms_in/lookup/prices (allowed: dst_number, dst_country_iso)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			err := tt.lookup(client, tt.filters)

			require.EqualError(t, err, tt.message)
			assert.Equal(t, 0, s.requestCount())
			assert.Equal(t, 0, s.mints())
		})
	}
}

func TestPricesLookupErrors(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		fixture    string
		retryAfter string
		sentinel   error
		code       string
		wait       time.Duration
		message    string
	}{
		{name: "400", status: http.StatusBadRequest, fixture: "prices/errors/filter_not_allowed.json",
			sentinel: ErrPricesBadRequest, code: "102", message: "didww: prices service: src_country_iso is not allowed."},
		{name: "403", status: http.StatusForbidden, body: `{"errors":[{"title":"Insufficient scope","code":"403","status":"403"}]}`,
			sentinel: ErrPricesAccessDenied, code: "403", message: "didww: prices service: Insufficient scope"},
		{name: "404 price_not_found", status: http.StatusNotFound, fixture: "prices/errors/price_not_found.json",
			sentinel: ErrPriceNotFound, code: "price_not_found", message: "didww: prices service: No effective price covers this number in the assigned plan."},
		{name: "404 price_plan_not_assigned", status: http.StatusNotFound, fixture: "prices/errors/price_plan_not_assigned.json",
			sentinel: ErrPricePlanNotAssigned, code: "price_plan_not_assigned", message: "didww: prices service: No price plan is assigned to this account."},
		{name: "404 other code", status: http.StatusNotFound, body: `{"errors":[{"title":"Not found","code":"not_found"}]}`,
			sentinel: ErrPricePlanNotAssigned, code: "not_found", message: "didww: prices service: Not found"},
		{name: "404 without body", status: http.StatusNotFound,
			sentinel: ErrPricePlanNotAssigned, code: "price_plan_not_assigned", message: "didww: prices service responded with status 404"},
		{name: "406", status: http.StatusNotAcceptable, body: `{"errors":[{"title":"Not acceptable","code":"406","status":"406"}]}`,
			sentinel: ErrPricesNotAcceptable, code: "406", message: "didww: prices service: Not acceptable"},
		{name: "429", status: http.StatusTooManyRequests, retryAfter: "7",
			body:     `{"errors":[{"title":"Too many requests","code":"too_many_requests","status":"429"}]}`,
			sentinel: ErrPricesTooManyRequests, code: "too_many_requests", wait: 7 * time.Second, message: "didww: prices service: Too many requests"},
		{name: "503 not retried on lookups", status: http.StatusServiceUnavailable, retryAfter: "1", fixture: "prices/errors/serving_failure.json",
			sentinel: ErrPricesServiceUnavailable, code: "serving_failure", wait: time.Second, message: "didww: prices service: The price file is not available yet."},
		{name: "500", status: http.StatusInternalServerError,
			sentinel: ErrPricesUnexpectedStatus, message: "didww: prices service responded with status 500"},
		{name: "500 with a non-JSON body", status: http.StatusInternalServerError, body: "<html>oops</html>",
			sentinel: ErrPricesUnexpectedStatus, message: "didww: prices service responded with status 500"},
		{name: "304", status: http.StatusNotModified,
			sentinel: ErrPricesUnexpectedStatus, message: "didww: prices service responded with status 304"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := []byte(tt.body)
			if tt.fixture != "" {
				body = loadFixture(t, tt.fixture)
			}
			s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				writePricesJSON(w, tt.status, body)
			})

			_, err := client.VoiceInPrices().Lookup(context.Background(), map[string]string{"number": "18005550123"})

			require.ErrorIs(t, err, tt.sentinel)
			var pricesErr *PricesError
			require.ErrorAs(t, err, &pricesErr)
			assert.Equal(t, tt.status, pricesErr.StatusCode)
			assert.Equal(t, tt.code, pricesErr.Code)
			assert.Equal(t, tt.wait, pricesErr.RetryAfter)
			assert.EqualError(t, err, tt.message)
			assert.Equal(t, 1, s.requestCount())
			assert.Equal(t, 1, s.mints())
		})
	}
}

func TestPricesLookupDoesNotFollowRedirects(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/v3/elsewhere")
		w.WriteHeader(http.StatusFound)
	})

	_, err := client.VoiceInPrices().Lookup(context.Background(), map[string]string{"number": "18005550123"})

	require.ErrorIs(t, err, ErrPricesUnexpectedStatus)
	var pricesErr *PricesError
	require.ErrorAs(t, err, &pricesErr)
	assert.Equal(t, http.StatusFound, pricesErr.StatusCode)
	assert.Equal(t, 1, s.requestCount())
}

func TestPricesLookupRetriesRefusedTokenOnce(t *testing.T) {
	fixture := loadFixture(t, "prices/voice_in/lookup/200.json")
	s, client := newPricesTestServer(t, func(n int, w http.ResponseWriter, _ *http.Request) {
		if n == 1 {
			writePricesJSON(w, http.StatusUnauthorized, []byte(`{"errors":[{"title":"Unauthorized","code":"401","status":"401"}]}`))
			return
		}
		writePricesJSON(w, http.StatusOK, fixture)
	})

	result, err := client.VoiceInPrices().Lookup(context.Background(), map[string]string{"number": "18005550123"})
	require.NoError(t, err)

	assert.Equal(t, 2, result.TotalRecords)
	require.Equal(t, 2, s.requestCount())
	assert.Equal(t, "Bearer token-1", s.request(0).Header.Get("Authorization"))
	assert.Equal(t, "Bearer token-2", s.request(1).Header.Get("Authorization"))
	assert.Equal(t, []string{"api-vin.didww.com", "api-vin.didww.com"}, s.mintAudiences(t))
}

func TestPricesLookupNotAuthorizedAfterRetry(t *testing.T) {
	s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writePricesJSON(w, http.StatusUnauthorized, []byte(`{"errors":[{"title":"Unauthorized","code":"401","status":"401"}]}`))
	})

	_, err := client.VoiceInPrices().Lookup(context.Background(), map[string]string{"number": "18005550123"})

	require.ErrorIs(t, err, ErrPricesNotAuthorized)
	var pricesErr *PricesError
	require.ErrorAs(t, err, &pricesErr)
	assert.Equal(t, http.StatusUnauthorized, pricesErr.StatusCode)
	assert.Equal(t, "401", pricesErr.Code)
	assert.Equal(t, 2, s.requestCount())
	assert.Equal(t, 2, s.mints())
}

func TestPricesLookupConnectionError(t *testing.T) {
	auth := newAuthTestServer(t, nil, nil)
	client := newTestClient(t, auth.URL, WithVoiceInPricesURL(closedServerURL()))

	_, err := client.VoiceInPrices().Lookup(context.Background(), map[string]string{"number": "18005550123"})

	require.ErrorIs(t, err, ErrPricesConnection)
	var pricesErr *PricesError
	require.ErrorAs(t, err, &pricesErr)
	assert.Equal(t, 0, pricesErr.StatusCode)
	assert.Error(t, pricesErr.Err)
}

func TestPricesLookupReturnsTokenErrorAsIs(t *testing.T) {
	var apiCalls int
	var mu sync.Mutex
	auth := newAuthTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
		writeTokenError(t, w, http.StatusUnauthorized, "token/invalid_client.json")
	}, func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		apiCalls++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	client := newTestClient(t, auth.URL, WithVoiceInPricesURL(auth.URL))

	_, err := client.VoiceInPrices().Lookup(context.Background(), map[string]string{"number": "18005550123"})

	require.ErrorIs(t, err, ErrInvalidClient)
	var authErr *AuthError
	require.ErrorAs(t, err, &authErr)
	var pricesErr *PricesError
	assert.False(t, errors.As(err, &pricesErr))
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 0, apiCalls)
}

func TestPricesErrorMessage(t *testing.T) {
	transportErr := errors.New("connection refused")
	tests := []struct {
		name     string
		err      *PricesError
		expected string
	}{
		{"detail over title", &PricesError{StatusCode: 400, Errors: []ErrorDetail{{Title: "Bad", Detail: "number is invalid"}, {Title: "Also bad"}}},
			"didww: prices service: number is invalid; Also bad"},
		{"status fallback", &PricesError{StatusCode: 202}, "didww: prices service responded with status 202"},
		{"transport error", &PricesError{Err: transportErr, kind: ErrPricesConnection}, "didww: prices service: connection refused"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.err, tt.expected)
		})
	}

	err := &PricesError{Err: transportErr, kind: ErrPricesConnection}
	assert.ErrorIs(t, err, ErrPricesConnection)
	assert.ErrorIs(t, err, transportErr)
	assert.NotErrorIs(t, err, ErrPricesUnexpectedStatus)
}

func TestPricesLookupUnparsableBody(t *testing.T) {
	for _, body := range []string{"<html>maintenance</html>", `{}`, `{"data":null}`, `{"data":[],"meta":{"total_records":0}}`} {
		t.Run(body, func(t *testing.T) {
			s, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, _ *http.Request) {
				writePricesJSON(w, http.StatusOK, []byte(body))
			})

			_, err := client.VoiceOutPrices().Lookup(context.Background(), map[string]string{"number": "442071234567"})

			var clientErr *ClientError
			require.ErrorAs(t, err, &clientErr)
			assert.False(t, errors.As(err, new(*PricesError)))
			assert.Equal(t, 1, s.requestCount())
		})
	}
}

func TestPricesServicesCallTheirOwnHost(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	srv := newAuthTestServer(t, nil, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	client := newTestClient(t, srv.URL,
		WithVoiceInPricesURL(srv.URL+"/vin"), WithVoiceOutPricesURL(srv.URL+"/vout"), WithSMSPricesURL(srv.URL+"/sms"))

	heads := []func(context.Context, time.Time) (*PriceFile, error){
		client.VoiceInPrices().Head, client.VoiceOutPrices().Head, client.SMSOutA2PPrices().Head,
		client.SMSOutP2PPrices().Head, client.SMSOutOTPPrices().Head, client.SMSInPrices().Head,
	}
	for _, head := range heads {
		_, err := head(context.Background(), time.Time{})
		require.NoError(t, err)
	}

	assert.Equal(t, []string{
		"/vin/v3/prices.csv.gz", "/vout/v3/prices.csv.gz", "/sms/v3/sms_out/a2p/prices.csv.gz",
		"/sms/v3/sms_out/p2p/prices.csv.gz", "/sms/v3/sms_out/otp/prices.csv.gz", "/sms/v3/sms_in/prices.csv.gz",
	}, paths)
}
