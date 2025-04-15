package prompts

import (
	"fmt"
)

func QueryPrompt(promptName string, arguments map[string]string) string {
	// Print the input parameters
	promptInfo := fmt.Sprintf("Prompt Name: %s\nArguments: %v\n\n", promptName, arguments)
	
	return promptInfo + `You are an AI assistant that helps users query information from various sources.
Please respond to the user's query in a helpful, accurate, and concise manner.
If you don't know the answer, it's better to say so than to make up information.
Always cite your sources when providing factual information.`
}
