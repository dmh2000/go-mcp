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

// --- ResourcesRead Request ---

// ResourcesReadRequest represents the full JSON-RPC request for the "resources/read" method.
// Note: The 'uri' field is at the top level, not within 'params', based on the example.
type ResourcesReadRequest struct {
	JSONRPC string      `json:"jsonrpc"` // Should be "2.0"
	ID      interface{} `json:"id"`      // Request ID (string or number)
	Method  string      `json:"method"`  // Should be "resources/read"
	URI     string      `json:"uri"`     // URI of the resource to read.
}

// MarshalJSON implements the json.Marshaler interface for ResourcesReadRequest.
func (r *ResourcesReadRequest) MarshalJSON() ([]byte, error) {
	type Alias ResourcesReadRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for ResourcesReadRequest.
func (r *ResourcesReadRequest) UnmarshalJSON(data []byte) error {
	type Alias ResourcesReadRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}

// --- ResourcesRead Response ---

// ResourceContent represents the content of a single resource read.
type ResourceContent struct {
	URI      string  `json:"uri"`                // URI of the resource.
	MimeType string  `json:"mimeType,omitempty"` // Optional MIME type.
	Text     *string `json:"text"`               // Text content (use pointer for null).
	Base64   *string `json:"base64"`             // Base64 encoded content (use pointer for null).
}

// ResourcesReadResponse represents the full JSON-RPC response for the "resources/read" method.
// Note: The 'contents' field is at the top level, not within 'result', based on the example.
type ResourcesReadResponse struct {
	JSONRPC  string            `json:"jsonrpc"`         // Should be "2.0"
	ID       interface{}       `json:"id"`              // Response ID (matches request ID)
	Contents []ResourceContent `json:"contents"`        // List of resource contents read.
	Error    *interface{}      `json:"error,omitempty"` // Error object, if any
}

// MarshalJSON implements the json.Marshaler interface for ResourcesReadResponse.
func (r *ResourcesReadResponse) MarshalJSON() ([]byte, error) {
	type Alias ResourcesReadResponse
	// Ensure Contents slice is not nil before marshaling
	if r.Contents == nil {
		r.Contents = []ResourceContent{}
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for ResourcesReadResponse.
func (r *ResourcesReadResponse) UnmarshalJSON(data []byte) error {
	type Alias ResourcesReadResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Ensure Contents slice is not nil after unmarshaling
	if r.Contents == nil {
		r.Contents = []ResourceContent{}
	}
	return nil
}
