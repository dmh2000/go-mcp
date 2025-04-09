package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	// Define the range of printable ASCII characters (inclusive)
	asciiStart = 32  // Space
	asciiEnd   = 126 // Tilde (~)
	asciiRange = asciiEnd - asciiStart + 1

	// Define the maximum allowed length for random data generation
	maxRandomDataLength = 1024
)

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
