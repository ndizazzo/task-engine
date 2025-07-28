package docker

import (
	"context"
	"fmt"
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

type ImageSpec struct {
	Image        string
	Tag          string
	Architecture string
}

type MultiArchImageSpec struct {
	Image         string
	Tag           string
	Architectures []string
}

func NewDockerPullAction(logger *slog.Logger, images map[string]ImageSpec, options ...DockerPullOption) *task_engine.Action[*DockerPullAction] {
	action := &DockerPullAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		Images:           images,
		MultiArchImages:  make(map[string]MultiArchImageSpec),
		AllTags:          false,
		Quiet:            false,
		Platform:         "",
		CommandProcessor: command.NewDefaultCommandRunner(),
	}

	for _, option := range options {
		option(action)
	}

	return &task_engine.Action[*DockerPullAction]{
		ID:      "docker-pull-action",
		Wrapped: action,
	}
}

func NewDockerPullMultiArchAction(logger *slog.Logger, multiArchImages map[string]MultiArchImageSpec, options ...DockerPullOption) *task_engine.Action[*DockerPullAction] {
	action := &DockerPullAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		Images:           make(map[string]ImageSpec),
		MultiArchImages:  multiArchImages,
		AllTags:          false,
		Quiet:            false,
		Platform:         "",
		CommandProcessor: command.NewDefaultCommandRunner(),
	}

	for _, option := range options {
		option(action)
	}

	return &task_engine.Action[*DockerPullAction]{
		ID:      "docker-pull-multiarch-action",
		Wrapped: action,
	}
}

type DockerPullOption func(*DockerPullAction)

func WithAllTags() DockerPullOption {
	return func(a *DockerPullAction) {
		a.AllTags = true
	}
}

func WithPullQuietOutput() DockerPullOption {
	return func(a *DockerPullAction) {
		a.Quiet = true
	}
}

func WithPullPlatform(platform string) DockerPullOption {
	return func(a *DockerPullAction) {
		a.Platform = platform
	}
}

type DockerPullAction struct {
	task_engine.BaseAction
	Images           map[string]ImageSpec
	MultiArchImages  map[string]MultiArchImageSpec
	AllTags          bool
	Quiet            bool
	Platform         string
	CommandProcessor command.CommandRunner
	Output           string
	PulledImages     []string
	FailedImages     []string
}

func (a *DockerPullAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerPullAction) Execute(execCtx context.Context) error {
	totalImages := len(a.Images) + len(a.MultiArchImages)
	if totalImages == 0 {
		return fmt.Errorf("no images specified for pulling")
	}

	a.Logger.Info("Starting Docker pull operation", "image_count", totalImages)
	a.PulledImages = []string{}
	a.FailedImages = []string{}

	for name, spec := range a.Images {
		if err := a.pullImage(execCtx, name, spec); err != nil {
			a.FailedImages = append(a.FailedImages, name)
			a.Logger.Error("Failed to pull image", "name", name, "error", err)
		} else {
			a.PulledImages = append(a.PulledImages, name)
		}
	}

	for name, spec := range a.MultiArchImages {
		if err := a.pullMultiArchImage(execCtx, name, spec); err != nil {
			a.FailedImages = append(a.FailedImages, name)
			a.Logger.Error("Failed to pull multi-arch image", "name", name, "error", err)
		} else {
			a.PulledImages = append(a.PulledImages, name)
		}
	}

	a.Output = fmt.Sprintf("Pulled %d images, failed %d images", len(a.PulledImages), len(a.FailedImages))

	if len(a.FailedImages) > 0 {
		return fmt.Errorf("failed to pull %d images: %v", len(a.FailedImages), a.FailedImages)
	}

	a.Logger.Info("Docker pull operation completed successfully", "pulled_count", len(a.PulledImages))
	return nil
}

func (a *DockerPullAction) pullImage(ctx context.Context, name string, spec ImageSpec) error {
	args := []string{"pull"}

	if a.Quiet {
		args = append(args, "--quiet")
	}

	platform := a.Platform
	if platform == "" && spec.Architecture != "" {
		platform = spec.Architecture
	}

	if platform != "" {
		args = append(args, "--platform", platform)
	}

	imageRef := a.buildImageReference(spec)
	args = append(args, imageRef)

	a.Logger.Info("Pulling Docker image", "name", name, "image", imageRef, "architecture", spec.Architecture)

	output, err := a.CommandProcessor.RunCommandWithContext(ctx, "docker", args...)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w. Output: %s", imageRef, err, output)
	}

	a.Logger.Info("Successfully pulled image", "name", name, "image", imageRef)
	return nil
}

func (a *DockerPullAction) pullMultiArchImage(ctx context.Context, name string, spec MultiArchImageSpec) error {
	var lastError error
	successCount := 0

	for _, arch := range spec.Architectures {
		imageSpec := ImageSpec{
			Image:        spec.Image,
			Tag:          spec.Tag,
			Architecture: arch,
		}

		args := []string{"pull"}

		if a.Quiet {
			args = append(args, "--quiet")
		}

		platform := a.Platform
		if platform == "" {
			platform = arch
		}

		if platform != "" {
			args = append(args, "--platform", platform)
		}

		imageRef := a.buildImageReference(imageSpec)
		args = append(args, imageRef)

		a.Logger.Info("Pulling multi-arch Docker image", "name", name, "image", imageRef, "architecture", arch)

		output, err := a.CommandProcessor.RunCommandWithContext(ctx, "docker", args...)
		if err != nil {
			lastError = fmt.Errorf("failed to pull image %s for architecture %s: %w. Output: %s", imageRef, arch, err, output)
			a.Logger.Error("Failed to pull multi-arch image", "name", name, "architecture", arch, "error", err)
			continue
		}

		successCount++
		a.Logger.Info("Successfully pulled multi-arch image", "name", name, "image", imageRef, "architecture", arch)
	}

	if successCount == 0 {
		return fmt.Errorf("failed to pull any architecture for image %s: %w", name, lastError)
	}

	if successCount < len(spec.Architectures) {
		a.Logger.Warn("Partially successful multi-arch pull", "name", name, "successful", successCount, "total", len(spec.Architectures))
	}

	return nil
}

func (a *DockerPullAction) buildImageReference(spec ImageSpec) string {
	if spec.Tag == "" {
		return spec.Image
	}
	return fmt.Sprintf("%s:%s", spec.Image, spec.Tag)
}

func (a *DockerPullAction) GetPulledImages() []string {
	return a.PulledImages
}

func (a *DockerPullAction) GetFailedImages() []string {
	return a.FailedImages
}

func (a *DockerPullAction) GetOutput() string {
	return a.Output
}
