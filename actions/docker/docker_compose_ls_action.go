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

// NewDockerComposeLsAction creates an action to list Docker Compose stacks
func NewDockerComposeLsAction(logger *slog.Logger, options ...DockerComposeLsOption) *task_engine.Action[*DockerComposeLsAction] {
	action := &DockerComposeLsAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		All:              false,
		Filter:           "",
		Format:           "",
		Quiet:            false,
		WorkingDir:       "",
		CommandProcessor: command.NewDefaultCommandRunner(),
	}

	// Apply options
	for _, option := range options {
		option(action)
	}

	return &task_engine.Action[*DockerComposeLsAction]{
		ID:      "docker-compose-ls-action",
		Wrapped: action,
	}
}

// DockerComposeLsOption is a function type for configuring DockerComposeLsAction
type DockerComposeLsOption func(*DockerComposeLsAction)

// WithComposeAll shows all stacks (default hides stopped stacks)
func WithComposeAll() DockerComposeLsOption {
	return func(a *DockerComposeLsAction) {
		a.All = true
	}
}

// WithComposeFilter filters output based on conditions provided
func WithComposeFilter(filter string) DockerComposeLsOption {
	return func(a *DockerComposeLsAction) {
		a.Filter = filter
	}
}

// WithComposeFormat uses a custom template for output
func WithComposeFormat(format string) DockerComposeLsOption {
	return func(a *DockerComposeLsAction) {
		a.Format = format
	}
}

// WithComposeQuiet only show stack names
func WithComposeQuiet() DockerComposeLsOption {
	return func(a *DockerComposeLsAction) {
		a.Quiet = true
	}
}

// WithWorkingDir sets the working directory for the compose command
func WithWorkingDir(workingDir string) DockerComposeLsOption {
	return func(a *DockerComposeLsAction) {
		a.WorkingDir = workingDir
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
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerComposeLsAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerComposeLsAction) Execute(execCtx context.Context) error {
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
