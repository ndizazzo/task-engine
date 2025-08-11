package mocks

import (
	"context"
	"io"
	"log/slog"

	engine "github.com/ndizazzo/task-engine"
	"github.com/stretchr/testify/mock"
)

// MockCommandRunner is a mock implementation of CommandRunner for testing
type MockCommandRunner struct {
	mock.Mock
}

// RunCommand mocks the RunCommand method
func (m *MockCommandRunner) RunCommand(command string, args ...string) (string, error) {
	arguments := make([]interface{}, len(args)+1)
	arguments[0] = command
	for i, arg := range args {
		arguments[i+1] = arg
	}

	ret := m.Called(arguments...)
	return ret.String(0), ret.Error(1)
}

// RunCommandWithContext mocks the RunCommandWithContext method
func (m *MockCommandRunner) RunCommandWithContext(ctx context.Context, command string, args ...string) (string, error) {
	arguments := make([]interface{}, len(args)+2)
	arguments[0] = ctx
	arguments[1] = command
	for i, arg := range args {
		arguments[i+2] = arg
	}

	ret := m.Called(arguments...)
	return ret.String(0), ret.Error(1)
}

// RunCommandInDir mocks the RunCommandInDir method
func (m *MockCommandRunner) RunCommandInDir(workingDir string, command string, args ...string) (string, error) {
	arguments := make([]interface{}, len(args)+2)
	arguments[0] = workingDir
	arguments[1] = command
	for i, arg := range args {
		arguments[i+2] = arg
	}

	ret := m.Called(arguments...)
	return ret.String(0), ret.Error(1)
}

// RunCommandInDirWithContext mocks the RunCommandInDirWithContext method
func (m *MockCommandRunner) RunCommandInDirWithContext(ctx context.Context, workingDir string, command string, args ...string) (string, error) {
	arguments := make([]interface{}, len(args)+3)
	arguments[0] = ctx
	arguments[1] = workingDir
	arguments[2] = command
	for i, arg := range args {
		arguments[i+3] = arg
	}

	ret := m.Called(arguments...)
	return ret.String(0), ret.Error(1)
}

// MockActionParameter is a mock implementation of ActionParameter for testing
type MockActionParameter struct {
	ResolveFunc func(ctx context.Context, gc *engine.GlobalContext) (interface{}, error)
}

// Resolve implements the ActionParameter interface
func (m *MockActionParameter) Resolve(ctx context.Context, gc *engine.GlobalContext) (interface{}, error) {
	if m.ResolveFunc != nil {
		return m.ResolveFunc(ctx, gc)
	}
	return nil, nil
}

// NewDiscardLogger creates a logger that discards all output for testing
func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
