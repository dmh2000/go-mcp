package mcp

import "encoding/json"

// Method names for tool operations.
const (
	MethodListTools = "tools/list"
	MethodCallTool  = "tools/call"
)

// ToolInputSchema defines the expected parameters for a tool, represented as a JSON Schema object.
// Using map[string]interface{} for flexibility, but could be a more specific struct if the schema structure is fixed.
type ToolInputSchema map[string]interface{}

// Tool defines a tool the client can call.
type Tool struct {
	// Description is a human-readable description of the tool.
	Description string `json:"description,omitempty"`
	// InputSchema is a JSON Schema object defining the expected parameters.
	InputSchema ToolInputSchema `json:"inputSchema"`
	// Name is the name of the tool.
	Name string `json:"name"`
}

// ListToolsParams defines the parameters for a "tools/list" request.
type ListToolsParams struct {
	// Cursor is an opaque token for pagination.
	Cursor string `json:"cursor,omitempty"`
}

// ListToolsResult defines the result structure for a "tools/list" response.
type ListToolsResult struct {
	// Meta contains reserved protocol metadata.
	Meta map[string]interface{} `json:"_meta,omitempty"`
	// NextCursor is an opaque token for the next page of results.
	NextCursor string `json:"nextCursor,omitempty"`
	// Tools is the list of tools found.
	Tools []Tool `json:"tools"`
}

// CallToolParams defines the parameters for a "tools/call" request.
type CallToolParams struct {
	// Arguments are the parameters to pass to the tool.
	// Using map[string]interface{} for flexibility as argument types can vary.
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	// Name is the name of the tool to call.
	Name string `json:"name"`
}

// EmbeddedResource represents resource contents embedded in a message.
// Note: Duplicated from prompts.go, consider consolidating.
type EmbeddedResource struct {
	Annotations *Annotations    `json:"annotations,omitempty"`
	Resource    json.RawMessage `json:"resource"` // Can be TextResourceContents or BlobResourceContents
	Type        string          `json:"type"`     // Should be "resource"
}

// CallToolResult defines the result structure for a "tools/call" response.
type CallToolResult struct {
	// Meta contains reserved protocol metadata.
	Meta map[string]interface{} `json:"_meta,omitempty"`
	// Content holds the tool's output data (TextContent, ImageContent, or EmbeddedResource).
	// Each element needs to be unmarshaled into the specific type based on the "type" field
	// after initial unmarshaling into json.RawMessage.
	Content []json.RawMessage `json:"content"`
	// IsError indicates if the tool call resulted in an error. Defaults to false.
	IsError bool `json:"isError,omitempty"`
}

// Note: Standard json.Marshal and json.Unmarshal can be used for these types.
// For CallToolResult.Content and EmbeddedResource.Resource, further processing is needed after unmarshaling
// to determine the concrete type.
