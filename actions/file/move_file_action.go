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

func NewMoveFileAction(source string, destination string, createDirs bool, logger *slog.Logger) *task_engine.Action[*MoveFileAction] {
	if logger == nil {
		logger = slog.Default()
	}
	if source == "" {
		return nil
	}
	if destination == "" {
		return nil
	}
	if source == destination {
		return nil
	}

	id := fmt.Sprintf("move-file-%s", strings.ReplaceAll(filepath.Base(source), "/", "-"))

	return &task_engine.Action[*MoveFileAction]{
		ID: id,
		Wrapped: &MoveFileAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			Source:        source,
			Destination:   destination,
			CreateDirs:    createDirs,
			commandRunner: command.NewDefaultCommandRunner(),
		},
	}
}

type MoveFileAction struct {
	task_engine.BaseAction
	Source        string
	Destination   string
	CreateDirs    bool
	commandRunner command.CommandRunner
}

func (a *MoveFileAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *MoveFileAction) Execute(execCtx context.Context) error {
	if _, err := os.Stat(a.Source); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", a.Source)
	}

	if a.CreateDirs {
		destDir := filepath.Dir(a.Destination)
		if err := os.MkdirAll(destDir, 0750); err != nil {
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
