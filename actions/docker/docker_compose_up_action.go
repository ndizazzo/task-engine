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

type DockerComposeUpAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	// Parameter-only inputs
	WorkingDirParam task_engine.ActionParameter
	ServicesParam   task_engine.ActionParameter

	// Execution dependency
	commandRunner command.CommandRunner

	// Resolved/output fields
	ResolvedWorkingDir string
	ResolvedServices   []string
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *DockerComposeUpAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

// NewDockerComposeUpAction creates the action instance (modern constructor)
func NewDockerComposeUpAction(logger *slog.Logger) *DockerComposeUpAction {
	return &DockerComposeUpAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		commandRunner:     command.NewDefaultCommandRunner(),
	}
}

// WithParameters sets inputs and returns the wrapped action
func (a *DockerComposeUpAction) WithParameters(workingDirParam, servicesParam task_engine.ActionParameter) (*task_engine.Action[*DockerComposeUpAction], error) {
	if workingDirParam == nil || servicesParam == nil {
		return nil, fmt.Errorf("parameters cannot be nil")
	}
	a.WorkingDirParam = workingDirParam
	a.ServicesParam = servicesParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*DockerComposeUpAction](a.Logger)
	return constructor.WrapAction(a, "Docker Compose Up", "docker-compose-up-action"), nil
}

func (a *DockerComposeUpAction) Execute(execCtx context.Context) error {
	// Resolve working directory parameter
	workingDirValue, err := a.ResolveStringParameter(execCtx, a.WorkingDirParam, "working directory")
	if err != nil {
		return err
	}
	a.ResolvedWorkingDir = workingDirValue

	// Resolve services parameter
	servicesValue, err := a.ResolveParameter(execCtx, a.ServicesParam, "services")
	if err != nil {
		return err
	}
	if servicesSlice, ok := servicesValue.([]string); ok {
		a.ResolvedServices = servicesSlice
	} else if servicesStr, ok := servicesValue.(string); ok {
		if strings.Contains(servicesStr, ",") {
			a.ResolvedServices = strings.Split(servicesStr, ",")
		} else {
			a.ResolvedServices = strings.Fields(servicesStr)
		}
	} else {
		return fmt.Errorf("services parameter is not a string slice or string, got %T", servicesValue)
	}

	args := []string{"compose", "up", "-d"}
	args = append(args, a.ResolvedServices...)

	a.Logger.Info("Executing docker compose up", "services", a.ResolvedServices, "workingDir", a.ResolvedWorkingDir)

	var output string
	if a.ResolvedWorkingDir != "" {
		output, err = a.commandRunner.RunCommandInDirWithContext(execCtx, a.ResolvedWorkingDir, "docker", args...)
	} else {
		output, err = a.commandRunner.RunCommandWithContext(execCtx, "docker", args...)
	}

	if err != nil {
		a.Logger.Error("Failed to run docker compose up", "error", err, "output", output)
		return fmt.Errorf("failed to run docker compose up for services %v in dir %s: %w. Output: %s", a.ResolvedServices, a.ResolvedWorkingDir, err, output)
	}
	a.Logger.Info("Docker compose up finished successfully", "output", output)
	return nil
}

// GetOutput returns details about the compose up execution
func (a *DockerComposeUpAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"services":   a.ResolvedServices,
		"workingDir": a.ResolvedWorkingDir,
	})
}
