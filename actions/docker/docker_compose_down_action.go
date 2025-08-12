package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerComposeDownAction creates a DockerComposeDownAction instance
func NewDockerComposeDownAction(logger *slog.Logger) *DockerComposeDownAction {
	return &DockerComposeDownAction{
		BaseAction:    task_engine.BaseAction{Logger: logger},
		commandRunner: command.NewDefaultCommandRunner(),
	}
}

// DockerComposeDownAction runs docker compose down
// It can target specific services or all services if none are provided.
type DockerComposeDownAction struct {
	task_engine.BaseAction
	commandRunner command.CommandRunner

	// Parameter-only fields
	WorkingDirParam task_engine.ActionParameter
	ServicesParam   task_engine.ActionParameter
}

// WithParameters sets the parameters and returns a wrapped Action
func (a *DockerComposeDownAction) WithParameters(workingDirParam, servicesParam task_engine.ActionParameter) (*task_engine.Action[*DockerComposeDownAction], error) {
	if workingDirParam == nil || servicesParam == nil {
		return nil, fmt.Errorf("parameters cannot be nil")
	}

	a.WorkingDirParam = workingDirParam
	a.ServicesParam = servicesParam

	return &task_engine.Action[*DockerComposeDownAction]{
		ID:      "docker-compose-down-action",
		Name:    "Docker Compose Down",
		Wrapped: a,
	}, nil
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *DockerComposeDownAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *DockerComposeDownAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve working directory parameter
	var workingDir string
	if a.WorkingDirParam != nil {
		workingDirValue, err := a.WorkingDirParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve working directory parameter: %w", err)
		}
		if workingDirStr, ok := workingDirValue.(string); ok {
			workingDir = workingDirStr
		} else {
			return fmt.Errorf("working directory parameter is not a string, got %T", workingDirValue)
		}
	}

	// Resolve services parameter
	var services []string
	if a.ServicesParam != nil {
		servicesValue, err := a.ServicesParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve services parameter: %w", err)
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
	return map[string]interface{}{
		"success": true,
	}
}
