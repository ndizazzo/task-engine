package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// DockerGenericActionConstructor provides the new constructor pattern
type DockerGenericActionConstructor struct {
	logger *slog.Logger
}

// NewDockerGenericAction creates a new DockerGenericAction constructor
func NewDockerGenericAction(logger *slog.Logger) *DockerGenericActionConstructor {
	return &DockerGenericActionConstructor{
		logger: logger,
	}
}

// WithParameters creates a DockerGenericAction with the specified parameters
func (c *DockerGenericActionConstructor) WithParameters(
	dockerCmdParam task_engine.ActionParameter,
) (*task_engine.Action[*DockerGenericAction], error) {
	action := &DockerGenericAction{
		BaseAction:       task_engine.NewBaseAction(c.logger),
		DockerCmd:        []string{},
		CommandProcessor: command.NewDefaultCommandRunner(),
		DockerCmdParam:   dockerCmdParam,
	}

	id := "docker-generic-action"
	return &task_engine.Action[*DockerGenericAction]{
		ID:      id,
		Name:    "Docker Generic",
		Wrapped: action,
	}, nil
}

// DockerGenericAction runs a generic docker command and stores its output
// NOTE: This is designed to be pretty simple... more advanced stuff with error handling for specific docker commands
// should be separate actions
type DockerGenericAction struct {
	task_engine.BaseAction
	DockerCmd        []string
	CommandProcessor command.CommandRunner
	Output           string

	// Parameter-aware fields
	DockerCmdParam task_engine.ActionParameter
}

func (a *DockerGenericAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve docker command parameter if it exists
	if a.DockerCmdParam != nil {
		dockerCmdValue, err := a.DockerCmdParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve docker command parameter: %w", err)
		}
		if dockerCmdSlice, ok := dockerCmdValue.([]string); ok {
			a.DockerCmd = dockerCmdSlice
		} else if dockerCmdStr, ok := dockerCmdValue.(string); ok {
			// If it's a single string, split by space
			a.DockerCmd = strings.Fields(dockerCmdStr)
		} else {
			return fmt.Errorf("docker command parameter is not a string slice or string, got %T", dockerCmdValue)
		}
	}

	a.Logger.Info("Executing docker command", "command", a.DockerCmd)
	output, err := a.CommandProcessor.RunCommand("docker", a.DockerCmd...)
	a.Output = strings.TrimSpace(output)

	if err != nil {
		a.Logger.Error("Failed to run docker command", "error", err, "output", output)
		return fmt.Errorf("failed to run docker command %v: %w. Output: %s", a.DockerCmd, err, output)
	}
	a.Logger.Info("Docker command finished successfully", "output", a.Output)
	return nil
}

// GetOutput returns the raw output and command metadata
func (a *DockerGenericAction) GetOutput() interface{} {
	return map[string]interface{}{
		"command": a.DockerCmd,
		"output":  a.Output,
		"success": a.Output != "",
	}
}
