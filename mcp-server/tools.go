package main

import (
	"encoding/json"
	"fmt"
	"time"

	ping "sqirvy/mcp/mcp-server/tools"
	"sqirvy/mcp/pkg/mcp"
)

const (
	pingTargetIP = "192.168.5.4"
	pingTimeout  = 5 * time.Second // Timeout for the ping command
	pingToolName = "ping"
)

// handlePingTool handles the "tools/call" request specifically for the "ping" tool.
// It executes the ping command and returns the result or an error.
func (s *Server) handlePingTool(id mcp.RequestID, params mcp.CallToolParams) ([]byte, error) {
	s.logger.Printf("Handle  : tools/call request for '%s' (ID: %v)", params.Name, id)

	// Execute the ping command
	output, err := ping.PingHost(pingTargetIP, pingTimeout)

	var result mcp.CallToolResult
	var content mcp.TextContent

	if err != nil {
		s.logger.Printf("Error executing ping to %s: %v", pingTargetIP, err)
		// Ping failed, return the error message in the content
		content = mcp.TextContent{
			Type: "text",
			Text: fmt.Sprintf("Error pinging %s: %v", pingTargetIP, err),
		}
		result.IsError = true // Indicate it's a tool-level error
	} else {
		s.logger.Printf("Ping to %s successful. Output:\n%s", pingTargetIP, output)
		content = mcp.TextContent{
			Type: "text",
			Text: output,
		}
		result.IsError = false
	}

	// Marshal the content into json.RawMessage
	contentBytes, marshalErr := json.Marshal(content)
	if marshalErr != nil {
		err = fmt.Errorf("failed to marshal ping result content: %w", marshalErr)
		s.logger.Println(err.Error())
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInternalError, err.Error(), nil)
		return s.marshalErrorResponse(id, rpcErr) // Return marshalled JSON-RPC error
	}

	result.Content = []json.RawMessage{json.RawMessage(contentBytes)}

	// Marshal the successful (or tool-error) CallToolResult response
	return s.marshalResponse(id, result)
}
