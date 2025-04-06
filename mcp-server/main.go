package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Constants for server configuration and behavior
const (
	// Service information
	serviceName    = "Go MCP Server"
	serviceVersion = "1.0.0"

	// File paths
	logFilePath = "mcp-server.log"

	// Random string generation
	maxRandomStringLength = 1024
	defaultStringLength   = 10
	randomStringCharset   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// JSON-RPC
	jsonRPCVersion = "2.0"
	shutdownMsgID  = 999
	initMsgID      = 0

	// Timeouts and delays
	shutdownDelay = 500 * time.Millisecond
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
	log.Println("MCP Server initialized")
	log.Println("Initialize method called")
	*reply = InitResponse{
		Name:    serviceName,
		Version: serviceVersion,
		Capabilities: []string{
			"RandomString",
		},
	}
	return nil
}

// RandomString generates and returns a random string of specified length
func (s *MCPService) RandomString(args *RandomStringArgs, reply *RandomStringResponse) error {
	log.Printf("Generating random string of length %d", args.Length)
	log.Printf("RandomString method called with length: %d", args.Length)
	// Use the predefined charset
	length := args.Length
	if length <= 0 {
		length = defaultStringLength // Default length
	} else if length > maxRandomStringLength {
		log.Printf("Requested length %d exceeds maximum allowed length %d", length, maxRandomStringLength)
		return fmt.Errorf("requested length %d exceeds maximum allowed length %d", length, maxRandomStringLength)
	}

	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Errorf("failed to generate random string: %w", err)
	}

	// Map the random bytes to the charset
	for i := range b {
		b[i] = randomStringCharset[int(b[i])%len(randomStringCharset)]
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
	log.Printf("DEBUG: Received request: method=%s, params=%s",
		request.Method, string(request.Params))

	// Split the method name (e.g., "MCPService.RandomString" -> "MCPService", "RandomString")
	parts := strings.Split(request.Method, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid method name: %s", request.Method)
	}

	r.ServiceMethod = request.Method

	// Extract and validate ID field - critical for matching requests and responses in JSON-RPC
	if len(request.ID) > 0 {
		// Try to unmarshal as different types, as JSON-RPC 2.0 allows string, number, or null
		var idNum float64
		var idStr string

		// First try as number
		if err := json.Unmarshal(request.ID, &idNum); err == nil {
			// Validate the number is positive (for uint64)
			if idNum < 0 {
				log.Printf("Warning: Received negative ID value %f, using 0 instead", idNum)
				r.Seq = 0
			} else if idNum > float64(^uint64(0)) {
				log.Printf("Warning: ID value %f too large for uint64, using 0 instead", idNum)
				r.Seq = 0
			} else {
				r.Seq = uint64(idNum)
			}
		} else if err := json.Unmarshal(request.ID, &idStr); err == nil {
			// Try as string - convert to integer if possible, or hash it
			if intVal, err := strconv.ParseUint(idStr, 10, 64); err == nil {
				r.Seq = intVal
			} else {
				// Use a hash of the string as ID
				h := fnv.New64()
				h.Write([]byte(idStr))
				r.Seq = h.Sum64()
				log.Printf("Converting string ID '%s' to hash: %d", idStr, r.Seq)
			}
		} else {
			log.Printf("Warning: Failed to parse ID field: %v, using 0", err)
			r.Seq = 0
		}
	} else {
		// Empty ID or null ID (for notifications)
		r.Seq = 0
		log.Printf("Request has no ID (notification or empty ID)")
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
	log.Printf("DEBUG: Trying to unmarshal params: %s", string(c.requestParams))

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
		JSONRPC: jsonRPCVersion,
		ID:      r.Seq,
	}

	if r.Error == "" {
		resp.Result = x
	} else {
		resp.Error = new(interface{})
		*resp.Error = r.Error
	}

	// Log the response
	respJSON, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}
	log.Printf("DEBUG: Sending response: %s", string(respJSON))

	return c.encoder.Encode(resp)
}

// Close closes the connection
func (c *LoggingServerCodec) Close() error {
	return c.conn.Close()
}

func main() {
	// Add logging setup to log to a file in the same directory as the binary
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	var initResponse InitResponse
	service.Initialize(&InitArgs{}, &initResponse)

	initJSON, err := json.Marshal(map[string]interface{}{
		"jsonrpc": jsonRPCVersion,
		"result":  initResponse,
		"id":      initMsgID,
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
	log.Println("DEBUG: Custom JSON-RPC codec created, waiting for client requests")

	// Create a channel to signal when to stop serving
	done := make(chan struct{})

	// Start serving in a goroutine
	go func() {
		log.Println("DEBUG: Starting to serve requests")
		for {
			err := rpc.ServeRequest(codec)
			if err != nil {
				log.Printf("ERROR serving request: %v", err)
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
		// Send shutdown message to stdout
		shutdownMsg, err := json.Marshal(map[string]interface{}{
			"jsonrpc": jsonRPCVersion,
			"result":  map[string]string{"status": "shutting_down"},
			"id":      shutdownMsgID,
		})
		if err != nil {
			log.Printf("Error marshaling shutdown message: %v", err)
		} else {
			fmt.Println(string(shutdownMsg))
		}

		// Give some time for cleanup
		time.Sleep(shutdownDelay)
		os.Exit(0)
	case <-done:
		log.Println("RPC server stopped normally")
	}

	log.Println("MCP Server shutdown complete")
}
