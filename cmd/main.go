// Anthropic API Client
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Println("ANTHROPIC_API_KEY environment variable not set")
		return
	}

	// Initialize the Anthropic client with the API key from environment variable
	var client = anthropic.NewClient()

	// Example usage of QueryText method
	ctx := context.Background()
	prompts := []string{"Hello, how are you?"}
	model := "claude-3-7-sonnet-latest"

	response, err := QueryText(ctx, &client, prompts, model)
	if err != nil {
		fmt.Println("Error querying text:", err)
		return
	}

	fmt.Println("Response:", response)
}
