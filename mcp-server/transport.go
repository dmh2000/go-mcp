package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"os" // Added for os.File type assertion
	"strconv"
	"strings"

	// Use the absolute module path
	"sqirvy/mcp/pkg/mcp"
)

const (
	headerContentLength = "Content-Length"
)

// writeMessage marshals the message, adds necessary headers, and writes to the writer.
// DEPRECATED in favor of sending raw bytes from server/handlers, but kept for reference
// or potential future use if sending structured data directly.
func writeMessage(writer io.Writer, message interface{}, logger *log.Logger) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Log the message being sent *before* writing headers/payload
	logger.Printf("Sending message: %s", string(payload))

	header := fmt.Sprintf("%s: %d\r\n\r\n", headerContentLength, len(payload))

	if _, err := writer.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write message header: %w", err)
	}
	if _, err := writer.Write(payload); err != nil {
		return fmt.Errorf("failed to write message payload: %w", err)
	}

	// Flush if the writer is buffered (like os.Stdout often is implicitly)
	if f, ok := writer.(*os.File); ok {
		// Best effort sync; might not be necessary depending on OS buffering
		_ = f.Sync()
	} else if flusher, ok := writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			logger.Printf("Warning: failed to flush writer: %v", err)
		}
	}

	return nil
}

// readMessage reads headers and the corresponding payload from the reader.
func readMessage(reader *bufio.Reader, logger *log.Logger) ([]byte, error) {

	tpReader := textproto.NewReader(reader)

	// Read headers
	mimeHeader, err := tpReader.ReadMIMEHeader()
	if err != nil {
		// EOF is expected on clean shutdown, return it directly
		if err == io.EOF {
			return nil, io.EOF
		}
		// Check for unexpected EOF, often happens if client disconnects abruptly
		if err == io.ErrUnexpectedEOF || strings.Contains(err.Error(), "unexpected EOF") {
			logger.Println("Received unexpected EOF while reading headers, client likely disconnected.")
			return nil, io.EOF // Treat as clean EOF for loop termination
		}
		return nil, fmt.Errorf("failed to read MIME header: %w", err)
	}
	logger.Printf("Received headers: %s", mimeHeader)

	// Get Content-Length
	contentLengthStr := mimeHeader.Get(headerContentLength)
	if contentLengthStr == "" {
		// Log the headers received for debugging
		var headers strings.Builder
		for k, v := range mimeHeader {
			headers.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", ")))
		}
		logger.Printf("Received headers without Content-Length:\n%s", headers.String())
		return nil, fmt.Errorf("missing or empty %s header", headerContentLength)
	}
	logger.Printf("Content-Length header: %s", contentLengthStr)

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid %s header value '%s': %w", headerContentLength, contentLengthStr, err)
	}
	if contentLength <= 0 {
		return nil, fmt.Errorf("invalid non-positive %s: %d", headerContentLength, contentLength)
	}

	// Read the exact number of bytes for the payload
	payload := make([]byte, contentLength)
	n, err := io.ReadFull(reader, payload)
	if err != nil {
		// EOF or UnexpectedEOF likely means client closed connection during payload read
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			logger.Printf("Received EOF/UnexpectedEOF while reading payload (expected %d bytes, got %d), client likely disconnected.", contentLength, n)
			return nil, io.EOF // Treat as clean EOF
		}
		return nil, fmt.Errorf("failed to read payload (expected %d bytes, got %d): %w", contentLength, n, err)
	}

	// Log the raw payload received
	logger.Printf("Received raw message payload (%d bytes): %s", len(payload), string(payload))

	return payload, nil
}

// peekMessageType attempts to unmarshal just enough to get the method/id/error.
// This is useful for logging before full unmarshalling and handling.
func peekMessageType(payload []byte) (method string, id mcp.RequestID, isNotification bool, isResponse bool, isError bool) {
	var base struct {
		Method  string          `json:"method"`
		ID      mcp.RequestID   `json:"id"`      // Can be string, number, or null/absent
		Error   json.RawMessage `json:"error"`   // Check if non-null
		Result  json.RawMessage `json:"result"`  // Check if non-null
		Params  json.RawMessage `json:"params"`  // Needed to differentiate req/notification
		JSONRPC string          `json:"jsonrpc"` // Check for presence
	}

	// Use a decoder to ignore unknown fields gracefully
	decoder := json.NewDecoder(bytes.NewReader(payload))
	// decoder.DisallowUnknownFields() // Use this for stricter parsing if needed

	if err := decoder.Decode(&base); err != nil {
		// Cannot determine type if basic unmarshal fails
		return "", nil, false, false, false
	}

	// Basic JSON-RPC validation
	if base.JSONRPC != "2.0" {
		return "", nil, false, false, false // Not a valid JSON-RPC 2.0 message
	}

	id = base.ID // Store the ID (can be nil)
	method = base.Method

	// Determine message type based on fields present according to JSON-RPC 2.0 spec
	hasID := base.ID != nil
	hasMethod := base.Method != ""
	hasResult := len(base.Result) > 0 && string(base.Result) != "null"
	hasError := len(base.Error) > 0 && string(base.Error) != "null"
	// Params check isn't strictly necessary for type determination but good practice
	// hasParams := len(base.Params) > 0 && string(base.Params) != "null"

	isNotification = !hasID && hasMethod          // Notification: MUST NOT have id, MUST have method
	isResponse = hasID && (hasResult || hasError) // Response: MUST have id, MUST have result OR error (but not both)
	isError = hasID && hasError                   // Error Response: MUST have id, MUST have error

	// If it's not a notification or response, it should be a request
	// isRequest := hasID && hasMethod && !hasResult && !hasError

	return method, id, isNotification, isResponse, isError
}
