package examples

import (
	"fmt"
	"os"

	didww "github.com/didww/didww-api-3-go-sdk/v3"
)

// ClientFromEnv creates a DIDWW client using the DIDWW_API_KEY environment variable.
func ClientFromEnv() *didww.Client {
	apiKey := os.Getenv("DIDWW_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "DIDWW_API_KEY environment variable is required")
		os.Exit(1)
	}

	client, err := didww.NewClient(apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create client: %v\n", err)
		os.Exit(1)
	}

	return client
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}
