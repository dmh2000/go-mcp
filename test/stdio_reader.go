package main

import (
	"fmt"
	"io"
	"log" // Added log package
	"os"
)

const logFileName = "stdio_reader.log"

func main() {
	// --- Logger Setup ---
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Fallback to stderr if log file cannot be opened
		fmt.Fprintf(os.Stderr, "stdio_reader: Error opening log file %s: %v. Logging to stderr.\n", logFileName, err)
		log.SetOutput(os.Stderr)
	} else {
		defer logFile.Close()
		log.SetOutput(logFile) // Direct log output to the file
	}
	log.SetPrefix("STDIO_READER: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Add timestamp and file/line info

	log.Println("Starting up. Reading from stdin...")

	// Read all data from standard input
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading from stdin: %v", err) // Use Fatalf to log and exit
	}

	log.Printf("Read %d bytes from stdin. Logging received data.", len(inputData))

	// Log the raw data read from stdin
	log.Printf("Received Data:\n---\n%s\n---", string(inputData))

	log.Println("Finished processing stdin. Exiting.")
}
