package docker_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig())
	suite.NoError(err)

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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		docker.NewDockerComposeLsConfig(
			docker.WithComposeAll(),
			docker.WithComposeFilter("name=myapp"),
			docker.WithComposeFormat("table {{.Name}}\t{{.Status}}"),
			docker.WithComposeLsQuiet(),
			docker.WithWorkingDir("/path/to/compose"),
		),
	)
	suite.NoError(err)

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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 2)
	suite.Equal("myapp", action.Wrapped.Stacks[0].Name)
	suite.Equal("running", action.Wrapped.Stacks[0].Status)
	suite.Equal("/path/to/docker-compose.yml", action.Wrapped.Stacks[0].ConfigFiles)
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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig(docker.WithComposeAll()))
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig(docker.WithComposeFilter("name=myapp")))
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

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
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--format", "table {{.Name}}\t{{.Status}}\t{{.ConfigFiles}}").Return(expectedOutput, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig(docker.WithComposeFormat("table {{.Name}}\t{{.Status}}\t{{.ConfigFiles}}")))
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig(docker.WithComposeLsQuiet()))
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(ctx)

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

	// Create a mock runner that returns our test output
	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(output, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	// Execute the action to trigger parseStacks internally
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	suite.Len(action.Wrapped.Stacks, 3)
	suite.Equal("myapp", action.Wrapped.Stacks[0].Name)
	suite.Equal("running", action.Wrapped.Stacks[0].Status)
	suite.Equal("/path/to/docker-compose.yml", action.Wrapped.Stacks[0].ConfigFiles)
	suite.Equal("testapp", action.Wrapped.Stacks[1].Name)
	suite.Equal("stopped", action.Wrapped.Stacks[1].Status)
	suite.Equal("/path/to/compose.yml,/path/to/override.yml", action.Wrapped.Stacks[1].ConfigFiles)
	suite.Equal("devapp", action.Wrapped.Stacks[2].Name)
	suite.Equal("created", action.Wrapped.Stacks[2].Status)
	suite.Equal("/path/to/dev-compose.yml", action.Wrapped.Stacks[2].ConfigFiles)

	mockRunner.AssertExpectations(suite.T())
}

// Note: parseStackLine is an unexported method, so we test it indirectly through the public Execute method

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_Execute_EmptyOutput() {
	logger := slog.Default()
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

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

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 2)
	mockRunner.AssertExpectations(suite.T())
}

// Parameter-aware constructor tests
func (suite *DockerComposeLsActionTestSuite) TestNewDockerComposeLsActionWithParams() {
	logger := slog.Default()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)

	suite.NotNil(action)
	suite.Equal("docker-compose-ls-with-params-action", action.ID)
	suite.NotNil(action.Wrapped)
	suite.NotNil(action.Wrapped.WorkingDirParam)
	suite.Empty(action.Wrapped.WorkingDir)
}

func (suite *DockerComposeLsActionTestSuite) TestNewDockerComposeLsActionWithParams_WithOptions() {
	logger := slog.Default()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(
		workingDirParam,
		docker.NewDockerComposeLsConfig(
			docker.WithComposeAll(),
			docker.WithComposeFilter("name=myapp"),
			docker.WithComposeFormat("table {{.Name}}\t{{.Status}}"),
			docker.WithComposeLsQuiet(),
		),
	)
	suite.NoError(err)

	suite.NotNil(action)
	suite.True(action.Wrapped.All)
	suite.Equal("name=myapp", action.Wrapped.Filter)
	suite.Equal("table {{.Name}}\t{{.Status}}", action.Wrapped.Format)
	suite.True(action.Wrapped.Quiet)
	suite.Empty(action.Wrapped.WorkingDir)
	suite.NotNil(action.Wrapped.WorkingDirParam)
}

// Parameter resolution tests
func (suite *DockerComposeLsActionTestSuite) TestExecute_WithStaticParameter() {
	logger := slog.Default()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal("/tmp/test-dir", action.Wrapped.WorkingDir)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 1)
	suite.Equal("myapp", action.Wrapped.Stacks[0].Name)
	suite.Equal("running", action.Wrapped.Stacks[0].Status)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestExecute_WithActionOutputParameter() {
	logger := slog.Default()

	// Create a mock global context with action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/from-action",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	expectedOutput := `NAME                STATUS              CONFIG FILES
api-service         running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(err)
	suite.Equal("/tmp/from-action", action.Wrapped.WorkingDir)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 1)
	suite.Equal("api-service", action.Wrapped.Stacks[0].Name)
	suite.Equal("running", action.Wrapped.Stacks[0].Status)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestExecute_WithTaskOutputParameter() {
	logger := slog.Default()

	// Create a mock global context with task output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"workingDir": "/tmp/from-task",
	})

	workingDirParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "workingDir",
	}
	expectedOutput := `NAME                STATUS              CONFIG FILES
frontend-service    running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(err)
	suite.Equal("/tmp/from-task", action.Wrapped.WorkingDir)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 1)
	suite.Equal("frontend-service", action.Wrapped.Stacks[0].Name)
	suite.Equal("running", action.Wrapped.Stacks[0].Status)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestExecute_WithEntityOutputParameter() {
	logger := slog.Default()

	// Create a mock global context with entity output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("docker-build", map[string]interface{}{
		"buildDir": "/tmp/from-build",
	})

	workingDirParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "buildDir",
	}
	expectedOutput := `NAME                STATUS              CONFIG FILES
cache-service       running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(err)
	suite.Equal("/tmp/from-build", action.Wrapped.WorkingDir)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 1)
	suite.Equal("cache-service", action.Wrapped.Stacks[0].Name)
	suite.Equal("running", action.Wrapped.Stacks[0].Status)
	mockRunner.AssertExpectations(suite.T())
}

// Error handling tests
func (suite *DockerComposeLsActionTestSuite) TestExecute_WithInvalidActionOutputParameter() {
	logger := slog.Default()

	// Create a mock global context without the referenced action
	globalContext := task_engine.NewGlobalContext()

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "non-existent-action",
		OutputKey: "workingDir",
	}

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve working directory parameter")
}

func (suite *DockerComposeLsActionTestSuite) TestExecute_WithInvalidOutputKey() {
	logger := slog.Default()

	// Create a mock global context with action output but missing key
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"otherKey": "/tmp/other",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir", // This key doesn't exist
	}

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve working directory parameter")
}

func (suite *DockerComposeLsActionTestSuite) TestExecute_WithEmptyActionID() {
	logger := slog.Default()

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "", // Empty ActionID
		OutputKey: "workingDir",
	}

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve working directory parameter")
}

func (suite *DockerComposeLsActionTestSuite) TestExecute_WithNonMapOutput() {
	logger := slog.Default()

	// Create a mock global context with non-map action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", "not-a-map")

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve working directory parameter")
}

// Parameter type validation tests
func (suite *DockerComposeLsActionTestSuite) TestExecute_WithNonStringWorkingDirParameter() {
	logger := slog.Default()
	workingDirParam := task_engine.StaticParameter{Value: 123} // Not a string

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "working directory parameter resolved to non-string value")
}

// Complex scenario tests
func (suite *DockerComposeLsActionTestSuite) TestExecute_WithMixedParameterTypes() {
	logger := slog.Default()
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/from-action",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	expectedOutput := `NAME                STATUS              CONFIG FILES
static-service      running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls", "--all", "--filter", "name=static-service").Return(expectedOutput, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(
		workingDirParam,
		docker.NewDockerComposeLsConfig(
			docker.WithComposeAll(),
			docker.WithComposeFilter("name=static-service"),
		),
	)
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(err)
	suite.Equal("/tmp/from-action", action.Wrapped.WorkingDir) // From action output
	suite.True(action.Wrapped.All)                             // From static option
	suite.Equal("name=static-service", action.Wrapped.Filter)  // From static option
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 1)
	suite.Equal("static-service", action.Wrapped.Stacks[0].Name)
	mockRunner.AssertExpectations(suite.T())
}

// Backward compatibility tests
func (suite *DockerComposeLsActionTestSuite) TestBackwardCompatibility_OriginalConstructor() {
	logger := slog.Default()
	workingDir := "/tmp/test-dir"

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, docker.NewDockerComposeLsConfig(docker.WithWorkingDir(workingDir)))
	suite.NoError(err)

	suite.NotNil(action)
	suite.Equal(workingDir, action.Wrapped.WorkingDir)
	suite.Nil(action.Wrapped.WorkingDirParam)
}

func (suite *DockerComposeLsActionTestSuite) TestBackwardCompatibility_ExecuteWithoutGlobalContext() {
	logger := slog.Default()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	expectedOutput := `NAME                STATUS              CONFIG FILES
myapp               running             /path/to/docker-compose.yml`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ls").Return(expectedOutput, nil)

	action, err := docker.NewDockerComposeLsAction(logger).WithParameters(workingDirParam, docker.NewDockerComposeLsConfig())
	suite.NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal("/tmp/test-dir", action.Wrapped.WorkingDir)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Stacks, 1)
	suite.Equal("myapp", action.Wrapped.Stacks[0].Name)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeLsActionTestSuite) TestDockerComposeLsAction_GetOutput() {
	action := &docker.DockerComposeLsAction{
		Output: "raw output",
		Stacks: []docker.ComposeStack{
			{Name: "app1", Status: "running", ConfigFiles: "/path/to/compose.yml"},
			{Name: "app2", Status: "stopped", ConfigFiles: "/path/to/compose2.yml"},
		},
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(2, m["count"])
	suite.Equal("raw output", m["output"])
	suite.Equal(true, m["success"])
	suite.Len(m["stacks"], 2)
}
