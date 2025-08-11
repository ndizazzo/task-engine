package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
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

// DockerComposeLsActionBuilder provides a fluent interface for building DockerComposeLsAction
type DockerComposeLsActionBuilder struct {
	logger *slog.Logger
}

// NewDockerComposeLsAction creates a fluent builder for DockerComposeLsAction
func NewDockerComposeLsAction(logger *slog.Logger) *DockerComposeLsActionBuilder {
	return &DockerComposeLsActionBuilder{logger: logger}
}

// WithParameters sets the parameters for working directory and configuration
func (b *DockerComposeLsActionBuilder) WithParameters(workingDirParam task_engine.ActionParameter, config DockerComposeLsConfig) *task_engine.Action[*DockerComposeLsAction] {
	// Determine whether to treat the provided parameter as active
	// - Non-empty static string: active (resolve at runtime)
	// - Non-string static parameter: active (so Execute will error as tests expect)
	// - Any non-static parameter: active
	// - Empty string static parameter: inactive (back-compat original constructor)
	useParam := false
	if sp, ok := workingDirParam.(task_engine.StaticParameter); ok {
		switch v := sp.Value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				useParam = true
			}
		default:
			// Non-string value should still attempt resolution (and then fail)
			useParam = true
		}
	} else if workingDirParam != nil {
		useParam = true
	}

	action := &DockerComposeLsAction{
		BaseAction:       task_engine.NewBaseAction(b.logger),
		All:              config.All,
		Filter:           config.Filter,
		Format:           config.Format,
		Quiet:            config.Quiet,
		WorkingDir:       config.WorkingDir,
		CommandProcessor: command.NewDefaultCommandRunner(),
	}
	if useParam {
		action.WorkingDirParam = workingDirParam
	}

	id := "docker-compose-ls-action"
	if useParam {
		id = "docker-compose-ls-with-params-action"
	}
	return &task_engine.Action[*DockerComposeLsAction]{
		ID:      id,
		Name:    "Docker Compose LS",
		Wrapped: action,
	}
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
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve working directory parameter if it exists
	if a.WorkingDirParam != nil {
		workingDirValue, err := a.WorkingDirParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve working directory parameter: %w", err)
		}
		if workingDirStr, ok := workingDirValue.(string); ok {
			a.WorkingDir = workingDirStr
		} else {
			return fmt.Errorf("working directory parameter is not a string, got %T", workingDirValue)
		}
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
	return map[string]interface{}{
		"stacks":  a.Stacks,
		"count":   len(a.Stacks),
		"output":  a.Output,
		"success": true,
	}
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
