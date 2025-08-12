package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerComposeExecAction creates an action instance (modern constructor pattern)
func NewDockerComposeExecAction(logger *slog.Logger) *DockerComposeExecAction {
	return &DockerComposeExecAction{
		BaseAction:    task_engine.BaseAction{Logger: logger},
		commandRunner: command.NewDefaultCommandRunner(),
	}
}

type DockerComposeExecAction struct {
	task_engine.BaseAction
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

// WithParameters validates and attaches parameters, returning the wrapped action
func (a *DockerComposeExecAction) WithParameters(
	workingDirParam task_engine.ActionParameter,
	serviceParam task_engine.ActionParameter,
	commandArgsParam task_engine.ActionParameter,
) (*task_engine.Action[*DockerComposeExecAction], error) {
	if workingDirParam == nil || serviceParam == nil || commandArgsParam == nil {
		return nil, fmt.Errorf("parameters cannot be nil")
	}
	a.WorkingDirParam = workingDirParam
	a.ServiceParam = serviceParam
	a.CommandArgsParam = commandArgsParam

	return &task_engine.Action[*DockerComposeExecAction]{
		ID:      "docker-compose-exec-action",
		Name:    "Docker Compose Exec",
		Wrapped: a,
	}, nil
}

func (a *DockerComposeExecAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve working directory parameter
	workingDirValue, err := a.WorkingDirParam.Resolve(execCtx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve working directory parameter: %w", err)
	}
	if workingDirStr, ok := workingDirValue.(string); ok {
		a.ResolvedWorkingDir = workingDirStr
	} else {
		return fmt.Errorf("working directory parameter is not a string, got %T", workingDirValue)
	}

	// Resolve service parameter
	serviceValue, err := a.ServiceParam.Resolve(execCtx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve service parameter: %w", err)
	}
	if serviceStr, ok := serviceValue.(string); ok {
		a.ResolvedService = serviceStr
	} else {
		return fmt.Errorf("service parameter is not a string, got %T", serviceValue)
	}

	// Resolve command arguments parameter
	commandArgsValue, err := a.CommandArgsParam.Resolve(execCtx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve command arguments parameter: %w", err)
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
	return map[string]interface{}{
		"service":    a.ResolvedService,
		"workingDir": a.ResolvedWorkingDir,
		"command":    a.ResolvedCommandArgs,
		"success":    true,
	}
}
