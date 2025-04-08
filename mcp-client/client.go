package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"

	// Use the absolute module path
	"sqirvy/mcp/pkg/mcp"
)

// Client handles MCP communication from the client side.
type Client struct {
	reader        *bufio.Reader
	writer        io.Writer
	logger        *log.Logger
	mu            sync.Mutex // Protects writer access
	requestID     atomic.Int64
	pending       sync.Map // Map of RequestID -> chan *mcp.RPCResponse
	clientVersion string
}

// NewClient creates a new MCP client instance.
func NewClient(serverStdout io.Reader, serverStdin io.Writer, logger *log.Logger) *Client {
	c := &Client{
		reader:        bufio.NewReader(serverStdout),
		writer:        serverStdin,
		logger:        logger,
		clientVersion: "2024-11-05", // Client's supported protocol version
		pending:       sync.Map{},
	}
	// Start a goroutine to read responses from the server
	// go c.readLoop() // TODO: Implement read loop if client needs to handle async responses/notifications
	return c
}

// nextID generates a unique sequential request ID.
func (c *Client) nextID() mcp.RequestID {
	return c.requestID.Add(1) // Returns the new value
}

// Initialize performs the MCP initialization handshake.
func (c *Client) Initialize(protocolVersion string, clientInfo mcp.Implementation, caps mcp.ClientCapabilities) (*mcp.ServerCapabilities, error) {
	c.logger.Println("Sending initialize request...")
	id := c.nextID()
	params := mcp.InitializeParams{
		ProtocolVersion: protocolVersion,
		ClientInfo:      clientInfo,
		Capabilities:    caps,
	}

	reqBytes, err := mcp.MarshalInitializeRequest(id, params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal initialize request: %w", err)
	}

	// Send the request
	if err := c.sendRaw(reqBytes); err != nil {
		return nil, fmt.Errorf("failed to send initialize request: %w", err)
	}

	// Read the response
	respBytes, err := readMessage(c.reader, c.logger) // Using transport function directly for now
	if err != nil {
		return nil, fmt.Errorf("failed to read initialize response: %w", err)
	}

	// Unmarshal the response
	result, respID, rpcErr, parseErr := mcp.UnmarshalInitializeResponse(respBytes)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse initialize response: %w", parseErr)
	}
	if rpcErr != nil {
		return nil, fmt.Errorf("received RPC error during initialize: %w", rpcErr)
	}

	// request id's can be int64 or string, but json marshalling return float64, so fix it
	match := compareRequestIDs(id, respID)
	if !match {
		c.logger.Printf("Warning: Initialize response ID mismatch. Got %v (%T), expected %v (%T)", respID, respID, id, id)
		return nil, fmt.Errorf("initialize response ID mismatch: got %v, expected %v", respID, id)
	}

	c.logger.Printf("Received initialize response (ID: %v). Result: %+v", respID, result)

	// Send initialized notification
	if err := c.NotifyInitialized(); err != nil {
		// Log the error but don't fail initialization just because notification failed
		c.logger.Printf("Warning: Failed to send initialized notification: %v", err)
	}

	return &result.Capabilities, nil
}

// compareRequestIDs checks if the received response ID matches the expected request ID,
// handling potential type differences (e.g., int64 vs float64 from JSON).
func compareRequestIDs(id mcp.RequestID, receivedID mcp.RequestID) bool {
	var match = true
	switch rid := receivedID.(type) {
	case float64:
		// convert float to int64
		newID := mcp.RequestID(int64(rid))
		if newID != id {
			match = false
		}
	case int64:
		match = true
	case string:
		match = true
	default:
		// Type mismatch or unexpected type (e.g., nil)
		match = false
	}
	return match
}

// NotifyInitialized sends the initialized notification to the server.
func (c *Client) NotifyInitialized() error {
	c.logger.Println("Sending initialized notification...")
	// Notifications have no ID and expect no response.
	// Params can be an empty object or nil depending on the marshaller.
	// Let's use an empty struct which marshals to {}.
	params := struct{}{}
	method := "initialized" // Or "notifications/initialized" depending on server expectation

	// Manually construct the notification JSON
	// TODO: Add a MarshalNotification helper in pkg/mcp?
	notif := map[string]interface{}{
		"jsonrpc": mcp.JSONRPCVersion,
		"method":  method,
		"params":  params,
	}
	notifBytes, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized notification: %w", err)
	}

	if err := c.sendRaw(notifBytes); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}
	c.logger.Println("Initialized notification sent.")
	return nil
}

// sendRaw sends pre-marshalled bytes with headers to the server.
func (c *Client) sendRaw(payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Log the raw payload being sent *before* writing
	c.logger.Printf("Sending raw message payload (%d bytes): %s", len(payload), string(payload))

	header := fmt.Sprintf("%s: %d\r\n\r\n", headerContentLength, len(payload))

	if _, err := c.writer.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write message header: %w", err)
	}
	if _, err := c.writer.Write(payload); err != nil {
		return fmt.Errorf("failed to write message payload: %w", err)
	}

	// Flush if the writer supports it (e.g., pipe might need it)
	if flusher, ok := c.writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			c.logger.Printf("Warning: failed to flush writer after sending message: %v", err)
		}
	} else if f, ok := c.writer.(interface{ Sync() error }); ok {
		// Pipes might implement Sync
		_ = f.Sync()
	}

	return nil
}

// TODO: Implement readLoop if needed for handling asynchronous server messages
// func (c *Client) readLoop() {
// 	for {
// 		payload, err := readMessage(c.reader, c.logger)
// 		if err != nil {
// 			if err == io.EOF {
// 				c.logger.Println("Server closed connection (EOF). Exiting read loop.")
// 				// TODO: Handle server disconnect gracefully
// 				return
// 			}
// 			c.logger.Printf("Error reading message in readLoop: %v. Exiting.", err)
// 			// TODO: Handle read errors
// 			return
// 		}
//
// 		// Process the message (check if response or notification)
// 		// ... peek message type ...
// 		// ... if response, find pending channel and send ...
// 		// ... if notification, handle it ...
// 	}
// }

// Close cleans up client resources (e.g., closes writer).
// Note: Closing the writer (server's stdin) often signals the server to exit.
func (c *Client) Close() error {
	c.logger.Println("Closing client writer (server stdin)...")
	if closer, ok := c.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
