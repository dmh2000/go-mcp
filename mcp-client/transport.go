package main

// Contains functions for reading/writing JSON-RPC messages over stdio pipes.
// This is very similar to mcp-server/transport.go but adapted for the client.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/textproto" // Not strictly needed here, but kept for consistency
	"strconv"
	"strings"

	// Use the absolute module path
	"sqirvy/mcp/pkg/mcp"
)

const (
	headerContentLength = "Content-Length"
)

// writeMessage marshals the message, adds necessary headers, and writes to the writer.
// Used internally by client methods like sendRaw.
func writeMessage(writer io.Writer, message interface{}, logger *log.Logger) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Log the message being sent *before* writing headers/payload
	logger.Printf("Sending message: %s", string(payload))

	header := fmt.Sprintf("%s: %d\r\n\r\n", headerContentLength, len(payload))

	// Use a buffer to write header and payload atomically if possible
	var buf bytes.Buffer
	buf.WriteString(header)
	buf.Write(payload)

	n, err := writer.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write message (wrote %d/%d bytes): %w", n, buf.Len(), err)
	}
	if n != buf.Len() {
		return fmt.Errorf("incomplete message write (wrote %d/%d bytes)", n, buf.Len())
	}

	// Flush if the writer supports it (e.g., pipe might need it)
	if flusher, ok := writer.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			logger.Printf("Warning: failed to flush writer after sending message: %v", err)
		}
	} else if f, ok := writer.(interface{ Sync() error }); ok {
		// Pipes might implement Sync
		_ = f.Sync()
	}

	return nil
}

// readMessage reads headers and the corresponding payload from the reader.
func readMessage(reader *bufio.Reader, logger *log.Logger) ([]byte, error) {
	tpReader := textproto.NewReader(reader)

	// Read headers
	mimeHeader, err := tpReader.ReadMIMEHeader()
	if err != nil {
		// EOF is expected on clean shutdown
		if err == io.EOF {
			return nil, io.EOF
		}
		// Check for unexpected EOF or pipe errors
		if err == io.ErrUnexpectedEOF || strings.Contains(err.Error(), "pipe") || strings.Contains(err.Error(), "connection reset") {
			logger.Println("Received EOF/pipe error while reading headers, server likely disconnected.")
			return nil, io.EOF // Treat as clean EOF for loop termination
		}
		return nil, fmt.Errorf("failed to read MIME header: %w", err)
	}

	// Get Content-Length
	contentLengthStr := mimeHeader.Get(headerContentLength)
	if contentLengthStr == "" {
		var headers strings.Builder
		for k, v := range mimeHeader {
			headers.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", ")))
		}
		logger.Printf("Received headers without Content-Length:\n%s", headers.String())
		return nil, fmt.Errorf("missing or empty %s header", headerContentLength)
	}

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
		// EOF or UnexpectedEOF likely means server closed connection during payload read
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			logger.Printf("Received EOF/UnexpectedEOF while reading payload (expected %d bytes, got %d), server likely disconnected.", contentLength, n)
			return nil, io.EOF // Treat as clean EOF
		}
		return nil, fmt.Errorf("failed to read payload (expected %d bytes, got %d): %w", contentLength, n, err)
	}

	// Log the raw payload received
	logger.Printf("Received raw message payload (%d bytes): %s", len(payload), string(payload))

	return payload, nil
}

// peekMessageType attempts to unmarshal just enough to get the method/id/error.
// Useful for logging before full unmarshalling and handling.
// (Identical to server's version)
func peekMessageType(payload []byte) (method string, id mcp.RequestID, isNotification bool, isResponse bool, isError bool) {
	var base struct {
		Method  string          `json:"method"`
		ID      mcp.RequestID   `json:"id"` // Can be string, number, or null/absent
		Error   json.RawMessage `json:"error"`
		Result  json.RawMessage `json:"result"`
		Params  json.RawMessage `json:"params"`
		JSONRPC string          `json:"jsonrpc"`
	}

	decoder := json.NewDecoder(bytes.NewReader(payload))
	if err := decoder.Decode(&base); err != nil {
		return "", nil, false, false, false
	}

	if base.JSONRPC != "2.0" {
		return "", nil, false, false, false
	}

	id = base.ID
	method = base.Method
	hasID := base.ID != nil
	hasMethod := base.Method != ""
	hasResult := len(base.Result) > 0 && string(base.Result) != "null"
	hasError := len(base.Error) > 0 && string(base.Error) != "null"

	isNotification = !hasID && hasMethod
	isResponse = hasID && (hasResult || hasError)
	isError = hasID && hasError

	return method, id, isNotification, isResponse, isError
}
