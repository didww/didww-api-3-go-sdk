// Fetches and prints current account balance and credit.
//
// Usage: DIDWW_API_KEY=your_api_key go run ./examples/balance/
package main

import (
	"context"
	"fmt"

	"github.com/didww/didww-api-3-go-sdk/v3/examples"
)

func main() {
	client := examples.ClientFromEnv()

	balance, err := client.Balance().Find(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("Total Balance:", balance.TotalBalance)
	fmt.Println("Balance:", balance.Balance)
	fmt.Println("Credit:", balance.Credit)
}
