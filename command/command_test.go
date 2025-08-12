package command

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultCommandRunner(t *testing.T) {
	runner := NewDefaultCommandRunner()
	assert.NotNil(t, runner)
	assert.IsType(t, &DefaultCommandRunner{}, runner)
}

func TestRunCommand(t *testing.T) {
	runner := NewDefaultCommandRunner()

	// Test successful command
	result, err := runner.RunCommand("echo", "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)

	// Test command with no args
	result, err = runner.RunCommand("echo")
	assert.NoError(t, err)
	assert.Equal(t, "", result)

	// Test failing command
	result, err = runner.RunCommand("nonexistentcommand")
	assert.Error(t, err)
	// On macOS, the error message varies, so we just check that there's an error
}

func TestRunCommandWithContext(t *testing.T) {
	runner := NewDefaultCommandRunner()
	ctx := context.Background()

	// Test successful command
	result, err := runner.RunCommandWithContext(ctx, "echo", "hello world")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)

	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err = runner.RunCommandWithContext(ctx, "sleep", "5")
	assert.Error(t, err)
	// Context timeout behavior varies by OS, just check for error
}

func TestRunCommandInDir(t *testing.T) {
	runner := NewDefaultCommandRunner()

	// Test command in current directory
	result, err := runner.RunCommandInDir(".", "pwd")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Test command in non-existent directory
	result, err = runner.RunCommandInDir("/nonexistent/dir", "pwd")
	assert.Error(t, err)
	// Error message varies by OS, just check for error
}

func TestRunCommandInDirWithContext(t *testing.T) {
	runner := NewDefaultCommandRunner()
	ctx := context.Background()

	// Test successful command
	result, err := runner.RunCommandInDirWithContext(ctx, ".", "pwd")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err = runner.RunCommandInDirWithContext(ctx, ".", "sleep", "5")
	assert.Error(t, err)
	// Context timeout behavior varies by OS, just check for error
}

func TestCommandRunnerInterface(t *testing.T) {
	var runner CommandRunner = NewDefaultCommandRunner()
	assert.NotNil(t, runner)

	// Test interface methods
	result, err := runner.RunCommand("echo", "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", result)

	result, err = runner.RunCommandWithContext(context.Background(), "echo", "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", result)

	result, err = runner.RunCommandInDir(".", "echo", "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", result)

	result, err = runner.RunCommandInDirWithContext(context.Background(), ".", "echo", "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestCommandOutputTrimming(t *testing.T) {
	runner := NewDefaultCommandRunner()

	// Test that output is properly trimmed
	result, err := runner.RunCommand("echo", "-n", "  hello world  ")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result) // echo -n output gets trimmed by TrimSpace

	// Test with echo that adds newline
	result, err = runner.RunCommand("echo", "  hello world  ")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", result) // TrimSpace removes newline and whitespace
}

func TestCommandErrorHandling(t *testing.T) {
	runner := NewDefaultCommandRunner()

	// Test command that returns error but has output
	result, err := runner.RunCommand("sh", "-c", "echo 'error message'; exit 1")
	assert.Error(t, err)
	assert.Equal(t, "error message\n", result) // Should still get output even with error (includes newline)

	// Test command that fails immediately
	result, err = runner.RunCommand("false")
	assert.Error(t, err)
	assert.Equal(t, "", result) // No output for immediate failure
}
