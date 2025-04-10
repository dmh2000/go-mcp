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
	"math/big" // Added for crypto/rand.Int
)

const (
	// Define the set of allowed characters (alphanumeric)
	allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
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

// RandomData generates a cryptographically secure random string of alphanumeric characters
// (a-z, A-Z, 0-9) of the specified length.
// Returns an error if length <= 0, length exceeds maxRandomDataLength, or if generating random indices fails.
func RandomData(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}
	if length > maxRandomDataLength {
		return "", fmt.Errorf("requested length %d exceeds maximum allowed length %d", length, maxRandomDataLength)
	}

	result := make([]byte, length)
	numChars := big.NewInt(int64(len(allowedChars)))

	for i := 0; i < length; i++ {
		// Generate a random index within the bounds of the allowed character set
		randomIndex, err := rand.Int(rand.Reader, numChars)
		if err != nil {
			return "", fmt.Errorf("failed to generate random index: %w", err)
		}
		// Select the character at the random index
		result[i] = allowedChars[randomIndex.Int64()]
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
