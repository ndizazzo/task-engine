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

	// Normalize Windows-style separators so validation treats them as path separators
	normalized := strings.ReplaceAll(path, "\\", "/")

	// Cleaned path for absolute checks and final normalization
	cleanPath := filepath.Clean(normalized)

	// Quick absolute allow
	if filepath.IsAbs(cleanPath) {
		return nil
	}

	// Count traversal segments on the ORIGINAL normalized path to catch multi-level traversal
	// even when filepath.Clean collapses them.
	rawParts := strings.Split(normalized, "/")
	traversalCount := 0
	for _, segment := range rawParts {
		if segment == ".." {
			traversalCount++
		}
	}

	// Reject any path that attempts to traverse more than one directory up
	if traversalCount > 1 {
		return fmt.Errorf("invalid %s path: %s (contains potentially dangerous path traversal)", pathType, path)
	}

	// Split cleaned path into components for subsequent checks
	parts := strings.Split(cleanPath, "/")

	// Allow current directory references like "." or "./foo", but still forbid traversal beyond parent directories within the path
	if parts[0] == "." {
		// Disallow any ".." segments beyond leading "."
		for _, segment := range parts[1:] {
			if segment == ".." {
				return fmt.Errorf("invalid %s path: %s (contains potentially dangerous path traversal)", pathType, path)
			}
		}
		return nil
	}

	// Permit a single leading ".." (one parent up) to satisfy allowed use-cases,
	// but reject if the single ".." occurs anywhere except at the start
	if parts[0] == ".." {
		// traversalCount already guards against multiple traversals
		return nil
	}
	// If the single traversal appears not at the start, reject
	if traversalCount == 1 {
		return fmt.Errorf("invalid %s path: %s (contains potentially dangerous path traversal)", pathType, path)
	}

	// For other relative paths, ensure they don't contain any ".." traversal
	for _, segment := range parts {
		if segment == ".." {
			return fmt.Errorf("invalid %s path: %s (contains potentially dangerous path traversal)", pathType, path)
		}
	}

	// Allow simple relative paths without traversal (e.g., "file.txt", "dir/file.txt")
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

	// Normalize Windows-style separators to forward slashes for consistent behavior
	normalized := strings.ReplaceAll(path, "\\", "/")
	cleanPath := filepath.Clean(normalized)

	// Count traversal segments on original normalized path
	rawParts := strings.Split(normalized, "/")
	traversalCount := 0
	for _, segment := range rawParts {
		if segment == ".." {
			traversalCount++
		}
	}
	if traversalCount > 1 {
		return "", fmt.Errorf("invalid path: %s (contains potentially dangerous path traversal)", path)
	}

	// Absolute paths: return cleaned path
	if filepath.IsAbs(cleanPath) {
		return cleanPath, nil
	}

	parts := strings.Split(cleanPath, "/")

	// Handle current-dir relative paths: forbid any ".." segments beyond the leading "."
	if parts[0] == "." {
		for _, segment := range parts[1:] {
			if segment == ".." {
				return "", fmt.Errorf("invalid path: %s (contains potentially dangerous path traversal)", path)
			}
		}
		// Remove the leading "./" for a cleaner representation, but keep lone "." intact
		if cleanPath == "." {
			return cleanPath, nil
		}
		return strings.TrimPrefix(cleanPath, "./"), nil
	}

	// Allow a single leading ".." but no additional traversal segments
	if parts[0] == ".." {
		return cleanPath, nil
	}

	// Disallow any other occurrences of ".." or a single traversal not at the start
	if traversalCount == 1 {
		return "", fmt.Errorf("invalid path: %s (contains potentially dangerous path traversal)", path)
	}
	// Extra guard on cleaned components
	for _, segment := range parts {
		if segment == ".." {
			return "", fmt.Errorf("invalid path: %s (contains potentially dangerous path traversal)", path)
		}
	}

	return cleanPath, nil
}
