// Package mcp defines types and functions related to the Model Context Protocol (MCP).
package mcp

import "encoding/json"

// ToolsListRequest represents the full JSON-RPC request for the "tools/list" method.
type ToolsListRequest struct {
	JSONRPC string      `json:"jsonrpc"` // Should be "2.0"
	ID      interface{} `json:"id"`      // Request ID (string or number)
	Method  string      `json:"method"`  // Should be "tools/list"
	Params  struct{}    `json:"params"`  // Empty params object for tools/list
}

// MarshalJSON implements the json.Marshaler interface for ToolsListRequest.
func (r *ToolsListRequest) MarshalJSON() ([]byte, error) {
	type Alias ToolsListRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for ToolsListRequest.
func (r *ToolsListRequest) UnmarshalJSON(data []byte) error {
	type Alias ToolsListRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}

// --- ToolsList Response ---

// ToolDefinition describes a single tool available on the server.
type ToolDefinition struct {
	Name        string      `json:"name"`                  // The name of the tool.
	Description string      `json:"description,omitempty"` // Optional description of the tool.
	InputSchema interface{} `json:"inputSchema,omitempty"` // Optional JSON schema for the tool's input parameters.
}

// ToolsListResult holds the actual result data for the "tools/list" response.
type ToolsListResult struct {
	Tools []ToolDefinition `json:"tools"` // A list of available tools.
}

// ToolsListResponse represents the full JSON-RPC response for the "tools/list" method.
type ToolsListResponse struct {
	JSONRPC string          `json:"jsonrpc"`         // Should be "2.0"
	ID      interface{}     `json:"id"`              // Response ID (matches request ID)
	Result  ToolsListResult `json:"result"`          // The actual result payload
	Error   *interface{}    `json:"error,omitempty"` // Error object, if any
}

// MarshalJSON implements the json.Marshaler interface for ToolsListResponse.
func (r *ToolsListResponse) MarshalJSON() ([]byte, error) {
	type Alias ToolsListResponse
	// Ensure Tools slice in Result is not nil before marshaling
	if r.Result.Tools == nil {
		r.Result.Tools = []ToolDefinition{}
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for ToolsListResponse.
func (r *ToolsListResponse) UnmarshalJSON(data []byte) error {
	type Alias ToolsListResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Ensure Tools slice in Result is not nil after unmarshaling
	if r.Result.Tools == nil {
		r.Result.Tools = []ToolDefinition{}
	}
	return nil
}

// --- ToolsCall Request ---

// ToolsCallParams defines the parameters for the "tools/call" method.
type ToolsCallParams struct {
	Name      string      `json:"name"`      // The name of the tool to call.
	Arguments interface{} `json:"arguments"` // The arguments for the tool call, structure depends on the tool.
}

// ToolsCallRequest represents the full JSON-RPC request for the "tools/call" method.
type ToolsCallRequest struct {
	JSONRPC string          `json:"jsonrpc"` // Should be "2.0"
	ID      interface{}     `json:"id"`      // Request ID (string or number)
	Method  string          `json:"method"`  // Should be "tools/call"
	Params  ToolsCallParams `json:"params"`  // Parameters for the tool call
}

// MarshalJSON implements the json.Marshaler interface for ToolsCallRequest.
func (r *ToolsCallRequest) MarshalJSON() ([]byte, error) {
	type Alias ToolsCallRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for ToolsCallRequest.
func (r *ToolsCallRequest) UnmarshalJSON(data []byte) error {
	type Alias ToolsCallRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}

// --- ToolsCall Response ---

// ContentItem represents a piece of content within the tools/call response.
type ContentItem struct {
	Type string `json:"type"` // Type of content, e.g., "text".
	Text string `json:"text"` // The actual text content.
}

// ToolsCallResult holds the actual result data for the "tools/call" response.
type ToolsCallResult struct {
	Content []ContentItem `json:"content"` // A list of content items resulting from the tool call.
}

// ToolsCallResponse represents the full JSON-RPC response for the "tools/call" method.
type ToolsCallResponse struct {
	JSONRPC string          `json:"jsonrpc"`         // Should be "2.0"
	ID      interface{}     `json:"id"`              // Response ID (matches request ID)
	Result  ToolsCallResult `json:"result"`          // The actual result payload
	Error   *interface{}    `json:"error,omitempty"` // Error object, if any
}

// MarshalJSON implements the json.Marshaler interface for ToolsCallResponse.
func (r *ToolsCallResponse) MarshalJSON() ([]byte, error) {
	type Alias ToolsCallResponse
	// Ensure Content slice in Result is not nil before marshaling
	if r.Result.Content == nil {
		r.Result.Content = []ContentItem{}
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for ToolsCallResponse.
func (r *ToolsCallResponse) UnmarshalJSON(data []byte) error {
	type Alias ToolsCallResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Ensure Content slice in Result is not nil after unmarshaling
	if r.Result.Content == nil {
		r.Result.Content = []ContentItem{}
	}
	return nil
}
