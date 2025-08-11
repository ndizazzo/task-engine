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

// DockerImageListActionConstructor provides the new constructor pattern
type DockerImageListActionConstructor struct {
	logger *slog.Logger
}

// NewDockerImageListAction creates a new DockerImageListAction constructor
func NewDockerImageListAction(logger *slog.Logger) *DockerImageListActionConstructor {
	return &DockerImageListActionConstructor{
		logger: logger,
	}
}

// WithParameters creates a DockerImageListAction with the specified parameters
func (c *DockerImageListActionConstructor) WithParameters(
	allParam task_engine.ActionParameter,
	digestsParam task_engine.ActionParameter,
	filterParam task_engine.ActionParameter,
	formatParam task_engine.ActionParameter,
	noTruncParam task_engine.ActionParameter,
	quietParam task_engine.ActionParameter,
) (*task_engine.Action[*DockerImageListAction], error) {
	action := &DockerImageListAction{
		BaseAction:       task_engine.NewBaseAction(c.logger),
		All:              false,
		Digests:          false,
		Filter:           "",
		Format:           "",
		NoTrunc:          false,
		Quiet:            false,
		CommandProcessor: command.NewDefaultCommandRunner(),
		AllParam:         allParam,
		DigestsParam:     digestsParam,
		FilterParam:      filterParam,
		FormatParam:      formatParam,
		NoTruncParam:     noTruncParam,
		QuietParam:       quietParam,
	}

	id := "docker-image-list-action"
	return &task_engine.Action[*DockerImageListAction]{
		ID:      id,
		Name:    "Docker Image List",
		Wrapped: action,
	}, nil
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

	// Parameter-aware fields
	AllParam     task_engine.ActionParameter
	DigestsParam task_engine.ActionParameter
	FilterParam  task_engine.ActionParameter
	FormatParam  task_engine.ActionParameter
	NoTruncParam task_engine.ActionParameter
	QuietParam   task_engine.ActionParameter
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerImageListAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

// SetOptions applies configuration options to the action
func (a *DockerImageListAction) SetOptions(options ...DockerImageListOption) {
	for _, option := range options {
		option(a)
	}
}

func (a *DockerImageListAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve All parameter if provided
	if a.AllParam != nil {
		v, err := a.AllParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve all parameter: %w", err)
		}
		if allBool, ok := v.(bool); ok {
			a.All = allBool
		} else {
			return fmt.Errorf("all parameter is not a bool, got %T", v)
		}
	}

	// Resolve Digests parameter if provided
	if a.DigestsParam != nil {
		v, err := a.DigestsParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve digests parameter: %w", err)
		}
		if digestsBool, ok := v.(bool); ok {
			a.Digests = digestsBool
		} else {
			return fmt.Errorf("digests parameter is not a bool, got %T", v)
		}
	}

	// Resolve Filter parameter if provided
	if a.FilterParam != nil {
		v, err := a.FilterParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve filter parameter: %w", err)
		}
		if filterStr, ok := v.(string); ok {
			if strings.TrimSpace(filterStr) != "" {
				a.Filter = filterStr
			}
		} else {
			return fmt.Errorf("filter parameter is not a string, got %T", v)
		}
	}

	// Resolve Format parameter if provided
	if a.FormatParam != nil {
		v, err := a.FormatParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve format parameter: %w", err)
		}
		if formatStr, ok := v.(string); ok {
			if strings.TrimSpace(formatStr) != "" {
				a.Format = formatStr
			}
		} else {
			return fmt.Errorf("format parameter is not a string, got %T", v)
		}
	}

	// Resolve NoTrunc parameter if provided
	if a.NoTruncParam != nil {
		v, err := a.NoTruncParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve noTrunc parameter: %w", err)
		}
		if noTruncBool, ok := v.(bool); ok {
			a.NoTrunc = noTruncBool
		} else {
			return fmt.Errorf("noTrunc parameter is not a bool, got %T", v)
		}
	}

	// Resolve Quiet parameter if provided
	if a.QuietParam != nil {
		v, err := a.QuietParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve quiet parameter: %w", err)
		}
		if quietBool, ok := v.(bool); ok {
			a.Quiet = quietBool
		} else {
			return fmt.Errorf("quiet parameter is not a bool, got %T", v)
		}
	}

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

// GetOutput returns parsed image information and raw output metadata
func (a *DockerImageListAction) GetOutput() interface{} {
	return map[string]interface{}{
		"images":  a.Images,
		"count":   len(a.Images),
		"output":  a.Output,
		"success": true,
	}
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
