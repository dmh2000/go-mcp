package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"

	"sqirvy/mcp/pkg/mcp" // Use the correct module path
)

const (
	protocolVersion       = "2024-11-05" // Match the server/spec version
	clientName            = "GoMCPExampleClient"
	clientVersion         = "0.1.0"
	notificationInitialized = "initialized" // Method name for the initialized notification
)

// Client handles the MCP client logic.
type Client struct {
	transport *StdioTransport
	logger    *log.Logger
	requestID atomic.Int64 // Safely incrementing request ID
}

// NewClient creates a new MCP client instance.
func NewClient(transport *StdioTransport, logger *log.Logger) *Client {
	return &Client{
		transport: transport,
		logger:    logger,
	}
}

// nextID generates the next request ID.
func (c *Client) nextID() int64 {
	return c.requestID.Add(1)
}

// Run performs the initial MCP handshake: initialize -> initialized notification.
func (c *Client) Run() error {
	defer c.transport.Close() // Ensure transport is closed when Run finishes

	// 1. Send Initialize Request
	initID := c.nextID()
	initParams := mcp.InitializeParams{
		ProtocolVersion: protocolVersion,
		ClientInfo: mcp.Implementation{
			Name:    clientName,
			Version: clientVersion,
		},
		Capabilities: mcp.ClientCapabilities{
			// Define any specific client capabilities here if needed
			// Example:
			// Roots: &struct { ListChanged bool `json:"listChanged,omitempty"` }{ListChanged: true},
		},
	}

	initRequestBytes, err := mcp.MarshalInitializeRequest(initID, initParams)
	if err != nil {
		c.logger.Printf("Failed to marshal initialize request: %v", err)
		return fmt.Errorf("failed to marshal initialize request: %w", err)
	}

	c.logger.Println("Sending initialize request...")
	if err := c.transport.WriteMessage(initRequestBytes); err != nil {
		c.logger.Printf("Failed to send initialize request: %v", err)
		return fmt.Errorf("failed to send initialize request: %w", err)
	}

	// 2. Wait for Initialize Response
	c.logger.Println("Waiting for initialize response...")
	initResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read initialize response: %v", err)
		return fmt.Errorf("failed to read initialize response: %w", err)
	}
	c.logger.Printf("Received initialize response JSON: %s", string(initResponseBytes)) // Log the raw JSON

	// 3. Process Initialize Response
	initResult, respID, rpcErr, parseErr := mcp.UnmarshalInitializeResponse(initResponseBytes)
	if parseErr != nil {
		c.logger.Printf("Failed to parse initialize response: %v", parseErr)
		return fmt.Errorf("failed to parse initialize response: %w", parseErr)
	}
	// Basic ID check (type might differ float64 vs int64, so compare values)
	if fmt.Sprintf("%v", respID) != fmt.Sprintf("%v", initID) {
		c.logger.Printf("Initialize response ID mismatch. Got: %v (%T), Want: %v (%T)", respID, respID, initID, initID)
		return fmt.Errorf("initialize response ID mismatch. Got: %v, Want: %v", respID, initID)
	}
	if rpcErr != nil {
		c.logger.Printf("Received RPC error in initialize response: Code=%d, Message=%s, Data=%v", rpcErr.Code, rpcErr.Message, rpcErr.Data)
		return fmt.Errorf("received RPC error in initialize response: %w", rpcErr)
	}
	if initResult == nil {
		c.logger.Println("Initialize response contained no result.")
		return fmt.Errorf("initialize response contained no result")
	}

	c.logger.Printf("Server initialized successfully. ProtocolVersion: %s", initResult.ProtocolVersion)
	c.logger.Printf("Server Info: Name=%s, Version=%s", initResult.ServerInfo.Name, initResult.ServerInfo.Version)
	// Log capabilities (consider pretty printing if complex)
	capsBytes, _ := json.MarshalIndent(initResult.Capabilities, "", "  ")
	c.logger.Printf("Server Capabilities:\n%s", string(capsBytes))

	// 4. Send Initialized Notification
	// Notifications have no ID.
	initializedNotification := mcp.RPCRequest{
		JSONRPC: mcp.JSONRPCVersion,
		Method:  notificationInitialized,
		Params:  map[string]interface{}{}, // Empty params object as per spec
		// ID field is omitted for notifications
	}

	initializedBytes, err := json.Marshal(initializedNotification)
	if err != nil {
		c.logger.Printf("Failed to marshal initialized notification: %v", err)
		return fmt.Errorf("failed to marshal initialized notification: %w", err)
	}

	c.logger.Println("Sending initialized notification...")
	if err := c.transport.WriteMessage(initializedBytes); err != nil {
		c.logger.Printf("Failed to send initialized notification: %v", err)
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}
	c.logger.Println("MCP handshake complete.")

	// 5. Call the 'ping' tool
	pingID := c.nextID()
	pingParams := mcp.CallToolParams{
		Name: "ping",
		// No arguments needed for this specific ping tool
	}
	pingRequestBytes, err := mcp.MarshalCallToolRequest(pingID, pingParams)
	if err != nil {
		c.logger.Printf("Failed to marshal ping request: %v", err)
		return fmt.Errorf("failed to marshal ping request: %w", err)
	}

	c.logger.Println("Sending ping tool request...")
	if err := c.transport.WriteMessage(pingRequestBytes); err != nil {
		c.logger.Printf("Failed to send ping request: %v", err)
		return fmt.Errorf("failed to send ping request: %w", err)
	}

	// 6. Wait for Ping Response
	c.logger.Println("Waiting for ping response...")
	pingResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read ping response: %v", err)
		return fmt.Errorf("failed to read ping response: %w", err)
	}
	c.logger.Printf("Received ping response JSON: %s", string(pingResponseBytes))

	// 7. Process Ping Response
	pingResult, pingRespID, pingRPCErr, pingParseErr := mcp.UnmarshalCallToolResponse(pingResponseBytes)
	if pingParseErr != nil {
		c.logger.Printf("Failed to parse ping response: %v", pingParseErr)
		return fmt.Errorf("failed to parse ping response: %w", pingParseErr)
	}
	if fmt.Sprintf("%v", pingRespID) != fmt.Sprintf("%v", pingID) {
		c.logger.Printf("Ping response ID mismatch. Got: %v (%T), Want: %v (%T)", pingRespID, pingRespID, pingID, pingID)
		return fmt.Errorf("ping response ID mismatch. Got: %v, Want: %v", pingRespID, pingID)
	}
	if pingRPCErr != nil {
		c.logger.Printf("Received RPC error in ping response: Code=%d, Message=%s, Data=%v", pingRPCErr.Code, pingRPCErr.Message, pingRPCErr.Data)
		return fmt.Errorf("received RPC error in ping response: %w", pingRPCErr)
	}
	if pingResult == nil {
		c.logger.Println("Ping response contained no result.")
		return fmt.Errorf("ping response contained no result")
	}

	// 8. Log Ping Output from Content
	if len(pingResult.Content) > 0 {
		// Assuming the first content item is the text output
		var textContent mcp.TextContent
		if err := json.Unmarshal(pingResult.Content[0], &textContent); err != nil {
			c.logger.Printf("Failed to unmarshal ping result content into TextContent: %v", err)
			// Log the raw content as fallback
			c.logger.Printf("Raw ping result content[0]: %s", string(pingResult.Content[0]))
		} else {
			if pingResult.IsError {
				c.logger.Printf("Ping tool reported an error: %s", textContent.Text)
			} else {
				c.logger.Printf("Ping tool output:\n%s", textContent.Text)
			}
		}
	} else {
		c.logger.Println("Ping response result contained no content.")
	}

	c.logger.Println("Ping tool call complete. Client will now terminate.")
	return nil // Success
}
