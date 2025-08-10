package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// DockerImage represents a Docker image with its metadata
type DockerImage struct {
	Repository string
	Tag        string
	ImageID    string
	Size       string
	Created    string
}

// NewDockerImageListAction creates an action to list all Docker images
func NewDockerImageListAction(logger *slog.Logger, options ...DockerImageListOption) *task_engine.Action[*DockerImageListAction] {
	action := &DockerImageListAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		All:              false,
		Digests:          false,
		Filter:           "",
		Format:           "",
		NoTrunc:          false,
		Quiet:            false,
		CommandProcessor: command.NewDefaultCommandRunner(),
	}

	// Apply options
	for _, option := range options {
		option(action)
	}

	return &task_engine.Action[*DockerImageListAction]{
		ID:      "docker-image-list-action",
		Wrapped: action,
	}
}

// DockerImageListOption is a function type for configuring DockerImageListAction
type DockerImageListOption func(*DockerImageListAction)

// WithAll shows all images (default hides intermediate images)
func WithAll() DockerImageListOption {
	return func(a *DockerImageListAction) {
		a.All = true
	}
}

// WithDigests shows digests
func WithDigests() DockerImageListOption {
	return func(a *DockerImageListAction) {
		a.Digests = true
	}
}

// WithFilter filters output based on conditions provided
func WithFilter(filter string) DockerImageListOption {
	return func(a *DockerImageListAction) {
		a.Filter = filter
	}
}

// WithFormat uses a custom template for output
func WithFormat(format string) DockerImageListOption {
	return func(a *DockerImageListAction) {
		a.Format = format
	}
}

// WithNoTrunc don't truncate output
func WithNoTrunc() DockerImageListOption {
	return func(a *DockerImageListAction) {
		a.NoTrunc = true
	}
}

// WithQuietOutput only show image IDs
func WithQuietOutput() DockerImageListOption {
	return func(a *DockerImageListAction) {
		a.Quiet = true
	}
}

// DockerImageListAction lists Docker images
type DockerImageListAction struct {
	task_engine.BaseAction
	All              bool
	Digests          bool
	Filter           string
	Format           string
	NoTrunc          bool
	Quiet            bool
	CommandProcessor command.CommandRunner
	Output           string
	Images           []DockerImage // Stores the parsed images
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerImageListAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerImageListAction) Execute(execCtx context.Context) error {
	args := []string{"image", "ls"}

	if a.All {
		args = append(args, "--all")
	}
	if a.Digests {
		args = append(args, "--digests")
	}
	if a.Filter != "" {
		args = append(args, "--filter", a.Filter)
	}
	if a.Format != "" {
		args = append(args, "--format", a.Format)
	}
	if a.NoTrunc {
		args = append(args, "--no-trunc")
	}
	if a.Quiet {
		args = append(args, "--quiet")
	}

	a.Logger.Info("Executing docker image ls",
		"all", a.All,
		"digests", a.Digests,
		"filter", a.Filter,
		"format", a.Format,
		"noTrunc", a.NoTrunc,
		"quiet", a.Quiet,
	)

	output, err := a.CommandProcessor.RunCommand("docker", args...)
	if err != nil {
		a.Logger.Error("Failed to list Docker images", "error", err.Error(), "output", output)
		return fmt.Errorf("failed to list Docker images: %w", err)
	}

	a.Output = output
	a.parseImages(output)

	a.Logger.Info("Docker image ls finished successfully",
		"imageCount", len(a.Images),
		"output", output,
	)

	return nil
}

// parseImages parses the docker image ls output and populates the Images slice
func (a *DockerImageListAction) parseImages(output string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		a.Images = []DockerImage{}
		return
	}

	// Skip header line
	if len(lines) > 0 && strings.Contains(lines[0], "REPOSITORY") {
		lines = lines[1:]
	}

	var images []DockerImage
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse the line
		image := a.parseImageLine(line)
		if image != nil {
			images = append(images, *image)
		}
	}

	a.Images = images
}

// parseImageLine parses a single line from docker image ls output
func (a *DockerImageListAction) parseImageLine(line string) *DockerImage {
	// Format: REPOSITORY TAG IMAGE ID CREATED SIZE
	// Example: nginx               latest              sha256:abc123def456 2 weeks ago         133MB
	// Example: <none>              <none>              sha256:def456ghi789 3 weeks ago         0B
	// Example: docker.io/library/ubuntu   20.04               sha256:ghi789jkl012 1 month ago         72.8MB

	// Split by whitespace
	parts := strings.Fields(line)
	if len(parts) < 5 {
		return nil
	}

	// Find the image ID (starts with sha256:)
	imageIDIndex := -1
	for i, part := range parts {
		if strings.HasPrefix(part, "sha256:") {
			imageIDIndex = i
			break
		}
	}

	if imageIDIndex == -1 {
		return nil
	}

	imageID := parts[imageIDIndex]

	// Everything before the image ID is repository and tag
	repoTagParts := parts[:imageIDIndex]
	if len(repoTagParts) < 2 {
		return nil
	}

	// The last part is the tag, everything else is repository
	tag := repoTagParts[len(repoTagParts)-1]
	repository := strings.Join(repoTagParts[:len(repoTagParts)-1], " ")

	// Note: <none> values are preserved as literal strings, not converted to empty strings

	// The parts after image ID should be: CREATED SIZE
	// Created time can be "2 weeks ago" or "1 month ago" etc.
	remainingParts := parts[imageIDIndex+1:]
	if len(remainingParts) < 2 {
		return nil
	}

	// Find the size (last part)
	size := remainingParts[len(remainingParts)-1]

	// Everything between image ID and size is the created time
	created := strings.Join(remainingParts[:len(remainingParts)-1], " ")

	return &DockerImage{
		Repository: repository,
		Tag:        tag,
		ImageID:    imageID,
		Size:       size,
		Created:    created,
	}
}
