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

// ComposeStack represents a Docker Compose stack with its metadata
type ComposeStack struct {
	Name        string
	Status      string
	ConfigFiles string
}

// DockerComposeLsConfig holds configuration for Docker Compose ls action
type DockerComposeLsConfig struct {
	All        bool
	Filter     string
	Format     string
	Quiet      bool
	WorkingDir string
}

// DockerComposeLsActionBuilder provides the new constructor pattern
type DockerComposeLsActionBuilder struct {
	common.BaseConstructor[*DockerComposeLsAction]
}

// NewDockerComposeLsAction creates a new DockerComposeLsAction builder
func NewDockerComposeLsAction(logger *slog.Logger) *DockerComposeLsActionBuilder {
	return &DockerComposeLsActionBuilder{
		BaseConstructor: *common.NewBaseConstructor[*DockerComposeLsAction](logger),
	}
}

// WithParameters creates a DockerComposeLsAction with the specified parameters
func (b *DockerComposeLsActionBuilder) WithParameters(
	workingDirParam task_engine.ActionParameter,
	config DockerComposeLsConfig,
) (*task_engine.Action[*DockerComposeLsAction], error) {
	action := &DockerComposeLsAction{
		BaseAction:       task_engine.NewBaseAction(b.GetLogger()),
		All:              config.All,
		Filter:           config.Filter,
		Format:           config.Format,
		Quiet:            config.Quiet,
		WorkingDir:       config.WorkingDir,
		CommandProcessor: command.NewDefaultCommandRunner(),
		Output:           "",
		Stacks:           []ComposeStack{},
		WorkingDirParam:  nil, // Default to nil for backward compatibility
	}

	// Only set the parameter if it has a meaningful value
	if workingDirParam != nil {
		if sp, ok := workingDirParam.(task_engine.StaticParameter); ok {
			if v, ok2 := sp.Value.(string); ok2 && strings.TrimSpace(v) != "" {
				action.WorkingDirParam = workingDirParam
			} else if !ok2 {
				// Set non-string static parameters so Execute can fail as expected
				action.WorkingDirParam = workingDirParam
			}
			// Don't set empty string static parameters for backward compatibility
		} else {
			// Non-static parameters should always be set
			action.WorkingDirParam = workingDirParam
		}
	}

	// Generate custom ID based on whether parameters are provided
	id := "docker-compose-ls-action"
	if action.WorkingDirParam != nil {
		// Only use custom ID for meaningful string parameters
		if sp, ok := action.WorkingDirParam.(task_engine.StaticParameter); ok {
			if v, ok2 := sp.Value.(string); ok2 && strings.TrimSpace(v) != "" {
				id = "docker-compose-ls-with-params-action"
			}
		} else {
			// Non-static parameters should use the custom ID
			id = "docker-compose-ls-with-params-action"
		}
	}

	return b.WrapAction(action, "Docker Compose LS", id), nil
}

// DockerComposeLsOption is a function type for configuring DockerComposeLsAction
type DockerComposeLsOption func(*DockerComposeLsConfig)

// NewDockerComposeLsConfig creates a config with options
func NewDockerComposeLsConfig(options ...DockerComposeLsOption) DockerComposeLsConfig {
	config := DockerComposeLsConfig{
		All:        false,
		Filter:     "",
		Format:     "",
		Quiet:      false,
		WorkingDir: "",
	}
	for _, option := range options {
		option(&config)
	}

	return config
}

// WithComposeAll shows all stacks (default hides stopped stacks)
func WithComposeAll() DockerComposeLsOption {
	return func(c *DockerComposeLsConfig) {
		c.All = true
	}
}

// WithComposeFilter filters output based on conditions provided
func WithComposeFilter(filter string) DockerComposeLsOption {
	return func(c *DockerComposeLsConfig) {
		c.Filter = filter
	}
}

// WithComposeFormat uses a custom template for output
func WithComposeFormat(format string) DockerComposeLsOption {
	return func(c *DockerComposeLsConfig) {
		c.Format = format
	}
}

// WithComposeQuiet only show stack names
func WithComposeLsQuiet() DockerComposeLsOption {
	return func(c *DockerComposeLsConfig) {
		c.Quiet = true
	}
}

// WithWorkingDir sets the working directory for the compose command
func WithWorkingDir(workingDir string) DockerComposeLsOption {
	return func(c *DockerComposeLsConfig) {
		c.WorkingDir = workingDir
	}
}

// DockerComposeLsAction lists Docker Compose stacks
type DockerComposeLsAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	All              bool
	Filter           string
	Format           string
	Quiet            bool
	WorkingDir       string
	CommandProcessor command.CommandRunner
	Output           string
	Stacks           []ComposeStack // Stores the parsed stacks

	// Parameter-aware fields
	WorkingDirParam task_engine.ActionParameter
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerComposeLsAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerComposeLsAction) Execute(execCtx context.Context) error {
	// Resolve working directory parameter if it exists
	if a.WorkingDirParam != nil {
		workingDirValue, err := a.ResolveStringParameter(execCtx, a.WorkingDirParam, "working directory")
		if err != nil {
			return err
		}
		a.WorkingDir = workingDirValue
	}

	args := []string{"compose", "ls"}

	if a.All {
		args = append(args, "--all")
	}
	if a.Filter != "" {
		args = append(args, "--filter", a.Filter)
	}
	if a.Format != "" {
		args = append(args, "--format", a.Format)
	}
	if a.Quiet {
		args = append(args, "--quiet")
	}

	a.Logger.Info("Executing docker compose ls",
		"all", a.All,
		"filter", a.Filter,
		"format", a.Format,
		"quiet", a.Quiet,
		"workingDir", a.WorkingDir,
	)

	output, err := a.CommandProcessor.RunCommand("docker", args...)
	if err != nil {
		a.Logger.Error("Failed to list Docker Compose stacks", "error", err.Error(), "output", output)
		return fmt.Errorf("failed to list Docker Compose stacks: %w", err)
	}

	a.Output = output
	a.parseStacks(output)

	a.Logger.Info("Docker compose ls finished successfully",
		"stackCount", len(a.Stacks),
		"output", output,
	)

	return nil
}

// GetOutput returns parsed stack information and raw output metadata.
// This enables other actions to reference the output of this action
// using ActionOutputParameter references.
func (a *DockerComposeLsAction) GetOutput() interface{} {
	return a.BuildOutputWithCount(a.Stacks, true, map[string]interface{}{
		"stacks": a.Stacks,
		"output": a.Output,
	})
}

// parseStacks parses the docker compose ls output and populates the Stacks slice
func (a *DockerComposeLsAction) parseStacks(output string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		a.Stacks = []ComposeStack{}
		return
	}

	// Skip header line
	if len(lines) > 0 && strings.Contains(lines[0], "NAME") {
		lines = lines[1:]
	}

	var stacks []ComposeStack
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse the line
		stack := a.parseStackLine(line)
		if stack != nil {
			stacks = append(stacks, *stack)
		}
	}

	a.Stacks = stacks
}

// parseStackLine parses a single line from docker compose ls output
func (a *DockerComposeLsAction) parseStackLine(line string) *ComposeStack {
	// Format: NAME STATUS CONFIG FILES
	// Example: myapp running /path/to/docker-compose.yml
	// Example: testapp stopped /path/to/compose.yml,/path/to/override.yml
	// Quiet format: just the stack name
	// Example: myapp

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	// If quiet mode, we only have the stack name
	if a.Quiet {
		return &ComposeStack{
			Name:        parts[0],
			Status:      "",
			ConfigFiles: "",
		}
	}

	// Regular mode requires at least 3 fields
	if len(parts) < 3 {
		return nil
	}

	name := parts[0]
	status := parts[1]
	configFiles := strings.Join(parts[2:], " ")

	return &ComposeStack{
		Name:        name,
		Status:      status,
		ConfigFiles: configFiles,
	}
}
