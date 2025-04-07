package mcp

import "encoding/json"

// Role defines the sender or recipient of messages and data.
type Role string

const (
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
)

// Annotations provide optional metadata for client interpretation.
type Annotations struct {
	// Audience describes the intended customer (e.g., "user", "assistant").
	Audience []Role `json:"audience,omitempty"`
	// Priority indicates importance (1=most important, 0=least important).
	Priority *float64 `json:"priority,omitempty"` // Use pointer for optional 0 value
}

// Resource represents a known resource the server can read.
type Resource struct {
	Annotations *Annotations `json:"annotations,omitempty"`
	// Description is a human-readable description of the resource.
	Description string `json:"description,omitempty"`
	// MimeType is the MIME type of the resource, if known.
	MimeType string `json:"mimeType,omitempty"`
	// Name is a human-readable name for the resource.
	Name string `json:"name"`
	// Size is the raw size in bytes, if known.
	Size *int `json:"size,omitempty"` // Use pointer for optional 0 value
	// URI is the unique identifier for the resource.
	URI string `json:"uri"`
}

// ListResourcesParams defines the parameters for a "resources/list" request.
type ListResourcesParams struct {
	// Cursor is an opaque token for pagination.
	Cursor string `json:"cursor,omitempty"`
}

// ListResourcesResult defines the result structure for a "resources/list" response.
type ListResourcesResult struct {
	// Meta contains reserved protocol metadata.
	Meta map[string]interface{} `json:"_meta,omitempty"`
	// NextCursor is an opaque token for the next page of results.
	NextCursor string `json:"nextCursor,omitempty"`
	// Resources is the list of resources found.
	Resources []Resource `json:"resources"`
}

// ReadResourceParams defines the parameters for a "resources/read" request.
type ReadResourceParams struct {
	// URI is the identifier of the resource to read.
	URI string `json:"uri"`
}

// TextResourceContents represents the text content of a resource.
type TextResourceContents struct {
	// MimeType is the MIME type of the resource, if known.
	MimeType string `json:"mimeType,omitempty"`
	// Text is the content of the resource.
	Text string `json:"text"`
	// URI is the identifier of the resource.
	URI string `json:"uri"`
}

// BlobResourceContents represents the binary content of a resource.
type BlobResourceContents struct {
	// Blob is the base64-encoded binary data.
	Blob string `json:"blob"`
	// MimeType is the MIME type of the resource, if known.
	MimeType string `json:"mimeType,omitempty"`
	// URI is the identifier of the resource.
	URI string `json:"uri"`
}

// ReadResourceResult defines the result structure for a "resources/read" response.
type ReadResourceResult struct {
	// Meta contains reserved protocol metadata.
	Meta map[string]interface{} `json:"_meta,omitempty"`
	// Contents holds the resource data, which can be text or blob.
	// Each element needs to be unmarshaled into either TextResourceContents or BlobResourceContents.
	// Example: Check for the presence of "text" or "blob" field after unmarshaling into json.RawMessage.
	Contents []json.RawMessage `json:"contents"`
}

// Note: Standard json.Marshal and json.Unmarshal can be used for these types.
// For ReadResourceResult.Contents, further processing is needed after unmarshaling
// to determine the concrete type (TextResourceContents or BlobResourceContents) of each element.
