package file

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	engine "github.com/ndizazzo/task-engine"
)

// NewReadFileAction creates a new ReadFileAction with the given logger
func NewReadFileAction(logger *slog.Logger) *ReadFileAction {
	return &ReadFileAction{
		BaseAction: engine.NewBaseAction(logger),
	}
}

// WithParameters sets the parameters for file path and output buffer
func (a *ReadFileAction) WithParameters(pathParam engine.ActionParameter, outputBuffer *[]byte) (*engine.Action[*ReadFileAction], error) {
	if outputBuffer == nil {
		return nil, fmt.Errorf("invalid parameter: outputBuffer cannot be nil")
	}

	a.PathParam = pathParam
	a.OutputBuffer = outputBuffer

	return &engine.Action[*ReadFileAction]{
		ID:      "read-file-action",
		Name:    "Read File",
		Wrapped: a,
	}, nil
}

// ReadFileAction reads content from a file and stores it in the provided buffer
type ReadFileAction struct {
	engine.BaseAction
	FilePath     string
	OutputBuffer *[]byte                // Pointer to buffer where file contents will be stored
	PathParam    engine.ActionParameter // optional parameter to resolve path
}

func (a *ReadFileAction) Execute(execCtx context.Context) error {
	// Resolve path parameter if provided
	effectivePath := a.FilePath
	if a.PathParam != nil {
		if gc, ok := execCtx.Value(engine.GlobalContextKey).(*engine.GlobalContext); ok {
			v, err := a.PathParam.Resolve(execCtx, gc)
			if err != nil {
				return fmt.Errorf("failed to resolve path parameter: %w", err)
			}
			if s, ok := v.(string); ok {
				effectivePath = s
			} else {
				return fmt.Errorf("resolved path parameter is not a string: %T", v)
			}
		} else {
			if sp, ok := a.PathParam.(engine.StaticParameter); ok {
				if s, ok2 := sp.Value.(string); ok2 {
					effectivePath = s
				} else {
					return fmt.Errorf("static path parameter is not a string: %T", sp.Value)
				}
			} else {
				return fmt.Errorf("global context not available for dynamic path resolution")
			}
		}
	}

	// Sanitize path to prevent path traversal attacks
	sanitizedPath, err := SanitizePath(effectivePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	a.Logger.Info("Attempting to read file", "path", sanitizedPath)
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

// GetOutput returns the file content and metadata
func (a *ReadFileAction) GetOutput() interface{} {
	if a.OutputBuffer == nil {
		return map[string]interface{}{
			"content":  nil,
			"fileSize": 0,
			"filePath": a.FilePath,
			"success":  false,
		}
	}

	return map[string]interface{}{
		"content":  *a.OutputBuffer,
		"fileSize": len(*a.OutputBuffer),
		"filePath": a.FilePath,
		"success":  true,
	}
}
