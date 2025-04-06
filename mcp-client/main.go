// Package main implements an example Model Context Protocol (MCP) client
//
// This client demonstrates how to implement a client for the Model Context Protocol,
// which allows communication with an AI model through a standardized JSON-RPC interface.
// The client launches a server process, communicates with it over stdin/stdout using
// JSON-RPC, and provides a clean API for making calls to the server's capabilities.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Constants for configuration - centralizing these values makes the code more maintainable
const (
	// JSON-RPC protocol version - MCP uses JSON-RPC 2.0 spec
	jsonRPCVersion = "2.0"

	// Initial request ID - we increment this for each request to match responses
	initialRequestID = 1

	// Method names - service methods available on the MCP server
	methodRandomString = "MCPService.RandomString"

	// Timeouts and delays - prevent hanging when server is unresponsive
	readTimeoutDuration     = 10 * time.Second // Maximum time to wait for a server response
	shutdownTimeoutDuration = 3 * time.Second  // Time to allow for graceful shutdown before forcing

	// Default test parameters
	defaultRandomStringLength = 20 // Default length for random string tests

	// Server path - location of the MCP server executable
	defaultServerPath = "../bin/mcp-server" // Path is relative to the client executable
)

// InitResponse represents the initialization response from the MCP server
// This is the first message received from the server after startup
type InitResponse struct {
	Name         string   `json:"name"`         // Server name
	Version      string   `json:"version"`      // Server version
	Capabilities []string `json:"capabilities"` // List of capabilities the server supports
}

// RPCRequest represents a JSON-RPC request sent to the server
// Following the JSON-RPC 2.0 specification: https://www.jsonrpc.org/specification
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"` // Must be "2.0" for JSON-RPC 2.0
	Method  string      `json:"method"`  // Method name to call on the server (e.g., "MCPService.RandomString")
	Params  interface{} `json:"params"`  // Parameters for the method call, will be marshaled to JSON
	ID      int         `json:"id"`      // Request ID, used to match responses to requests
}

// RPCResponse represents a JSON-RPC response received from the server
// Following the JSON-RPC 2.0 specification
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`         // Must be "2.0" for JSON-RPC 2.0
	Result  json.RawMessage `json:"result"`          // Raw JSON result, to be unmarshaled by the caller
	Error   *RPCError       `json:"error,omitempty"` // Error, if any
	ID      int             `json:"id"`              // Response ID, should match the request ID
}

// RPCError represents a JSON-RPC error object in a response
// Following the JSON-RPC 2.0 specification
type RPCError struct {
	Code    int    `json:"code"`    // Error code
	Message string `json:"message"` // Error message
}

// RandomStringArgs contains arguments for the RandomString method
// This demonstrates how to create a typed argument struct for a specific method
type RandomStringArgs struct {
	Length int `json:"length"` // The requested length of the random string
}

// RandomStringResponse contains the response for the RandomString method
// This demonstrates how to create a typed response struct for a specific method
type RandomStringResponse struct {
	Result string `json:"result"` // The generated random string
}

// MCPClient represents a client for the Model Context Protocol
// It manages a connection to an MCP server subprocess and provides methods
// to communicate with it via JSON-RPC over stdin/stdout
type MCPClient struct {
	cmd       *exec.Cmd      // The server subprocess
	stdin     io.WriteCloser // Pipe to server's stdin for sending requests
	stdout    *bufio.Reader  // Buffered reader for server's stdout for reading responses
	nextID    int            // Counter for generating unique request IDs
	mutex     sync.Mutex     // Mutex to protect nextID when multiple goroutines make requests
	serverCap []string       // Server capabilities cached from initialization
	logger    *log.Logger    // Logger for debug messages
}

// NewMCPClient creates a new MCP client and starts the server process
// It sets up pipes for stdin/stdout communication and initializes the client
// serverPath: path to the MCP server executable
func NewMCPClient(serverPath string) (*MCPClient, error) {
	// Create command to run the server executable
	cmd := exec.Command(serverPath)

	// Set up stdin pipe to send requests to the server
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Set up stdout pipe to receive responses from the server
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// Forward server's stderr to our stderr for debugging
	cmd.Stderr = os.Stderr

	// Start the server process
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Create and return the client instance
	return &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout), // Buffered reader for better performance
		nextID: initialRequestID,        // Start with initial request ID
		logger: log.New(os.Stdout, "DEBUG - ", log.Ldate|log.Ltime),
	}, nil
}

// Close shuts down the MCP client and server
// It attempts a graceful shutdown by closing stdin (which signals EOF to the server),
// then waits for the process to exit. If the process doesn't exit within the timeout,
// it forcefully kills it.
func (c *MCPClient) Close() error {
	// First, close stdin to signal EOF to the server
	// This is the preferred way to gracefully shut down the server
	if c.stdin != nil {
		if err := c.stdin.Close(); err != nil {
			// Log but continue with shutdown - don't return yet as we still need to clean up
			c.logger.Printf("Error closing stdin: %v", err)
		}
	}

	// Only proceed with process handling if we have a valid process
	if c.cmd != nil && c.cmd.Process != nil {
		// Wait for the process to exit naturally (with timeout)
		// We use a channel to handle the wait asynchronously
		done := make(chan error, 1)
		go func() {
			// Wait retrieves process state after completion
			// This will block until the process exits
			done <- c.cmd.Wait()
		}()

		// Give the process some time to exit gracefully
		// We use select to handle either completion or timeout
		select {
		case err := <-done:
			// Process has exited, check if it was successful
			if err != nil {
				// Process exited with an error
				if exitErr, ok := err.(*exec.ExitError); ok {
					c.logger.Printf("Server process exited with status: %v", exitErr.ExitCode())
				}
				return fmt.Errorf("server terminated with error: %w", err)
			}
			// Process exited normally
			return nil
		case <-time.After(shutdownTimeoutDuration):
			// Process didn't exit within timeout, force kill
			c.logger.Println("Server didn't exit within timeout, forcing termination")
			if err := c.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill server process: %w", err)
			}
			return fmt.Errorf("server process killed after timeout")
		}
	}

	return nil
}

// Initialize sends an initialize request to the server and processes the response
// This must be called after creating a new client and before making any RPC calls
// It sends a request and reads the response that contains server information and capabilities
func (c *MCPClient) Initialize() (*InitResponse, error) {
	// Send an initialize request to the server
	c.logger.Printf("Sending initialize request to server")
	
	// Create initialize args with client info
	initArgs := struct {
		ClientInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
	}{
		ClientInfo: struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "MCP Test Client",
			Version: "1.0.0",
		},
	}
	
	// Prepare a struct to receive the result
	var initResp InitResponse
	
	// Make the RPC call to initialize the server
	if err := c.Call("initialize", initArgs, &initResp); err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}
	
	// Send initialized notification
	c.logger.Printf("Sending initialized notification to server")
	if err := c.Call("initialized", struct{}{}, nil); err != nil {
		c.logger.Printf("Warning: failed to send initialized notification: %v", err)
		// Continue anyway, this is just a notification
	}
	
	// Store server capabilities for later reference
	// This allows us to check if a capability is available before calling it
	c.serverCap = initResp.Capabilities
	
	c.logger.Printf("Server initialized successfully with capabilities: %v", c.serverCap)
	
	return &initResp, nil
}

// Call makes an RPC call to the MCP server
// This is the core method that sends requests and receives responses
//
// Parameters:
//   - method: the RPC method to call (e.g., "MCPService.RandomString")
//   - params: the parameters to pass to the method (will be marshaled to JSON)
//   - result: a pointer to store the result (will be unmarshaled from JSON)
//
// Example usage:
//
//	var result RandomStringResponse
//	err := client.Call("MCPService.RandomString", RandomStringArgs{Length: 10}, &result)
func (c *MCPClient) Call(method string, params interface{}, result interface{}) error {
	// Get a unique ID for this request
	// We use a mutex to ensure thread safety when multiple goroutines make calls
	c.mutex.Lock()
	id := c.nextID
	c.nextID++
	c.mutex.Unlock()

	// Create the JSON-RPC request object according to the spec
	req := RPCRequest{
		JSONRPC: jsonRPCVersion, // Always "2.0" for JSON-RPC 2.0
		Method:  method,         // The method to call
		Params:  params,         // Parameters to pass to the method
		ID:      id,             // Unique ID to match response with request
	}

	// Marshal the request to JSON
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug output
	c.logger.Printf("Sending request: %s", string(reqBytes))

	// Send the request to the server via stdin
	// We append a newline to separate JSON objects in the stream
	if _, err := c.stdin.Write(append(reqBytes, '\n')); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Read the response from the server
	var rpcResp RPCResponse
	if err := c.readResponse(&rpcResp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Debug output
	c.logger.Printf("Received response: %+v", rpcResp)

	// Validate the response ID matches the request ID
	// This is crucial for ensuring we're processing the correct response
	if rpcResp.ID != id {
		return fmt.Errorf("response ID mismatch: got %d, expected %d (possible out-of-order response)",
			rpcResp.ID, id)
	}

	// Check for RPC error in the response
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error: %d - %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// Unmarshal the result into the provided result pointer
	// This converts the JSON result into the user's target struct
	if err := json.Unmarshal(rpcResp.Result, result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}

// readResponse reads a JSON-RPC response from the server with a timeout
// This is an internal helper method used by Initialize and Call
// It implements timeout handling to prevent hanging if the server doesn't respond
func (c *MCPClient) readResponse(resp *RPCResponse) error {
	// Create a channel to signal when the read operation is complete
	// We use a struct to carry both the result and any error
	// The channel is buffered to prevent goroutine leaks if the select
	// chooses the timeout case before the read completes
	responseCh := make(chan struct {
		line string
		err  error
	}, 1)

	// Use the global timeout constant for consistency
	readTimeout := readTimeoutDuration

	// Perform the read operation in a separate goroutine
	// This allows us to implement a timeout for the read operation
	go func() {
		// ReadString will block until a newline character is read
		// or there's an error (like EOF when the server terminates)
		line, err := c.stdout.ReadString('\n')
		responseCh <- struct {
			line string
			err  error
		}{line, err}
	}()

	// Wait for either the response or a timeout
	// Using select to handle multiple channel operations
	select {
	case result := <-responseCh:
		// Check for errors during the read operation
		if result.err != nil {
			// Specifically check for EOF and provide a clearer error message
			// EOF indicates the server process has terminated
			if result.err == io.EOF {
				return fmt.Errorf("server process terminated unexpectedly (EOF received)")
			}

			// Check for other connection-related errors
			// These often indicate that the server process died or was killed
			if strings.Contains(result.err.Error(), "closed") ||
				strings.Contains(result.err.Error(), "broken pipe") {
				return fmt.Errorf("connection to server closed unexpectedly: %w", result.err)
			}

			// General error case for any other read errors
			return fmt.Errorf("failed to read from stdout: %w", result.err)
		}

		// Parse the JSON response
		// This converts the raw JSON string into the RPCResponse struct
		if err := json.Unmarshal([]byte(result.line), resp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		return nil

	case <-time.After(readTimeout):
		// Timeout case - this prevents the client from hanging indefinitely
		// if the server is unresponsive
		return fmt.Errorf("timeout waiting for server response after %v", readTimeout)
	}
}

// RandomString calls the RandomString method on the MCP server
// This is a convenience wrapper around the Call method for the RandomString capability
// It handles creating the arguments and parsing the result
//
// Parameters:
//   - length: the desired length of the random string
//
// Returns:
//   - the generated random string and any error that occurred
func (c *MCPClient) RandomString(length int) (string, error) {
	// Create the arguments for the RandomString method
	args := RandomStringArgs{Length: length}

	// Prepare a struct to receive the result
	var result RandomStringResponse

	// Make the RPC call
	if err := c.Call(methodRandomString, args, &result); err != nil {
		return "", err
	}

	// Return just the string result
	return result.Result, nil
}

// HasCapability checks if the server has a specific capability
// This allows clients to check if a server capability is available before calling it
//
// Parameters:
//   - capability: the name of the capability to check for
//
// Returns:
//   - true if the server has the capability, false otherwise
func (c *MCPClient) HasCapability(capability string) bool {
	// Loop through the server capabilities
	for _, cap := range c.serverCap {
		if cap == capability {
			return true
		}
	}
	return false
}

// main demonstrates how to use the MCP client
// It provides a complete example of:
// 1. Parsing command line arguments
// 2. Creating an MCP client
// 3. Initializing the connection
// 4. Checking server capabilities
// 5. Making RPC calls
// 6. Properly cleaning up resources
func main() {
	// Define command line flags for customization
	// The -server flag allows specifying a custom server executable path
	serverPathFlag := flag.String("server", "", "Path to the MCP server executable (default: use built-in path)")
	flag.Parse()

	// Determine the server path: use command line argument if provided, otherwise use default
	serverPath := defaultServerPath
	if *serverPathFlag != "" {
		serverPath = *serverPathFlag
		fmt.Printf("Using server path from command line: %s\n", serverPath)
	} else {
		fmt.Printf("Using default server path: %s\n", serverPath)
	}

	// Create the MCP client with the selected server path
	// This launches the server subprocess and establishes pipes
	client, err := NewMCPClient(serverPath)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	// Ensure we always clean up resources, even if there's an error
	defer client.Close()

	// Process the initialization message from the server
	// This is required before making any RPC calls
	initResp, err := client.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Print information about the connected server
	fmt.Printf("Connected to MCP server:\n")
	fmt.Printf("  Name: %s\n", initResp.Name)
	fmt.Printf("  Version: %s\n", initResp.Version)
	fmt.Printf("  Capabilities: %v\n", initResp.Capabilities)

	// Test the RandomString capability if available
	// This demonstrates checking for a capability before using it
	if client.HasCapability("RandomString") {
		fmt.Println("\nTesting RandomString capability:")

		// Call the RandomString method with our default length
		randomStr, err := client.RandomString(defaultRandomStringLength)
		if err != nil {
			log.Fatalf("Failed to call RandomString: %v", err)
		}

		// Display the result
		client.logger.Printf("Random string (%d chars): %s", defaultRandomStringLength, randomStr)
	}

	// The client will be automatically closed by the defer statement above
	// This ensures proper cleanup even if there's a panic or early return
}
