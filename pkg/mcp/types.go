package mcp

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// JSONRPCVersion is the fixed JSON-RPC version string.
const JSONRPCVersion = "2.0"

// RequestID represents the ID field in a JSON-RPC request/response, which can be a string or number.
type RequestID interface{}

// RPCRequest defines the structure for a JSON-RPC request.
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      RequestID   `json:"id"`
}

// RPCResponse defines the structure for a JSON-RPC response.
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      RequestID       `json:"id"`
}

// RPCError defines the structure for a JSON-RPC error object.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface for RPCError.
func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

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

// jsonEqual compares two byte slices containing JSON, ignoring whitespace differences.
// Useful for comparing marshaled JSON in tests.
func jsonEqual(a, b []byte) (bool, error) {
	var j1, j2 interface{}
	if err := json.Unmarshal(a, &j1); err != nil {
		return false, fmt.Errorf("failed to unmarshal first JSON: %w", err)
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, fmt.Errorf("failed to unmarshal second JSON: %w", err)
	}
	return reflect.DeepEqual(j1, j2), nil
}
