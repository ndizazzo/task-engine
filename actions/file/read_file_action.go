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
func NewReadFileAction(filePath string, outputBuffer *[]byte, logger *slog.Logger) *engine.Action[*ReadFileAction] {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	if filePath == "" {
		logger.Error("Invalid parameter: filePath cannot be empty")
		return nil
	}
	if outputBuffer == nil {
		logger.Error("Invalid parameter: outputBuffer cannot be nil")
		return nil
	}

	id := fmt.Sprintf("read-file-%s", filepath.Base(filePath))
	return &engine.Action[*ReadFileAction]{
		ID: id,
		Wrapped: &ReadFileAction{
			BaseAction:   engine.BaseAction{Logger: logger},
			FilePath:     filePath,
			OutputBuffer: outputBuffer,
		},
	}
}

// ReadFileAction reads content from a file and stores it in the provided buffer
type ReadFileAction struct {
	engine.BaseAction
	FilePath     string
	OutputBuffer *[]byte // Pointer to buffer where file contents will be stored
}

func (a *ReadFileAction) Execute(execCtx context.Context) error {
	a.Logger.Info("Attempting to read file", "path", a.FilePath)

	// Check if file exists
	fileInfo, err := os.Stat(a.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			errMsg := fmt.Sprintf("file %s does not exist", a.FilePath)
			a.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
		a.Logger.Error("Failed to stat file", "path", a.FilePath, "error", err)
		return fmt.Errorf("failed to stat file %s: %w", a.FilePath, err)
	}

	// Check if it's a regular file
	if fileInfo.IsDir() {
		errMsg := fmt.Sprintf("path %s is a directory, not a file", a.FilePath)
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	// Read the file contents
	content, err := os.ReadFile(a.FilePath)
	if err != nil {
		a.Logger.Error("Failed to read file", "path", a.FilePath, "error", err)
		return fmt.Errorf("failed to read file %s: %w", a.FilePath, err)
	}

	// Store the content in the output buffer
	*a.OutputBuffer = content

	a.Logger.Info("Successfully read file", "path", a.FilePath, "size", len(content))
	return nil
}
