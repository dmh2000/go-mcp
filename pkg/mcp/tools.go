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
