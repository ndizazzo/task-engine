package command

import (
	"context"
	"os/exec"
	"strings"
)

// CommandRunner interface for executing system commands
type CommandRunner interface {
	RunCommand(command string, args ...string) (string, error)
	RunCommandWithContext(ctx context.Context, command string, args ...string) (string, error)
	RunCommandInDir(workingDir string, command string, args ...string) (string, error)
	RunCommandInDirWithContext(ctx context.Context, workingDir string, command string, args ...string) (string, error)
}

// DefaultCommandRunner is the default implementation of CommandRunner
type DefaultCommandRunner struct{}

// NewDefaultCommandRunner creates a new default command runner
func NewDefaultCommandRunner() *DefaultCommandRunner {
	return &DefaultCommandRunner{}
}

// RunCommand executes a command and returns the output
func (r *DefaultCommandRunner) RunCommand(command string, args ...string) (string, error) {
	return r.RunCommandWithContext(context.Background(), command, args...)
}

// RunCommandWithContext executes a command with context and returns the output
func (r *DefaultCommandRunner) RunCommandWithContext(ctx context.Context, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return strings.TrimSpace(string(output)), nil
}

// RunCommandInDir executes a command in a specific working directory
func (r *DefaultCommandRunner) RunCommandInDir(workingDir string, command string, args ...string) (string, error) {
	return r.RunCommandInDirWithContext(context.Background(), workingDir, command, args...)
}

// RunCommandInDirWithContext executes a command in a specific working directory with context
func (r *DefaultCommandRunner) RunCommandInDirWithContext(ctx context.Context, workingDir string, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return strings.TrimSpace(string(output)), nil
}
