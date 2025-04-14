package main

import (
	"flag"
	"log"
	"os"
	// Use the absolute module path based on go.mod
	// No third-party libraries needed for this basic client yet.
)

func main() {
	// --- Command Line Flags ---
	// Default path assumes 'mcp-client' is run from the repository root.
	serverPath := flag.String("server-path", "bin/mcp-server", "Path to the mcp-server executable")
	serverLog := flag.String("server-log", "mcp-server-from-client.log", "Log file for the server subprocess")
	flag.Parse()

	// --- Logger Setup ---
	// Log directly to stdout for the client
	logger := log.New(os.Stdout, "MCP-CLIENT: ", log.LstdFlags|log.Lshortfile)
	logger.Println("--------------------------------------------------")
	logger.Println("MCP Client starting...")
	logger.Printf("Server executable: %s", *serverPath)
	logger.Printf("Server log file: %s", *serverLog)

	// --- Initialize Transport ---
	logger.Println("Initializing stdio transport...")
	transport, err := NewStdioTransport(*serverPath, *serverLog, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize transport: %v", err)
	}
	// Transport closing is handled by client.Run() via defer

	// --- Initialize and Run Client ---
	logger.Println("Creating MCP client...")
	client := NewClient(transport, logger)

	logger.Println("Running client handshake...")
	if err := client.Run(); err != nil {
		logger.Printf("Client run failed: %v", err)
		logger.Println("--------------------------------------------------")
		// Attempt to close transport even on error, logging any further issues
		if closeErr := transport.Close(); closeErr != nil {
			logger.Printf("Error closing transport after client failure: %v", closeErr)
		}
		os.Exit(1) // Exit with error status
	}

	// --- Shutdown ---
	logger.Println("Client finished successfully.")
	logger.Println("--------------------------------------------------")
	// Transport is closed via defer in client.Run()
	// No explicit exit needed here, main will return 0
}
