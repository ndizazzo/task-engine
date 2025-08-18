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

type DockerComposeDownAction struct {
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
func (a *DockerComposeDownAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

// NewDockerComposeDownAction creates the action instance (modern constructor)
func NewDockerComposeDownAction(logger *slog.Logger) *DockerComposeDownAction {
	return &DockerComposeDownAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		commandRunner:     command.NewDefaultCommandRunner(),
	}
}

// WithParameters sets inputs and returns the wrapped action
func (a *DockerComposeDownAction) WithParameters(workingDirParam, servicesParam task_engine.ActionParameter) (*task_engine.Action[*DockerComposeDownAction], error) {
	if workingDirParam == nil || servicesParam == nil {
		return nil, fmt.Errorf("parameters cannot be nil")
	}
	a.WorkingDirParam = workingDirParam
	a.ServicesParam = servicesParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*DockerComposeDownAction](a.Logger)
	return constructor.WrapAction(a, "Docker Compose Down", "docker-compose-down-action"), nil
}

func (a *DockerComposeDownAction) Execute(execCtx context.Context) error {
	// Resolve working directory parameter
	var workingDir string
	if a.WorkingDirParam != nil {
		workingDirValue, err := a.ResolveStringParameter(execCtx, a.WorkingDirParam, "working directory")
		if err != nil {
			return err
		}
		workingDir = workingDirValue
	}

	// Resolve services parameter
	var services []string
	if a.ServicesParam != nil {
		servicesValue, err := a.ResolveParameter(execCtx, a.ServicesParam, "services")
		if err != nil {
			return err
		}
		if servicesSlice, ok := servicesValue.([]string); ok {
			services = servicesSlice
		} else if servicesStr, ok := servicesValue.(string); ok {
			// If it's a single string, split by comma or space
			if strings.Contains(servicesStr, ",") {
				services = strings.Split(servicesStr, ",")
			} else {
				services = strings.Fields(servicesStr)
			}
		} else {
			return fmt.Errorf("services parameter is not a string slice or string, got %T", servicesValue)
		}
	}

	args := []string{"compose", "down"}
	if len(services) > 0 {
		args = append(args, services...)
	}

	a.Logger.Info("Executing docker compose down", "services", services, "workingDir", workingDir)

	var output string
	var err error
	if workingDir != "" {
		output, err = a.commandRunner.RunCommandInDirWithContext(execCtx, workingDir, "docker", args...)
	} else {
		output, err = a.commandRunner.RunCommandWithContext(execCtx, "docker", args...)
	}

	if err != nil {
		a.Logger.Error("Failed to run docker compose down", "error", err, "output", output, "services", services)
		return fmt.Errorf("failed to run docker compose down for services %v: %w. Output: %s", services, err, output)
	}

	a.Logger.Info("Docker compose down finished successfully", "output", output)
	return nil
}

// GetOutput returns details about the compose down execution
func (a *DockerComposeDownAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, nil)
}
