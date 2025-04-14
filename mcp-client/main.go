package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	// Use the absolute module path based on go.mod
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

}
