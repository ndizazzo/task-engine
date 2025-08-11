package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewChangeOwnershipAction creates a new ChangeOwnershipAction with the given logger
func NewChangeOwnershipAction(logger *slog.Logger) *ChangeOwnershipAction {
	if logger == nil {
		logger = slog.Default()
	}
	return &ChangeOwnershipAction{
		BaseAction:    task_engine.NewBaseAction(logger),
		commandRunner: command.NewDefaultCommandRunner(),
	}
}

type ChangeOwnershipAction struct {
	task_engine.BaseAction

	// Parameters
	PathParam  task_engine.ActionParameter
	OwnerParam task_engine.ActionParameter
	GroupParam task_engine.ActionParameter
	Recursive  bool

	// Runtime resolved values
	Path  string
	Owner string
	Group string

	commandRunner command.CommandRunner
}

// WithParameters sets the parameters for path, owner, and group and returns a wrapped Action
func (a *ChangeOwnershipAction) WithParameters(pathParam, ownerParam, groupParam task_engine.ActionParameter, recursive bool) (*task_engine.Action[*ChangeOwnershipAction], error) {
	a.PathParam = pathParam
	a.OwnerParam = ownerParam
	a.GroupParam = groupParam
	a.Recursive = recursive

	id := "change-ownership-action"
	return &task_engine.Action[*ChangeOwnershipAction]{
		ID:      id,
		Name:    "Change Ownership",
		Wrapped: a,
	}, nil
}

func (a *ChangeOwnershipAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *ChangeOwnershipAction) Execute(execCtx context.Context) error {
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

	if a.OwnerParam != nil {
		ownerValue, err := a.OwnerParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve owner parameter: %w", err)
		}
		if ownerStr, ok := ownerValue.(string); ok {
			a.Owner = ownerStr
		} else {
			return fmt.Errorf("owner parameter is not a string, got %T", ownerValue)
		}
	}

	if a.GroupParam != nil {
		groupValue, err := a.GroupParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve group parameter: %w", err)
		}
		if groupStr, ok := groupValue.(string); ok {
			a.Group = groupStr
		} else {
			return fmt.Errorf("group parameter is not a string, got %T", groupValue)
		}
	}

	if a.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if a.Owner == "" && a.Group == "" {
		return fmt.Errorf("at least one of owner or group must be specified")
	}
	if _, err := os.Stat(a.Path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", a.Path)
	}

	// Build chown arguments
	var ownerSpec string
	switch {
	case a.Owner != "" && a.Group != "":
		ownerSpec = a.Owner + ":" + a.Group
	case a.Owner != "":
		ownerSpec = a.Owner
	default:
		ownerSpec = ":" + a.Group
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

	a.Logger.Info("Successfully changed ownership", "path", a.Path, "owner", a.Owner, "group", a.Group)
	return nil
}

// GetOutput returns metadata about the ownership change
func (a *ChangeOwnershipAction) GetOutput() interface{} {
	return map[string]interface{}{
		"path":      a.Path,
		"owner":     a.Owner,
		"group":     a.Group,
		"recursive": a.Recursive,
		"success":   true,
	}
}
