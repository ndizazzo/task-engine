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
	"github.com/ndizazzo/task-engine/actions/common"
)

// NewWriteFileAction creates a new WriteFileAction with the given logger
func NewWriteFileAction(logger *slog.Logger) *WriteFileAction {
	return &WriteFileAction{
		BaseAction:        engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for file path, content, overwrite flag, and input buffer
func (a *WriteFileAction) WithParameters(pathParam, content engine.ActionParameter, overwrite bool, inputBuffer *bytes.Buffer) (*engine.Action[*WriteFileAction], error) {
	if inputBuffer == nil && content == nil {
		return nil, fmt.Errorf("invalid parameter: either content or inputBuffer must be provided")
	}

	a.PathParam = pathParam
	a.Content = content
	a.Overwrite = overwrite
	a.InputBuffer = inputBuffer

	return &engine.Action[*WriteFileAction]{
		ID:      "write-file-action",
		Name:    "Write File",
		Wrapped: a,
	}, nil
}

// WriteFileAction writes specified content to a file
// It creates parent directories if needed
// By default (Overwrite=false), it will not overwrite the file if it exists
// If InputBuffer is set, its content will be used.
type WriteFileAction struct {
	engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	FilePath    string
	Content     engine.ActionParameter // Now supports ActionParameter
	Overwrite   bool
	InputBuffer *bytes.Buffer // Optional buffer to read content from
	// Internal state for output
	writtenContent []byte
	writeError     error
	PathParam      engine.ActionParameter // optional path parameter
}

func (a *WriteFileAction) Execute(execCtx context.Context) error {
	// Resolve path parameter if provided using the ParameterResolver
	effectivePath := a.FilePath
	if a.PathParam != nil {
		pathValue, err := a.ResolveStringParameter(execCtx, a.PathParam, "path")
		if err != nil {
			a.writeError = err
			return a.writeError
		}
		effectivePath = pathValue
	}

	// Sanitize path to prevent path traversal attacks
	sanitizedPath, err := SanitizePath(effectivePath)
	if err != nil {
		a.writeError = fmt.Errorf("invalid file path: %w", err)
		return a.writeError
	}

	var contentToWrite []byte

	// Resolve content parameter if provided using the ParameterResolver
	if a.Content != nil {
		resolvedContent, err := a.ResolveParameter(execCtx, a.Content, "content")
		if err != nil {
			a.writeError = err
			return a.writeError
		}

		// Convert resolved content to bytes
		switch v := resolvedContent.(type) {
		case []byte:
			contentToWrite = v
		case string:
			contentToWrite = []byte(v)
		case *[]byte:
			if v != nil {
				contentToWrite = *v
			}
		default:
			a.writeError = fmt.Errorf("unsupported content type: %T", resolvedContent)
			return a.writeError
		}
	}

	// Use input buffer if no content parameter resolved or if content is empty
	if a.InputBuffer != nil && len(contentToWrite) == 0 {
		contentToWrite = a.InputBuffer.Bytes()
		a.Logger.Debug("Using content from input buffer", "buffer_length", len(contentToWrite))
	} else if len(contentToWrite) > 0 {
		a.Logger.Debug("Using resolved content", "content_length", len(contentToWrite))
	}

	// Allow empty content (empty files are valid)
	// Store the content that was written for output
	a.writtenContent = make([]byte, len(contentToWrite))
	copy(a.writtenContent, contentToWrite)

	a.Logger.Info("Attempting to write file", "path", sanitizedPath, "content_length", len(contentToWrite), "overwrite", a.Overwrite)

	if !a.Overwrite {
		if _, err := os.Stat(sanitizedPath); err == nil {
			errMsg := fmt.Sprintf("file %s already exists and overwrite is set to false", sanitizedPath)
			a.Logger.Error(errMsg)
			a.writeError = errors.New(errMsg)
			return a.writeError
		} else if !os.IsNotExist(err) {
			a.Logger.Error("Failed to check if file exists", "path", sanitizedPath, "error", err)
			a.writeError = fmt.Errorf("failed to stat file %s before writing: %w", sanitizedPath, err)
			return a.writeError
		}
	}

	dir := filepath.Dir(sanitizedPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		a.Logger.Error("Failed to create parent directory for file", "path", dir, "error", err)
		a.writeError = fmt.Errorf("failed to create directory %s for file: %w", dir, err)
		return a.writeError
	}

	// Write the determined content
	if err := os.WriteFile(sanitizedPath, contentToWrite, 0o600); err != nil {
		a.Logger.Error("Failed to write file", "path", sanitizedPath, "error", err)
		a.writeError = fmt.Errorf("failed to write file %s: %w", sanitizedPath, err)
		return a.writeError
	}

	a.Logger.Info("Successfully wrote file", "path", sanitizedPath)
	return nil
}

// GetOutput returns information about the write operation
func (a *WriteFileAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, a.writeError == nil, map[string]interface{}{
		"filePath":      a.FilePath,
		"contentLength": len(a.writtenContent),
		"overwrite":     a.Overwrite,
		"error":         a.writeError,
	})
}
