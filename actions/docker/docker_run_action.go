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

// DockerRunActionBuilder provides a fluent interface for building DockerRunAction
type DockerRunActionBuilder struct {
	logger       *slog.Logger
	imageParam   task_engine.ActionParameter
	outputBuffer *bytes.Buffer
	runArgs      []string
}

// NewDockerRunAction creates a fluent builder for DockerRunAction
func NewDockerRunAction(logger *slog.Logger) *DockerRunActionBuilder {
	return &DockerRunActionBuilder{
		logger: logger,
	}
}

// WithParameters sets the parameters for image, output buffer, and run arguments
func (b *DockerRunActionBuilder) WithParameters(imageParam task_engine.ActionParameter, outputBuffer *bytes.Buffer, runArgs ...string) (*task_engine.Action[*DockerRunAction], error) {
	b.imageParam = imageParam
	b.outputBuffer = outputBuffer
	b.runArgs = runArgs

	id := "docker-run-action"
	return &task_engine.Action[*DockerRunAction]{
		ID:   id,
		Name: "Docker Run",
		Wrapped: &DockerRunAction{
			BaseAction:    task_engine.NewBaseAction(b.logger),
			Image:         "",
			RunArgs:       b.runArgs,
			OutputBuffer:  b.outputBuffer,
			commandRunner: command.NewDefaultCommandRunner(),
			ImageParam:    b.imageParam,
		},
	}, nil
}

// NOTE: Command arguments for inside the container should be part of RunArgs
type DockerRunAction struct {
	task_engine.BaseAction
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
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve image via parameter if provided
	effectiveImage := a.Image
	if a.ImageParam != nil {
		v, err := a.ImageParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve image parameter: %w", err)
		}
		if s, ok := v.(string); ok {
			effectiveImage = s
		} else {
			return fmt.Errorf("resolved image parameter is not a string: %T", v)
		}
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
	return map[string]interface{}{
		"image":   a.Image,
		"args":    a.RunArgs,
		"output":  a.Output,
		"success": a.Output != "",
	}
}
