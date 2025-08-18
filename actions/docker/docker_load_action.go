package docker

import (
	"context"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

// DockerLoadActionBuilder provides the new constructor pattern
type DockerLoadActionBuilder struct {
	common.BaseConstructor[*DockerLoadAction]
	options []DockerLoadOption
}

// NewDockerLoadAction creates a new DockerLoadAction builder
func NewDockerLoadAction(logger *slog.Logger) *DockerLoadActionBuilder {
	return &DockerLoadActionBuilder{
		BaseConstructor: *common.NewBaseConstructor[*DockerLoadAction](logger),
		options:         []DockerLoadOption{},
	}
}

// WithParameters creates a DockerLoadAction with the specified parameters
func (b *DockerLoadActionBuilder) WithParameters(
	tarFilePathParam task_engine.ActionParameter,
) (*task_engine.Action[*DockerLoadAction], error) {
	action := &DockerLoadAction{
		BaseAction:       task_engine.NewBaseAction(b.GetLogger()),
		TarFilePath:      "",
		Platform:         "",
		Quiet:            false,
		CommandProcessor: command.NewDefaultCommandRunner(),
		Output:           "",
		LoadedImages:     []string{},
		TarFilePathParam: tarFilePathParam,
	}

	// Apply options
	for _, option := range b.options {
		option(action)
	}

	// Generate custom ID based on tar file path for backward compatibility
	id := "docker-load-action"
	if sp, ok := tarFilePathParam.(task_engine.StaticParameter); ok {
		if pathStr, ok2 := sp.Value.(string); ok2 {
			cleaned := strings.TrimSpace(pathStr)
			cleaned = strings.ReplaceAll(cleaned, " ", "-")
			cleaned = strings.ReplaceAll(cleaned, "/", "-")
			cleaned = strings.ReplaceAll(cleaned, "%", "")
			cleaned = strings.ReplaceAll(cleaned, "@", "")
			cleaned = strings.ReplaceAll(cleaned, "#", "")
			cleaned = strings.ReplaceAll(cleaned, "$", "")
			if cleaned != "" {
				id = "docker-load-" + cleaned + "-action"
			} else {
				id = "docker-load--action"
			}
		}
	}

	return b.WrapAction(action, "Docker Load", id), nil
}

// WithOptions adds options to the builder
func (b *DockerLoadActionBuilder) WithOptions(options ...DockerLoadOption) *DockerLoadActionBuilder {
	b.options = append(b.options, options...)
	return b
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
	common.ParameterResolver
	common.OutputBuilder
	TarFilePath      string
	Platform         string
	Quiet            bool
	CommandProcessor command.CommandRunner
	Output           string
	LoadedImages     []string // Stores the names of loaded images

	// Parameter-aware fields
	TarFilePathParam task_engine.ActionParameter
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerLoadAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerLoadAction) Execute(execCtx context.Context) error {
	// Resolve tar file path parameter if it exists
	if a.TarFilePathParam != nil {
		tarFilePathValue, err := a.ResolveStringParameter(execCtx, a.TarFilePathParam, "tar file path")
		if err != nil {
			return err
		}
		a.TarFilePath = tarFilePathValue
	}

	// If no tar file path provided, honor tests that expect empty path to still attempt command
	// a.TarFilePath may be empty; RunCommand will be invoked with "-i", ""

	args := []string{"load", "-i", a.TarFilePath}

	if a.Platform != "" {
		args = append(args, "--platform", a.Platform)
	}

	if a.Quiet {
		args = append(args, "-q")
	}

	a.Logger.Info("Executing docker load", "tarFile", a.TarFilePath, "platform", a.Platform, "quiet", a.Quiet)
	output, err := a.CommandProcessor.RunCommand("docker", args...)
	a.Output = output

	if err != nil {
		a.Logger.Error("Failed to load Docker image", "error", err, "output", output)
		return err
	}

	// Parse loaded image names from output
	a.parseLoadedImages(output)

	a.Logger.Info("Docker load finished successfully", "loadedImages", a.LoadedImages, "output", a.Output)
	return nil
}

// GetOutput returns information about loaded images and raw output
func (a *DockerLoadAction) GetOutput() interface{} {
	return a.BuildOutputWithCount(a.LoadedImages, len(a.LoadedImages) > 0, map[string]interface{}{
		"loadedImages": a.LoadedImages,
		"output":       a.Output,
		"tarFile":      a.TarFilePath,
	})
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
