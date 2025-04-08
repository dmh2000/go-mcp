package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	// Use the absolute module path based on go.mod
	"sqirvy/mcp/pkg/mcp"
)

func main() {
	// --- Command Line Flags ---
	serverPath := flag.String("server-path", "bin/mcp-server", "Path to the mcp-server executable")
	serverLog := flag.String("server-log", "mcp-server-from-client.log", "Log file for the server subprocess")
	flag.Parse()

	// --- Logger Setup ---
	// Log directly to stdout for the client
	logger := log.New(os.Stdout, "MCP-CLIENT: ", log.LstdFlags|log.Lshortfile)
	logger.Println("--------------------------------------------------")
	logger.Println("MCP Client starting...")
	logger.Printf("Attempting to start server: %s", *serverPath)
	logger.Printf("Server log file: %s", *serverLog)

	// --- Start Server Subprocess ---
	// Pass the --log argument to the server
	cmd := exec.Command(*serverPath, "--log", *serverLog)

	// Get pipes for stdin and stdout of the server process
	serverStdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Fatalf("Failed to get server stdin pipe: %v", err)
	}
	defer serverStdin.Close() // Ensure pipe is closed

	serverStdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Fatalf("Failed to get server stdout pipe: %v", err)
	}
	defer serverStdout.Close() // Ensure pipe is closed

	// Redirect server's stderr to a file or client's stderr if desired
	// For now, let it inherit client's stderr or discard
	// cmd.Stderr = os.Stderr // Example: Redirect server stderr to client stderr

	// Start the server process
	if err := cmd.Start(); err != nil {
		logger.Fatalf("Failed to start server process '%s': %v", *serverPath, err)
	}
	logger.Printf("Server process started (PID: %d)", cmd.Process.Pid)

	// Ensure server process is killed eventually
	defer func() {
		logger.Println("Ensuring server process termination...")
		if cmd.Process != nil {
			// Attempt graceful shutdown first
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				logger.Printf("Failed to send SIGTERM to server process: %v. Attempting SIGKILL.", err)
				if killErr := cmd.Process.Kill(); killErr != nil {
					logger.Printf("Failed to send SIGKILL to server process: %v", killErr)
				}
			} else {
				logger.Println("Sent SIGTERM to server process.")
				// Optionally wait for a short period before force killing
				// time.Sleep(1 * time.Second)
				// if err := cmd.Process.Kill(); err != nil {
				// 	logger.Printf("Failed to send SIGKILL after SIGTERM: %v", err)
				// }
			}
		}
		// Wait for the process to exit and release resources
		if err := cmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				logger.Printf("Server process exited with error: %v (Stderr: %s)", err, string(exitErr.Stderr))
			} else {
				logger.Printf("Error waiting for server process: %v", err)
			}
		} else {
			logger.Println("Server process exited cleanly.")
		}
	}()

	// Handle Ctrl+C (SIGINT) to gracefully shut down server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Printf("Received signal: %v. Shutting down...", sig)
		// Signal the defer function to run by exiting main
		os.Exit(1)
	}()

	// --- Client Initialization ---
	client := NewClient(serverStdout, serverStdin, logger)
	logger.Println("Client created. Starting initialization...")

	// Define client capabilities (example)
	clientCaps := mcp.ClientCapabilities{
		// Add capabilities the client supports, e.g.:
		// Roots: &struct { ListChanged bool `json:"listChanged,omitempty"` }{ListChanged: false},
	}
	clientInfo := mcp.Implementation{
		Name:    "GoMCPExampleClient",
		Version: "0.1.0",
	}

	// Perform the initialize handshake
	serverCaps, err := client.Initialize("2024-11-05", clientInfo, clientCaps)
	if err != nil {
		logger.Fatalf("Initialization handshake failed: %v", err)
	}

	logger.Printf("Initialization successful. Server capabilities: %+v", serverCaps)

	// --- List Resources ---
	logger.Println("Requesting resource list from server...")
	listResult, err := client.ListResources(nil) // Pass nil for default params
	if err != nil {
		logger.Fatalf("Failed to list resources: %v", err)
	}
	logger.Printf("Successfully listed resources:")
	if len(listResult.Resources) == 0 {
		logger.Println("  (No resources reported by server)")
	} else {
		for _, resource := range listResult.Resources {
			logger.Printf("  - Name: %s, URI: %s, Description: %s, MimeType: %s",
				resource.Name, resource.URI, resource.Description, resource.MimeType)
		}
	}
	if listResult.NextCursor != "" {
		logger.Printf("  (Pagination cursor available: %s)", listResult.NextCursor)
	}

	logger.Println("Client finished.")
	logger.Println("--------------------------------------------------")

	// Allow a brief moment for logs to flush before exit/defer runs
	time.Sleep(100 * time.Millisecond)
}
