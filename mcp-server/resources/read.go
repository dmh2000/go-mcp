package resources

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

	// Ensure the path is absolute and clean
	filePath = filepath.Clean(filePath)
	if !filepath.IsAbs(filePath) && runtime.GOOS != "windows" { // Windows paths might start with drive letter
		return nil, "", fmt.Errorf("file URI path must be absolute: %s", parsedURI.Path)
	}
	// TODO: Add security checks here to prevent accessing files outside allowed directories (path traversal).

	logger.Printf("Attempting to read file: %s", filePath)

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

// Helper function needed for path cleaning on Windows
import "runtime"
