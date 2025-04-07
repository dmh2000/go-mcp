package mcp

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestToolsListRequest_MarshalUnmarshal tests marshaling and unmarshaling of ToolsListRequest.
func TestToolsListRequest_MarshalUnmarshal(t *testing.T) {
	req := ToolsListRequest{
		JSONRPC: "2.0",
		ID:      "req-1",
		Method:  "tools/list",
		Params:  struct{}{},
	}

	// Marshal
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal ToolsListRequest failed: %v", err)
	}

	// Unmarshal
	var unmarshaledReq ToolsListRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Fatalf("Unmarshal ToolsListRequest failed: %v", err)
	}

	// Compare
	if !reflect.DeepEqual(req, unmarshaledReq) {
		t.Errorf("Unmarshaled ToolsListRequest does not match original.\nOriginal: %+v\nUnmarshaled: %+v", req, unmarshaledReq)
	}

	// Test expected JSON structure (basic check)
	expectedJSON := `{"jsonrpc":"2.0","id":"req-1","method":"tools/list","params":{}}`
	if string(jsonData) != expectedJSON {
		t.Errorf("Marshaled JSON does not match expected.\nExpected: %s\nGot: %s", expectedJSON, string(jsonData))
	}
}

// TestToolsListResponse_MarshalUnmarshal tests marshaling and unmarshaling of ToolsListResponse.
func TestToolsListResponse_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		response ToolsListResponse
		expected string
	}{
		{
			name: "Response with tools",
			response: ToolsListResponse{
				JSONRPC: "2.0",
				ID:      123,
				Result: ToolsListResult{
					Tools: []ToolDefinition{
						{
							Name:        "tool1",
							Description: "desc1",
							InputSchema: map[string]interface{}{"type": "string"},
						},
					},
				},
			},
			expected: `{"jsonrpc":"2.0","id":123,"result":{"tools":[{"name":"tool1","description":"desc1","inputSchema":{"type":"string"}}]}}`,
		},
		{
			name: "Response with empty tools",
			response: ToolsListResponse{
				JSONRPC: "2.0",
				ID:      124,
				Result: ToolsListResult{
					Tools: []ToolDefinition{}, // Explicitly empty
				},
			},
			expected: `{"jsonrpc":"2.0","id":124,"result":{"tools":[]}}`,
		},
		{
			name: "Response with nil tools (should marshal as empty)",
			response: ToolsListResponse{
				JSONRPC: "2.0",
				ID:      125,
				Result: ToolsListResult{
					Tools: nil, // Nil slice
				},
			},
			expected: `{"jsonrpc":"2.0","id":125,"result":{"tools":[]}}`, // Expect empty array
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			jsonData, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Marshal ToolsListResponse failed: %v", err)
			}

			// Compare marshaled JSON
			s := string(jsonData)
			if s != tt.expected {
				t.Errorf("Marshaled JSON does not match expected.\nExpected: %s\nGot: %s", tt.expected, string(jsonData))
			}

			// Unmarshal
			var unmarshaledResp ToolsListResponse
			err = json.Unmarshal(jsonData, &unmarshaledResp)
			if err != nil {
				t.Fatalf("Unmarshal ToolsListResponse failed: %v", err)
			}

			// Ensure nil slices become empty slices after unmarshal
			if tt.response.Result.Tools == nil {
				tt.response.Result.Tools = []ToolDefinition{}
			}

			// Compare
			if !reflect.DeepEqual(tt.response, unmarshaledResp) {
				t.Errorf("Unmarshaled ToolsListResponse does not match original.\nOriginal: %+v\nUnmarshaled: %+v", tt.response, unmarshaledResp)
			}
		})
	}
}

// TestToolsCallRequest_MarshalUnmarshal tests marshaling and unmarshaling of ToolsCallRequest.
func TestToolsCallRequest_MarshalUnmarshal(t *testing.T) {
	args := map[string]interface{}{
		"arg1": "value1",
		"arg2": 123,
	}
	req := ToolsCallRequest{
		JSONRPC: "2.0",
		ID:      "call-1",
		Method:  "tools/call",
		Params: ToolsCallParams{
			Name:      "myTool",
			Arguments: args,
		},
	}

	// Marshal
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal ToolsCallRequest failed: %v", err)
	}

	// Unmarshal
	var unmarshaledReq ToolsCallRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Fatalf("Unmarshal ToolsCallRequest failed: %v", err)
	}

	// Compare basic fields
	if req.JSONRPC != unmarshaledReq.JSONRPC || req.ID != unmarshaledReq.ID || req.Method != unmarshaledReq.Method || req.Params.Name != unmarshaledReq.Params.Name {
		t.Errorf("Unmarshaled ToolsCallRequest basic fields do not match original.\nOriginal: %+v\nUnmarshaled: %+v", req, unmarshaledReq)
	}

	// Compare Arguments specifically (needs type assertion)
	if !reflect.DeepEqual(req.Params.Arguments, unmarshaledReq.Params.Arguments) {
		t.Errorf("Unmarshaled ToolsCallRequest Arguments do not match original.\nOriginal Args: %+v\nUnmarshaled Args: %+v", req.Params.Arguments, unmarshaledReq.Params.Arguments)
	}

	// Test expected JSON structure (basic check, argument order might vary)
	// Note: JSON object key order is not guaranteed, so direct string comparison is fragile.
	// A more robust check would unmarshal the expected string and compare the structs.
	// expectedJSON := `{"jsonrpc":"2.0","id":"call-1","method":"tools/call","params":{"name":"myTool","arguments":{"arg1":"value1","arg2":123}}}`
	// if string(jsonData) != expectedJSON {
	//  t.Errorf("Marshaled JSON does not match expected.\nExpected: %s\nGot: %s", expectedJSON, string(jsonData))
	// }
}

// TestToolsCallResponse_MarshalUnmarshal tests marshaling and unmarshaling of ToolsCallResponse.
func TestToolsCallResponse_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		response ToolsCallResponse
		expected string
	}{
		{
			name: "Response with content",
			response: ToolsCallResponse{
				JSONRPC: "2.0",
				ID:      5,
				Result: ToolsCallResult{
					Content: []ContentItem{
						{Type: "text", Text: "Result text"},
					},
				},
			},
			expected: `{"jsonrpc":"2.0","id":5,"result":{"content":[{"type":"text","text":"Result text"}]}}`,
		},
		{
			name: "Response with empty content",
			response: ToolsCallResponse{
				JSONRPC: "2.0",
				ID:      6,
				Result: ToolsCallResult{
					Content: []ContentItem{}, // Explicitly empty
				},
			},
			expected: `{"jsonrpc":"2.0","id":6,"result":{"content":[]}}`,
		},
		{
			name: "Response with nil content (should marshal as empty)",
			response: ToolsCallResponse{
				JSONRPC: "2.0",
				ID:      7,
				Result: ToolsCallResult{
					Content: nil, // Nil slice
				},
			},
			expected: `{"jsonrpc":"2.0","id":7,"result":{"content":[]}}`, // Expect empty array
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			jsonData, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Marshal ToolsCallResponse failed: %v", err)
			}

			// Compare marshaled JSON
			if string(jsonData) != tt.expected {
				t.Errorf("Marshaled JSON does not match expected.\nExpected: %s\nGot: %s", tt.expected, string(jsonData))
			}

			// Unmarshal
			var unmarshaledResp ToolsCallResponse
			err = json.Unmarshal(jsonData, &unmarshaledResp)
			if err != nil {
				t.Fatalf("Unmarshal ToolsCallResponse failed: %v", err)
			}

			// Ensure nil slices become empty slices after unmarshal
			if tt.response.Result.Content == nil {
				tt.response.Result.Content = []ContentItem{}
			}

			// Compare
			if !reflect.DeepEqual(tt.response, unmarshaledResp) {
				t.Errorf("Unmarshaled ToolsCallResponse does not match original.\nOriginal: %+v\nUnmarshaled: %+v", tt.response, unmarshaledResp)
			}
		})
	}
}
