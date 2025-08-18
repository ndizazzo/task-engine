package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

// NewChangeOwnershipAction creates a new ChangeOwnershipAction with the given logger
func NewChangeOwnershipAction(logger *slog.Logger) *ChangeOwnershipAction {
	return &ChangeOwnershipAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		commandRunner:     command.NewDefaultCommandRunner(),
	}
}

type ChangeOwnershipAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder

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

// WithParameters sets the parameters for ownership change and returns a wrapped Action
func (a *ChangeOwnershipAction) WithParameters(
	pathParam task_engine.ActionParameter,
	ownerParam task_engine.ActionParameter,
	groupParam task_engine.ActionParameter,
	recursive bool,
) (*task_engine.Action[*ChangeOwnershipAction], error) {
	a.PathParam = pathParam
	a.OwnerParam = ownerParam
	a.GroupParam = groupParam
	a.Recursive = recursive

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*ChangeOwnershipAction](a.Logger)
	return constructor.WrapAction(a, "Change Ownership", "change-ownership-action"), nil
}

func (a *ChangeOwnershipAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *ChangeOwnershipAction) Execute(execCtx context.Context) error {
	// Resolve parameters using the ParameterResolver
	if a.PathParam != nil {
		pathValue, err := a.ResolveStringParameter(execCtx, a.PathParam, "path")
		if err != nil {
			return err
		}
		a.Path = pathValue
	}

	if a.OwnerParam != nil {
		ownerValue, err := a.ResolveStringParameter(execCtx, a.OwnerParam, "owner")
		if err != nil {
			return err
		}
		a.Owner = ownerValue
	}

	if a.GroupParam != nil {
		groupValue, err := a.ResolveStringParameter(execCtx, a.GroupParam, "group")
		if err != nil {
			return err
		}
		a.Group = groupValue
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
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"path":      a.Path,
		"owner":     a.Owner,
		"group":     a.Group,
		"recursive": a.Recursive,
	})
}
