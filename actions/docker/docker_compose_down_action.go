package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerComposeDownAction creates an action to run docker compose down, optionally for specific services.
// Modified to accept workingDir
func NewDockerComposeDownAction(logger *slog.Logger, workingDir string, services ...string) *task_engine.Action[*DockerComposeDownAction] {
	var id string
	if len(services) == 0 {
		id = "docker-compose-down-all-action"
	} else {
		// Use a hash or similar if service order matters for ID uniqueness, like in UpAction
		id = fmt.Sprintf("docker-compose-down-%s-action", strings.Join(services, "_"))
	}

	return &task_engine.Action[*DockerComposeDownAction]{
		ID: id,
		Wrapped: &DockerComposeDownAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			WorkingDir:    workingDir, // Store workingDir
			Services:      services,
			commandRunner: command.NewDefaultCommandRunner(), // Create default runner
		},
	}
}

// DockerComposeDownAction runs docker compose down
// It can target specific services or all services if none are provided.
type DockerComposeDownAction struct {
	task_engine.BaseAction
	WorkingDir    string // Added WorkingDir
	Services      []string
	commandRunner command.CommandRunner
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *DockerComposeDownAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *DockerComposeDownAction) Execute(execCtx context.Context) error {
	args := []string{"compose", "down"}
	if len(a.Services) > 0 {
		args = append(args, a.Services...)
	}

	a.Logger.Info("Executing docker compose down", "services", a.Services, "workingDir", a.WorkingDir)

	var output string
	var err error
	if a.WorkingDir != "" {
		output, err = a.commandRunner.RunCommandInDirWithContext(execCtx, a.WorkingDir, "docker", args...)
	} else {
		output, err = a.commandRunner.RunCommandWithContext(execCtx, "docker", args...)
	}

	if err != nil {
		a.Logger.Error("Failed to run docker compose down", "error", err, "output", output, "services", a.Services)
		return fmt.Errorf("failed to run docker compose down for services %v: %w. Output: %s", a.Services, err, output)
	}

	a.Logger.Info("Docker compose down finished successfully", "output", output)
	return nil
}
