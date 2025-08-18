package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

// NewMoveFileAction creates a new MoveFileAction with the given logger
func NewMoveFileAction(logger *slog.Logger) *MoveFileAction {
	if logger == nil {
		logger = slog.Default()
	}
	return &MoveFileAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		commandRunner:     command.NewDefaultCommandRunner(),
	}
}

// WithParameters sets the parameters for source, destination, and create directories flag
func (a *MoveFileAction) WithParameters(
	sourceParam task_engine.ActionParameter,
	destinationParam task_engine.ActionParameter,
	createDirs bool,
) (*task_engine.Action[*MoveFileAction], error) {
	a.SourceParam = sourceParam
	a.DestinationParam = destinationParam
	a.CreateDirs = createDirs

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*MoveFileAction](a.Logger)
	return constructor.WrapAction(a, "Move File", "move-file-action"), nil
}

// MoveFileAction moves a file from source to destination
type MoveFileAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	Source      string
	Destination string
	CreateDirs  bool

	// Parameter-aware fields
	SourceParam      task_engine.ActionParameter
	DestinationParam task_engine.ActionParameter

	// Execution dependency
	commandRunner command.CommandRunner
}

func (a *MoveFileAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *MoveFileAction) Execute(execCtx context.Context) error {
	// Resolve parameters using the ParameterResolver
	if a.SourceParam != nil {
		sourceValue, err := a.ResolveStringParameter(execCtx, a.SourceParam, "source")
		if err != nil {
			return err
		}
		a.Source = sourceValue
	}

	if a.DestinationParam != nil {
		destValue, err := a.ResolveStringParameter(execCtx, a.DestinationParam, "destination")
		if err != nil {
			return err
		}
		a.Destination = destValue
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
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"source":      a.Source,
		"destination": a.Destination,
		"createDirs":  a.CreateDirs,
	})
}
