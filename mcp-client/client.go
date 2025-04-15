package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"

	"sqirvy/mcp/pkg/mcp" // Use the correct module path
)

const (
	protocolVersion         = "2024-11-05" // Match the server/spec version
	clientName              = "GoMCPExampleClient"
	clientVersion           = "0.1.0"
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

	// Call Ping Tool
	if err := c.callPingTool(); err != nil {
		return err // Error already logged in callPingTool
	}

	// Read Random Data Resource
	if err := c.readRandomDataResource(); err != nil {
		return err // Error already logged in readRandomDataResource
	}
	// Get Sqirvy Query Prompt
	if err := c.getSqirvyQueryPrompt(); err != nil {
		return err // Error already logged in getSqirvyQueryPrompt
	}

	// List Tools
	if err := c.listTools(); err != nil {
		return err // Error already logged
	}

	// List Resource Templates
	if err := c.listResourceTemplates(); err != nil {
		return err // Error already logged
	}

	// List Prompts
	if err := c.listPrompts(); err != nil {
		return err // Error already logged
	}

	// Read File Resource
	if err := c.readFileResource("file:///documents/example.txt"); err != nil {
		return err // Error already logged
	}

	c.logger.Println("All client operations complete. Client will now terminate.")
	return nil // Success
}

// --- Helper Functions for MCP Calls ---

// callPingTool sends a tools/call request for the 'ping' tool and processes the response.
func (c *Client) callPingTool() error {
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

	c.logger.Println("Waiting for ping response...")
	pingResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read ping response: %v", err)
		return fmt.Errorf("failed to read ping response: %w", err)
	}
	c.logger.Printf("Received ping response JSON: %s", string(pingResponseBytes))

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

	if len(pingResult.Content) > 0 {
		var textContent mcp.TextContent
		if err := json.Unmarshal(pingResult.Content[0], &textContent); err != nil {
			c.logger.Printf("Failed to unmarshal ping result content into TextContent: %v", err)
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
	c.logger.Println("Ping tool call complete.")
	return nil
}

// readRandomDataResource sends a resources/read request for 'data://random_data' and processes the response.
func (c *Client) readRandomDataResource() error {
	readID := c.nextID()
	readParams := mcp.ReadResourceParams{
		URI: "data://random_data?length=10", // Request 10 random characters
	}
	readRequestBytes, err := mcp.MarshalReadResourcesRequest(readID, readParams)
	if err != nil {
		c.logger.Printf("Failed to marshal read resource request: %v", err)
		return fmt.Errorf("failed to marshal read resource request: %w", err)
	}

	c.logger.Printf("Sending read resource request for URI: %s", readParams.URI)
	if err := c.transport.WriteMessage(readRequestBytes); err != nil {
		c.logger.Printf("Failed to send read resource request: %v", err)
		return fmt.Errorf("failed to send read resource request: %w", err)
	}

	c.logger.Println("Waiting for read resource response...")
	readResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read resource response: %v", err)
		return fmt.Errorf("failed to read resource response: %w", err)
	}
	c.logger.Printf("Received read resource response JSON: %s", string(readResponseBytes))

	readResult, readRespID, readRPCErr, readParseErr := mcp.UnmarshalReadResourcesResponse(readResponseBytes)
	if readParseErr != nil {
		c.logger.Printf("Failed to parse read resource response: %v", readParseErr)
		return fmt.Errorf("failed to parse read resource response: %w", readParseErr)
	}
	if fmt.Sprintf("%v", readRespID) != fmt.Sprintf("%v", readID) {
		c.logger.Printf("Read resource response ID mismatch. Got: %v (%T), Want: %v (%T)", readRespID, readRespID, readID, readID)
		return fmt.Errorf("read resource response ID mismatch. Got: %v, Want: %v", readRespID, readID)
	}
	if readRPCErr != nil {
		c.logger.Printf("Received RPC error in read resource response: Code=%d, Message=%s, Data=%v", readRPCErr.Code, readRPCErr.Message, readRPCErr.Data)
		return fmt.Errorf("received RPC error in read resource response: %w", readRPCErr)
	}
	if readResult == nil {
		c.logger.Println("Read resource response contained no result.")
		return fmt.Errorf("read resource response contained no result")
	}

	if len(readResult.Contents) > 0 {
		var textContent mcp.TextResourceContents
		if err := json.Unmarshal(readResult.Contents[0], &textContent); err != nil {
			c.logger.Printf("Failed to unmarshal read resource result content into TextResourceContents: %v", err)
			c.logger.Printf("Raw read resource result content[0]: %s", string(readResult.Contents[0]))
		} else {
			if textContent.URI != readParams.URI {
				c.logger.Printf("Warning: Read resource response URI mismatch. Got: %s, Want: %s", textContent.URI, readParams.URI)
			}
			c.logger.Printf("Random data resource (%s) content:\n%s", textContent.URI, textContent.Text)
		}
	} else {
		c.logger.Println("Read resource response result contained no content.")
	}
	c.logger.Println("Read resource call complete.")
	return nil
}

// readFileResource sends a resources/read request for a given file URI and processes the response.
func (c *Client) readFileResource(fileURI string) error {
	readID := c.nextID()
	readParams := mcp.ReadResourceParams{
		URI: fileURI,
	}
	readRequestBytes, err := mcp.MarshalReadResourcesRequest(readID, readParams)
	if err != nil {
		c.logger.Printf("Failed to marshal read file resource request for %s: %v", fileURI, err)
		return fmt.Errorf("failed to marshal read file resource request: %w", err)
	}

	c.logger.Printf("Sending read resource request for URI: %s", readParams.URI)
	if err := c.transport.WriteMessage(readRequestBytes); err != nil {
		c.logger.Printf("Failed to send read file resource request: %v", err)
		return fmt.Errorf("failed to send read file resource request: %w", err)
	}

	c.logger.Println("Waiting for read file resource response...")
	readResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read file resource response: %v", err)
		return fmt.Errorf("failed to read file resource response: %w", err)
	}
	c.logger.Printf("Received read file resource response JSON: %s", string(readResponseBytes))

	readResult, readRespID, readRPCErr, readParseErr := mcp.UnmarshalReadResourcesResponse(readResponseBytes)
	if readParseErr != nil {
		c.logger.Printf("Failed to parse read file resource response: %v", readParseErr)
		return fmt.Errorf("failed to parse read file resource response: %w", readParseErr)
	}
	if fmt.Sprintf("%v", readRespID) != fmt.Sprintf("%v", readID) {
		c.logger.Printf("Read file resource response ID mismatch. Got: %v (%T), Want: %v (%T)", readRespID, readRespID, readID, readID)
		return fmt.Errorf("read file resource response ID mismatch. Got: %v, Want: %v", readRespID, readID)
	}
	if readRPCErr != nil {
		// Log the specific RPC error received from the server
		c.logger.Printf("Received RPC error in read file resource response: Code=%d, Message=%s, Data=%v", readRPCErr.Code, readRPCErr.Message, readRPCErr.Data)
		// Check if the error indicates the file wasn't found (using InvalidParams code as per server logic)
		if readRPCErr.Code == mcp.ErrorCodeInvalidParams && strings.Contains(readRPCErr.Message, "not found") {
			c.logger.Printf("Server reported file not found for URI: %s", fileURI)
			// Return nil here as the client successfully communicated, but the file doesn't exist on the server side.
			// Or return the RPC error if the client should treat this as a failure. For now, log and continue.
			return nil
		}
		// Return other RPC errors as client failures
		return fmt.Errorf("received RPC error in read file resource response: %w", readRPCErr)
	}
	if readResult == nil {
		c.logger.Println("Read file resource response contained no result.")
		return fmt.Errorf("read file resource response contained no result")
	}

	if len(readResult.Contents) > 0 {
		// Attempt to unmarshal as TextResourceContents first
		var textContent mcp.TextResourceContents
		if err := json.Unmarshal(readResult.Contents[0], &textContent); err == nil && textContent.Text != "" {
			if textContent.URI != readParams.URI {
				c.logger.Printf("Warning: Read file resource response URI mismatch. Got: %s, Want: %s", textContent.URI, readParams.URI)
			}
			c.logger.Printf("File resource (%s, Mime: %s) content:\n%s", textContent.URI, textContent.MimeType, textContent.Text)
		} else {
			// If not text, try BlobResourceContents (or log raw if neither)
			// TODO: Add BlobResourceContents handling if needed
			c.logger.Printf("Received non-text or unparseable content for file resource %s.", fileURI)
			c.logger.Printf("Raw read file resource result content[0]: %s", string(readResult.Contents[0]))
		}
	} else {
		c.logger.Printf("Read file resource response result for %s contained no content.", fileURI)
	}
	c.logger.Printf("Read file resource call for %s complete.", fileURI)
	return nil
}

// getSqirvyQueryPrompt sends a prompts/get request for 'sqirvy_query' and processes the response.
func (c *Client) getSqirvyQueryPrompt() error {
	promptID := c.nextID()
	promptParams := mcp.GetPromptParams{
		Name: "sqirvy_query",
		Arguments: map[string]string{
			"query": "What is the Model Context Protocol?", // Provide the required argument
		},
	}
	promptRequestBytes, err := mcp.MarshalGetPromptRequest(promptID, promptParams)
	if err != nil {
		c.logger.Printf("Failed to marshal get prompt request: %v", err)
		return fmt.Errorf("failed to marshal get prompt request: %w", err)
	}

	c.logger.Printf("Sending get prompt request for prompt: %s", promptParams.Name)
	if err := c.transport.WriteMessage(promptRequestBytes); err != nil {
		c.logger.Printf("Failed to send get prompt request: %v", err)
		return fmt.Errorf("failed to send get prompt request: %w", err)
	}

	c.logger.Println("Waiting for get prompt response...")
	promptResponseBytes, err := c.transport.ReadMessage()
	if err != nil {
		c.logger.Printf("Failed to read prompt response: %v", err)
		return fmt.Errorf("failed to read prompt response: %w", err)
	}
	c.logger.Printf("Received get prompt response JSON: %s", string(promptResponseBytes))

	promptResult, promptRespID, promptRPCErr, promptParseErr := mcp.UnmarshalGetPromptResponse(promptResponseBytes)
	if promptParseErr != nil {
		c.logger.Printf("Failed to parse get prompt response: %v", promptParseErr)
		return fmt.Errorf("failed to parse get prompt response: %w", promptParseErr)
	}
	if fmt.Sprintf("%v", promptRespID) != fmt.Sprintf("%v", promptID) {
		c.logger.Printf("Get prompt response ID mismatch. Got: %v (%T), Want: %v (%T)", promptRespID, promptRespID, promptID, promptID)
		return fmt.Errorf("get prompt response ID mismatch. Got: %v, Want: %v", promptRespID, promptID)
	}
	if promptRPCErr != nil {
		c.logger.Printf("Received RPC error in get prompt response: Code=%d, Message=%s, Data=%v", promptRPCErr.Code, promptRPCErr.Message, promptRPCErr.Data)
		return fmt.Errorf("received RPC error in get prompt response: %w", promptRPCErr)
	}
	if promptResult == nil {
		c.logger.Println("Get prompt response contained no result.")
		return fmt.Errorf("get prompt response contained no result")
	}

	if len(promptResult.Messages) > 0 {
		var textContent mcp.TextContent
		if err := json.Unmarshal(promptResult.Messages[0].Content, &textContent); err != nil {
			c.logger.Printf("Failed to unmarshal prompt message content into TextContent: %v", err)
			c.logger.Printf("Raw prompt message content[0]: %s", string(promptResult.Messages[0].Content))
		} else {
			c.logger.Printf("Prompt '%s' (Role: %s) content:\n%s", promptParams.Name, promptResult.Messages[0].Role, textContent.Text)
		}
	} else {
		c.logger.Println("Get prompt response result contained no messages.")
	}

	c.logger.Println("Get prompt call complete.")
	return nil
}
