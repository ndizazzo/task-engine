package file

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

// NewReadFileAction creates a new ReadFileAction with the given logger
func NewReadFileAction(logger *slog.Logger) *ReadFileAction {
	return &ReadFileAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for file path and output buffer
func (a *ReadFileAction) WithParameters(
	pathParam task_engine.ActionParameter,
	outputBuffer *[]byte,
) (*task_engine.Action[*ReadFileAction], error) {
	if outputBuffer == nil {
		return nil, fmt.Errorf("invalid parameter: outputBuffer cannot be nil")
	}

	a.PathParam = pathParam
	a.OutputBuffer = outputBuffer

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*ReadFileAction](a.Logger)
	return constructor.WrapAction(a, "Read File", "read-file-action"), nil
}

// ReadFileAction reads content from a file and stores it in the provided buffer
type ReadFileAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	FilePath     string
	OutputBuffer *[]byte                     // Pointer to buffer where file contents will be stored
	PathParam    task_engine.ActionParameter // optional parameter to resolve path
}

func (a *ReadFileAction) Execute(execCtx context.Context) error {
	// Resolve path parameter if provided using the ParameterResolver
	effectivePath := a.FilePath
	if a.PathParam != nil {
		pathValue, err := a.ResolveStringParameter(execCtx, a.PathParam, "path")
		if err != nil {
			return err
		}
		effectivePath = pathValue
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
		return a.BuildStandardOutput(nil, false, map[string]interface{}{
			"content":  nil,
			"fileSize": 0,
			"filePath": a.FilePath,
		})
	}

	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"content":  *a.OutputBuffer,
		"fileSize": len(*a.OutputBuffer),
		"filePath": a.FilePath,
	})
}
