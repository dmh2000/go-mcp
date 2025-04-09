package main

import (
	"bufio" // Added bufio
	"fmt"
	"io"
	"log" // Added log package
	"net/textproto" // Added textproto
	"os"
	"strconv" // Added strconv
	"strings" // Added strings
)

const logFileName = "stdio_reader.log"
const headerContentLength = "Content-Length" // Copied from transport.go

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

	log.Println("Starting up. Reading MCP messages from stdin...")

	// Use a buffered reader for stdin
	stdinReader := bufio.NewReader(os.Stdin)

	// Loop reading messages according to MCP/LSP framing
	for {
		log.Println("Waiting for next message...")
		payload, err := readMCPMessage(stdinReader, log.Default()) // Use local read function
		if err != nil {
			if err == io.EOF {
				log.Println("EOF received from stdin. Exiting.")
				break // Exit loop cleanly on EOF
			}
			// Log other errors and exit
			log.Fatalf("Error reading MCP message: %v", err)
		}

		// Log the successfully read message payload
		log.Printf("Successfully read message (%d bytes):\n---\n%s\n---", len(payload), string(payload))

		// Add a small delay if needed for testing/observation
		// time.Sleep(100 * time.Millisecond)
	}

	log.Println("Finished reading messages. Exiting.")
}

// readMCPMessage reads headers and the corresponding payload from the reader.
// (Simplified version based on mcp-server/transport.go readMessage)
func readMCPMessage(reader *bufio.Reader, logger *log.Logger) ([]byte, error) {
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

	return payload, nil
}
