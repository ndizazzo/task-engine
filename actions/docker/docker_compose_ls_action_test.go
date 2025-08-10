package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

// DockerComposeLsActionTestSuite tests the DockerComposeLsAction
type DockerComposeLsActionTestSuite struct {
	suite.Suite
}

// TestDockerComposeLsActionTestSuite runs the DockerComposeLsAction test suite
func TestDockerComposeLsActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeLsActionTestSuite))
}

func (suite *DockerComposeLsActionTestSuite) TestNewDockerComposeLsAction() {
	logger := slog.Default()

	action := NewDockerComposeLsAction(logger)

	suite.NotNil(action)
	suite.Equal("docker-compose-ls-action", action.ID)
	suite.False(action.Wrapped.All)
	suite.Empty(action.Wrapped.Filter)
	suite.Empty(action.Wrapped.Format)
	suite.False(action.Wrapped.Quiet)
	suite.Empty(action.Wrapped.WorkingDir)
}

func (suite *DockerComposeLsActionTestSuite) TestNewDockerComposeLsActionWithOptions() {
	logger := slog.Default()

	action := NewDockerComposeLsAction(logger,
		WithComposeAll(),
		WithComposeFilter("name=myapp"),
		WithComposeFormat("table {{.Name}}\t{{.Status}}"),
		WithComposeQuiet(),
		WithWorkingDir("/path/to/compose"),
	)

	suite.NotNil(action)
	suite.True(action.Wrapped.All)
	suite.Equal("name=myapp", action.Wrapped.Filter)
	suite.Equal("table {{.Name}}\t{{.Status}}", action.Wrapped.Format)
	suite.True(action.Wrapped.Quiet)
	suite.Equal("/path/to/compose", action.Wrapped.WorkingDir)
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_Success() {
	logger := slog.Default()
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml
testapp             stopped             /path/to/compose.yml,/path/to/override.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 2)

	// Check first stack
	suite.Equal("myapp", action.Wrapped.Stacks[0].Name)
	suite.Equal("running", action.Wrapped.Stacks[0].Status)
	suite.Equal("/path/to/docker-compose.yml", action.Wrapped.Stacks[0].ConfigFiles)

	// Check second stack
	suite.Equal("testapp", action.Wrapped.Stacks[1].Name)
	suite.Equal("stopped", action.Wrapped.Stacks[1].Status)
	suite.Equal("/path/to/compose.yml,/path/to/override.yml", action.Wrapped.Stacks[1].ConfigFiles)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_WithAll() {
	logger := slog.Default()
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml
testapp             stopped             /path/to/compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--all").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger, WithComposeAll())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 2)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_WithFilter() {
	logger := slog.Default()
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--filter", "name=myapp").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger, WithComposeFilter("name=myapp"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 1)
	suite.Equal("myapp", action.Wrapped.Stacks[0].Name)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_WithFormat() {
	logger := slog.Default()
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--format", "table {{.Name}}\t{{.Status}}").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger, WithComposeFormat("table {{.Name}}\t{{.Status}}"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 1)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_WithQuiet() {
	logger := slog.Default()
	expectedOutput := `myapp
testapp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--quiet").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger, WithComposeQuiet())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 2)
	suite.Equal("myapp", action.Wrapped.Stacks[0].Name)
	suite.Equal("testapp", action.Wrapped.Stacks[1].Name)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_CommandError() {
	logger := slog.Default()
	expectedError := errors.New("docker compose command failed")

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return("", expectedError)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "docker compose command failed", "Error should contain the command failure message")
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.Stacks)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_ContextCancellation() {
	logger := slog.Default()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return("", context.Canceled)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(ctx)

	suite.Error(err)
	suite.Contains(err.Error(), "context canceled", "Error should contain the context cancellation message")
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.Stacks)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_parseStacks() {
	logger := slog.Default()
	output := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml
testapp             stopped             /path/to/compose.yml,/path/to/override.yml
devapp              created             /path/to/dev-compose.yml`

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.Output = output
	action.Wrapped.parseStacks(output)

	suite.Len(action.Wrapped.Stacks, 3)

	// Check first stack
	suite.Equal("myapp", action.Wrapped.Stacks[0].Name)
	suite.Equal("running", action.Wrapped.Stacks[0].Status)
	suite.Equal("/path/to/docker-compose.yml", action.Wrapped.Stacks[0].ConfigFiles)

	// Check second stack
	suite.Equal("testapp", action.Wrapped.Stacks[1].Name)
	suite.Equal("stopped", action.Wrapped.Stacks[1].Status)
	suite.Equal("/path/to/compose.yml,/path/to/override.yml", action.Wrapped.Stacks[1].ConfigFiles)

	// Check third stack
	suite.Equal("devapp", action.Wrapped.Stacks[2].Name)
	suite.Equal("created", action.Wrapped.Stacks[2].Status)
	suite.Equal("/path/to/dev-compose.yml", action.Wrapped.Stacks[2].ConfigFiles)
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_parseStackLine() {
	logger := slog.Default()
	action := NewDockerComposeLsAction(logger)

	// Test normal line
	line := "myapp               running             /path/to/docker-compose.yml"
	stack := action.Wrapped.parseStackLine(line)

	suite.Equal("myapp", stack.Name)
	suite.Equal("running", stack.Status)
	suite.Equal("/path/to/docker-compose.yml", stack.ConfigFiles)

	// Test line with multiple config files
	line = "testapp             stopped             /path/to/compose.yml,/path/to/override.yml"
	stack = action.Wrapped.parseStackLine(line)

	suite.Equal("testapp", stack.Name)
	suite.Equal("stopped", stack.Status)
	suite.Equal("/path/to/compose.yml,/path/to/override.yml", stack.ConfigFiles)

	// Test line with extra whitespace
	line = "  devapp    created    /path/to/dev-compose.yml  "
	stack = action.Wrapped.parseStackLine(line)

	suite.Equal("devapp", stack.Name)
	suite.Equal("created", stack.Status)
	suite.Equal("/path/to/dev-compose.yml", stack.ConfigFiles)

	// Test line with tab separators
	line = "prodapp\tstopped\t/path/to/prod-compose.yml"
	stack = action.Wrapped.parseStackLine(line)

	suite.Equal("prodapp", stack.Name)
	suite.Equal("stopped", stack.Status)
	suite.Equal("/path/to/prod-compose.yml", stack.ConfigFiles)

	// Test line with mixed separators
	line = "mixedapp\t running \t/path/to/mixed-compose.yml,/path/to/override.yml"
	stack = action.Wrapped.parseStackLine(line)

	suite.Equal("mixedapp", stack.Name)
	suite.Equal("running", stack.Status)
	suite.Equal("/path/to/mixed-compose.yml,/path/to/override.yml", stack.ConfigFiles)
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_EmptyOutput() {
	logger := slog.Default()
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Empty(action.Wrapped.Stacks)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_OutputWithTrailingWhitespace() {
	logger := slog.Default()
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml
testapp             stopped             /path/to/compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action := NewDockerComposeLsAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 2)
	mockRunner.AssertExpectations(suite.T())
}
