package tasks

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
)

// ExampleParameterPassingTask demonstrates the new parameter passing system
// This task reads a file, processes its content, and writes the result to a new file
func NewExampleParameterPassingTask(config ExampleParameterPassingConfig, logger *slog.Logger) *engine.Task {
	return &engine.Task{
		ID:   "example-parameter-passing",
		Name: "Example Parameter Passing Between Actions",
		Actions: []engine.ActionWrapper{
			// Action 1: Read source file - will produce output with file content
			createReadFileAction(config.SourcePath, logger),

			// Action 2: Process content using output from Action 1
			createProcessContentAction(logger),

			// Action 3: Write processed content to destination using output from Action 2
			createWriteFileAction(config.DestinationPath, logger),
		},
		Logger: logger,
	}
}

// ExampleParameterPassingConfig holds configuration for the parameter passing example
type ExampleParameterPassingConfig struct {
	SourcePath      string
	DestinationPath string
}

// createReadFileAction creates a read file action with a specific ID for parameter reference
func createReadFileAction(filePath string, logger *slog.Logger) *engine.Action[*file.ReadFileAction] {
	// Create a buffer to store the file content
	var outputBuffer []byte
	action, err := file.NewReadFileAction(logger).WithParameters(engine.StaticParameter{Value: filePath}, &outputBuffer)
	if err != nil {
		logger.Error("Failed to create read file action", "error", err)
		panic(err)
	}
	// Set a specific ID for parameter reference
	action.ID = "read-source-file"
	return action
}

// createProcessContentAction creates an action that processes content from the read action
func createProcessContentAction(logger *slog.Logger) engine.ActionWrapper {
	// Create a custom action that processes content in memory
	// This action will use the content from the read action
	action := &engine.Action[*ContentProcessingAction]{
		ID: "process-content",
		Wrapped: &ContentProcessingAction{
			BaseAction: engine.BaseAction{Logger: logger},
		},
	}
	return action
}

// ContentProcessingAction is a custom action that processes content in memory
type ContentProcessingAction struct {
	engine.BaseAction
	processedContent []byte
	processingError  error
}

func (a *ContentProcessingAction) Execute(ctx context.Context) error {
	// Get the global context from the context
	globalCtx, ok := ctx.Value(engine.GlobalContextKey).(*engine.GlobalContext)
	if !ok {
		a.processingError = fmt.Errorf("global context not found in context")
		return a.processingError
	}

	// Get the content from the read action
	readOutput, exists := globalCtx.ActionOutputs["read-source-file"]
	if !exists {
		a.processingError = fmt.Errorf("read action output not found")
		return a.processingError
	}

	// Extract the content from the read action output
	readOutputMap, ok := readOutput.(map[string]interface{})
	if !ok {
		a.processingError = fmt.Errorf("read action output is not a map")
		return a.processingError
	}

	content, exists := readOutputMap["content"]
	if !exists {
		a.processingError = fmt.Errorf("content field not found in read action output")
		return a.processingError
	}

	// Process the content (simple example: convert to uppercase)
	switch v := content.(type) {
	case []byte:
		a.processedContent = bytes.ToUpper(v)
	case string:
		a.processedContent = []byte(strings.ToUpper(v))
	default:
		a.processingError = fmt.Errorf("unsupported content type: %T", content)
		return a.processingError
	}

	a.Logger.Info("Content processed successfully", "originalLength", getContentLength(content), "processedLength", len(a.processedContent))
	return nil
}

// getContentLength safely gets the length of content regardless of its type
func getContentLength(content interface{}) int {
	switch v := content.(type) {
	case []byte:
		return len(v)
	case string:
		return len(v)
	default:
		return 0
	}
}

func (a *ContentProcessingAction) GetOutput() interface{} {
	return map[string]interface{}{
		"processedContent": a.processedContent,
		"success":          a.processingError == nil,
		"error":            a.processingError,
	}
}

// createWriteFileAction creates a write file action that uses content from the read action
func createWriteFileAction(destinationPath string, logger *slog.Logger) *engine.Action[*file.WriteFileAction] {
	// This action will use the processed content from the content processing action
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: destinationPath},
		engine.ActionOutputField("process-content", "processedContent"),
		true,
		nil,
	)
	if err != nil {
		logger.Error("Failed to create write file action", "error", err)
		panic(err)
	}
	action.ID = "write-destination-file"
	return action
}

// ExampleCrossTaskParameterPassing demonstrates parameter passing between different tasks
func NewExampleCrossTaskParameterPassing(config CrossTaskConfig, logger *slog.Logger) *engine.Task {
	return &engine.Task{
		ID:      "example-cross-task-parameter-passing",
		Name:    "Example Cross-Task Parameter Passing",
		Actions: []engine.ActionWrapper{
			// This task will reference outputs from other tasks
			// Implementation will be enhanced in future iterations
		},
		Logger: logger,
	}
}

// CrossTaskConfig holds configuration for cross-task parameter passing
type CrossTaskConfig struct {
	SourceTaskID      string
	DestinationTaskID string
}
