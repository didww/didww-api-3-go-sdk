package didww

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/didww/didww-api-3-go-sdk/v4/jsonapi"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
)

const (
	priceFileName            = "prices.csv.gz"
	priceLookupPath          = "lookup/prices"
	codePricePlanNotAssigned = "price_plan_not_assigned"
	codePriceNotFound        = "price_not_found"
	maxUnavailableRetries    = 2
	maxUnavailableRetryAfter = 60 * time.Second
	priceFileMode            = 0o644
)

// priceFilters lists the lookup filters a prices service accepts.
type priceFilters struct {
	required []string
	optional []string
	anyOf    []string
}

func (f priceFilters) allowed() []string {
	return slices.Concat(f.required, f.optional, f.anyOf)
}

var (
	voiceInPriceFilters  = priceFilters{required: []string{"number"}, optional: []string{"src_country_iso"}}
	voiceOutPriceFilters = priceFilters{required: []string{"number"}}
	smsPriceFilters      = priceFilters{anyOf: []string{"dst_number", "dst_country_iso"}}
)

// PricesService reads one price list of a prices service: its price file, a
// gzipped CSV, and its lookup. Requests use a token for the audience of the
// service host. Without a price plan of the service every call fails with
// ErrPricePlanNotAssigned.
//
// Failed calls return a *PricesError (see the ErrPrices* sentinels), a
// *AuthError when no token could be obtained, or a plain error for invalid
// arguments. Lookup returns a *ClientError when a 200 answer is not a JSON:API
// document it can parse or holds no prices, and Download a *ClientError for a
// nil io.Writer. DownloadToFile and Fetch also return the os and io/fs errors of
// the file at the path (stat, temp file, file mode and times, rename) as is,
// and Download returns an error of its io.Writer as is. When ctx ends while a
// 503 Retry-After is being waited out, ctx's error is returned as is.
type PricesService[T any] struct {
	client   *Client
	baseURL  string
	audience string
	prefix   string
	filters  priceFilters
	sleep    func(ctx context.Context, d time.Duration) error
}

func newPricesService[T any](c *Client, baseURL, audience, prefix string, filters priceFilters) *PricesService[T] {
	return &PricesService[T]{
		client:   c,
		baseURL:  baseURL,
		audience: audience,
		prefix:   prefix,
		filters:  filters,
		sleep:    sleepContext,
	}
}

// VoiceInPrices returns the Voice IN price list. Lookup filters: number
// (required), src_country_iso.
func (c *Client) VoiceInPrices() *PricesService[resource.VoiceInPrice] {
	return newPricesService[resource.VoiceInPrice](c, c.voiceInPricesURL, audienceVoiceIn, "", voiceInPriceFilters)
}

// VoiceOutPrices returns the Voice OUT price list. Lookup filters: number (required).
func (c *Client) VoiceOutPrices() *PricesService[resource.VoiceOutPrice] {
	return newPricesService[resource.VoiceOutPrice](c, c.voiceOutPricesURL, audienceVoiceOut, "", voiceOutPriceFilters)
}

// SMSOutA2PPrices returns the SMS OUT A2P price list. Lookup filters: one of
// dst_number and dst_country_iso, or both.
func (c *Client) SMSOutA2PPrices() *PricesService[resource.SMSOutPrice] {
	return newPricesService[resource.SMSOutPrice](c, c.smsPricesURL, audienceSMS, "sms_out/a2p/", smsPriceFilters)
}

// SMSOutP2PPrices returns the SMS OUT P2P price list. Lookup filters: one of
// dst_number and dst_country_iso, or both.
func (c *Client) SMSOutP2PPrices() *PricesService[resource.SMSOutPrice] {
	return newPricesService[resource.SMSOutPrice](c, c.smsPricesURL, audienceSMS, "sms_out/p2p/", smsPriceFilters)
}

// SMSOutOTPPrices returns the SMS OUT OTP price list. Lookup filters: one of
// dst_number and dst_country_iso, or both.
func (c *Client) SMSOutOTPPrices() *PricesService[resource.SMSOutPrice] {
	return newPricesService[resource.SMSOutPrice](c, c.smsPricesURL, audienceSMS, "sms_out/otp/", smsPriceFilters)
}

// SMSInPrices returns the SMS IN price list. Lookup filters: one of
// dst_number and dst_country_iso, or both.
func (c *Client) SMSInPrices() *PricesService[resource.SMSInPrice] {
	return newPricesService[resource.SMSInPrice](c, c.smsPricesURL, audienceSMS, "sms_in/", smsPriceFilters)
}

func (s *PricesService[T]) filePath() string   { return s.prefix + priceFileName }
func (s *PricesService[T]) lookupPath() string { return s.prefix + priceLookupPath }

// PriceLookup is the answer of a price lookup. It always holds at least one
// row: no match is ErrPriceNotFound.
//
// Row IDs of SMS prices are derived from the row content and change with the
// price, so they do not track a destination over time.
type PriceLookup[T any] struct {
	Rows         []*T
	TotalRecords int
}

// Lookup returns the prices matching filters, such as
// {"number": "18005550123"}. Empty values are dropped. A missing or unknown
// filter fails before any request; the formats of the values are checked by
// the server (ErrPricesBadRequest).
func (s *PricesService[T]) Lookup(ctx context.Context, filters map[string]string) (*PriceLookup[T], error) {
	query, err := s.lookupQuery(filters)
	if err != nil {
		return nil, err
	}
	u := joinV3(s.baseURL, s.lookupPath()) + "?" + query.Encode()

	resp, err := s.client.send(ctx, s.client.noRedirectClient, s.audience, func() (*http.Request, error) {
		req, buildErr := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
		if buildErr != nil {
			return nil, &ClientError{Message: fmt.Sprintf("failed to create price lookup request: %v", buildErr)}
		}
		req.Header.Set("Accept", jsonapiMediaType)
		return req, nil
	})
	if err != nil {
		return nil, pricesSendError(err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	if resp.StatusCode != http.StatusOK {
		return nil, newPricesError(resp.StatusCode, resp.Header, readErrorBody(resp.Body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &PricesError{StatusCode: resp.StatusCode, Err: err, kind: ErrPricesConnection}
	}
	return parsePriceLookup[T](body)
}

// lookupQuery checks filters against the filters of the service and encodes them.
func (s *PricesService[T]) lookupQuery(filters map[string]string) (url.Values, error) {
	names := make([]string, 0, len(filters))
	for name, value := range filters {
		if strings.TrimSpace(value) != "" {
			names = append(names, name)
		}
	}
	slices.Sort(names)

	allowed := s.filters.allowed()
	var unknown, missing []string
	for _, name := range names {
		if !slices.Contains(allowed, name) {
			unknown = append(unknown, name)
		}
	}
	for _, name := range s.filters.required {
		if !slices.Contains(names, name) {
			missing = append(missing, name)
		}
	}
	anyOfGiven := slices.ContainsFunc(s.filters.anyOf, func(name string) bool { return slices.Contains(names, name) })

	switch {
	case len(unknown) > 0:
		return nil, fmt.Errorf("filter %s is not allowed on %s (allowed: %s)",
			strings.Join(unknown, ", "), s.lookupPath(), strings.Join(allowed, ", "))
	case len(missing) > 0:
		return nil, fmt.Errorf("filter %s is required on %s", strings.Join(missing, ", "), s.lookupPath())
	case len(s.filters.anyOf) > 0 && !anyOfGiven:
		return nil, fmt.Errorf("one of filters %s is required on %s", strings.Join(s.filters.anyOf, ", "), s.lookupPath())
	}

	query := url.Values{}
	for _, name := range names {
		query.Set("filter["+name+"]", filters[name])
	}
	return query, nil
}

func parsePriceLookup[T any](body []byte) (*PriceLookup[T], error) {
	rows, err := jsonapi.UnmarshalMany[T](body)
	if err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to parse price lookup: %v", err)}
	}
	if len(rows) == 0 {
		return nil, &ClientError{Message: "failed to parse price lookup: the answer holds no prices"}
	}
	var doc struct {
		Meta struct {
			TotalRecords int `json:"total_records"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, &ClientError{Message: fmt.Sprintf("failed to parse price lookup meta: %v", err)}
	}
	return &PriceLookup[T]{Rows: rows, TotalRecords: doc.Meta.TotalRecords}, nil
}

// PriceFile describes the price file of a HEAD or download answer.
type PriceFile struct {
	StatusCode    int       // 200, or 304 when the file has not changed since If-Modified-Since
	LastModified  time.Time // zero when absent
	ContentLength int64     // -1 when unknown, e.g. on 304
	Filename      string    // from Content-Disposition, e.g. "voice_in_prices.csv.gz"
	BytesWritten  int64     // bytes written by a download; 0 for HEAD and 304
}

// NotModified reports whether the file has not changed since the
// If-Modified-Since date sent (304).
func (f *PriceFile) NotModified() bool {
	return f.StatusCode == http.StatusNotModified
}

// Head returns the headers of the price file without its content. A non-zero
// ifModifiedSince is sent as If-Modified-Since.
func (s *PricesService[T]) Head(ctx context.Context, ifModifiedSince time.Time) (*PriceFile, error) {
	return s.requestFile(ctx, http.MethodHead, ifModifiedSince, nil)
}

// Download writes the price file to w exactly as served: gzipped, never
// decompressed (see EachPriceRow). A non-zero ifModifiedSince is sent as
// If-Modified-Since; a 304 answer writes nothing.
//
// Only a 200 answer is written, but w may hold partial data when the
// transfer fails. An error returned by w is returned as is.
func (s *PricesService[T]) Download(ctx context.Context, w io.Writer, ifModifiedSince time.Time) (*PriceFile, error) {
	if w == nil {
		return nil, &ClientError{Message: "destination writer is required"}
	}
	return s.requestFile(ctx, http.MethodGet, ifModifiedSince, w)
}

// DownloadToFile downloads the price file to path. The file is written to a
// temporary file next to path, which replaces path only when the download
// succeeds, so a failed download leaves an existing file as it was. The new
// file gets mode 0644 and the Last-Modified of the answer as its
// modification time.
func (s *PricesService[T]) DownloadToFile(ctx context.Context, path string) (*PriceFile, error) {
	return s.downloadToFile(ctx, path, time.Time{})
}

// Fetch downloads the price file to path like DownloadToFile unless the file
// at path is current: the modification time of an existing file is sent as
// If-Modified-Since, and a 304 answer leaves the file untouched. Fetch trusts
// the modification time of any file at path.
func (s *PricesService[T]) Fetch(ctx context.Context, path string) (*PriceFile, error) {
	var ifModifiedSince time.Time
	info, err := os.Stat(path)
	switch {
	case err == nil:
		ifModifiedSince = info.ModTime()
	case !errors.Is(err, fs.ErrNotExist):
		return nil, err
	}
	return s.downloadToFile(ctx, path, ifModifiedSince)
}

func (s *PricesService[T]) downloadToFile(ctx context.Context, path string, ifModifiedSince time.Time) (*PriceFile, error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) //nolint:errcheck // already renamed on success

	file, err := s.requestFile(ctx, http.MethodGet, ifModifiedSince, tmp)
	if err == nil && file.StatusCode == http.StatusOK {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	if file.NotModified() {
		return file, nil
	}
	if err := os.Chmod(tmpPath, priceFileMode); err != nil {
		return nil, err
	}
	if !file.LastModified.IsZero() {
		if err := os.Chtimes(tmpPath, file.LastModified, file.LastModified); err != nil {
			return nil, err
		}
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return nil, err
	}
	return file, nil
}

// requestFile sends a HEAD or GET for the price file, retrying a 503 whose
// Retry-After is at most a minute up to maxUnavailableRetries times.
func (s *PricesService[T]) requestFile(ctx context.Context, method string, ifModifiedSince time.Time, w io.Writer) (*PriceFile, error) {
	for attempt := 1; ; attempt++ {
		file, wait, err := s.fileAttempt(ctx, method, ifModifiedSince, w, attempt)
		if file != nil || err != nil {
			return file, err
		}
		if err := s.sleep(ctx, wait); err != nil {
			return nil, err
		}
	}
}

// fileAttempt sends one file request. It returns neither a file nor an error
// when the request is to be retried after wait.
func (s *PricesService[T]) fileAttempt(ctx context.Context, method string, ifModifiedSince time.Time, w io.Writer, attempt int) (file *PriceFile, wait time.Duration, err error) {
	u := joinV3(s.baseURL, s.filePath())
	resp, err := s.client.send(ctx, s.client.noRedirectClient, s.audience, func() (*http.Request, error) {
		req, buildErr := http.NewRequestWithContext(ctx, method, u, http.NoBody)
		if buildErr != nil {
			return nil, &ClientError{Message: fmt.Sprintf("failed to create price file request: %v", buildErr)}
		}
		if !ifModifiedSince.IsZero() {
			req.Header.Set("If-Modified-Since", ifModifiedSince.UTC().Format(http.TimeFormat))
		}
		return req, nil
	})
	if err != nil {
		return nil, 0, pricesSendError(err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNotModified:
		file = newPriceFile(resp)
		if resp.StatusCode == http.StatusOK && w != nil {
			if file.BytesWritten, err = copyPriceFile(w, resp.Body); err != nil {
				return nil, 0, err
			}
		}
		return file, 0, nil
	case http.StatusServiceUnavailable:
		if d, ok := retryAfterSeconds(resp.Header); ok && attempt <= maxUnavailableRetries && d <= maxUnavailableRetryAfter {
			drainAndClose(resp.Body)
			return nil, d, nil
		}
	}
	return nil, 0, newPricesError(resp.StatusCode, resp.Header, readErrorBody(resp.Body))
}

// copyPriceFile copies body to w. An error of w is returned as is; a read
// error is a connection error.
func copyPriceFile(w io.Writer, body io.Reader) (int64, error) {
	rec := &writeRecorder{w: w}
	n, err := io.Copy(rec, body)
	switch {
	case err == nil:
		return n, nil
	case rec.err != nil:
		return n, rec.err
	case errors.Is(err, io.ErrShortWrite):
		return n, err
	default:
		return n, &PricesError{StatusCode: http.StatusOK, Err: err, kind: ErrPricesConnection}
	}
}

func newPriceFile(resp *http.Response) *PriceFile {
	file := &PriceFile{StatusCode: resp.StatusCode, ContentLength: -1}
	if t, err := http.ParseTime(resp.Header.Get("Last-Modified")); err == nil {
		file.LastModified = t
	}
	if n, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64); err == nil && n >= 0 {
		file.ContentLength = n
	}
	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil {
		file.Filename = params["filename"]
	}
	return file
}

// EachPriceRow reads a price file as downloaded (gzipped CSV) and calls fn with
// each row, keyed by the header line. It stops at the first error; an error
// returned by fn is returned as is.
func EachPriceRow(r io.Reader, fn func(row map[string]string) error) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("read price file: %w", err)
	}
	defer gz.Close() //nolint:errcheck // read-only

	br := bufio.NewReader(gz)
	if bom, _ := br.Peek(3); bytes.Equal(bom, []byte("\uFEFF")) {
		_, _ = br.Discard(3)
	}
	cr := csv.NewReader(br)
	header, err := cr.Read()
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read price file: %w", err)
	}

	for {
		record, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read price file: %w", err)
		}
		row := make(map[string]string, len(header))
		for i, name := range header {
			row[name] = record[i]
		}
		if err := fn(row); err != nil {
			return err
		}
	}
}

// pricesSendError returns token and client errors as is and turns transport
// errors into a connection error.
func pricesSendError(err error) error {
	var authErr *AuthError
	var clientErr *ClientError
	if errors.As(err, &authErr) || errors.As(err, &clientErr) {
		return err
	}
	return &PricesError{Err: err, kind: ErrPricesConnection}
}

// newPricesError maps an answer that is not a success to a *PricesError.
func newPricesError(status int, header http.Header, body []byte) *PricesError {
	e := &PricesError{StatusCode: status, Errors: parsePricesErrors(body)}
	if len(e.Errors) > 0 {
		e.Code = e.Errors[0].Code
	}
	if d, ok := retryAfterSeconds(header); ok {
		e.RetryAfter = d
	}
	switch status {
	case http.StatusBadRequest:
		e.kind = ErrPricesBadRequest
	case http.StatusUnauthorized:
		e.kind = ErrPricesNotAuthorized
	case http.StatusForbidden:
		e.kind = ErrPricesAccessDenied
	case http.StatusNotFound:
		if e.Code == "" {
			e.Code = codePricePlanNotAssigned
		}
		e.kind = ErrPricePlanNotAssigned
		if e.Code == codePriceNotFound {
			e.kind = ErrPriceNotFound
		}
	case http.StatusNotAcceptable:
		e.kind = ErrPricesNotAcceptable
	case http.StatusTooManyRequests:
		e.kind = ErrPricesTooManyRequests
	case http.StatusServiceUnavailable:
		e.kind = ErrPricesServiceUnavailable
	default:
		e.kind = ErrPricesUnexpectedStatus
	}
	return e
}

// parsePricesErrors returns the error objects of a JSON:API errors document,
// or nil for any other body.
func parsePricesErrors(body []byte) []ErrorDetail {
	var doc struct {
		Errors []json.RawMessage `json:"errors"`
	}
	if json.Unmarshal(body, &doc) != nil {
		return nil
	}
	var details []ErrorDetail
	for _, raw := range doc.Errors {
		if !strings.HasPrefix(strings.TrimSpace(string(raw)), "{") {
			continue
		}
		var d ErrorDetail
		if json.Unmarshal(raw, &d) == nil {
			details = append(details, d)
		}
	}
	return details
}

// retryAfterSeconds parses a Retry-After given in seconds; an HTTP date is ignored.
func retryAfterSeconds(header http.Header) (time.Duration, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(header.Get("Retry-After")), 10, 32)
	if err != nil || n < 0 {
		return 0, false
	}
	return time.Duration(n) * time.Second, true
}

func readErrorBody(body io.Reader) []byte {
	b, _ := io.ReadAll(io.LimitReader(body, maxErrorBodySize))
	return b
}
