package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

func NewChangePermissionsAction(path string, permissions string, recursive bool, logger *slog.Logger) *task_engine.Action[*ChangePermissionsAction] {
	if logger == nil {
		logger = slog.Default()
	}
	if path == "" {
		return nil
	}
	if permissions == "" {
		return nil
	}

	id := fmt.Sprintf("change-permissions-%s", strings.ReplaceAll(path, "/", "-"))

	return &task_engine.Action[*ChangePermissionsAction]{
		ID: id,
		Wrapped: &ChangePermissionsAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			Path:          path,
			Permissions:   permissions,
			Recursive:     recursive,
			commandRunner: command.NewDefaultCommandRunner(),
		},
	}
}

type ChangePermissionsAction struct {
	task_engine.BaseAction
	Path          string
	Permissions   string
	Recursive     bool
	commandRunner command.CommandRunner
}

func (a *ChangePermissionsAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *ChangePermissionsAction) Execute(execCtx context.Context) error {
	if _, err := os.Stat(a.Path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", a.Path)
	}

	args := []string{a.Permissions, a.Path}
	if a.Recursive {
		args = append([]string{"-R"}, args...)
	}

	a.Logger.Info("Changing permissions", "path", a.Path, "permissions", a.Permissions, "recursive", a.Recursive)

	output, err := a.commandRunner.RunCommandWithContext(execCtx, "chmod", args...)
	if err != nil {
		a.Logger.Error("Failed to change permissions", "error", err, "output", output)
		return fmt.Errorf("failed to change permissions of %s to %s: %w. Output: %s", a.Path, a.Permissions, err, output)
	}

	a.Logger.Info("Successfully changed permissions", "path", a.Path, "permissions", a.Permissions)
	return nil
}
