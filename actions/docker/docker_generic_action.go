package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerGenericAction creates an action to run an arbitrary docker command
func NewDockerGenericAction(logger *slog.Logger, dockerCmd ...string) *task_engine.Action[*DockerGenericAction] {
	id := fmt.Sprintf("docker-generic-%s-action", strings.Join(dockerCmd, "-"))
	return &task_engine.Action[*DockerGenericAction]{
		ID: id,
		Wrapped: &DockerGenericAction{
			BaseAction:       task_engine.BaseAction{Logger: logger},
			DockerCmd:        dockerCmd,
			CommandProcessor: command.NewDefaultCommandRunner(),
		},
	}
}

// DockerGenericAction runs a generic docker command and stores its output
// NOTE: This is desgiend to be pretty simple... more advanced stuff with error handling for specific docker commands
// should be separate actions
type DockerGenericAction struct {
	task_engine.BaseAction
	DockerCmd        []string
	CommandProcessor command.CommandRunner
	Output           string
}

func (a *DockerGenericAction) Execute(execCtx context.Context) error {
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
