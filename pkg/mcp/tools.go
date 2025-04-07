// Package mcp defines types and functions related to the Model Context Protocol (MCP).
package mcp

import (
	"encoding/json"
	"fmt"
)

// ToolsListRequest represents the request for the "tools/list" method.
// It currently has no parameters.
type ToolsListRequest struct{}

// MarshalJSON implements the json.Marshaler interface for ToolsListRequest.
// Since there are no parameters, it returns an empty JSON object "{}".
func (r *ToolsListRequest) MarshalJSON() ([]byte, error) {
	// Representing no parameters, typically an empty object or null might be used.
	// The example shows no "params" field, so we handle this during the full RPC request construction.
	// This marshaller is for the *parameters* part, which is empty here.
	return []byte("{}"), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for ToolsListRequest.
// It expects an empty JSON object "{}" or null.
func (r *ToolsListRequest) UnmarshalJSON(data []byte) error {
	// Check if the input is an empty object or null, otherwise return an error.
	s := string(data)
	if s != "{}" && s != "null" {
		return fmt.Errorf("expected empty object or null for ToolsListRequest params, got %s", s)
	}
	return nil
}

// --- ToolsList Response ---

// ToolDefinition describes a single tool available on the server.
type ToolDefinition struct {
	Name        string      `json:"name"`                  // The name of the tool.
	Description string      `json:"description,omitempty"` // Optional description of the tool.
	InputSchema interface{} `json:"inputSchema,omitempty"` // Optional JSON schema for the tool's input parameters.
}

// ToolsListResponse represents the response for the "tools/list" method.
type ToolsListResponse struct {
	Tools []ToolDefinition `json:"tools"` // A list of available tools.
}

// MarshalJSON implements the json.Marshaler interface for ToolsListResponse.
func (r *ToolsListResponse) MarshalJSON() ([]byte, error) {
	// Use a temporary struct to handle potential nil slice correctly
	type Alias ToolsListResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	// Ensure Tools is not nil for marshaling, represent empty list as []
	if aux.Tools == nil {
		aux.Tools = []ToolDefinition{}
	}
	return json.Marshal(aux)
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
	// Ensure Tools is not nil after unmarshaling if it was missing or null in JSON
	if r.Tools == nil {
		r.Tools = []ToolDefinition{}
	}
	return nil
}
