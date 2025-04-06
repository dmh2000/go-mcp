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

// Constants for configuration
const (
	// JSON-RPC protocol version
	jsonRPCVersion = "2.0"

	// Initial request ID
	initialRequestID = 1

	// Method names
	methodRandomString = "MCPService.RandomString"

	// Timeouts and delays
	readTimeoutDuration     = 10 * time.Second
	shutdownTimeoutDuration = 3 * time.Second

	// Default test parameters
	defaultRandomStringLength = 20

	// Server path
	defaultServerPath = "../bin/mcp-server"
)

// InitResponse represents the initialization response from the MCP server
type InitResponse struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

// RPCResponse represents a JSON-RPC response
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RandomStringArgs contains arguments for the RandomString method
type RandomStringArgs struct {
	Length int `json:"length"`
}

// RandomStringResponse contains the response for the RandomString method
type RandomStringResponse struct {
	Result string `json:"result"`
}

// MCPClient represents a client for the Model Context Protocol
type MCPClient struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    *bufio.Reader
	nextID    int
	mutex     sync.Mutex
	serverCap []string
}

// NewMCPClient creates a new MCP client and starts the server
func NewMCPClient(serverPath string) (*MCPClient, error) {
	cmd := exec.Command(serverPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr // Forward server's stderr to our stderr

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	return &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		nextID: initialRequestID,
	}, nil
}

// Close shuts down the MCP client and server
func (c *MCPClient) Close() error {
	// First, close stdin to signal EOF to the server
	if c.stdin != nil {
		if err := c.stdin.Close(); err != nil {
			// Log but continue with shutdown
			fmt.Fprintf(os.Stderr, "Error closing stdin: %v\n", err)
		}
	}

	// Only force kill if we have a process and no other option
	if c.cmd != nil && c.cmd.Process != nil {
		// Wait for the process to exit naturally (with timeout)
		done := make(chan error, 1)
		go func() {
			// Wait retrieves process state after completion
			done <- c.cmd.Wait()
		}()

		// Give the process some time to exit gracefully
		select {
		case err := <-done:
			if err != nil {
				// Process exited with an error
				if exitErr, ok := err.(*exec.ExitError); ok {
					fmt.Fprintf(os.Stderr, "Server process exited with status: %v\n", exitErr.ExitCode())
				}
				return fmt.Errorf("server terminated with error: %w", err)
			}
			// Process exited normally
			return nil
		case <-time.After(shutdownTimeoutDuration):
			// Process didn't exit within timeout, force kill
			fmt.Fprintf(os.Stderr, "Server didn't exit within timeout, forcing termination\n")
			if err := c.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill server process: %w", err)
			}
			return fmt.Errorf("server process killed after timeout")
		}
	}

	return nil
}

// Initialize processes the initialization message from the server
func (c *MCPClient) Initialize() (*InitResponse, error) {
	// Read the initialization message from the server
	var rpcResp RPCResponse
	if err := c.readResponse(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to read initialization response: %w", err)
	}

	// Parse the initialization response
	var initResp InitResponse
	if err := json.Unmarshal(rpcResp.Result, &initResp); err != nil {
		return nil, fmt.Errorf("failed to parse initialization response: %w", err)
	}

	// Store server capabilities
	c.serverCap = initResp.Capabilities

	return &initResp, nil
}

// Call makes an RPC call to the MCP server
func (c *MCPClient) Call(method string, params interface{}, result interface{}) error {
	c.mutex.Lock()
	id := c.nextID
	c.nextID++
	c.mutex.Unlock()

	// Create the RPC request
	req := RPCRequest{
		JSONRPC: jsonRPCVersion,
		Method:  method,
		Params:  params,
		ID:      id,
	}

	// Marshal and send the request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Fprintf(os.Stderr, "DEBUG - Sending request: %s\n", string(reqBytes))

	if _, err := c.stdin.Write(append(reqBytes, '\n')); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Read the response
	var rpcResp RPCResponse
	if err := c.readResponse(&rpcResp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Fprintf(os.Stderr, "DEBUG - Received response: %+v\n", rpcResp)

	// Validate the response ID matches the request ID
	if rpcResp.ID != id {
		return fmt.Errorf("response ID mismatch: got %d, expected %d (possible out-of-order response)",
			rpcResp.ID, id)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error: %d - %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// Unmarshal the result
	if err := json.Unmarshal(rpcResp.Result, result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}

// readResponse reads a JSON-RPC response from the server with a timeout
func (c *MCPClient) readResponse(resp *RPCResponse) error {
	// Create a channel to signal when the read operation is complete
	responseCh := make(chan struct {
		line string
		err  error
	}, 1)

	// Default timeout for reads (can be made configurable)
	// Use the global constant for timeout
	readTimeout := readTimeoutDuration

	// Perform the read operation in a separate goroutine
	go func() {
		line, err := c.stdout.ReadString('\n')
		responseCh <- struct {
			line string
			err  error
		}{line, err}
	}()

	// Wait for either the response or a timeout
	select {
	case result := <-responseCh:
		if result.err != nil {
			// Specifically check for EOF and provide a clearer error message
			if result.err == io.EOF {
				return fmt.Errorf("server process terminated unexpectedly (EOF received)")
			}
			// Check for unexpected closed pipe or similar conditions
			if strings.Contains(result.err.Error(), "closed") ||
				strings.Contains(result.err.Error(), "broken pipe") {
				return fmt.Errorf("connection to server closed unexpectedly: %w", result.err)
			}
			// General error case
			return fmt.Errorf("failed to read from stdout: %w", result.err)
		}

		if err := json.Unmarshal([]byte(result.line), resp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		return nil
	case <-time.After(readTimeout):
		return fmt.Errorf("timeout waiting for server response after %v", readTimeout)
	}
}

// RandomString calls the RandomString method on the MCP server
func (c *MCPClient) RandomString(length int) (string, error) {
	args := RandomStringArgs{Length: length}
	var result RandomStringResponse

	if err := c.Call(methodRandomString, args, &result); err != nil {
		return "", err
	}

	return result.Result, nil
}

// HasCapability checks if the server has a specific capability
func (c *MCPClient) HasCapability(capability string) bool {
	for _, cap := range c.serverCap {
		if cap == capability {
			return true
		}
	}
	return false
}

func main() {
	// Define command line flags
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
	client, err := NewMCPClient(serverPath)
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	defer client.Close()

	// Process the initialization message
	initResp, err := client.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Print the initialization response
	fmt.Printf("Connected to MCP server:\n")
	fmt.Printf("  Name: %s\n", initResp.Name)
	fmt.Printf("  Version: %s\n", initResp.Version)
	fmt.Printf("  Capabilities: %v\n", initResp.Capabilities)

	// Test the RandomString capability if available
	if client.HasCapability("RandomString") {
		fmt.Println("\nTesting RandomString capability:")
		randomStr, err := client.RandomString(defaultRandomStringLength)
		if err != nil {
			log.Fatalf("Failed to call RandomString: %v", err)
		}
		fmt.Printf("  Random string (%d chars): %s\n", defaultRandomStringLength, randomStr)
	}
}
