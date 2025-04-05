// Package api provides integration with Anthropic's Claude AI models.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// QueryText sends a text query to the specified Anthropic model and returns the response.
func QueryText(ctx context.Context, client *anthropic.Client, prompts []string, model string) (string, error) {
	if ctx.Err() != nil {
		return "", fmt.Errorf("request context error %w", ctx.Err())
	}

	if len(prompts) == 0 {
		prompts = []string{"Hello, how are you?"}
	}

	// prompts are user messages
	messages := make([]anthropic.MessageParam, 0, len(prompts))
	for _, p := range prompts {
		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(p)))
	}

	// Create new message request with the provided prompt and temperature
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 4096,
		Model:     anthropic.Model(model),
		System: []anthropic.TextBlockParam{
			{Text: "You are a helpful assistant."},
		},
		Messages: messages,
	})

	if err != nil {
		return "", fmt.Errorf("failed to create message: %w", err)
	}

	// Verify we got a non-empty response
	if len(message.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// Build response using strings.Builder for better performance
	var response strings.Builder
	for _, content := range message.Content {
		response.WriteString(content.Text)
	}
	return response.String(), nil
}
