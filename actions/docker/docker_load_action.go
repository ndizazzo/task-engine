package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerLoadAction creates an action to load a Docker image from a tar archive file
func NewDockerLoadAction(logger *slog.Logger, tarFilePath string, options ...DockerLoadOption) *task_engine.Action[*DockerLoadAction] {
	id := fmt.Sprintf("docker-load-%s-action", strings.ReplaceAll(tarFilePath, "/", "-"))

	action := &DockerLoadAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		TarFilePath:      tarFilePath,
		Platform:         "",
		Quiet:            false,
		CommandProcessor: command.NewDefaultCommandRunner(),
	}

	// Apply options
	for _, option := range options {
		option(action)
	}

	return &task_engine.Action[*DockerLoadAction]{
		ID:      id,
		Wrapped: action,
	}
}

// DockerLoadOption is a function type for configuring DockerLoadAction
type DockerLoadOption func(*DockerLoadAction)

// WithPlatform sets the platform filter for loading specific platform variants
func WithPlatform(platform string) DockerLoadOption {
	return func(a *DockerLoadAction) {
		a.Platform = platform
	}
}

// WithQuiet suppresses the load output
func WithQuiet() DockerLoadOption {
	return func(a *DockerLoadAction) {
		a.Quiet = true
	}
}

// DockerLoadAction loads a Docker image from a tar archive file
type DockerLoadAction struct {
	task_engine.BaseAction
	TarFilePath      string
	Platform         string
	Quiet            bool
	CommandProcessor command.CommandRunner
	Output           string
	LoadedImages     []string // Stores the names of loaded images
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerLoadAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerLoadAction) Execute(execCtx context.Context) error {
	args := []string{"load", "-i", a.TarFilePath}

	if a.Platform != "" {
		args = append(args, "--platform", a.Platform)
	}

	if a.Quiet {
		args = append(args, "-q")
	}

	a.Logger.Info("Executing docker load", "tarFile", a.TarFilePath, "platform", a.Platform, "quiet", a.Quiet)
	output, err := a.CommandProcessor.RunCommand("docker", args...)
	a.Output = strings.TrimSpace(output)

	if err != nil {
		a.Logger.Error("Failed to load Docker image", "error", err, "output", output)
		return fmt.Errorf("failed to load Docker image from %s: %w. Output: %s", a.TarFilePath, err, output)
	}

	// Parse loaded image names from output
	a.parseLoadedImages(output)

	a.Logger.Info("Docker load finished successfully", "loadedImages", a.LoadedImages, "output", a.Output)
	return nil
}

// parseLoadedImages extracts image names from the docker load output
// Example output: "Loaded image: nginx:latest" or "Loaded image ID: sha256:abc123..."
func (a *DockerLoadAction) parseLoadedImages(output string) {
	lines := strings.Split(output, "\n")
	images := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Loaded image: ") {
			// Extract image name from "Loaded image: nginx:latest"
			imageName := strings.TrimPrefix(line, "Loaded image: ")
			if imageName != "" {
				images = append(images, imageName)
			}
		} else if strings.HasPrefix(line, "Loaded image ID: ") {
			// Extract image ID from "Loaded image ID: sha256:abc123..."
			imageID := strings.TrimPrefix(line, "Loaded image ID: ")
			if imageID != "" {
				images = append(images, imageID)
			}
		}
	}

	a.LoadedImages = images
}
