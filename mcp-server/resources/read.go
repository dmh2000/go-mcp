package resources

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings" // Added for HasPrefix and TrimPrefix
)

// ReadFileResource reads the content of a file specified by a file:// URI.
// It returns the content as bytes, the determined MIME type, and any error.
func ReadFileResource(uri string, logger *log.Logger) ([]byte, string, error) {
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return nil, "", fmt.Errorf("invalid URI format: %w", err)
	}

	if parsedURI.Scheme != "file" {
		return nil, "", fmt.Errorf("unsupported URI scheme: %s", parsedURI.Scheme)
	}

	// Convert file URI path to a system path.
	// Handle potential differences in path separators and encoding.
	// For file://hostname/path, Host is usually empty or localhost on Unix-like systems.
	// For file:///path, Path starts with /.
	filePath := parsedURI.Path
	if parsedURI.Host != "" && parsedURI.Host != "localhost" {
		// Handle UNC paths if necessary, though less common for typical file URIs
		// For simplicity, we'll assume standard file paths here.
		logger.Printf("Warning: file URI host '%s' ignored, treating path as '%s'", parsedURI.Host, filePath)
	}

	// Get the project root (current working directory)
	projectRoot, err := os.Getwd()
	if err != nil {
		logger.Printf("Error getting working directory: %v", err)
		return nil, "", fmt.Errorf("internal server error: could not determine project root")
	}
	logger.Printf("Project root directory: %s", projectRoot)

	// Treat the URI path as relative to the project root.
	// Strip leading '/' from the URI path.
	relativePath := strings.TrimPrefix(parsedURI.Path, "/")

	// Join the project root with the relative path and clean it.
	filePath = filepath.Join(projectRoot, relativePath)
	filePath = filepath.Clean(filePath)

	// Security Check: Ensure the final path is still within the project root.
	// This helps prevent path traversal attacks (e.g., file:///../outside_project).
	if !strings.HasPrefix(filePath, projectRoot) {
		logger.Printf("Security Alert: Attempt to access file outside project root. Requested URI: %s, Resolved Path: %s", uri, filePath)
		return nil, "", fmt.Errorf("permission denied: cannot access files outside project root")
	}

	logger.Printf("Attempting to read file relative to project root: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("file not found: %s", filePath)
		}
		if os.IsPermission(err) {
			return nil, "", fmt.Errorf("permission denied reading file: %s", filePath)
		}
		return nil, "", fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	// Basic MIME type detection (can be improved with libraries like net/http.DetectContentType)
	// For now, assume text/plain for simplicity.
	mimeType := "text/plain"
	// Example using http.DetectContentType (requires importing "net/http")
	// if len(content) > 0 {
	//     mimeType = http.DetectContentType(content)
	// }
	// logger.Printf("Detected MIME type for %s: %s", filePath, mimeType)

	return content, mimeType, nil
}

// findProjectRoot searches upwards from the executable's directory for go.mod
// to determine the project root. Falls back to CWD if go.mod is not found.
func findProjectRoot(logger *log.Logger) (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		logger.Printf("Error getting executable path: %v", err)
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}
	currentDir := filepath.Dir(executablePath)
	logger.Printf("Searching for project root starting from executable directory: %s", currentDir)

	// Walk up the directory tree
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		// logger.Printf("Checking for go.mod at: %s", goModPath) // Verbose logging
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, this directory is the project root
			logger.Printf("Found go.mod at %s, using %s as project root", goModPath, currentDir)
			return currentDir, nil
		} else if !os.IsNotExist(err) {
			// Some other error occurred trying to stat go.mod
			logger.Printf("Error checking for go.mod in %s: %v", currentDir, err)
			// Decide whether to continue or return error. Let's return error.
			return "", fmt.Errorf("error checking for go.mod in %s: %w", currentDir, err)
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the root directory without finding go.mod
			logger.Printf("Reached filesystem root, go.mod not found. Falling back to CWD.")
			// Fallback to CWD as a last resort
			cwd, err := os.Getwd()
			if err != nil {
				logger.Printf("Error getting CWD as fallback: %v", err)
				return "", fmt.Errorf("go.mod not found and could not get CWD: %w", err)
			}
			logger.Printf("Using CWD %s as project root", cwd)
			return cwd, nil
		}
		currentDir = parentDir
	}
}
