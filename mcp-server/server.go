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
	notificationInitialized = "initialized"    // Standard notification method from client after initialize response
	headerContentLength     = "Content-Length" // Duplicated from transport.go for sendRawMessage
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
	// Add state for resources, tools, prompts later
}

// NewServer creates a new MCP server instance.
func NewServer(reader io.Reader, writer io.Writer, logger *log.Logger) *Server {
	return &Server{
		reader:        bufio.NewReader(reader),
		writer:        writer,
		logger:        logger,
		initialized:   false,
		serverVersion: "2024-11-05", // Align with your spec/schema version
		serverInfo: mcp.Implementation{
			Name:    "GoMCPExampleServer",
			Version: "0.1.0", // Example version
		},
	}
}

// Run starts the server's main loop.
func (s *Server) Run() error {
	s.logger.Println("Server Run() started. Waiting for initialize request...")

	// 1. Handle Initialization
	if err := s.performInitialization(); err != nil {
		s.logger.Printf("Initialization failed: %v", err)
		return fmt.Errorf("initialization failed: %w", err)
	}

	s.logger.Println("Initialization complete. Entering main request loop.")
	s.initialized = true

	// 2. Main Request Loop
	for {
		payload, err := readMessage(s.reader, s.logger)
		if err != nil {
			if err == io.EOF {
				s.logger.Println("Client closed connection (EOF). Exiting.")
				return nil // Clean exit
			}
			s.logger.Printf("Error reading message: %v. Exiting.", err)
			// Treat read errors in the main loop as fatal for now
			return fmt.Errorf("error reading message: %w", err)
		}

		// Process the received message
		s.processMessage(payload)
	}
}

// performInitialization handles the initial handshake.
func (s *Server) performInitialization() error {
	// Wait for the first message, expecting "initialize"
	payload, err := readMessage(s.reader, s.logger)
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("client disconnected before sending initialize request")
		}
		return fmt.Errorf("error reading initial message: %w", err)
	}

	// Peek to check if it's an initialize request
	method, id, isNotification, isResponse, _ := peekMessageType(payload)

	// Strict check: Must be a request with method "initialize"
	if isNotification || isResponse || method != mcp.MethodInitialize || id == nil {
		err := fmt.Errorf("expected '%s' request first, but got method '%s' (notification: %t, response: %t, id: %v)", mcp.MethodInitialize, method, isNotification, isResponse, id)
		s.logger.Println(err.Error())
		// Try to send an error response if we have an ID (which we should for a request)
		if id != nil {
			rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidRequest, err.Error(), nil)
			// Marshal and send the error response directly here
			errorBytes, marshalErr := mcp.MarshalErrorResponse(id, rpcErr)
			if marshalErr == nil {
				_ = s.sendRawMessage(errorBytes) // Ignore send error here, as we are exiting anyway
			} else {
				s.logger.Printf("Failed to marshal error response for invalid initial message: %v", marshalErr)
			}
		}
		return err
	}

	// Handle the initialize request (which returns marshalled response/error bytes)
	responseBytes, handleErr := s.handleInitializeRequest(id, payload)
	if handleErr != nil {
		// Error should have been logged in handler.
		// If responseBytes is not nil here, it's a marshalled error response from the handler.
		if responseBytes != nil {
			_ = s.sendRawMessage(responseBytes) // Attempt to send the error response
		}
		return fmt.Errorf("failed to handle initialize request: %w", handleErr)
	}

	// Send the successful initialize response
	if err := s.sendRawMessage(responseBytes); err != nil {
		return fmt.Errorf("failed to send initialize response: %w", err)
	}

	s.logger.Println("Initialize response sent. Waiting for 'initialized' notification...")

	// Wait for the "initialized" notification (use constant)
	payload, err = readMessage(s.reader, s.logger)
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("client disconnected after sending initialize response, before sending '%s' notification", notificationInitialized)
		}
		return fmt.Errorf("error reading message after initialize response: %w", err)
	}

	method, _, isNotification, _, _ = peekMessageType(payload)
	// Check if it's the correct notification
	// Note: The spec uses "notifications/initialized", but examples sometimes use "initialized".
	// Be flexible or stick strictly to one. Let's use the constant `notificationInitialized`.
	if !isNotification || (method != notificationInitialized && method != "notifications/initialized") {
		err := fmt.Errorf("expected '%s' or 'notifications/initialized' notification, but got method '%s' (notification: %t)", notificationInitialized, method, isNotification)
		s.logger.Println(err.Error())
		// Cannot send error for notification, just log and potentially exit
		return err
	}

	s.logger.Printf("Received '%s' notification. Initialization sequence complete.", method)
	// No response needed for notifications

	return nil
}

// processMessage determines the type of message and routes it appropriately.
func (s *Server) processMessage(payload []byte) {
	method, id, isNotification, isResponse, isError := peekMessageType(payload)

	if isNotification {
		s.logger.Printf("Received Notification (Method: %s). No response needed.", method)
		// Handle specific notifications like $/cancel if needed
		// Example:
		// if method == "$/cancel" {
		//     s.handleCancelNotification(payload)
		// }
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
		// Should not happen after initialization, but handle defensively
		s.logger.Printf("Error: Received duplicate 'initialize' request (ID: %v) after initialization.", id)
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidRequest, "Duplicate initialize request", nil)
		responseBytes, handleErr = mcp.MarshalErrorResponse(id, rpcErr)

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
