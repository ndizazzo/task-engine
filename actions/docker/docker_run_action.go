package docker

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerRunAction creates an action to run a docker container
// Optionally accepts a buffer to write the command's stdout to.
func NewDockerRunAction(logger *slog.Logger, image string, outputBuffer *bytes.Buffer, runArgs ...string) *task_engine.Action[*DockerRunAction] {
	id := fmt.Sprintf("docker-run-%s-action", image)
	return &task_engine.Action[*DockerRunAction]{
		ID: id,
		Wrapped: &DockerRunAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			Image:         image,
			RunArgs:       runArgs,
			OutputBuffer:  outputBuffer, // Store buffer pointer
			commandRunner: command.NewDefaultCommandRunner(),
		},
	}
}

// NOTE: Command arguments for inside the container should be part of RunArgs
type DockerRunAction struct {
	task_engine.BaseAction
	Image         string
	RunArgs       []string
	commandRunner command.CommandRunner
	Output        string        // Stores trimmed output regardless of buffer
	OutputBuffer  *bytes.Buffer // Optional buffer to write output to
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *DockerRunAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *DockerRunAction) Execute(execCtx context.Context) error {
	args := []string{"run"}
	args = append(args, a.RunArgs...)

	a.Logger.Info("Executing docker run", "image", a.Image, "args", a.RunArgs)
	output, err := a.commandRunner.RunCommand("docker", args...)
	a.Output = strings.TrimSpace(output) // Store trimmed output internally

	// Write to buffer if provided
	if a.OutputBuffer != nil {
		a.OutputBuffer.Reset()
		_, writeErr := a.OutputBuffer.WriteString(output)
		if writeErr != nil {
			a.Logger.Error("Failed to write command output to shared buffer", "error", writeErr)
		}
	}

	if err != nil {
		a.Logger.Error("Failed to run docker container", "error", err, "output", output)
		return fmt.Errorf("failed to run docker container %s with args %v: %w. Output: %s", a.Image, a.RunArgs, err, output)
	}
	a.Logger.Info("Docker run finished successfully", "output", a.Output)

	return nil
}
