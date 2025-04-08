package main

import (
	"encoding/json"
	"fmt"

	// Keep log import
	// Use the absolute module path
	"sqirvy/mcp/pkg/mcp"
)

// --- Initialization Handler ---

// handleInitializeRequest handles the "initialize" request.
// It validates the request, performs capability negotiation (currently basic),
// and returns the marshalled InitializeResult response bytes or marshalled error response bytes.
func (s *Server) handleInitializeRequest(id mcp.RequestID, payload []byte) ([]byte, error) {
	var req mcp.RPCRequest // Use the base request type first
	if err := json.Unmarshal(payload, &req); err != nil {
		err = fmt.Errorf("failed to unmarshal base initialize request structure: %w", err)
		s.logger.Println(err.Error())
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeParseError, err.Error(), nil)
		// Marshal and return the error response bytes
		errorBytes, marshalErr := s.marshalErrorResponse(id, rpcErr)
		if marshalErr != nil {
			return nil, marshalErr // Return marshalling error if that fails too
		}
		return errorBytes, err // Return marshalled error and the original error
	}

	// Check if Params field is present and is a valid JSON object/array
	if req.Params == nil {
		err := fmt.Errorf("initialize request missing 'params' field")
		s.logger.Println(err.Error())
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidRequest, err.Error(), nil)
		errorBytes, marshalErr := s.marshalErrorResponse(id, rpcErr)
		if marshalErr != nil {
			return nil, marshalErr
		}
		return errorBytes, err
	}

	// Ensure req.Params is json.RawMessage before unmarshalling into specific type
	paramsRaw, ok := req.Params.(json.RawMessage)
	if !ok {
		// This might happen if params is not an object/array in the JSON
		// Attempt to remarshal and then treat as RawMessage if needed, or handle error
		tempParamsBytes, err := json.Marshal(req.Params)
		if err != nil {
			err = fmt.Errorf("initialize request 'params' field is not a valid JSON object/array (marshal check failed): %w", err)
			s.logger.Println(err.Error())
			rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidParams, err.Error(), nil)
			errorBytes, marshalErr := s.marshalErrorResponse(id, rpcErr)
			if marshalErr != nil {
				return nil, marshalErr
			}
			return errorBytes, err
		}
		paramsRaw = json.RawMessage(tempParamsBytes)
		s.logger.Printf("Initialize params were not RawMessage initially, re-marshalled: %s", string(paramsRaw))
	}

	// Now unmarshal params specifically into InitializeParams
	var params mcp.InitializeParams
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		err = fmt.Errorf("failed to unmarshal initialize params object: %w", err)
		s.logger.Println(err.Error())
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidParams, err.Error(), nil)
		errorBytes, marshalErr := s.marshalErrorResponse(id, rpcErr)
		if marshalErr != nil {
			return nil, marshalErr
		}
		return errorBytes, err
	}

	s.logger.Printf("Received Initialize Request (ID: %v): ClientInfo=%+v, ProtocolVersion=%s, Caps=%+v",
		id, params.ClientInfo, params.ProtocolVersion, params.Capabilities)

	// --- Capability Negotiation (Basic Example) ---
	if params.ProtocolVersion == "" {
		err := fmt.Errorf("client initialize request missing protocolVersion")
		s.logger.Println(err.Error())
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidParams, err.Error(), nil)
		errorBytes, marshalErr := s.marshalErrorResponse(id, rpcErr)
		if marshalErr != nil {
			return nil, marshalErr
		}
		return errorBytes, err
	}
	// Basic check: Log if client version differs, but proceed using our version.
	if params.ProtocolVersion != s.serverVersion {
		s.logger.Printf("Client requested protocol version '%s', server using '%s'", params.ProtocolVersion, s.serverVersion)
	}
	// TODO: Add more robust version negotiation if needed.
	// TODO: Inspect params.Capabilities and potentially enable/disable server features.

	// --- Prepare Response ---
	result := mcp.InitializeResult{
		ProtocolVersion: s.serverVersion,
		ServerInfo:      s.serverInfo,
		Capabilities:    mcp.ServerCapabilities{
			// Explicitly state no capabilities initially.
			// To enable later, uncomment and potentially configure:
			// Logging:   map[string]interface{}{}, // Empty object indicates basic support
			// Prompts:   &mcp.ServerCapabilitiesPrompts{ListChanged: false}, // Use pointer struct from schema/types
			// Resources: &mcp.ServerCapabilitiesResources{ListChanged: false, Subscribe: false}, // Use pointer struct
			// Tools:     &mcp.ServerCapabilitiesTools{ListChanged: false}, // Use pointer struct
		},
		Instructions: "Welcome to the Go MCP Example Server! Currently, no tools, prompts, or resources are enabled.", // Optional
	}

	// Marshal the successful response using the server's helper
	responseBytes, err := s.marshalResponse(id, result)
	if err != nil {
		// marshalResponse already logged the error and returns marshalled error bytes
		return responseBytes, err // Return the error bytes and the original marshalling error
	}

	s.logger.Printf("Prepared Initialize Response (ID: %v): ServerInfo=%+v, ProtocolVersion=%s, Caps=%+v",
		id, result.ServerInfo, result.ProtocolVersion, result.Capabilities)

	return responseBytes, nil // Return success response bytes and nil error
}

// --- Handlers for other methods ---
// These handlers now return the marshalled response/error bytes and any error encountered during marshalling.
// They no longer call sendResponse/sendErrorResponse directly.

func (s *Server) handleListTools(id mcp.RequestID, payload []byte) ([]byte, error) {
	s.logger.Printf("Handling tools/list request (ID: %v)", id)
	// TODO: Implement actual tool listing logic if/when tools are added.
	// For now, return empty list.
	result := mcp.ListToolsResult{
		Tools: []mcp.Tool{},
		// NextCursor: "", // Omit if no pagination needed yet
	}
	// Marshal the success response
	return s.marshalResponse(id, result)
}

func (s *Server) handleCallTool(id mcp.RequestID, payload []byte) ([]byte, error) {
	s.logger.Printf("Handling tools/call request (ID: %v) - Not Implemented", id)
	// TODO: Implement tool calling logic later.
	rpcErr := mcp.NewRPCError(mcp.ErrorCodeMethodNotFound, "Method 'tools/call' not implemented", nil)
	// Marshal the error response
	return s.marshalErrorResponse(id, rpcErr)
}

func (s *Server) handleListPrompts(id mcp.RequestID, payload []byte) ([]byte, error) {
	s.logger.Printf("Handling prompts/list request (ID: %v)", id)
	// TODO: Implement actual prompt listing logic.
	result := mcp.ListPromptsResult{
		Prompts: []mcp.Prompt{},
		// NextCursor: "",
	}
	return s.marshalResponse(id, result)
}

func (s *Server) handleGetPrompt(id mcp.RequestID, payload []byte) ([]byte, error) {
	s.logger.Printf("Handling prompts/get request (ID: %v) - Not Implemented", id)
	// TODO: Implement prompt retrieval logic.
	rpcErr := mcp.NewRPCError(mcp.ErrorCodeMethodNotFound, "Method 'prompts/get' not implemented", nil)
	return s.marshalErrorResponse(id, rpcErr)
}

func (s *Server) handleListResources(id mcp.RequestID, payload []byte) ([]byte, error) {
	s.logger.Printf("Handling resources/list request (ID: %v)", id)
	// TODO: Implement actual resource listing logic.
	result := mcp.ListResourcesResult{
		Resources: []mcp.Resource{},
		// NextCursor: "",
	}
	return s.marshalResponse(id, result)
}

func (s *Server) handleReadResource(id mcp.RequestID, payload []byte) ([]byte, error) {
	s.logger.Printf("Handling resources/read request (ID: %v) - Not Implemented", id)
	// TODO: Implement resource reading logic.
	rpcErr := mcp.NewRPCError(mcp.ErrorCodeMethodNotFound, "Method 'resources/read' not implemented", nil)
	return s.marshalErrorResponse(id, rpcErr)
}

// --- Helper Struct (Remove if MarshalInitializeResponse moves to pkg/mcp) ---
// MCPPackageHelper is a dummy struct to hang the MarshalInitializeResponse method on,
// simulating it being part of the mcp package. Remove this if that function is moved.
type MCPPackageHelper struct{}

// MarshalInitializeResponse is a helper to create the full RPCResponse structure
// for an InitializeResult. This function should ideally be in pkg/mcp but is here
// for simplicity as we are not modifying pkg/mcp directly.
// NOTE: This specific marshaller is actually not needed anymore as the generic
// s.marshalResponse() in server.go handles this case. Kept for history, can be removed.
// func (m *MCPPackageHelper) MarshalInitializeResponse(id mcp.RequestID, result mcp.InitializeResult) ([]byte, error) {
// 	resultBytes, err := json.Marshal(result)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal InitializeResult: %w", err)
// 	}
// 	resp := mcp.RPCResponse{
// 		JSONRPC: mcp.JSONRPCVersion,
// 		Result:  resultBytes,
// 		ID:      id,
// 	}
// 	return json.Marshal(resp)
// }
