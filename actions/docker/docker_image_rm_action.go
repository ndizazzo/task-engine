package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerImageRmByNameAction creates an action to remove a Docker image by name and tag
func NewDockerImageRmByNameAction(logger *slog.Logger, imageName string, options ...DockerImageRmOption) *task_engine.Action[*DockerImageRmAction] {
	id := fmt.Sprintf("docker-image-rm-%s-action", strings.ReplaceAll(imageName, "/", "-"))

	action := &DockerImageRmAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		ImageName:        imageName,
		ImageID:          "",
		RemoveByID:       false,
		Force:            false,
		NoPrune:          false,
		CommandProcessor: command.NewDefaultCommandRunner(),
	}

	// Apply options
	for _, option := range options {
		option(action)
	}

	return &task_engine.Action[*DockerImageRmAction]{
		ID:      id,
		Wrapped: action,
	}
}

// NewDockerImageRmByIDAction creates an action to remove a Docker image by ID
func NewDockerImageRmByIDAction(logger *slog.Logger, imageID string, options ...DockerImageRmOption) *task_engine.Action[*DockerImageRmAction] {
	id := fmt.Sprintf("docker-image-rm-id-%s-action", imageID)

	action := &DockerImageRmAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		ImageName:        "",
		ImageID:          imageID,
		RemoveByID:       true,
		Force:            false,
		NoPrune:          false,
		CommandProcessor: command.NewDefaultCommandRunner(),
	}

	// Apply options
	for _, option := range options {
		option(action)
	}

	return &task_engine.Action[*DockerImageRmAction]{
		ID:      id,
		Wrapped: action,
	}
}

// DockerImageRmOption is a function type for configuring DockerImageRmAction
type DockerImageRmOption func(*DockerImageRmAction)

// WithForce forces the removal of the image
func WithForce() DockerImageRmOption {
	return func(a *DockerImageRmAction) {
		a.Force = true
	}
}

// WithNoPrune prevents removal of untagged parent images
func WithNoPrune() DockerImageRmOption {
	return func(a *DockerImageRmAction) {
		a.NoPrune = true
	}
}

// DockerImageRmAction removes Docker images by name/tag or ID
type DockerImageRmAction struct {
	task_engine.BaseAction
	ImageName        string
	ImageID          string
	RemoveByID       bool
	Force            bool
	NoPrune          bool
	CommandProcessor command.CommandRunner
	Output           string
	RemovedImages    []string // Stores the IDs of removed images
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerImageRmAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerImageRmAction) Execute(execCtx context.Context) error {
	args := []string{"image", "rm"}

	// Add force flag if specified
	if a.Force {
		args = append(args, "-f")
	}

	// Add no-prune flag if specified
	if a.NoPrune {
		args = append(args, "--no-prune")
	}

	// Add the image identifier (name/tag or ID)
	if a.RemoveByID {
		args = append(args, a.ImageID)
	} else {
		args = append(args, a.ImageName)
	}

	identifier := a.ImageName
	if a.RemoveByID {
		identifier = a.ImageID
	}

	a.Logger.Info("Executing docker image rm", "identifier", identifier, "force", a.Force, "noPrune", a.NoPrune)
	output, err := a.CommandProcessor.RunCommand("docker", args...)
	a.Output = strings.TrimSpace(output)

	if err != nil {
		a.Logger.Error("Failed to remove Docker image", "error", err, "output", output)
		return fmt.Errorf("failed to remove Docker image %s: %w. Output: %s", identifier, err, output)
	}

	// Parse removed image IDs from output
	a.parseRemovedImages(output)

	a.Logger.Info("Docker image rm finished successfully", "removedImages", a.RemovedImages, "output", a.Output)
	return nil
}

// parseRemovedImages extracts image IDs from the docker image rm output
// Example output: "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"
func (a *DockerImageRmAction) parseRemovedImages(output string) {
	lines := strings.Split(output, "\n")
	images := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Deleted: ") {
			// Extract image ID from "Deleted: sha256:abc123def456789"
			imageID := strings.TrimPrefix(line, "Deleted: ")
			if imageID != "" {
				images = append(images, imageID)
			}
		} else if strings.HasPrefix(line, "Untagged: ") {
			// Extract image name from "Untagged: nginx:latest"
			imageName := strings.TrimPrefix(line, "Untagged: ")
			if imageName != "" {
				images = append(images, imageName)
			}
		}
	}

	a.RemovedImages = images
}
