// Package main implements an example Model Context Protocol (MCP) server
//
// This server demonstrates how to implement a server for the Model Context Protocol,
// which allows communication with clients through a standardized JSON-RPC interface.
// The server communicates over stdin/stdout, logs to a file, and provides capabilities
// that can be called by clients.
//
// This example shows:
// - Custom JSON-RPC over stdio implementation
// - Graceful server initialization and shutdown
// - Proper logging and error handling
// - Implementation of a simple MCP capability (RandomString)
// - Security considerations for random generation
package main

import (
	"crypto/rand"       // For secure random generation
	"encoding/json"     // For JSON encoding/decoding
	"fmt"               // For formatted I/O
	"hash/fnv"          // For hashing string IDs
	"io"                // For I/O interfaces
	"log"               // For logging
	"net/rpc"           // For RPC server implementation
	"os"                // For system calls
	"os/signal"         // For signal handling
	"strconv"           // For string conversions
	"strings"           // For string manipulation
	"syscall"           // For system call constants
	"time"              // For timeouts and delays
)

// Constants for server configuration and behavior
// Centralizing these values makes the code more maintainable and configurable
const (
	// Service information
	serviceName    = "Go MCP Server"    // Name of the MCP server
	serviceVersion = "1.0.0"            // Version of the MCP server implementation
	
	// File paths
	logFilePath = "mcp-server.log"      // Path to the log file
	
	// Random string generation
	maxRandomStringLength = 1024        // Maximum allowed random string length for security
	defaultStringLength   = 10          // Default length if not specified
	randomStringCharset   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // Character set for random strings
	
	// JSON-RPC constants
	jsonRPCVersion = "2.0"              // JSON-RPC 2.0 version string
	shutdownMsgID  = 999                // ID used for shutdown messages
	initMsgID      = 0                  // ID used for initialization messages
	
	// Timeouts and delays
	shutdownDelay = 500 * time.Millisecond // Time to wait before forcing shutdown
)

// MCPService represents the Model Context Protocol service
// This struct implements the RPC service methods that clients can call
// In a real implementation, this might contain configuration, state, or resources
type MCPService struct{}

// Initialization related types
// These are used during the initial handshake between client and server

// InitArgs contains arguments for the Initialize method
// This is empty in this implementation as no arguments are needed,
// but in a real implementation it could contain client information
type InitArgs struct{}

// InitResponse contains the server information sent during initialization
// This is the first message sent to clients and includes server capabilities
type InitResponse struct {
	Name         string   `json:"name"`         // Server name
	Version      string   `json:"version"`      // Server version
	Capabilities []string `json:"capabilities"` // List of available capabilities
}

// RandomStringArgs contains arguments for the RandomString method
// This demonstrates how to define a typed argument struct for a capability
type RandomStringArgs struct {
	Length int `json:"length"` // Desired length of the random string
}

// RandomStringResponse contains the response for the RandomString method
// This demonstrates how to define a typed response struct for a capability
type RandomStringResponse struct {
	Result string `json:"result"` // The generated random string
}

// Initialize returns information about the MCP server and its capabilities
// This is called by the client during initialization to get server information
// It follows the net/rpc convention of taking args and reply pointers
func (s *MCPService) Initialize(args *InitArgs, reply *InitResponse) error {
	// Log that the server was initialized - good for debugging
	log.Println("MCP Server initialized")
	log.Println("Initialize method called")
	
	// Populate the response with server information
	*reply = InitResponse{
		Name:    serviceName,    // Server name from constants
		Version: serviceVersion, // Server version from constants
		Capabilities: []string{
			"RandomString", // List of capabilities this server supports
		},
	}
	return nil
}

// RandomString generates and returns a random string of specified length
// This demonstrates implementing an MCP capability that clients can call
// The implementation uses cryptographically secure random generation with
// rejection sampling to ensure unbiased results
func (s *MCPService) RandomString(args *RandomStringArgs, reply *RandomStringResponse) error {
	// Log the request details for debugging
	log.Printf("Generating random string of length %d", args.Length)
	log.Printf("RandomString method called with length: %d", args.Length)
	
	// Validate and normalize the length parameter
	length := args.Length
	if length <= 0 {
		// Use default length if not specified or invalid
		length = defaultStringLength 
		log.Printf("Using default length of %d", defaultStringLength)
	} else if length > maxRandomStringLength {
		// Enforce maximum length for security reasons
		// This prevents potential DoS attacks requesting massive strings
		log.Printf("Requested length %d exceeds maximum allowed length %d", 
			length, maxRandomStringLength)
		return fmt.Errorf("requested length %d exceeds maximum allowed length %d", 
			length, maxRandomStringLength)
	}

	// Prepare result buffer to hold the random string
	result := make([]byte, length)
	
	// Implement rejection sampling for unbiased random generation
	// This is important for cryptographically secure random strings
	// where uniform distribution is required
	
	// Calculate the largest byte value that won't cause bias
	// We want random bytes in range [0, unbiased). Any value >= unbiased is rejected.
	charsetLen := len(randomStringCharset)
	// Find largest multiple of charset length not exceeding 256
	unbiased := 256 / charsetLen * charsetLen

	// Generate random bytes with rejection sampling
	randomBuf := make([]byte, 1) // Buffer for one random byte at a time
	for i := 0; i < length; i++ {
		// Keep trying until we get an unbiased byte
		for {
			// Get cryptographically secure random bytes
			_, err := rand.Read(randomBuf)
			if err != nil {
				return fmt.Errorf("failed to generate random string: %w", err)
			}
			
			// Reject if the value would cause bias (not uniformly distributed)
			if randomBuf[0] < byte(unbiased) {
				// No bias - use this byte to select a character
				result[i] = randomStringCharset[randomBuf[0]%byte(charsetLen)]
				break
			}
			// If biased, reject and try again
		}
	}

	// Set the result and return
	reply.Result = string(result)
	return nil
}

// Custom JSON-RPC codec that logs requests and responses
// This implements the rpc.ServerCodec interface to handle JSON-RPC over stdio
// with additional logging for debugging purposes
type LoggingServerCodec struct {
	decoder *json.Decoder     // For decoding incoming JSON requests
	encoder *json.Encoder     // For encoding outgoing JSON responses
	conn    io.ReadWriteCloser // The underlying connection (stdin/stdout)
	
	// Store the last parsed request params for passing between methods
	requestParams json.RawMessage
}

// NewLoggingServerCodec creates a new server codec with logging
// This factory function sets up a codec for the RPC server to use
// The conn parameter is typically a wrapper around stdin/stdout
func NewLoggingServerCodec(conn io.ReadWriteCloser) *LoggingServerCodec {
	return &LoggingServerCodec{
		decoder: json.NewDecoder(conn), // Create a decoder for incoming JSON
		encoder: json.NewEncoder(conn), // Create an encoder for outgoing JSON
		conn:    conn,                 // Store the connection for closing later
	}
}

// ReadRequestHeader reads the request header and logs it
// This is part of the rpc.ServerCodec interface implementation
// It parses the JSON-RPC request header and extracts method name and ID
func (c *LoggingServerCodec) ReadRequestHeader(r *rpc.Request) error {
	// Define a struct that matches the JSON-RPC 2.0 request format
	var request struct {
		JSONRPC string          `json:"jsonrpc"` // Should be "2.0"
		Method  string          `json:"method"`  // Method to call (e.g., "MCPService.RandomString")
		Params  json.RawMessage `json:"params"`  // Parameters as raw JSON
		ID      json.RawMessage `json:"id"`      // Request ID (can be number, string, or null)
	}

	// Decode the JSON-RPC request from the input stream
	if err := c.decoder.Decode(&request); err != nil {
		return err // Return any JSON parsing errors
	}

	// Log the request details for debugging
	log.Printf("DEBUG: Received request: method=%s, params=%s",
		request.Method, string(request.Params))

	// Validate the method name format (service.method)
	// This ensures the method can be properly dispatched to the right service
	parts := strings.Split(request.Method, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid method name: %s", request.Method)
	}

	// Set the service method in the RPC request
	r.ServiceMethod = request.Method

	// Extract and validate ID field - critical for matching requests and responses
	// JSON-RPC 2.0 allows IDs to be numbers, strings, or null
	if len(request.ID) > 0 {
		// Try to unmarshal as different types
		var idNum float64
		var idStr string

		// First try as number (most common case)
		if err := json.Unmarshal(request.ID, &idNum); err == nil {
			// Validate the number is in range for uint64
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
			// Try as string - either parse as number or hash it
			if intVal, err := strconv.ParseUint(idStr, 10, 64); err == nil {
				// String contains a valid number
				r.Seq = intVal
			} else {
				// String is not a number, hash it for a numeric ID
				h := fnv.New64()
				h.Write([]byte(idStr))
				r.Seq = h.Sum64()
				log.Printf("Converting string ID '%s' to hash: %d", idStr, r.Seq)
			}
		} else {
			// Couldn't parse as number or string
			log.Printf("Warning: Failed to parse ID field: %v, using 0", err)
			r.Seq = 0
		}
	} else {
		// Empty ID or null ID (used for notifications in JSON-RPC 2.0)
		r.Seq = 0
		log.Printf("Request has no ID (notification or empty ID)")
	}

	// Store params for later use in ReadRequestBody
	c.requestParams = request.Params

	return nil
}

// ReadRequestBody reads the request body and unmarshals the parameters
// This is part of the rpc.ServerCodec interface implementation
// It takes the previously stored raw JSON parameters and unmarshals them into the target struct
func (c *LoggingServerCodec) ReadRequestBody(x interface{}) error {
	// If there's no target or no parameters, nothing to do
	if x == nil || len(c.requestParams) == 0 {
		return nil
	}

	// Log the raw JSON being unmarshaled for debugging
	log.Printf("DEBUG: Trying to unmarshal params: %s", string(c.requestParams))

	// Unmarshal the raw JSON into the target struct
	// This populates the method's argument struct with the client-provided values
	return json.Unmarshal(c.requestParams, x)
}

// WriteResponse writes the JSON-RPC response back to the client
// This is part of the rpc.ServerCodec interface implementation
// It takes the RPC response and encodes it as a JSON-RPC 2.0 response
func (c *LoggingServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	// Create a struct that matches the JSON-RPC 2.0 response format
	resp := struct {
		JSONRPC string       `json:"jsonrpc"`      // Always "2.0" for JSON-RPC 2.0
		ID      uint64       `json:"id"`           // Must match the request ID
		Result  interface{}  `json:"result,omitempty"` // Result (if no error)
		Error   *interface{} `json:"error,omitempty"`  // Error (if any)
	}{
		JSONRPC: jsonRPCVersion, // From constants
		ID:      r.Seq,          // Match the request ID
	}

	// Set either Result or Error based on the response
	if r.Error == "" {
		// Success case - set the result
		resp.Result = x
	} else {
		// Error case - create an error object
		resp.Error = new(interface{})
		*resp.Error = r.Error
	}

	// Log the response for debugging
	respJSON, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}
	log.Printf("DEBUG: Sending response: %s", string(respJSON))

	// Encode and send the response to the client
	return c.encoder.Encode(resp)
}

// Close closes the underlying connection
// This is part of the rpc.ServerCodec interface implementation
// It's called when the RPC server is done with the codec
func (c *LoggingServerCodec) Close() error {
	// Simply close the underlying connection
	return c.conn.Close()
}

// main is the entry point for the MCP server
// It sets up logging, registers the service, and handles the JSON-RPC communication loop
func main() {
	// Set up logging to a file for debugging and troubleshooting
	// This is important since we're using stdin/stdout for communication
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// If we can't open the log file, we log to stderr and exit
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close() // Ensure we close the log file when the program exits
	log.SetOutput(logFile) // Redirect all log output to the file
	log.Println("MCP Server log initialized")

	// Set up signal handling for graceful shutdown
	// This allows the server to clean up when terminated with Ctrl+C or kill
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create and register the MCP service with the RPC server
	// This makes our service methods available to be called
	service := new(MCPService)
	err = rpc.Register(service)
	if err != nil {
		log.Fatalf("Error registering MCPService: %v", err)
	}

	// Log that we're starting the server
	log.Println("MCP Server starting, using stdio for communication")

	// Prepare and send the initialization response
	// This is the first message sent to clients upon connection
	// We send it before entering the main serve loop
	var initResponse InitResponse
	service.Initialize(&InitArgs{}, &initResponse)

	// Format the init response as a JSON-RPC message
	initJSON, err := json.Marshal(map[string]interface{}{
		"jsonrpc": jsonRPCVersion, // Always "2.0" for JSON-RPC 2.0
		"result":  initResponse,   // The initialization information
		"id":      initMsgID,      // ID 0 for the init message
	})
	if err != nil {
		log.Fatalf("Error marshaling initialization response: %v", err)
	}
	
	// Send the init JSON to stdout (to the client)
	fmt.Println(string(initJSON))

	// Log the initialization response for debugging
	log.Printf("Initialization response: %+v\n", initResponse)

	// Create a combined Reader/Writer/Closer for stdin/stdout communication
	// This serves as the transport for our JSON-RPC server
	rw := struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: os.Stdin,  // Read requests from stdin
		Writer: os.Stdout, // Write responses to stdout
		Closer: os.Stdin,  // Close stdin when we're done
	}

	// Create our custom JSON-RPC codec that adds logging
	codec := NewLoggingServerCodec(rw)

	// Log that we're ready to serve requests
	log.Println("DEBUG: Custom JSON-RPC codec created, waiting for client requests")

	// Create a channel to signal when the serve loop stops
	// This allows us to detect when the client disconnects
	done := make(chan struct{})

	// Start the main request-serving loop in a separate goroutine
	// This allows us to handle signals and perform graceful shutdown
	go func() {
		log.Println("DEBUG: Starting to serve requests")
		for {
			// ServeRequest processes a single request
			// It blocks until a request is received
			err := rpc.ServeRequest(codec)
			if err != nil {
				// Log errors but only exit the loop on EOF
				log.Printf("ERROR serving request: %v", err)
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					// EOF means the client has closed the connection
					break
				}
			}
		}
		// Signal that we're done serving requests
		close(done)
	}()

	// Wait for either a termination signal or client disconnection
	select {
	case sig := <-sigChan:
		// We received a termination signal (e.g., Ctrl+C)
		log.Printf("Received signal: %v, shutting down server...", sig)
		
		// Send a graceful shutdown message to the client
		shutdownMsg, err := json.Marshal(map[string]interface{}{
			"jsonrpc": jsonRPCVersion,
			"result":  map[string]string{"status": "shutting_down"},
			"id":      shutdownMsgID, // Special ID for shutdown messages
		})
		if err != nil {
			log.Printf("Error marshaling shutdown message: %v", err)
		} else {
			// Send the shutdown message to stdout (to the client)
			fmt.Println(string(shutdownMsg))
		}

		// Wait a bit to allow pending operations to complete
		log.Printf("Waiting %v for cleanup before exiting", shutdownDelay)
		time.Sleep(shutdownDelay)
		os.Exit(0)
		
	case <-done:
		// The serve loop has stopped (client disconnected)
		log.Println("RPC server stopped normally (client disconnected)")
	}

	// Final log message before exit
	// (Note: this may not be reached if os.Exit is called above)
	log.Println("MCP Server shutdown complete")
}
