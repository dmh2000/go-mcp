package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os" // Added for os.File type assertion
	"sync"

	// Use the absolute module path
	"sqirvy/mcp/pkg/mcp"
)

const (
	notificationInitialized = "initialized" // Standard notification method from client after initialize response
	// headerContentLength is defined in transport.go
)

// Server handles the MCP communication logic.
type Server struct {
	reader        *bufio.Reader
	writer        io.Writer // Using io.Writer for flexibility, though likely os.Stdout
	logger        *log.Logger
	mu            sync.Mutex // Protects writer access
	initialized   bool
	serverVersion string
	serverInfo    mcp.Implementation
	incomingMessages chan []byte // Channel for incoming message payloads
	shutdown      chan struct{} // Channel to signal shutdown
	// Add state for resources, tools, prompts later
}

// NewServer creates a new MCP server instance.
func NewServer(reader io.Reader, writer io.Writer, logger *log.Logger) *Server {
	return &Server{
		reader:           bufio.NewReader(reader),
		writer:           writer,
		logger:           logger,
		initialized:      false,
		serverVersion:    "2024-11-05", // Align with your spec/schema version
		incomingMessages: make(chan []byte, 10), // Buffered channel
		shutdown:         make(chan struct{}),
		serverInfo: mcp.Implementation{
			Name:    "GoMCPExampleServer",
			Version: "0.1.0", // Example version
		},
	}
}

// Run starts the server's main loop.
func (s *Server) Run() error {
	s.logger.Println("Server Run() started.")
	s.initialized = false // Ensure server starts in non-initialized state

	// 1. Start background reader loop immediately
	go s.readLoop()

	// 3. Main processing loop
	s.logger.Println("Entering main processing loop.")
	for {
		select {
		case payload := <-s.incomingMessages:
			// Process the received message
			// Consider running in a separate goroutine for full concurrency: go s.processMessage(payload)
			// For simplicity now, process sequentially in the main loop.
			s.processMessage(payload)
		case <-s.shutdown:
			s.logger.Println("Shutdown signal received. Exiting processing loop.")
			return nil // Normal shutdown
		}
	}
}

// readLoop continuously reads messages from the transport and sends them to the incomingMessages channel.
// It exits when readMessage returns an error (like io.EOF).
func (s *Server) readLoop() {
	defer func() {
		s.logger.Println("Exiting read loop.")
		close(s.shutdown) // Signal the main loop to shut down when reading stops
	}()
	s.logger.Println("Starting read loop.")
	for {
		payload, err := readMessage(s.reader, s.logger)
		if err != nil {
			if err == io.EOF {
				s.logger.Println("Client closed connection (EOF). Read loop finished.")
				// Normal termination, signal shutdown
			} else {
				// Log other read errors
				s.logger.Printf("Error reading message in readLoop: %v.", err)
				// Depending on the error, might want to signal shutdown or try to recover
			}
			return // Exit loop on any error
		}

		// Send payload to the processing loop
		select {
		case s.incomingMessages <- payload:
			// Message sent successfully
		case <-s.shutdown:
			// Shutdown signal received while trying to send, discard message and exit
			s.logger.Println("Shutdown signal received during message send in readLoop. Discarding message.")
			return
		}
// processMessage determines the type of message and routes it appropriately.
// It also handles the initial state transitions (waiting for initialize, waiting for initialized).
func (s *Server) processMessage(payload []byte) {
	method, id, isNotification, isResponse, isError := peekMessageType(payload)

	// --- State Machine: Before Initialization ---
	if !s.initialized {
		// State 1: Waiting for "initialize" request
		if method == mcp.MethodInitialize && !isNotification && id != nil {
			s.logger.Printf("Received 'initialize' request (ID: %v) while not initialized.", id)
			responseBytes, handleErr := s.handleInitializeRequest(id, payload)
			// Send response (success or error marshalled by handler)
			if responseBytes != nil {
				if sendErr := s.sendRawMessage(responseBytes); sendErr != nil {
					s.logger.Printf("FATAL: Failed to send initialize response/error for request ID %v: %v", id, sendErr)
					// Consider signaling shutdown?
				} else {
					s.logger.Println("Initialize response sent. Waiting for 'initialized' notification...")
				}
			}
			if handleErr != nil {
				s.logger.Printf("Error handling initialize request (ID: %v): %v", id, handleErr)
				// Error already logged in handler, response (if possible) was sent.
			}
			// Do not set s.initialized = true yet. Wait for notification.
			return // Handled initialize request, wait for next message.
		}

		// State 2: Waiting for "initialized" notification (after sending initialize response)
		// Note: The spec uses "notifications/initialized", but examples sometimes use "initialized". Accept both.
		if isNotification && (method == notificationInitialized || method == "notifications/initialized") {
			s.logger.Printf("Received '%s' notification. Initialization sequence complete.", method)
			s.initialized = true // <-- SET INITIALIZED FLAG HERE
			return              // Handled notification, ready for normal operation.
		}

		// Unexpected message before initialization is complete
		s.logger.Printf("Error: Received unexpected message before initialization complete. Method: '%s', ID: %v, IsNotification: %t, IsResponse: %t", method, id, isNotification, isResponse)
		if id != nil && !isNotification && !isResponse { // If it was a request with an ID
			rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidRequest, "Server not initialized", nil)
			errorBytes, _ := mcp.MarshalErrorResponse(id, rpcErr) // Ignore marshal error, send if possible
			if errorBytes != nil {
				_ = s.sendRawMessage(errorBytes) // Ignore send error
			}
		}
		return // Ignore the unexpected message otherwise
	}

	// --- State Machine: Initialized ---
	// Handle messages received *after* initialization is complete.

	if isNotification {
		// Handle 'initialized' notification received *after* already initialized (benign)
		if method == notificationInitialized || method == "notifications/initialized" {
			s.logger.Printf("Warning: Received duplicate '%s' notification after already initialized. Ignoring.", method)
			return
		}
		s.logger.Printf("Received Notification (Method: %s). No response needed.", method)
		// Handle other specific notifications like $/cancel if needed
		return
	}

	if isResponse || isError {
		// Server shouldn't receive responses unless it sent requests (not implemented yet)
		s.logger.Printf("Warning: Received unexpected Response/Error message (ID: %v, Method: %s, IsError: %t). Ignoring.", id, method, isError)
		return
	}

	// It's a Request (must have ID and method, not result/error)
	if id == nil || method == "" {
		s.logger.Printf("Error: Received message that is not a valid Request, Notification, or Response. Payload: %s", string(payload))
		// Cannot send error response if ID is missing.
		return
	}

	s.logger.Printf("Received Request (ID: %v, Method: %s)", id, method)

	var responseBytes []byte
	var handleErr error // Error returned by the handler function itself

	// Route to the appropriate handler
	switch method {
	case mcp.MethodInitialize:
		// Handle duplicate 'initialize' request after initialization
		s.logger.Printf("Error: Received duplicate 'initialize' request (ID: %v) after initialization.", id)
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidRequest, "Server already initialized", nil)
		responseBytes, handleErr = s.marshalErrorResponse(id, rpcErr) // Use helper

	case mcp.MethodListTools:
		responseBytes, handleErr = s.handleListTools(id, payload)
	case mcp.MethodCallTool:
		responseBytes, handleErr = s.handleCallTool(id, payload)
	case mcp.MethodListPrompts:
		responseBytes, handleErr = s.handleListPrompts(id, payload)
	case mcp.MethodGetPrompt:
		responseBytes, handleErr = s.handleGetPrompt(id, payload)
	case mcp.MethodListResources:
		responseBytes, handleErr = s.handleListResources(id, payload)
	case mcp.MethodReadResource:
		responseBytes, handleErr = s.handleReadResource(id, payload)
	// Add cases for other supported methods like ping, logging/setLevel, etc.
	default:
		s.logger.Printf("Received unsupported method '%s' for request ID %v", method, id)
		responseBytes, handleErr = createMethodNotFoundResponse(id, method, s.logger)
	}

	// --- Response Sending ---
	if handleErr != nil {
		// The handler failed internally (e.g., failed to marshal its *intended* response/error).
		s.logger.Printf("Error during handling of request (ID: %v, Method: %s): %v", id, method, handleErr)
		// If responseBytes is not nil here, it means the handler *did* manage to marshal an error response despite the internal error.
		if responseBytes == nil {
			// If the handler couldn't even produce an error response, create a generic one.
			s.logger.Printf("Handler failed without producing an error response. Creating generic InternalError.")
			rpcErr := mcp.NewRPCError(mcp.ErrorCodeInternalError, fmt.Sprintf("Internal server error processing method %s", method), nil)
			responseBytes, _ = mcp.MarshalErrorResponse(id, rpcErr) // Ignore marshal error here, send if possible
		}
	}

	// Send the response (either success or error marshalled by the handler or the generic error)
	if responseBytes != nil {
		if sendErr := s.sendRawMessage(responseBytes); sendErr != nil {
			s.logger.Printf("FATAL: Failed to send response/error for request ID %v: %v", id, sendErr)
			// This is likely a fatal error (e.g., stdout closed).
			// Consider panic or exit? For now, just log. The main loop might exit on the next read error.
		}
	} else {
		// This case should ideally not happen if handlers always return marshalled bytes or an error
		s.logger.Printf("Warning: No response bytes generated for request (ID: %v, Method: %s), handleErr was: %v", id, method, handleErr)
	}
}

// sendRawMessage sends pre-marshalled bytes with headers.
func (s *Server) sendRawMessage(payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Log the raw payload being sent *before* writing
	s.logger.Printf("Sending raw message payload (%d bytes): %s", len(payload), string(payload))

	header := fmt.Sprintf("%s: %d\r\n\r\n", headerContentLength, len(payload))

	// Use a temporary buffer to write header and payload together
	// This can help prevent partial writes if the underlying writer buffers internally.
	var buf bytes.Buffer
	buf.WriteString(header)
	buf.Write(payload)

	n, err := s.writer.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write message (wrote %d/%d bytes): %w", n, buf.Len(), err)
	}
	if n != buf.Len() {
		return fmt.Errorf("incomplete message write (wrote %d/%d bytes)", n, buf.Len())
	}

	// Flush if the writer supports it (e.g., bufio.Writer, os.Stdout might need it)
	if f, ok := s.writer.(*os.File); ok {
		// Syncing stdout might be overkill/ineffective depending on OS/terminal.
		// It's generally safe but might not guarantee immediate visibility everywhere.
		_ = f.Sync()
	} else if flusher, ok := s.writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			// Log flush errors but don't treat them as fatal write errors usually
			s.logger.Printf("Warning: failed to flush writer after sending message: %v", err)
		}
	}

	return nil
}

// sendResponse marshals a successful result into a full RPCResponse and sends it.
// Returns the marshalled bytes and any error during marshalling.
// It does *not* send the bytes itself.
func (s *Server) marshalResponse(id mcp.RequestID, result interface{}) ([]byte, error) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		err = fmt.Errorf("failed to marshal result for response ID %v: %w", id, err)
		s.logger.Println(err.Error())
		// Return bytes for an internal error instead
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInternalError, "Failed to marshal response result", nil)
		errorBytes, marshalErr := mcp.MarshalErrorResponse(id, rpcErr)
		// If we can't even marshal the error, return the original error and nil bytes
		if marshalErr != nil {
			s.logger.Printf("CRITICAL: Failed to marshal error response for result marshalling failure: %v", marshalErr)
			return nil, err // Return the original marshalling error
		}
		return errorBytes, err // Return the marshalled error bytes and the original error
	}

	resp := mcp.RPCResponse{
		JSONRPC: mcp.JSONRPCVersion,
		Result:  resultBytes,
		ID:      id,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		// This is highly unlikely if result marshalling worked, but handle defensively
		err = fmt.Errorf("failed to marshal final response object for ID %v: %w", id, err)
		s.logger.Println(err.Error())
		// Return bytes for an internal error instead
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInternalError, "Failed to marshal final response object", nil)
		errorBytes, marshalErr := mcp.MarshalErrorResponse(id, rpcErr)
		if marshalErr != nil {
			s.logger.Printf("CRITICAL: Failed to marshal error response for final response marshalling failure: %v", marshalErr)
			return nil, err // Return the original marshalling error
		}
		return errorBytes, err // Return the marshalled error bytes and the original error
	}

	return respBytes, nil // Return marshalled success response bytes and nil error
}

// marshalErrorResponse marshals an RPCError into a full RPCResponse.
// Returns the marshalled bytes and any error during marshalling.
// It does *not* send the bytes itself.
func (s *Server) marshalErrorResponse(id mcp.RequestID, rpcErr *mcp.RPCError) ([]byte, error) {
	responseBytes, err := mcp.MarshalErrorResponse(id, rpcErr)
	if err != nil {
		// Log the failure to marshal the *intended* error
		s.logger.Printf("CRITICAL: Failed to marshal error response (Code: %d, Msg: %s) for ID %v: %v", rpcErr.Code, rpcErr.Message, id, err)

		// Try to marshal a more generic internal error response
		genericErr := mcp.NewRPCError(mcp.ErrorCodeInternalError, "Failed to marshal original error response", nil)
		responseBytes, err = mcp.MarshalErrorResponse(id, genericErr)
		if err != nil {
			// If we can't even marshal the generic error, log and return the error
			s.logger.Printf("CRITICAL: Failed to marshal generic error response for ID %v: %v. No error response possible.", id, err)
			return nil, fmt.Errorf("failed to marshal even generic error response: %w", err)
		}
		// Return the generic error bytes, but also the original error that occurred
		return responseBytes, fmt.Errorf("failed to marshal original error response (Code: %d), sending generic error instead: %w", rpcErr.Code, err)
	}
	// Return the successfully marshalled error response bytes and nil error
	return responseBytes, nil
}
