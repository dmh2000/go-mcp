package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"sqirvy/mcp/pkg/mcp"
)

const (
	// Define the range of printable ASCII characters (inclusive)
	asciiStart = 32  // Space
	asciiEnd   = 126 // Tilde (~)
	asciiRange = asciiEnd - asciiStart + 1

	// Define the maximum allowed length for random data generation
	maxRandomDataLength = 1024
)

// Define the random_data template
var RandomDataTemplate mcp.ResourceTemplate = mcp.ResourceTemplate{
	Name:        "random_data",
	URITemplate: "data://random_data?length={length}", // RFC 6570 template
	Description: "Returns a string of random ASCII characters. Use URI like 'data://random_data?length=N' in resources/read, where N is the desired length.",
	MimeType:    "text/plain",
}

// RandomData generates a cryptographically secure random string of ASCII characters
// of the specified length using rejection sampling on raw bytes.
// Returns an error if length <= 0, length exceeds maxRandomDataLength, or if reading random data fails.
func RandomData(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}
	if length > maxRandomDataLength {
		return "", fmt.Errorf("requested length %d exceeds maximum allowed length %d", length, maxRandomDataLength)
	}

	result := make([]byte, length)
	bytesNeeded := length // Start with assuming 1 byte per character
	idx := 0

	for idx < length {
		// Read a batch of random bytes. Adjust batch size as needed for efficiency.
		// Reading more bytes at once reduces the overhead of calling rand.Read.
		bufferSize := bytesNeeded * 2 // Read more than strictly needed to reduce calls
		if bufferSize < 16 {
			bufferSize = 16
		} // Minimum buffer size
		randomBytes := make([]byte, bufferSize)

		n, err := io.ReadFull(rand.Reader, randomBytes)
		if err != nil {
			// Handle EOF specifically if it might occur (e.g., limited entropy source)
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return "", fmt.Errorf("failed to read enough random data (read %d bytes): %w", n, err)
			}
			return "", fmt.Errorf("failed to read random bytes: %w", err)
		}

		for _, b := range randomBytes {
			// Rejection sampling: Only accept bytes within the desired ASCII range
			if b >= asciiStart && b <= asciiEnd {
				if idx < length { // Ensure we don't write past the end of result slice
					result[idx] = b
					idx++
				} else {
					break // We have enough characters
				}
			}
			// Bytes outside the range [asciiStart, asciiEnd] are rejected (ignored)
		}

		if idx >= length {
			break // Exit outer loop once we have enough characters
		}

		// Estimate remaining bytes needed (can be refined)
		bytesNeeded = length - idx
	}

	return string(result), nil
}

// handleRandomDataResource processes a read request specifically for the data://random_data URI.
// It extracts the length, generates data, and marshals the response or error.
func (s *Server) handleRandomDataResource(id mcp.RequestID, params mcp.ReadResourceParams, parsedURI *url.URL) ([]byte, error) {
	s.logger.Printf("Processing random_data resource for URI: %s", params.URI)

	// Get the length parameter
	lengthStr := parsedURI.Query().Get("length")
	if lengthStr == "" {
		err := fmt.Errorf("missing 'length' query parameter in URI: %s", params.URI)
		s.logger.Println(err.Error())
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidParams, err.Error(), nil)
		return s.marshalErrorResponse(id, rpcErr)
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		err = fmt.Errorf("invalid 'length' query parameter '%s': %w", lengthStr, err)
		s.logger.Println(err.Error())
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidParams, err.Error(), nil)
		return s.marshalErrorResponse(id, rpcErr)
	}

	// Generate random data using the function from resources.go
	randomString, err := RandomData(length)
	if err != nil {
		// RandomData already logs details, just wrap the error for the RPC response
		err = fmt.Errorf("failed to generate random data for URI %s: %w", params.URI, err)
		s.logger.Println(err.Error())
		// Check if the error was due to invalid length (positive, max)
		// Use errors.Is for specific error types if RandomData returns them, otherwise check message
		if strings.Contains(err.Error(), "length must be positive") || strings.Contains(err.Error(), "exceeds maximum allowed length") {
			rpcErr := mcp.NewRPCError(mcp.ErrorCodeInvalidParams, err.Error(), nil)
			return s.marshalErrorResponse(id, rpcErr)
		}
		// Otherwise, treat as internal error
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInternalError, err.Error(), nil)
		return s.marshalErrorResponse(id, rpcErr)
	}

	// Prepare the result content
	content := mcp.TextResourceContents{
		URI:      params.URI,
		MimeType: "text/plain",
		Text:     randomString,
	}
	contentBytes, err := json.Marshal(content)
	if err != nil {
		err = fmt.Errorf("failed to marshal TextResourceContents for %s: %w", params.URI, err)
		s.logger.Println(err.Error())
		rpcErr := mcp.NewRPCError(mcp.ErrorCodeInternalError, err.Error(), nil)
		return s.marshalErrorResponse(id, rpcErr)
	}

	result := mcp.ReadResourceResult{
		Contents: []json.RawMessage{json.RawMessage(contentBytes)},
	}
	return s.marshalResponse(id, result)
}
