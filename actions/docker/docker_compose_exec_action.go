package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerComposeExecAction creates a new DockerComposeExecAction with the given logger
func NewDockerComposeExecAction(logger *slog.Logger) *DockerComposeExecAction {
	return &DockerComposeExecAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		commandRunner:     command.NewDefaultCommandRunner(),
	}
}

type DockerComposeExecAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	// Parameter-only inputs
	WorkingDirParam  task_engine.ActionParameter
	ServiceParam     task_engine.ActionParameter
	CommandArgsParam task_engine.ActionParameter

	// Execution dependency
	commandRunner command.CommandRunner

	// Resolved/output fields
	ResolvedWorkingDir  string
	ResolvedService     string
	ResolvedCommandArgs []string
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *DockerComposeExecAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

// WithParameters sets the parameters for compose exec and returns a wrapped Action
func (a *DockerComposeExecAction) WithParameters(
	workingDirParam task_engine.ActionParameter,
	serviceParam task_engine.ActionParameter,
	commandParam task_engine.ActionParameter,
) (*task_engine.Action[*DockerComposeExecAction], error) {
	a.WorkingDirParam = workingDirParam
	a.ServiceParam = serviceParam
	a.CommandArgsParam = commandParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*DockerComposeExecAction](a.Logger)
	return constructor.WrapAction(a, "Docker Compose Exec", "docker-compose-exec-action"), nil
}

func (a *DockerComposeExecAction) Execute(execCtx context.Context) error {
	// Resolve working directory parameter
	workingDirValue, err := a.ResolveStringParameter(execCtx, a.WorkingDirParam, "working directory")
	if err != nil {
		return err
	}
	a.ResolvedWorkingDir = workingDirValue

	// Resolve service parameter
	serviceValue, err := a.ResolveStringParameter(execCtx, a.ServiceParam, "service")
	if err != nil {
		return err
	}
	a.ResolvedService = serviceValue

	// Resolve command arguments parameter
	commandArgsValue, err := a.ResolveParameter(execCtx, a.CommandArgsParam, "command arguments")
	if err != nil {
		return err
	}
	if commandArgsSlice, ok := commandArgsValue.([]string); ok {
		a.ResolvedCommandArgs = commandArgsSlice
	} else if commandArgsStr, ok := commandArgsValue.(string); ok {
		a.ResolvedCommandArgs = strings.Fields(commandArgsStr)
	} else {
		return fmt.Errorf("command arguments parameter is not a string slice or string, got %T", commandArgsValue)
	}

	args := []string{"compose", "exec", a.ResolvedService}
	args = append(args, a.ResolvedCommandArgs...)

	a.Logger.Info("Executing docker compose exec", "service", a.ResolvedService, "command", a.ResolvedCommandArgs, "workingDir", a.ResolvedWorkingDir)

	var output string
	if a.ResolvedWorkingDir != "" {
		output, err = a.commandRunner.RunCommandInDirWithContext(execCtx, a.ResolvedWorkingDir, "docker", args...)
	} else {
		output, err = a.commandRunner.RunCommandWithContext(execCtx, "docker", args...)
	}

	if err != nil {
		a.Logger.Error("Failed to run docker compose exec", "error", err, "output", output)
		return fmt.Errorf("failed to run docker compose exec on service %s with command %v in dir %s: %w. Output: %s", a.ResolvedService, a.ResolvedCommandArgs, a.ResolvedWorkingDir, err, output)
	}
	a.Logger.Info("Docker compose exec finished successfully", "output", output)
	return nil
}

// GetOutput returns details about the compose exec execution
func (a *DockerComposeExecAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"service":    a.ResolvedService,
		"workingDir": a.ResolvedWorkingDir,
		"command":    a.ResolvedCommandArgs,
	})
}
