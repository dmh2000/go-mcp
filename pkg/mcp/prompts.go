package mcp

import "encoding/json"

// Method names for prompt operations.
const (
	MethodListPrompts = "prompts/list"
	MethodGetPrompt   = "prompts/get"
)

// PromptArgument describes an argument that a prompt template can accept.
type PromptArgument struct {
	// Description is a human-readable description of the argument.
	Description string `json:"description,omitempty"`
	// Name is the name of the argument.
	Name string `json:"name"`
	// Required indicates whether this argument must be provided.
	Required bool `json:"required,omitempty"` // Defaults to false if omitted
}

// Prompt represents a prompt or prompt template offered by the server.
type Prompt struct {
	// Arguments is a list of arguments the prompt template accepts.
	Arguments []PromptArgument `json:"arguments,omitempty"`
	// Description is an optional description of what the prompt provides.
	Description string `json:"description,omitempty"`
	// Name is the unique name of the prompt or prompt template.
	Name string `json:"name"`
}

// TextContent represents text content within a prompt message.
// Note: Duplicated from resources.go for clarity, consider consolidating.
type TextContent struct {
	Annotations *Annotations `json:"annotations,omitempty"`
	Text        string       `json:"text"`
	Type        string       `json:"type"` // Should be "text"
}

// ImageContent represents image content within a prompt message.
// Note: Duplicated from resources.go for clarity, consider consolidating.
type ImageContent struct {
	Annotations *Annotations `json:"annotations,omitempty"`
	Data        string       `json:"data"` // base64 encoded
	MimeType    string       `json:"mimeType"`
	Type        string       `json:"type"` // Should be "image"
}

// PromptMessage describes a message returned as part of a prompt.
// It's similar to SamplingMessage but supports embedded resources.
type PromptMessage struct {
	// Content holds the message data (TextContent, ImageContent, or EmbeddedResource).
	// Needs to be unmarshaled into the specific type based on the "type" field
	// after initial unmarshaling into json.RawMessage.
	Content json.RawMessage `json:"content"`
	// Role indicates the sender of the message (user or assistant).
	Role Role `json:"role"`
}

// ListPromptsParams defines the parameters for a "prompts/list" request.
type ListPromptsParams struct {
	// Cursor is an opaque token for pagination.
	Cursor string `json:"cursor,omitempty"`
}

// ListPromptsResult defines the result structure for a "prompts/list" response.
type ListPromptsResult struct {
	// Meta contains reserved protocol metadata.
	Meta map[string]interface{} `json:"_meta,omitempty"`
	// NextCursor is an opaque token for the next page of results.
	NextCursor string `json:"nextCursor,omitempty"`
	// Prompts is the list of prompts found.
	Prompts []Prompt `json:"prompts"`
}

// GetPromptParams defines the parameters for a "prompts/get" request.
type GetPromptParams struct {
	// Arguments to use for templating the prompt.
	Arguments map[string]string `json:"arguments,omitempty"`
	// Name is the name of the prompt or prompt template to retrieve.
	Name string `json:"name"`
}

// GetPromptResult defines the result structure for a "prompts/get" response.
// Note: The schema defines this as GetPromptResponse, using Result here for consistency.
type GetPromptResult struct {
	// Meta contains reserved protocol metadata.
	Meta map[string]interface{} `json:"_meta,omitempty"`
	// Description is an optional description for the prompt.
	Description string `json:"description,omitempty"`
	// Messages is the sequence of messages constituting the prompt.
	Messages []PromptMessage `json:"messages"`
}

// Note: Standard json.Marshal and json.Unmarshal can be used for these types.
// For PromptMessage.Content, further processing is needed after unmarshaling
// to determine the concrete type (TextContent, ImageContent, or EmbeddedResource).
