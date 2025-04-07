// Package mcp defines types and functions related to the Model Context Protocol (MCP).
package mcp

import "encoding/json"

// --- ResourcesList Request ---

// ResourcesListRequest represents the full JSON-RPC request for the "resources/list" method.
type ResourcesListRequest struct {
	JSONRPC string      `json:"jsonrpc"` // Should be "2.0"
	ID      interface{} `json:"id"`      // Request ID (string or number)
	Method  string      `json:"method"`  // Should be "resources/list"
	Params  struct{}    `json:"params"`  // Empty params object for resources/list
}

// MarshalJSON implements the json.Marshaler interface for ResourcesListRequest.
func (r *ResourcesListRequest) MarshalJSON() ([]byte, error) {
	type Alias ResourcesListRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for ResourcesListRequest.
func (r *ResourcesListRequest) UnmarshalJSON(data []byte) error {
	type Alias ResourcesListRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}

// --- ResourcesList Response ---

// ResourceDefinition describes a single resource available on the server.
type ResourceDefinition struct {
	URI          string `json:"uri"`                    // URI of the resource (e.g., file://, http://).
	Name         string `json:"name,omitempty"`         // Optional human-readable name.
	MimeType     string `json:"mimeType,omitempty"`     // Optional MIME type of the resource content.
	LastModified string `json:"lastModified,omitempty"` // Optional last modified timestamp (RFC3339 format).
}

// ResourcesListResult holds the actual result data for the "resources/list" response.
type ResourcesListResult struct {
	Resources []ResourceDefinition `json:"resources"` // A list of available resources.
}

// ResourcesListResponse represents the full JSON-RPC response for the "resources/list" method.
type ResourcesListResponse struct {
	JSONRPC string              `json:"jsonrpc"`         // Should be "2.0"
	ID      interface{}         `json:"id"`              // Response ID (matches request ID)
	Result  ResourcesListResult `json:"result"`          // The actual result payload
	Error   *interface{}        `json:"error,omitempty"` // Error object, if any
}

// MarshalJSON implements the json.Marshaler interface for ResourcesListResponse.
func (r *ResourcesListResponse) MarshalJSON() ([]byte, error) {
	type Alias ResourcesListResponse
	// Ensure Resources slice in Result is not nil before marshaling
	if r.Result.Resources == nil {
		r.Result.Resources = []ResourceDefinition{}
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for ResourcesListResponse.
func (r *ResourcesListResponse) UnmarshalJSON(data []byte) error {
	type Alias ResourcesListResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Ensure Resources slice in Result is not nil after unmarshaling
	if r.Result.Resources == nil {
		r.Result.Resources = []ResourceDefinition{}
	}
	return nil
}
