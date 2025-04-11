package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sqirvy/mcp/pkg/mcp"
)

const (
	pingTargetIP = "192.168.5.4"
	pingTimeout  = 5 * time.Second // Timeout for the ping command
	pingToolName = "ping"
)

// pingHost executes the ping command against the specified host.
// It sends one packet (-c 1) and waits for a reply.
// Returns the combined stdout/stderr output and any execution error.
func pingHost(host string) (string, error) {
	// Use -c 1 for Linux/macOS to send only one packet
	// Use -W 1 for a 1-second wait time for the reply (adjust if needed)
	// Consider using platform-specific flags if necessary or a go ping library
	cmd := exec.Command("ping", "-c", "1", "-W", "1", host)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start ping command: %w", err)
	}

	// Wait for the command to finish or timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(pingTimeout):
		// Timeout occurred
		if err := cmd.Process.Kill(); err != nil {
			return "", fmt.Errorf("failed to kill ping process after timeout: %w", err)
		}
		return "", fmt.Errorf("ping command timed out after %v", pingTimeout)
	case err := <-done:
		// Command finished
		output := out.String() + stderr.String()
		if err != nil {
			// Ping might return non-zero exit code even if it gets output (e.g., packet loss)
			// We return the output along with the error in this case.
			return strings.TrimSpace(output), fmt.Errorf("ping command failed with exit code: %w. Output: %s", err, output)
		}
		return strings.TrimSpace(output), nil
	}
}

// handlePingTool handles the "tools/call" request specifically for the "ping" tool.
// It executes the ping command and returns the result or an error.
func (s *Server) handlePingTool(id mcp.RequestID, params mcp.CallToolParams) ([]byte, error) {
	s.logger.Printf("Handle  : tools/call request for '%s' (ID: %v)", params.Name, id)

	// Execute the ping command
	output, err := pingHost(pingTargetIP)

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
