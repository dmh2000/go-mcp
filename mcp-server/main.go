package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// MCPService represents the Model Context Protocol service
type MCPService struct{}

// Initialization related types
type InitArgs struct{}

type InitResponse struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

// RandomStringArgs contains arguments for the RandomString method
type RandomStringArgs struct {
	Length int `json:"length"`
}

// RandomStringResponse contains the response for the RandomString method
type RandomStringResponse struct {
	Result string `json:"result"`
}

// Initialize returns information about the MCP server and its capabilities
func (s *MCPService) Initialize(args *InitArgs, reply *InitResponse) error {
	fmt.Fprintf(os.Stderr, "MCP Server initialized\n")
	log.Println("Initialize method called")
	*reply = InitResponse{
		Name:    "Go MCP Server",
		Version: "1.0.0",
		Capabilities: []string{
			"RandomString",
		},
	}
	return nil
}

// RandomString generates and returns a random string of specified length
func (s *MCPService) RandomString(args *RandomStringArgs, reply *RandomStringResponse) error {
	fmt.Fprintf(os.Stderr, "Generating random string of length %d\n", args.Length)
	log.Printf("RandomString method called with length: %d", args.Length)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := args.Length
	if length <= 0 {
		length = 10 // Default length
	}

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	reply.Result = string(b)
	return nil
}

// Custom JSON-RPC codec that logs requests and responses
type LoggingServerCodec struct {
	decoder *json.Decoder
	encoder *json.Encoder
	conn    io.ReadWriteCloser
	// Store the last parsed request params
	requestParams json.RawMessage
}

// NewLoggingServerCodec creates a new server codec with logging
func NewLoggingServerCodec(conn io.ReadWriteCloser) *LoggingServerCodec {
	return &LoggingServerCodec{
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(conn),
		conn:    conn,
	}
}

// ReadRequestHeader reads the request header and logs it
func (c *LoggingServerCodec) ReadRequestHeader(r *rpc.Request) error {
	// Read raw JSON to log it before decoding
	var request struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      json.RawMessage `json:"id"`
	}

	if err := c.decoder.Decode(&request); err != nil {
		return err
	}

	// Log the request details
	fmt.Fprintf(os.Stderr, "DEBUG: Received request: method=%s, params=%s\n",
		request.Method, string(request.Params))
	log.Printf("Received request: method=%s, params=%s",
		request.Method, string(request.Params))

	// Split the method name (e.g., "MCPService.RandomString" -> "MCPService", "RandomString")
	parts := strings.Split(request.Method, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid method name: %s", request.Method)
	}

	r.ServiceMethod = request.Method

	// Extract ID (not critical for our implementation)
	if len(request.ID) > 0 {
		if err := json.Unmarshal(request.ID, &r.Seq); err != nil {
			r.Seq = 0
		}
	}

	// Store params for ReadRequestBody
	c.requestParams = request.Params

	return nil
}

// ReadRequestBody reads the request body
func (c *LoggingServerCodec) ReadRequestBody(x interface{}) error {
	if x == nil || len(c.requestParams) == 0 {
		return nil
	}

	// Log the raw JSON being unmarshaled
	fmt.Fprintf(os.Stderr, "DEBUG: Trying to unmarshal params: %s\n", string(c.requestParams))

	// Unmarshal the raw JSON into the target
	return json.Unmarshal(c.requestParams, x)
}

// WriteResponse writes the response
func (c *LoggingServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	resp := struct {
		JSONRPC string       `json:"jsonrpc"`
		ID      uint64       `json:"id"`
		Result  interface{}  `json:"result,omitempty"`
		Error   *interface{} `json:"error,omitempty"`
	}{
		JSONRPC: "2.0",
		ID:      r.Seq,
	}

	if r.Error == "" {
		resp.Result = x
	} else {
		resp.Error = new(interface{})
		*resp.Error = r.Error
	}

	// Log the response
	respJSON, _ := json.Marshal(resp)
	fmt.Fprintf(os.Stderr, "DEBUG: Sending response: %s\n", string(respJSON))
	log.Printf("Sending response: %s", string(respJSON))

	return c.encoder.Encode(resp)
}

// Close closes the connection
func (c *LoggingServerCodec) Close() error {
	return c.conn.Close()
}

func main() {
	// Add logging setup to log to a file in the same directory as the binary
	logFile, err := os.OpenFile("mcp-server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("MCP Server log initialized")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Register the MCPService
	service := new(MCPService)
	err = rpc.Register(service)
	if err != nil {
		log.Fatalf("Error registering MCPService: %v", err)
	}

	// Setup JSON-RPC server over stdio
	log.Println("MCP Server starting, using stdio for communication")

	// Send initialization response to stdout
	initResponse := InitResponse{
		Name:    "Go MCP Server",
		Version: "1.0.0",
		Capabilities: []string{
			"RandomString",
		},
	}
	initJSON, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  initResponse,
		"id":      0,
	})
	if err != nil {
		log.Fatalf("Error marshaling initialization response: %v", err)
	}
	fmt.Println(string(initJSON))

	// Log initialization response
	log.Printf("Initialization response: %+v\n", initResponse)

	// Create a struct that combines stdin/stdout
	rw := struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: os.Stdin,
		Writer: os.Stdout,
		Closer: os.Stdin,
	}

	// Create our custom codec
	codec := NewLoggingServerCodec(rw)

	// Log codec creation
	log.Println("Custom JSON-RPC codec created, waiting for client requests")
	fmt.Fprintf(os.Stderr, "DEBUG: Custom JSON-RPC codec created, waiting for client requests\n")

	// Create a channel to signal when to stop serving
	done := make(chan struct{})

	// Start serving in a goroutine
	go func() {
		fmt.Fprintf(os.Stderr, "DEBUG: Starting to serve requests\n")
		for {
			err := rpc.ServeRequest(codec)
			if err != nil {
				log.Printf("ERROR serving request: %v", err)
				fmt.Fprintf(os.Stderr, "DEBUG: Error serving request: %v\n", err)
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					break
				}
			}
		}
		close(done)
	}()

	// Wait for either a signal or completion
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v, shutting down server...", sig)
		fmt.Fprintf(os.Stderr, "Shutting down MCP Server...\n")
		// Send shutdown message to stdout
		shutdownMsg, _ := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  map[string]string{"status": "shutting_down"},
			"id":      999,
		})
		fmt.Println(string(shutdownMsg))

		// Give some time for cleanup
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	case <-done:
		log.Println("RPC server stopped normally")
	}

	log.Println("MCP Server shutdown complete")
}
