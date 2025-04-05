# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands
- Build: `go build ./cmd`
- Run: `go run ./cmd`
- Test: `go test ./...`
- Test a specific file: `go test ./path/to/file_test.go`
- Lint: `golangci-lint run`
- Format: `gofmt -w .` or `go fmt ./...`

## Code Style Guidelines
- **Formatting**: Use gofmt for consistent formatting
- **Imports**: Group standard library imports first, then third-party imports
- **Error handling**: Return errors with context using `fmt.Errorf("context: %w", err)`
- **Documentation**: Document all exported functions and types with comments
- **Context**: Pass context.Context as the first parameter to functions making external calls
- **Naming**: Use CamelCase for exported names, camelCase for unexported names
- **Performance**: Use strings.Builder for string concatenation
- **Error messages**: Start with lowercase, no trailing punctuation