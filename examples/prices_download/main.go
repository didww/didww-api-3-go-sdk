// Downloads the Voice IN price file, checks it again with If-Modified-Since,
// and reads a few rows.
//
// The file is written exactly as served, still gzipped, and dated with the
// response's Last-Modified. Fetch(ctx, path) sends the modification time of
// the file at path as If-Modified-Since, so while the price file is unchanged
// the second call, and any later run against the same file, is answered 304
// and keeps it.
//
// The file goes to dir, or to didww-prices in the user cache directory, created
// for the current user only when missing.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/prices_download/ [dir]
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	didww "github.com/didww/didww-api-3-go-sdk/v4"
	"github.com/didww/didww-api-3-go-sdk/v4/examples"
	"github.com/didww/didww-api-3-go-sdk/v4/resource"
)

const firstRows = 5

// errEnoughRows stops EachPriceRow after the first rows.
var errEnoughRows = errors.New("enough rows")

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	dir, err := priceDir()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "voice_in_prices.csv.gz")
	fmt.Println("Price files go to", dir)

	err = run(ctx, client.VoiceInPrices(), path)
	var pricesErr *didww.PricesError
	switch {
	case err == nil:
	case errors.Is(err, didww.ErrPricePlanNotAssigned):
		fmt.Println("No Voice IN price plan on this account:", err)
	case errors.Is(err, didww.ErrPricesServiceUnavailable) && errors.As(err, &pricesErr):
		fmt.Printf("The price file is not available yet, retry in %s: %v\n", pricesErr.RetryAfter, err)
	default:
		panic(err)
	}
}

// priceDir returns the directory given as the first argument, or didww-prices
// in the user cache directory, creating it with mode 0700 when missing.
func priceDir() (string, error) {
	if len(os.Args) > 1 {
		return os.Args[1], nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cache, "didww-prices")
	return dir, os.MkdirAll(dir, 0o700)
}

func run(ctx context.Context, prices *didww.PricesService[resource.VoiceInPrice], path string) error {
	fmt.Println("=== HEAD ===")
	head, err := prices.Head(ctx, time.Time{})
	if err != nil {
		return err
	}
	fmt.Printf("Last-Modified: %s, %d bytes, %s\n", head.LastModified.Format(http.TimeFormat), head.ContentLength, head.Filename)

	fmt.Printf("\n=== Fetch %s ===\n", path)
	if err := fetch(ctx, prices, path); err != nil {
		return err
	}

	fmt.Println("\n=== Fetch again ===")
	if err := fetch(ctx, prices, path); err != nil {
		return err
	}

	fmt.Println("\n=== First rows ===")
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck // read-only

	n := 0
	err = didww.EachPriceRow(f, func(row map[string]string) error {
		fmt.Printf("  %s %s %s: %s\n", row["Country"], row["Destination"], row["Source type"], row["Rate (USD, excl. VAT)"])
		n++
		if n == firstRows {
			return errEnoughRows
		}
		return nil
	})
	if errors.Is(err, errEnoughRows) {
		return nil
	}
	return err
}

func fetch(ctx context.Context, prices *didww.PricesService[resource.VoiceInPrice], path string) error {
	file, err := prices.Fetch(ctx, path)
	if err != nil {
		return err
	}
	if file.NotModified() {
		fmt.Println("Not modified, file kept")
	} else {
		fmt.Printf("Downloaded %d bytes\n", file.BytesWritten)
	}
	return nil
}
