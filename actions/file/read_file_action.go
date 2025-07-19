package file

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	engine "github.com/ndizazzo/task-engine"
)

// NewReadFileAction creates an action that reads content from a file.
// The file contents will be stored in the provided buffer.
func NewReadFileAction(filePath string, outputBuffer *[]byte, logger *slog.Logger) (*engine.Action[*ReadFileAction], error) {
	if err := ValidateSourcePath(filePath); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}
	if outputBuffer == nil {
		return nil, fmt.Errorf("invalid parameter: outputBuffer cannot be nil")
	}

	id := fmt.Sprintf("read-file-%s", filepath.Base(filePath))
	return &engine.Action[*ReadFileAction]{
		ID: id,
		Wrapped: &ReadFileAction{
			BaseAction:   engine.BaseAction{Logger: logger},
			FilePath:     filePath,
			OutputBuffer: outputBuffer,
		},
	}, nil
}

// ReadFileAction reads content from a file and stores it in the provided buffer
type ReadFileAction struct {
	engine.BaseAction
	FilePath     string
	OutputBuffer *[]byte // Pointer to buffer where file contents will be stored
}

func (a *ReadFileAction) Execute(execCtx context.Context) error {
	// Sanitize path to prevent path traversal attacks
	sanitizedPath, err := SanitizePath(a.FilePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	a.Logger.Info("Attempting to read file", "path", sanitizedPath)

	// Check if file exists
	fileInfo, err := os.Stat(sanitizedPath)
	if err != nil {
		if os.IsNotExist(err) {
			errMsg := fmt.Sprintf("file %s does not exist", sanitizedPath)
			a.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
		a.Logger.Error("Failed to stat file", "path", sanitizedPath, "error", err)
		return fmt.Errorf("failed to stat file %s: %w", sanitizedPath, err)
	}

	// Check if it's a regular file
	if fileInfo.IsDir() {
		errMsg := fmt.Sprintf("path %s is a directory, not a file", sanitizedPath)
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	// Read the file contents
	// nosec G304 - Path is sanitized by SanitizePath function
	content, err := os.ReadFile(sanitizedPath)
	if err != nil {
		a.Logger.Error("Failed to read file", "path", sanitizedPath, "error", err)
		return fmt.Errorf("failed to read file %s: %w", sanitizedPath, err)
	}

	// Store the content in the output buffer
	*a.OutputBuffer = content

	a.Logger.Info("Successfully read file", "path", sanitizedPath, "size", len(content))
	return nil
}
