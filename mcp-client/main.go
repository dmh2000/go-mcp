package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
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
		nextID: 1,
	}, nil
}

// Close shuts down the MCP client and server
func (c *MCPClient) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
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
		JSONRPC: "2.0",
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

// readResponse reads a JSON-RPC response from the server
func (c *MCPClient) readResponse(resp *RPCResponse) error {
	line, err := c.stdout.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read from stdout: %w", err)
	}

	if err := json.Unmarshal([]byte(line), resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// RandomString calls the RandomString method on the MCP server
func (c *MCPClient) RandomString(length int) (string, error) {
	args := RandomStringArgs{Length: length}
	var result RandomStringResponse

	if err := c.Call("MCPService.RandomString", args, &result); err != nil {
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
	// Update the server path to use the absolute path
	client, err := NewMCPClient("/home/dmh2000/projects/mcp/mcp-server/mcp-server")
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
		randomStr, err := client.RandomString(20)
		if err != nil {
			log.Fatalf("Failed to call RandomString: %v", err)
		}
		fmt.Printf("  Random string (20 chars): %s\n", randomStr)
	}
}
