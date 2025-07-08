package docker

import (
	"context"
	"fmt"
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerComposeExecAction creates an action to run docker compose exec
func NewDockerComposeExecAction(logger *slog.Logger, workingDir string, service string, cmdArgs ...string) *task_engine.Action[*DockerComposeExecAction] {
	id := fmt.Sprintf("docker-compose-exec-%s-%s-action", service, cmdArgs[0])
	return &task_engine.Action[*DockerComposeExecAction]{
		ID: id,
		Wrapped: &DockerComposeExecAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			WorkingDir:    workingDir,
			Service:       service,
			CommandArgs:   cmdArgs,
			commandRunner: command.NewDefaultCommandRunner(),
		},
	}
}

type DockerComposeExecAction struct {
	task_engine.BaseAction
	WorkingDir    string
	Service       string
	CommandArgs   []string
	commandRunner command.CommandRunner
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *DockerComposeExecAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *DockerComposeExecAction) Execute(execCtx context.Context) error {
	args := []string{"compose", "exec", a.Service}
	args = append(args, a.CommandArgs...)

	a.Logger.Info("Executing docker compose exec", "service", a.Service, "command", a.CommandArgs, "workingDir", a.WorkingDir)

	var output string
	var err error
	if a.WorkingDir != "" {
		output, err = a.commandRunner.RunCommandInDirWithContext(execCtx, a.WorkingDir, "docker", args...)
	} else {
		output, err = a.commandRunner.RunCommandWithContext(execCtx, "docker", args...)
	}

	if err != nil {
		a.Logger.Error("Failed to run docker compose exec", "error", err, "output", output)
		return fmt.Errorf("failed to run docker compose exec on service %s with command %v in dir %s: %w. Output: %s", a.Service, a.CommandArgs, a.WorkingDir, err, output)
	}
	a.Logger.Info("Docker compose exec finished successfully", "output", output)
	return nil
}
