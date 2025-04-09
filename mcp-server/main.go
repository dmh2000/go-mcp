package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath" // Added for path manipulation

	// Use the absolute module path
	"sqirvy/mcp/pkg/mcp"
)

func main() {
	// --- Command Line Flags ---
	logFilePath := flag.String("log", "mcp-server.log", "Path to the log file")
	flag.Parse()

	// --- Logger Setup ---
	// Ensure the directory for the log file exists
	logDir := filepath.Dir(*logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log directory %s: %v\n", logDir, err)
		os.Exit(1)
	}

	logFile, err := os.OpenFile(*logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file %s: %v\n", *logFilePath, err)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := log.New(logFile, "MCP-SERVER: ", log.LstdFlags|log.Lshortfile)
	logger.Println("--------------------------------------------------")
	logger.Println("MCP Server starting...")
	logger.Printf("Logging to file: %s", *logFilePath)

	// --- Server Initialization ---
	// Use standard input and output
	stdin := os.Stdin
	stdout := os.Stdout

	// Create and run the server
	server := NewServer(stdin, stdout, logger)
	err = server.Run()

	// --- Shutdown ---
	if err != nil {
		logger.Printf("Server exited with error: %v", err)
		fmt.Fprintf(os.Stderr, "Server exited with error: %v\n", err) // Also print to stderr
		logger.Println("--------------------------------------------------")
		os.Exit(1)
	}

	logger.Println("Server exited normally.")
	logger.Println("--------------------------------------------------")
}

// Helper function to create a standard MethodNotFound error response
func createMethodNotFoundResponse(id mcp.RequestID, method string, logger *log.Logger) ([]byte, error) {
	rpcErr := mcp.NewRPCError(mcp.ErrorCodeMethodNotFound, fmt.Sprintf("Method '%s' not found", method), nil)
	responseBytes, err := mcp.MarshalErrorResponse(id, rpcErr)
	if err != nil {
		logger.Printf("Error marshalling MethodNotFound error response for ID %v: %v", id, err)
		// Return a generic internal error if marshalling fails
		genericErr := mcp.NewRPCError(mcp.ErrorCodeInternalError, "Failed to marshal error response", nil)
		// We might not be able to marshal this either, but try
		responseBytes, _ = mcp.MarshalErrorResponse(id, genericErr)
		return responseBytes, fmt.Errorf("failed to marshal MethodNotFound error response: %w", err)
	}
	return responseBytes, nil
}
