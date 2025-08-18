package docker

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

// DockerRunActionBuilder provides the new constructor pattern
type DockerRunActionBuilder struct {
	common.BaseConstructor[*DockerRunAction]
}

// NewDockerRunAction creates a new DockerRunAction builder
func NewDockerRunAction(logger *slog.Logger) *DockerRunActionBuilder {
	return &DockerRunActionBuilder{
		BaseConstructor: *common.NewBaseConstructor[*DockerRunAction](logger),
	}
}

// WithParameters creates a DockerRunAction with the specified parameters
func (b *DockerRunActionBuilder) WithParameters(
	imageParam task_engine.ActionParameter,
	outputBuffer *bytes.Buffer,
	runArgs ...string,
) (*task_engine.Action[*DockerRunAction], error) {
	action := &DockerRunAction{
		BaseAction:    task_engine.NewBaseAction(b.GetLogger()),
		Image:         "",
		RunArgs:       runArgs,
		commandRunner: command.NewDefaultCommandRunner(),
		Output:        "",
		OutputBuffer:  outputBuffer,
		ImageParam:    imageParam,
	}

	return b.WrapAction(action, "Docker Run", "docker-run-action"), nil
}

// NOTE: Command arguments for inside the container should be part of RunArgs
type DockerRunAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	Image         string
	RunArgs       []string
	commandRunner command.CommandRunner
	Output        string                      // Stores trimmed output regardless of buffer
	OutputBuffer  *bytes.Buffer               // Optional buffer to write output to
	ImageParam    task_engine.ActionParameter // optional parameter for image
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *DockerRunAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *DockerRunAction) Execute(execCtx context.Context) error {
	// Resolve image via parameter if provided
	effectiveImage := a.Image
	if a.ImageParam != nil {
		s, err := a.ResolveStringParameter(execCtx, a.ImageParam, "image")
		if err != nil {
			return err
		}
		effectiveImage = s
	}

	args := []string{"run"}
	args = append(args, a.RunArgs...)

	a.Logger.Info("Executing docker run", "image", effectiveImage, "args", a.RunArgs)
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

// GetOutput returns information about the docker run execution
func (a *DockerRunAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, a.Output != "", map[string]interface{}{
		"image":  a.Image,
		"args":   a.RunArgs,
		"output": a.Output,
	})
}
