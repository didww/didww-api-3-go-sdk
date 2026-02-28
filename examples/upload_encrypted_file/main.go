// Encrypts and uploads a file using DIDWW public keys.
//
// Usage:
//
//	DIDWW_API_KEY=your_api_key go run ./examples/upload_encrypted_file/
//	DIDWW_API_KEY=your_api_key FILE_PATH=/path/to/file.pdf go run ./examples/upload_encrypted_file/
//
// If FILE_PATH is not set, the bundled example.pdf ("Hello From Go SDK") is used.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	didww "github.com/didww/didww-api-3-go-sdk"
	"github.com/didww/didww-api-3-go-sdk/examples"
)

func main() {
	client := examples.ClientFromEnv()
	ctx := context.Background()

	// Read file content
	filePath := os.Getenv("FILE_PATH")
	var fileContent []byte
	var originalName string
	if filePath == "" {
		// Use bundled example.pdf next to this source file
		_, thisFile, _, _ := runtime.Caller(0)
		filePath = filepath.Join(filepath.Dir(thisFile), "example.pdf")
		originalName = "example.pdf"
		fmt.Println("FILE_PATH not set, using bundled example.pdf")
	} else {
		originalName = filepath.Base(filePath)
	}

	var err error
	fileContent, err = os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Original file: %s (%d bytes)\n", originalName, len(fileContent))

	// Initialize encryption (fetches public keys from API without auth)
	enc, err := didww.NewEncrypt(ctx, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize encryption: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Fingerprint: %s\n", enc.Fingerprint())

	// Encrypt file content
	encryptedData, err := enc.Encrypt(fileContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encrypt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Encrypted size: %d bytes\n", len(encryptedData))

	// Upload encrypted file
	ids, err := client.UploadEncryptedFile(
		ctx,
		encryptedData,
		originalName+".enc",
		enc.Fingerprint(),
		originalName,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to upload: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Uploaded encrypted file IDs: %v\n", ids)
}
