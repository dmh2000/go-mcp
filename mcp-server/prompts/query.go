package prompts

import (
	"fmt"
)

func QueryPrompt(promptName string, arguments map[string]string) string {
	// Print the input parameters
	promptInfo := fmt.Sprintf("name: %s\n", promptName)

	for key, value := range arguments {
		promptInfo += fmt.Sprintf("%s: %s\n", key, value)
	}

	return promptInfo
}
