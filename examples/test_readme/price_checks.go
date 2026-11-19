package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
)

// priceSkip turns the answers that depend on the account's price plans into
// a skip.
func priceSkip(err error) error {
	switch {
	case errors.Is(err, didww.ErrPricePlanNotAssigned):
		return skip("the account has no price plan of this service")
	case errors.Is(err, didww.ErrPriceNotFound):
		return skip("the account's price plan has no price for the README number")
	}
	return err
}

// README: Price lookups
func (s *suite) priceLookups() {
	ctx, client := s.ctx, s.client
	s.section("Price lookups")

	s.check("Voice IN", func() error {
		result, err := client.VoiceInPrices().Lookup(ctx, map[string]string{
			"number":          "18005550123",
			"src_country_iso": "US",
		})
		if err != nil {
			return priceSkip(err)
		}
		s.logf("%d matching prices", result.TotalRecords)
		for _, price := range result.Rows {
			if price.Rate != nil {
				s.logf("%s %s: %s USD/min", price.Destination, price.SourceType, *price.Rate)
			}
		}
		return nil
	})

	s.check("Voice OUT", func() error {
		voiceOut, err := client.VoiceOutPrices().Lookup(ctx, map[string]string{"number": "493012345678"})
		if err != nil {
			return priceSkip(err)
		}
		s.logf("category %s", voiceOut.Rows[0].Category) // enums.VoiceOutPriceCategory
		return nil
	})

	s.check("SMS OUT A2P", func() error {
		a2p, err := client.SMSOutA2PPrices().Lookup(ctx, map[string]string{"dst_number": "4915123456789"})
		if err != nil {
			return priceSkip(err)
		}
		s.logf("%d matching prices", a2p.TotalRecords)
		return nil
	})

	s.check("SMS IN", func() error {
		smsIn, err := client.SMSInPrices().Lookup(ctx, map[string]string{"dst_country_iso": "US"})
		if err != nil {
			return priceSkip(err)
		}
		s.logf("%d matching prices", smsIn.TotalRecords)
		return nil
	})
}

// README: Price files
func (s *suite) priceFiles() {
	ctx, client := s.ctx, s.client
	s.section("Price files")
	s.check("Fetch, read, check and download the Voice OUT price file", func() error {
		dir, err := os.MkdirTemp("", "didww-readme-")
		if err != nil {
			return err
		}
		defer func() { _ = os.RemoveAll(dir) }()
		path := filepath.Join(dir, "voice_out_prices.csv.gz")

		prices := client.VoiceOutPrices()

		// Download unless the copy at the path is current
		file, err := prices.Fetch(ctx, path)
		if err != nil {
			return priceSkip(err)
		}
		if file.NotModified() {
			s.logf("answered 304, the file was kept")
		} else {
			s.logf("downloaded %d bytes, last modified %s", file.BytesWritten, file.LastModified)
		}

		// Read the rows
		err = s.readPriceRows(path)
		if err != nil {
			return err
		}

		// Lower-level calls
		status, err := prices.Head(ctx, file.LastModified) // sent as If-Modified-Since; a zero time sends none
		if err != nil {
			return err
		}
		s.logf("head: not modified %t, content length %d, filename %q", status.NotModified(), status.ContentLength, status.Filename)

		if file, err = prices.DownloadToFile(ctx, path); err != nil {
			return err
		}
		s.logf("downloaded to file: %d bytes", file.BytesWritten)
		if file, err = prices.Download(ctx, io.Discard, time.Time{}); err != nil { // any io.Writer
			return err
		}
		s.logf("downloaded to a writer: %d bytes", file.BytesWritten)
		return nil
	})
}

var errEnoughRows = errors.New("enough rows")

// readPriceRows prints the first rows of a downloaded price file.
func (s *suite) readPriceRows(path string) error {
	f, err := os.Open(path) //nolint:gosec // a path in the check's temp directory
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	rows := 0
	err = didww.EachPriceRow(f, func(row map[string]string) error {
		s.logf("%s: %s", row["Prefix"], row["Rate"])
		rows++
		if rows == 3 {
			return errEnoughRows
		}
		return nil
	})
	if errors.Is(err, errEnoughRows) {
		return nil
	}
	return err
}

// README: Error Handling, the token service and prices errors
func (s *suite) pricesErrors() {
	ctx, client := s.ctx, s.client
	s.check("Token service and prices errors", func() error {
		_, err := client.SMSOutA2PPrices().Lookup(ctx, map[string]string{"dst_number": "4915123456789"})

		var authErr *didww.AuthError
		var pricesErr *didww.PricesError
		switch {
		case errors.Is(err, didww.ErrRateLimited) && errors.As(err, &authErr):
			// Wait RetryAfter and try again; repeat while the error is ErrRateLimited.
			s.logf("Token service rate limit, retry after %s", authErr.RetryAfter)
		case errors.Is(err, didww.ErrInvalidClient):
			return err // the API key was refused
		case errors.Is(err, didww.ErrPriceNotFound):
			s.logf("No price for this number")
		case errors.Is(err, didww.ErrPricesTooManyRequests) && errors.As(err, &pricesErr):
			s.logf("Prices rate limit, retry after %s", pricesErr.RetryAfter)
		case err != nil:
			s.logf("Error: %v", err)
			return priceSkip(err) // only a missing price plan is not a failure
		default:
			s.logf("price found")
		}
		return nil
	})
}
