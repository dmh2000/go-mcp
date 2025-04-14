package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
)

// StdioTransport manages communication with a server subprocess over stdio.
type StdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
	writer io.Writer // Embed the writer for direct use
	logger *log.Logger
	mu     sync.Mutex // Protects writer access
}

// NewStdioTransport creates and starts a new server subprocess and establishes stdio pipes.
func NewStdioTransport(serverPath, serverLog string, logger *log.Logger) (*StdioTransport, error) {
	cmd := exec.Command(serverPath, "--log", serverLog)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get server stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close() // Close stdin pipe if stdout fails
		return nil, fmt.Errorf("failed to get server stdout pipe: %w", err)
	}

	// Start the server process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to start server process '%s': %w", serverPath, err)
	}

	logger.Printf("Server process started (PID: %d)", cmd.Process.Pid)

	return &StdioTransport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
		writer: stdin, // Use the stdin pipe directly as the writer
		logger: logger,
	}, nil
}

// WriteMessage sends a JSON message (as bytes) to the server's stdin.
// It appends a newline character as required by the line-based JSON protocol.
func (t *StdioTransport) WriteMessage(payload []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.logger.Printf("Send    : %s", string(payload)) // Log the message being sent

	if _, err := t.writer.Write(payload); err != nil {
		return fmt.Errorf("failed to write message payload: %w", err)
	}
	if _, err := t.writer.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}
	// Flushing is typically handled by the underlying pipe closing or OS buffering.
	// If explicit flushing is needed, check if t.writer implements http.Flusher or similar.
	return nil
}

// ReadMessage reads a single JSON message (a line ending in newline) from the server's stdout.
func (t *StdioTransport) ReadMessage() ([]byte, error) {
	// ReadBytes includes the delimiter, so we need to trim it later if needed.
	payload, err := t.reader.ReadBytes('\n')
	if err != nil {
		// Log EOF specifically, as it's often expected during shutdown
		if err == io.EOF {
			t.logger.Println("Read    : EOF received from server stdout.")
		} else {
			t.logger.Printf("Read Error: %v", err)
		}
		return nil, err // Return EOF or other errors
	}

	// Trim trailing newline characters for clean JSON parsing.
	trimmedPayload := bytes.TrimSpace(payload)
	if len(trimmedPayload) == 0 {
		t.logger.Println("Read    : Received empty line, continuing read.")
		// Recursively call ReadMessage to get the next non-empty line
		// Be cautious with recursion depth if the server sends many empty lines.
		return t.ReadMessage()
	}

	t.logger.Printf("Receive : %s", string(trimmedPayload)) // Log the received message
	return trimmedPayload, nil
}

// Close closes the stdin/stdout pipes and waits for the server process to exit.
func (t *StdioTransport) Close() error {
	var closeErr error
	var waitErr error

	t.logger.Println("Closing transport...")

	// Close stdin first to signal the server we're done sending.
	if err := t.stdin.Close(); err != nil {
		closeErr = fmt.Errorf("failed to close server stdin: %w", err)
		t.logger.Printf("Error closing stdin: %v", err)
	}

	// Closing stdout isn't usually necessary from the client side,
	// but we can ensure the reader is no longer used.
	// The underlying pipe will be closed when the process exits.

	// Wait for the server process to exit.
	if t.cmd != nil && t.cmd.Process != nil {
		t.logger.Printf("Waiting for server process (PID: %d) to exit...", t.cmd.Process.Pid)
		waitErr = t.cmd.Wait()
		if waitErr != nil {
			// Log wait errors (like non-zero exit status) but don't necessarily overwrite closeErr
			t.logger.Printf("Server process wait error: %v", waitErr)
			// Combine errors if both occurred
			if closeErr != nil {
				return fmt.Errorf("stdin close error: %v; server wait error: %w", closeErr, waitErr)
			}
			return fmt.Errorf("server wait error: %w", waitErr)
		}
		t.logger.Println("Server process exited.")
	} else {
		t.logger.Println("Server process already nil or not started.")
	}

	return closeErr // Return error from closing stdin if it occurred
}
