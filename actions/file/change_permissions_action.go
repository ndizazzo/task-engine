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

// NewChangePermissionsAction creates a new ChangePermissionsAction with the given logger
func NewChangePermissionsAction(logger *slog.Logger) *ChangePermissionsAction {
	return &ChangePermissionsAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		commandRunner:     command.NewDefaultCommandRunner(),
	}
}

type ChangePermissionsAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder

	// Parameters
	PathParam task_engine.ActionParameter
	ModeParam task_engine.ActionParameter
	Recursive bool

	// Runtime resolved values
	Path        string
	Permissions string

	commandRunner command.CommandRunner
}

// WithParameters sets the parameters for permission change and returns a wrapped Action
func (a *ChangePermissionsAction) WithParameters(
	pathParam task_engine.ActionParameter,
	modeParam task_engine.ActionParameter,
	recursive bool,
) (*task_engine.Action[*ChangePermissionsAction], error) {
	a.PathParam = pathParam
	a.ModeParam = modeParam
	a.Recursive = recursive

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*ChangePermissionsAction](a.Logger)
	return constructor.WrapAction(a, "Change Permissions", "change-permissions-action"), nil
}

func (a *ChangePermissionsAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *ChangePermissionsAction) Execute(execCtx context.Context) error {
	// Resolve parameters using the ParameterResolver
	if a.PathParam != nil {
		pathValue, err := a.ResolveStringParameter(execCtx, a.PathParam, "path")
		if err != nil {
			return err
		}
		a.Path = pathValue
	}

	if a.ModeParam != nil {
		modeValue, err := a.ResolveStringParameter(execCtx, a.ModeParam, "permissions")
		if err != nil {
			return err
		}
		a.Permissions = modeValue
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
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"path":        a.Path,
		"permissions": a.Permissions,
		"recursive":   a.Recursive,
	})
}
