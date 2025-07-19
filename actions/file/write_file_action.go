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
func NewWriteFileAction(filePath string, content []byte, overwrite bool, inputBuffer *bytes.Buffer, logger *slog.Logger) *engine.Action[*WriteFileAction] {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	if filePath == "" {
		logger.Error("Invalid parameter: filePath cannot be empty")
		return nil
	}
	if inputBuffer == nil && len(content) == 0 {
		logger.Error("Invalid parameter: either content or inputBuffer must be provided")
		return nil
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
	}
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
	contentToWrite := a.Content // Default to pre-defined content
	if a.InputBuffer != nil {
		contentToWrite = a.InputBuffer.Bytes()
		a.Logger.Debug("Using content from input buffer", "buffer_length", len(contentToWrite))
	} else {
		a.Logger.Debug("Using pre-defined content", "content_length", len(contentToWrite))
	}

	a.Logger.Info("Attempting to write file", "path", a.FilePath, "content_length", len(contentToWrite), "overwrite", a.Overwrite)

	if !a.Overwrite {
		if _, err := os.Stat(a.FilePath); err == nil {
			errMsg := fmt.Sprintf("file %s already exists and overwrite is set to false", a.FilePath)
			a.Logger.Error(errMsg)
			return errors.New(errMsg)
		} else if !os.IsNotExist(err) {
			a.Logger.Error("Failed to check if file exists", "path", a.FilePath, "error", err)
			return fmt.Errorf("failed to stat file %s before writing: %w", a.FilePath, err)
		}
	}

	dir := filepath.Dir(a.FilePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		a.Logger.Error("Failed to create parent directory for file", "path", dir, "error", err)
		return fmt.Errorf("failed to create directory %s for file: %w", dir, err)
	}

	// Write the determined content
	if err := os.WriteFile(a.FilePath, contentToWrite, 0600); err != nil {
		a.Logger.Error("Failed to write file", "path", a.FilePath, "error", err)
		return fmt.Errorf("failed to write file %s: %w", a.FilePath, err)
	}

	a.Logger.Info("Successfully wrote file", "path", a.FilePath)
	return nil
}
