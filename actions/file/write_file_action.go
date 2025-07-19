package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	engine "github.com/ndizazzo/task-engine"
)

// NewWriteFileAction creates an action that writes content to a file.
// If inputBuffer is provided, its content will be used.
// Otherwise, the provided static content argument is used.
func NewWriteFileAction(filePath string, content []byte, overwrite bool, inputBuffer *bytes.Buffer, logger *slog.Logger) (*engine.Action[*WriteFileAction], error) {
	if err := ValidateDestinationPath(filePath); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}
	if inputBuffer == nil && len(content) == 0 {
		return nil, fmt.Errorf("invalid parameter: either content or inputBuffer must be provided")
	}

	id := fmt.Sprintf("write-file-%s", filepath.Base(filePath))
	return &engine.Action[*WriteFileAction]{
		ID: id,
		Wrapped: &WriteFileAction{
			BaseAction:  engine.BaseAction{Logger: logger},
			FilePath:    filePath,
			Content:     content, // Static content (used if buffer is nil)
			Overwrite:   overwrite,
			InputBuffer: inputBuffer, // Store buffer pointer
		},
	}, nil
}

// WriteFileAction writes specified content to a file
// It creates parent directories if needed
// By default (Overwrite=false), it will not overwrite the file if it exists
// If InputBuffer is set, its content will be used.
type WriteFileAction struct {
	engine.BaseAction
	FilePath    string
	Content     []byte // Used if InputBuffer is nil
	Overwrite   bool
	InputBuffer *bytes.Buffer // Optional buffer to read content from
}

func (a *WriteFileAction) Execute(execCtx context.Context) error {
	// Sanitize path to prevent path traversal attacks
	sanitizedPath, err := SanitizePath(a.FilePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	contentToWrite := a.Content // Default to pre-defined content
	if a.InputBuffer != nil {
		contentToWrite = a.InputBuffer.Bytes()
		a.Logger.Debug("Using content from input buffer", "buffer_length", len(contentToWrite))
	} else {
		a.Logger.Debug("Using pre-defined content", "content_length", len(contentToWrite))
	}

	a.Logger.Info("Attempting to write file", "path", sanitizedPath, "content_length", len(contentToWrite), "overwrite", a.Overwrite)

	if !a.Overwrite {
		if _, err := os.Stat(sanitizedPath); err == nil {
			errMsg := fmt.Sprintf("file %s already exists and overwrite is set to false", sanitizedPath)
			a.Logger.Error(errMsg)
			return errors.New(errMsg)
		} else if !os.IsNotExist(err) {
			a.Logger.Error("Failed to check if file exists", "path", sanitizedPath, "error", err)
			return fmt.Errorf("failed to stat file %s before writing: %w", sanitizedPath, err)
		}
	}

	dir := filepath.Dir(sanitizedPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		a.Logger.Error("Failed to create parent directory for file", "path", dir, "error", err)
		return fmt.Errorf("failed to create directory %s for file: %w", dir, err)
	}

	// Write the determined content
	if err := os.WriteFile(sanitizedPath, contentToWrite, 0600); err != nil {
		a.Logger.Error("Failed to write file", "path", sanitizedPath, "error", err)
		return fmt.Errorf("failed to write file %s: %w", sanitizedPath, err)
	}

	a.Logger.Info("Successfully wrote file", "path", sanitizedPath)
	return nil
}
