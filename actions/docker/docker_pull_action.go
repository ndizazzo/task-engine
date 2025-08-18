package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
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

// DockerPullActionConstructor provides the new constructor pattern
type DockerPullActionConstructor struct {
	common.BaseConstructor[*DockerPullAction]
}

// NewDockerPullAction creates a new DockerPullAction constructor
func NewDockerPullAction(logger *slog.Logger) *DockerPullActionConstructor {
	return &DockerPullActionConstructor{
		BaseConstructor: *common.NewBaseConstructor[*DockerPullAction](logger),
	}
}

// WithParameters creates a DockerPullAction with the specified parameters
func (c *DockerPullActionConstructor) WithParameters(
	imagesParam task_engine.ActionParameter,
	multiArchImagesParam task_engine.ActionParameter,
	allTagsParam task_engine.ActionParameter,
	quietParam task_engine.ActionParameter,
	platformParam task_engine.ActionParameter,
) (*task_engine.Action[*DockerPullAction], error) {
	action := &DockerPullAction{
		BaseAction:           task_engine.NewBaseAction(c.GetLogger()),
		Images:               make(map[string]ImageSpec),
		MultiArchImages:      make(map[string]MultiArchImageSpec),
		AllTags:              false,
		Quiet:                false,
		Platform:             "",
		CommandProcessor:     command.NewDefaultCommandRunner(),
		Output:               "",
		PulledImages:         []string{},
		FailedImages:         []string{},
		ImagesParam:          imagesParam,
		MultiArchImagesParam: multiArchImagesParam,
		AllTagsParam:         allTagsParam,
		QuietParam:           quietParam,
		PlatformParam:        platformParam,
	}

	return c.WrapAction(action, "Docker Pull", "docker-pull-action"), nil
}

// Backward compatibility functions
func NewDockerPullActionLegacy(logger *slog.Logger, images map[string]ImageSpec, options ...DockerPullOption) *task_engine.Action[*DockerPullAction] {
	action := &DockerPullAction{
		BaseAction:       task_engine.NewBaseAction(logger),
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
		Name:    "Docker Pull",
		Wrapped: action,
	}
}

func NewDockerPullMultiArchActionLegacy(logger *slog.Logger, multiArchImages map[string]MultiArchImageSpec, options ...DockerPullOption) *task_engine.Action[*DockerPullAction] {
	action := &DockerPullAction{
		BaseAction:       task_engine.NewBaseAction(logger),
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
		Name:    "Docker Pull (Multi-Arch)",
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
	common.ParameterResolver
	common.OutputBuilder
	Images           map[string]ImageSpec
	MultiArchImages  map[string]MultiArchImageSpec
	AllTags          bool
	Quiet            bool
	Platform         string
	CommandProcessor command.CommandRunner
	Output           string
	PulledImages     []string
	FailedImages     []string

	// Parameter-aware fields
	ImagesParam          task_engine.ActionParameter
	MultiArchImagesParam task_engine.ActionParameter
	AllTagsParam         task_engine.ActionParameter
	QuietParam           task_engine.ActionParameter
	PlatformParam        task_engine.ActionParameter
}

func (a *DockerPullAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerPullAction) Execute(execCtx context.Context) error {
	// Resolve images via parameter if provided
	if a.ImagesParam != nil {
		v, err := a.ResolveParameter(execCtx, a.ImagesParam, "images")
		if err != nil {
			return err
		}
		switch typed := v.(type) {
		case map[string]ImageSpec:
			a.Images = typed
		case map[string]interface{}:
			// Attempt to coerce map[string]interface{} into map[string]ImageSpec when fields match
			converted := make(map[string]ImageSpec, len(typed))
			for k, vi := range typed {
				if ms, ok := vi.(map[string]interface{}); ok {
					spec := ImageSpec{}
					if img, ok := ms["Image"].(string); ok {
						spec.Image = img
					}
					if tag, ok := ms["Tag"].(string); ok {
						spec.Tag = tag
					}
					if arch, ok := ms["Architecture"].(string); ok {
						spec.Architecture = arch
					}
					converted[k] = spec
				}
			}
			a.Images = converted
		default:
			return fmt.Errorf("unsupported images parameter type: %T", v)
		}
	}

	// Resolve multiarch images via parameter if provided
	if a.MultiArchImagesParam != nil {
		v, err := a.ResolveParameter(execCtx, a.MultiArchImagesParam, "multiarch images")
		if err != nil {
			return err
		}
		switch typed := v.(type) {
		case map[string]MultiArchImageSpec:
			a.MultiArchImages = typed
		case map[string]interface{}:
			// Attempt to coerce map[string]interface{} into map[string]MultiArchImageSpec
			converted := make(map[string]MultiArchImageSpec, len(typed))
			for k, vi := range typed {
				if ms, ok := vi.(map[string]interface{}); ok {
					spec := MultiArchImageSpec{}
					if img, ok := ms["Image"].(string); ok {
						spec.Image = img
					}
					if tag, ok := ms["Tag"].(string); ok {
						spec.Tag = tag
					}
					if archs, ok := ms["Architectures"].([]string); ok {
						spec.Architectures = archs
					} else if archsI, ok := ms["Architectures"].([]interface{}); ok {
						// Convert []interface{} to []string
						archStrs := make([]string, len(archsI))
						for i, arch := range archsI {
							if archStr, ok := arch.(string); ok {
								archStrs[i] = archStr
							}
						}
						spec.Architectures = archStrs
					}
					converted[k] = spec
				}
			}
			a.MultiArchImages = converted
		default:
			return fmt.Errorf("unsupported multiarch images parameter type: %T", v)
		}
	}

	// Resolve AllTags parameter if provided
	if a.AllTagsParam != nil {
		v, err := a.ResolveParameter(execCtx, a.AllTagsParam, "allTags")
		if err != nil {
			return err
		}
		if allTagsBool, ok := v.(bool); ok {
			a.AllTags = allTagsBool
		} else {
			return fmt.Errorf("allTags parameter is not a bool, got %T", v)
		}
	}

	// Resolve Quiet parameter if provided
	if a.QuietParam != nil {
		v, err := a.ResolveParameter(execCtx, a.QuietParam, "quiet")
		if err != nil {
			return err
		}
		if quietBool, ok := v.(bool); ok {
			a.Quiet = quietBool
		} else {
			return fmt.Errorf("quiet parameter is not a bool, got %T", v)
		}
	}

	// Resolve Platform parameter if provided
	if a.PlatformParam != nil {
		v, err := a.ResolveParameter(execCtx, a.PlatformParam, "platform")
		if err != nil {
			return err
		}
		if platformStr, ok := v.(string); ok {
			if strings.TrimSpace(platformStr) != "" {
				a.Platform = platformStr
			}
		} else {
			return fmt.Errorf("platform parameter is not a string, got %T", v)
		}
	}

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

func (a *DockerPullAction) GetOutput() interface{} {
	return map[string]interface{}{
		"output":       a.Output,
		"pulledImages": a.PulledImages,
		"failedImages": a.FailedImages,
		"totalImages":  len(a.Images) + len(a.MultiArchImages),
		"success":      len(a.FailedImages) == 0,
	}
}
