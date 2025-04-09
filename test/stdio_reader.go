package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "stdio_reader: Reading from stdin...") // Log start to stderr

	// Read all data from standard input
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stdio_reader: Error reading from stdin: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "stdio_reader: Read %d bytes from stdin. Printing to stdout.\n", len(inputData)) // Log completion to stderr

	// Print the raw data read from stdin directly to stdout
	fmt.Print(string(inputData))

	fmt.Fprintln(os.Stderr, "stdio_reader: Finished printing to stdout. Exiting.") // Log exit to stderr
}
