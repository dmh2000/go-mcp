// Package mcp defines types and functions related to the Model Context Protocol (MCP).
package mcp

import "encoding/json"

// --- PromptsList Request ---

// PromptsListRequest represents the full JSON-RPC request for the "prompts/list" method.
// Note: The example request doesn't show a 'params' field, but we include an empty
// struct for consistency with other MCP requests.
type PromptsListRequest struct {
	JSONRPC string      `json:"jsonrpc"`          // Should be "2.0"
	ID      interface{} `json:"id"`               // Request ID (string or number)
	Method  string      `json:"method"`           // Should be "prompts/list"
	Params  *struct{}   `json:"params,omitempty"` // Optional empty params object
}

// MarshalJSON implements the json.Marshaler interface for PromptsListRequest.
func (r *PromptsListRequest) MarshalJSON() ([]byte, error) {
	// Use a temporary type to handle the optional Params field correctly
	type Alias PromptsListRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	// Ensure Params is omitted if nil, or set to empty object if not nil
	if r.Params == nil {
		// If you want to strictly follow the example and *never* include params:
		// return json.Marshal(map[string]interface{}{
		// 	"jsonrpc": r.JSONRPC,
		// 	"id":      r.ID,
		// 	"method":  r.Method,
		// })
	} else {
		// If params is present (even if empty struct), marshal it
	}

	return json.Marshal(aux)
}

// UnmarshalJSON implements the json.Unmarshaler interface for PromptsListRequest.
func (r *PromptsListRequest) UnmarshalJSON(data []byte) error {
	type Alias PromptsListRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	// Set default value for Params if needed, or handle its absence
	return json.Unmarshal(data, &aux)
}

// --- PromptsList Response ---

// PromptArgument defines an argument for a prompt.
type PromptArgument struct {
	Name     string `json:"name"`              // Name of the argument.
	Required bool   `json:"required"`          // Whether the argument is required.
	// Add Type, Description etc. if needed later
}

// PromptDefinition describes a single prompt available on the server.
type PromptDefinition struct {
	Name        string           `json:"name"`                  // Name of the prompt.
	Description string           `json:"description,omitempty"` // Optional description.
	Arguments   []PromptArgument `json:"arguments,omitempty"`   // List of arguments the prompt takes.
}

// PromptsListResponse represents the full JSON-RPC response for the "prompts/list" method.
// Note: The 'prompts' field is at the top level, not within 'result', based on the example.
type PromptsListResponse struct {
	JSONRPC string             `json:"jsonrpc"`         // Should be "2.0"
	ID      interface{}        `json:"id"`              // Response ID (matches request ID)
	Prompts []PromptDefinition `json:"prompts"`         // List of available prompts.
	Error   *interface{}       `json:"error,omitempty"` // Error object, if any
}

// MarshalJSON implements the json.Marshaler interface for PromptsListResponse.
func (r *PromptsListResponse) MarshalJSON() ([]byte, error) {
	type Alias PromptsListResponse
	// Ensure Prompts slice is not nil before marshaling
	if r.Prompts == nil {
		r.Prompts = []PromptDefinition{}
	}
	// Ensure Arguments slice within each prompt is not nil
	for i := range r.Prompts {
		if r.Prompts[i].Arguments == nil {
			r.Prompts[i].Arguments = []PromptArgument{}
		}
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for PromptsListResponse.
func (r *PromptsListResponse) UnmarshalJSON(data []byte) error {
	type Alias PromptsListResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Ensure Prompts slice is not nil after unmarshaling
	if r.Prompts == nil {
		r.Prompts = []PromptDefinition{}
	}
	// Ensure Arguments slice within each prompt is not nil
	for i := range r.Prompts {
		if r.Prompts[i].Arguments == nil {
			r.Prompts[i].Arguments = []PromptArgument{}
		}
	}
	return nil
}

// --- PromptsGet Request ---

// PromptsGetParams defines the parameters for the "prompts/get" method.
type PromptsGetParams struct {
	Name      string      `json:"name"`      // The name of the prompt to get.
	Arguments interface{} `json:"arguments"` // The arguments for the prompt, structure depends on the prompt definition.
}

// PromptsGetRequest represents the full JSON-RPC request for the "prompts/get" method.
type PromptsGetRequest struct {
	JSONRPC string           `json:"jsonrpc"` // Should be "2.0"
	ID      interface{}      `json:"id"`      // Request ID (string or number)
	Method  string           `json:"method"`  // Should be "prompts/get"
	Params  PromptsGetParams `json:"params"`  // Parameters for the prompt retrieval
}

// MarshalJSON implements the json.Marshaler interface for PromptsGetRequest.
func (r *PromptsGetRequest) MarshalJSON() ([]byte, error) {
	type Alias PromptsGetRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for PromptsGetRequest.
func (r *PromptsGetRequest) UnmarshalJSON(data []byte) error {
	type Alias PromptsGetRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}

// --- PromptsGet Response ---

// MessageContent represents the content of a message in the prompts/get response.
type MessageContent struct {
	Type string `json:"type"` // Type of content, e.g., "text".
	Text string `json:"text"` // The actual text content.
}

// Message represents a single message (system or user) in the prompts/get response.
type Message struct {
	Role    string         `json:"role"`    // Role of the message sender (e.g., "system", "user").
	Content MessageContent `json:"content"` // The content of the message.
}

// PromptsGetResult holds the actual result data for the "prompts/get" response.
type PromptsGetResult struct {
	Messages []Message `json:"messages"` // A list of messages constituting the retrieved prompt.
}

// PromptsGetResponse represents the full JSON-RPC response for the "prompts/get" method.
type PromptsGetResponse struct {
	JSONRPC string           `json:"jsonrpc"`         // Should be "2.0"
	ID      interface{}      `json:"id"`              // Response ID (matches request ID)
	Result  PromptsGetResult `json:"result"`          // The actual result payload
	Error   *interface{}     `json:"error,omitempty"` // Error object, if any
}

// MarshalJSON implements the json.Marshaler interface for PromptsGetResponse.
func (r *PromptsGetResponse) MarshalJSON() ([]byte, error) {
	type Alias PromptsGetResponse
	// Ensure Messages slice in Result is not nil before marshaling
	if r.Result.Messages == nil {
		r.Result.Messages = []Message{}
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for PromptsGetResponse.
func (r *PromptsGetResponse) UnmarshalJSON(data []byte) error {
	type Alias PromptsGetResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Ensure Messages slice in Result is not nil after unmarshaling
	if r.Result.Messages == nil {
		r.Result.Messages = []Message{}
	}
	return nil
}
