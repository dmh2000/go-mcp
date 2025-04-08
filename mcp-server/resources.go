package main

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
)

const (
	// Define the range of printable ASCII characters (inclusive)
	asciiStart = 32  // Space
	asciiEnd   = 126 // Tilde (~)
	asciiRange = asciiEnd - asciiStart + 1
)

// RandomData generates a cryptographically secure random string of ASCII characters
// of the specified length. It uses rejection sampling to ensure uniform distribution
// across the printable ASCII range (32-126).
// Returns an error if length <= 0 or if reading random data fails.
func RandomData(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	result := make([]byte, length)
	max := big.NewInt(int64(asciiRange)) // The number of possible characters

	for i := 0; i < length; {
		// Generate a random number within the range [0, asciiRange)
		// Use crypto/rand for secure random numbers
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// Check for specific errors like io.EOF or io.ErrUnexpectedEOF if needed,
			// but generally, any error from crypto/rand is serious.
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return "", fmt.Errorf("failed to read enough random data (EOF): %w", err)
			}
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}

		// Add the offset to map the random number to the printable ASCII range
		char := byte(n.Int64() + asciiStart)

		// Although rand.Int should give uniform distribution within [0, max),
		// double-check it's within our expected ASCII bounds just in case.
		// This check is technically redundant if rand.Int works as expected,
		// but adds a layer of safety.
		if char >= asciiStart && char <= asciiEnd {
			result[i] = char
			i++ // Only increment if we successfully added a character
		}
		// No explicit rejection needed here because rand.Int already provides
		// uniform distribution within the desired range [0, asciiRange).
		// If we were reading raw bytes and mapping, rejection sampling would be
		// necessary to avoid bias if the range wasn't a power of 2.
	}

	return string(result), nil
}

// --- Example of rejection sampling if reading raw bytes ---
// This version reads raw bytes and uses rejection sampling.
// It's slightly less efficient than using rand.Int directly for this specific range.
func randomDataWithRejectionSampling(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	result := make([]byte, length)
	bytesNeeded := length // Start with assuming 1 byte per character
	idx := 0

	for idx < length {
		// Read a batch of random bytes. Adjust batch size as needed for efficiency.
		// Reading more bytes at once reduces the overhead of calling rand.Read.
		bufferSize := bytesNeeded * 2 // Read more than strictly needed to reduce calls
		if bufferSize < 16 { bufferSize = 16 } // Minimum buffer size
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
