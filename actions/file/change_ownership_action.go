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

func NewChangeOwnershipAction(path string, owner string, group string, recursive bool, logger *slog.Logger) *task_engine.Action[*ChangeOwnershipAction] {
	if logger == nil {
		logger = slog.Default()
	}
	if path == "" {
		return nil
	}
	if owner == "" && group == "" {
		return nil
	}

	id := fmt.Sprintf("change-ownership-%s", strings.ReplaceAll(path, "/", "-"))

	return &task_engine.Action[*ChangeOwnershipAction]{
		ID: id,
		Wrapped: &ChangeOwnershipAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			Path:          path,
			Owner:         owner,
			Group:         group,
			Recursive:     recursive,
			commandRunner: command.NewDefaultCommandRunner(),
		},
	}
}

type ChangeOwnershipAction struct {
	task_engine.BaseAction
	Path          string
	Owner         string
	Group         string
	Recursive     bool
	commandRunner command.CommandRunner
}

func (a *ChangeOwnershipAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *ChangeOwnershipAction) Execute(execCtx context.Context) error {
	if _, err := os.Stat(a.Path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", a.Path)
	}

	var ownerSpec string
	switch {
	case a.Owner != "" && a.Group != "":
		ownerSpec = fmt.Sprintf("%s:%s", a.Owner, a.Group)
	case a.Owner != "":
		ownerSpec = a.Owner
	default:
		ownerSpec = fmt.Sprintf(":%s", a.Group)
	}

	args := []string{ownerSpec, a.Path}
	if a.Recursive {
		args = append([]string{"-R"}, args...)
	}

	a.Logger.Info("Changing ownership", "path", a.Path, "owner", a.Owner, "group", a.Group, "recursive", a.Recursive)

	output, err := a.commandRunner.RunCommandWithContext(execCtx, "chown", args...)
	if err != nil {
		a.Logger.Error("Failed to change ownership", "error", err, "output", output)
		return fmt.Errorf("failed to change ownership of %s to %s: %w. Output: %s", a.Path, ownerSpec, err, output)
	}

	a.Logger.Info("Successfully changed ownership", "path", a.Path, "owner", ownerSpec)
	return nil
}
