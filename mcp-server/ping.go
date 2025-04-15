package main

import (
	"sqirvy/mcp/pkg/mcp"
	"sqirvy/mcp/pkg/utils" // Import the custom logger
)

// handlePingRequest handles the "ping" request.
// It simply returns an empty result object as per the spec.
func (s *Server) handlePingRequest(id mcp.RequestID) ([]byte, error) {
	s.logger.Printf(utils.LevelDebug, "Handle  : ping request (ID: %v)", id)

	// The result for ping is just an empty object.
	result := map[string]interface{}{} // Empty map represents empty JSON object {}

	// Marshal the successful response using the server's helper
	responseBytes, err := s.marshalResponse(id, result)
	if err != nil {
		// marshalResponse already logged the error and returns marshalled error bytes
		return responseBytes, err // Return the error bytes and the original marshalling error
	}

	return responseBytes, nil // Return success response bytes and nil error
}
