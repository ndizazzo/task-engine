package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewMoveFileAction creates a new MoveFileAction with the given logger
func NewMoveFileAction(logger *slog.Logger) *MoveFileAction {
	if logger == nil {
		logger = slog.Default()
	}
	return &MoveFileAction{
		BaseAction:    task_engine.NewBaseAction(logger),
		commandRunner: command.NewDefaultCommandRunner(),
	}
}

// WithParameters sets the parameters for source, destination, and create directories flag
func (a *MoveFileAction) WithParameters(sourceParam, destinationParam task_engine.ActionParameter, createDirs bool) (*task_engine.Action[*MoveFileAction], error) {
	a.SourceParam = sourceParam
	a.DestinationParam = destinationParam
	a.CreateDirs = createDirs

	return &task_engine.Action[*MoveFileAction]{
		ID:      "move-file-action",
		Name:    "Move File",
		Wrapped: a,
	}, nil
}

type MoveFileAction struct {
	task_engine.BaseAction
	Source        string
	Destination   string
	CreateDirs    bool
	commandRunner command.CommandRunner

	// Parameter-aware fields
	SourceParam      task_engine.ActionParameter
	DestinationParam task_engine.ActionParameter
}

func (a *MoveFileAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *MoveFileAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve parameters if they exist
	if a.SourceParam != nil {
		sourceValue, err := a.SourceParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve source parameter: %w", err)
		}
		if sourceStr, ok := sourceValue.(string); ok {
			a.Source = sourceStr
		} else {
			return fmt.Errorf("source parameter is not a string, got %T", sourceValue)
		}
	}

	if a.DestinationParam != nil {
		destValue, err := a.DestinationParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve destination parameter: %w", err)
		}
		if destStr, ok := destValue.(string); ok {
			a.Destination = destStr
		} else {
			return fmt.Errorf("destination parameter is not a string, got %T", destValue)
		}
	}

	// Basic validations before any external calls
	if strings.TrimSpace(a.Source) == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	if strings.TrimSpace(a.Destination) == "" {
		return fmt.Errorf("destination path cannot be empty")
	}
	if a.Source == a.Destination {
		return fmt.Errorf("source and destination paths cannot be the same")
	}

	if _, err := os.Stat(a.Source); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", a.Source)
	}

	if a.CreateDirs {
		destDir := filepath.Dir(a.Destination)
		if err := os.MkdirAll(destDir, 0o750); err != nil {
			a.Logger.Error("Failed to create destination directory", "dir", destDir, "error", err)
			return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
		}
	}

	a.Logger.Info("Moving file/directory", "source", a.Source, "destination", a.Destination, "createDirs", a.CreateDirs)

	output, err := a.commandRunner.RunCommandWithContext(execCtx, "mv", a.Source, a.Destination)
	if err != nil {
		a.Logger.Error("Failed to move file/directory", "error", err, "output", output)
		return fmt.Errorf("failed to move %s to %s: %w. Output: %s", a.Source, a.Destination, err, output)
	}

	a.Logger.Info("Successfully moved file/directory", "source", a.Source, "destination", a.Destination)
	return nil
}

// GetOutput returns metadata about the move operation
func (a *MoveFileAction) GetOutput() interface{} {
	return map[string]interface{}{
		"source":      a.Source,
		"destination": a.Destination,
		"createDirs":  a.CreateDirs,
		"success":     true,
	}
}
