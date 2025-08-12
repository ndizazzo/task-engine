package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewChangePermissionsAction creates a new ChangePermissionsAction with the given logger
func NewChangePermissionsAction(logger *slog.Logger) *ChangePermissionsAction {
	if logger == nil {
		logger = slog.Default()
	}
	return &ChangePermissionsAction{
		BaseAction:    task_engine.NewBaseAction(logger),
		commandRunner: command.NewDefaultCommandRunner(),
	}
}

type ChangePermissionsAction struct {
	task_engine.BaseAction

	// Parameters
	PathParam        task_engine.ActionParameter
	PermissionsParam task_engine.ActionParameter
	Recursive        bool

	// Runtime resolved values
	Path        string
	Permissions string

	commandRunner command.CommandRunner
}

// WithParameters sets the parameters for path and permissions and returns a wrapped Action
func (a *ChangePermissionsAction) WithParameters(pathParam, permissionsParam task_engine.ActionParameter, recursive bool) (*task_engine.Action[*ChangePermissionsAction], error) {
	a.PathParam = pathParam
	a.PermissionsParam = permissionsParam
	a.Recursive = recursive

	id := "change-permissions-action"
	return &task_engine.Action[*ChangePermissionsAction]{
		ID:      id,
		Name:    "Change Permissions",
		Wrapped: a,
	}, nil
}

func (a *ChangePermissionsAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *ChangePermissionsAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve parameters if they exist
	if a.PathParam != nil {
		pathValue, err := a.PathParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve path parameter: %w", err)
		}
		if pathStr, ok := pathValue.(string); ok {
			a.Path = pathStr
		} else {
			return fmt.Errorf("path parameter is not a string, got %T", pathValue)
		}
	}

	if a.PermissionsParam != nil {
		permissionsValue, err := a.PermissionsParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve permissions parameter: %w", err)
		}
		if permissionsStr, ok := permissionsValue.(string); ok {
			a.Permissions = permissionsStr
		} else {
			return fmt.Errorf("permissions parameter is not a string, got %T", permissionsValue)
		}
	}

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

// GetOutput returns metadata about the permission change
func (a *ChangePermissionsAction) GetOutput() interface{} {
	return map[string]interface{}{
		"path":        a.Path,
		"permissions": a.Permissions,
		"recursive":   a.Recursive,
		"success":     true,
	}
}
