package file

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePath validates that a path is safe to use for file operations
// It ensures the path is either absolute or starts with a relative path indicator
func ValidatePath(path, pathType string) error {
	if path == "" {
		return fmt.Errorf("%s path cannot be empty", pathType)
	}

	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)

	// Check if it's an absolute path
	if filepath.IsAbs(cleanPath) {
		return nil
	}

	// Check if it starts with a relative path indicator
	if strings.HasPrefix(cleanPath, ".") {
		return nil
	}

	// For relative paths, ensure they don't contain dangerous path traversal
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid %s path: %s (contains potentially dangerous path traversal)", pathType, path)
	}

	// Allow simple relative paths without .. (e.g., "file.txt", "dir/file.txt")
	return nil
}

// ValidateSourcePath validates a source path for file operations
func ValidateSourcePath(path string) error {
	return ValidatePath(path, "source")
}

// ValidateDestinationPath validates a destination path for file operations
func ValidateDestinationPath(path string) error {
	return ValidatePath(path, "destination")
}

// SanitizePath cleans and validates a path, returning the cleaned path if valid
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	cleanPath := filepath.Clean(path)

	// Check if it's an absolute path
	if filepath.IsAbs(cleanPath) {
		return cleanPath, nil
	}

	// Check if it starts with a relative path indicator
	if strings.HasPrefix(cleanPath, ".") {
		return cleanPath, nil
	}

	// For relative paths, ensure they don't contain dangerous path traversal
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("invalid path: %s (contains potentially dangerous path traversal)", path)
	}

	// Allow simple relative paths without .. (e.g., "file.txt", "dir/file.txt")
	return cleanPath, nil
}
