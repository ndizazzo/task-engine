package docker_test

import (
	"context"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const testUpWorkingDir = "/tmp/docker-up-test" // #nosec G101 - This is a test directory path, not credentials

type DockerComposeUpTestSuite struct {
	suite.Suite
	mockProcessor *command_mock.MockCommandRunner
}

func (suite *DockerComposeUpTestSuite) SetupTest() {
	suite.mockProcessor = new(command_mock.MockCommandRunner)
}

func (suite *DockerComposeUpTestSuite) TestExecuteSuccessNoServices() {
	logger := command_mock.NewDiscardLogger()
	dummyWorkingDir := testUpWorkingDir

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: []string{}},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "up", "-d").Return("Network up-test_default created...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecuteSuccessWithServices() {
	logger := command_mock.NewDiscardLogger()
	services := []string{"web", "db"}
	dummyWorkingDir := testUpWorkingDir

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: services},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "up", "-d", "web", "db").Return("Container up-test-web-1 Started...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecuteCommandFailure() {
	logger := command_mock.NewDiscardLogger()
	services := []string{"web"}
	dummyWorkingDir := testUpWorkingDir

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: services},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "up", "-d", "web").Return("error output", assert.AnError)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to run docker compose up")
	suite.Contains(execErr.Error(), "error output")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecuteWithEmptyWorkingDir() {
	logger := command_mock.NewDiscardLogger()
	emptyWorkingDir := ""

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(
		task_engine.StaticParameter{Value: emptyWorkingDir},
		task_engine.StaticParameter{Value: []string{}},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "compose", "up", "-d").Return("Network default created...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

// ===== PARAMETER-AWARE CONSTRUCTOR TESTS =====

func (suite *DockerComposeUpTestSuite) TestNewDockerComposeUpActionWithParams() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: []string{"web", "db"}}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("docker-compose-up-action", action.ID)
	suite.Equal("Docker Compose Up", action.Name)
	suite.NotNil(action.Wrapped.WorkingDirParam)
	suite.NotNil(action.Wrapped.ServicesParam)
	suite.Equal(workingDirParam, action.Wrapped.WorkingDirParam)
	suite.Equal(servicesParam, action.Wrapped.ServicesParam)
	// Resolved values computed at execution time
}

// ===== PARAMETER RESOLUTION TESTS =====

func (suite *DockerComposeUpTestSuite) TestExecute_WithStaticParameters() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: []string{"web", "db"}}

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "up", "-d", "web", "db").Return("Container up-test-web-1 Started...", nil)

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"web", "db"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithStringServicesParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: "web,db"}

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "up", "-d", "web", "db").Return("Container up-test-web-1 Started...", nil)

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"web", "db"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithSpaceSeparatedServicesParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: "web db redis"}

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "up", "-d", "web", "db", "redis").Return("Container up-test-web-1 Started...", nil)

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"web", "db", "redis"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithActionOutputParameter() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context with action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/from-action",
		"services":   []string{"api", "database"},
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "services",
	}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	// Set up mock expectation with the context that will actually be used
	suite.mockProcessor.On("RunCommandInDirWithContext", ctx, "/tmp/from-action", "docker", "compose", "up", "-d", "api", "database").Return("Container up-test-api-1 Started...", nil)

	execErr := action.Wrapped.Execute(ctx)

	suite.NoError(execErr)
	suite.Equal("/tmp/from-action", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"api", "database"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithTaskOutputParameter() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context with task output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"deployDir": "/tmp/from-task",
		"services":  "frontend,backend",
	})

	workingDirParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "deployDir",
	}
	servicesParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "services",
	}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	// Set up mock expectation with the context that will actually be used
	suite.mockProcessor.On("RunCommandInDirWithContext", ctx, "/tmp/from-task", "docker", "compose", "up", "-d", "frontend", "backend").Return("Container up-test-frontend-1 Started...", nil)

	execErr := action.Wrapped.Execute(ctx)

	suite.NoError(execErr)
	suite.Equal("/tmp/from-task", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"frontend", "backend"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithEntityOutputParameter() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context with entity output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("docker-build", map[string]interface{}{
		"buildDir":      "/tmp/from-build",
		"buildServices": []string{"builder", "cache"},
	})

	workingDirParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "buildDir",
	}
	servicesParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "buildServices",
	}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	// Set up mock expectation with the context that will actually be used
	suite.mockProcessor.On("RunCommandInDirWithContext", ctx, "/tmp/from-build", "docker", "compose", "up", "-d", "builder", "cache").Return("Container up-test-builder-1 Started...", nil)

	execErr := action.Wrapped.Execute(ctx)

	suite.NoError(execErr)
	suite.Equal("/tmp/from-build", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"builder", "cache"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

// ===== PARAMETER ERROR HANDLING TESTS =====

func (suite *DockerComposeUpTestSuite) TestExecute_WithInvalidActionOutputParameter() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context without the referenced action
	globalContext := task_engine.NewGlobalContext()

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "non-existent-action",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "action 'non-existent-action' not found in context")
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithInvalidOutputKey() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context with action output but missing key
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"otherField": "value",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir", // This key doesn't exist in the output
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "output key 'workingDir' not found in action 'config-action'")
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithEmptyActionID() {
	logger := command_mock.NewDiscardLogger()

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "", // Empty ActionID
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "ActionID cannot be empty")
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithNonMapOutput() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context with non-map action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", "not-a-map")

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "action 'config-action' output is not a map, cannot extract key 'workingDir'")
}

// ===== PARAMETER TYPE VALIDATION TESTS =====

func (suite *DockerComposeUpTestSuite) TestExecute_WithNonStringWorkingDirParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: 123}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "working directory parameter is not a string, got int")
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithInvalidServicesParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: 456} // Not a string or slice

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "services parameter is not a string slice or string, got int")
}

// ===== COMPLEX PARAMETER SCENARIOS =====

func (suite *DockerComposeUpTestSuite) TestExecute_WithMixedParameterTypes() {
	logger := command_mock.NewDiscardLogger()
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/from-config",
	})

	// Create action with dynamic working directory but static services
	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(
		task_engine.ActionOutputParameter{ActionID: "config-action", OutputKey: "workingDir"},
		task_engine.StaticParameter{Value: []string{"static-service"}},
	)
	suite.Require().NoError(err)

	action.Wrapped.SetCommandRunner(suite.mockProcessor)
	// Resolved values will be computed at execution time
	suite.NotNil(action.Wrapped.WorkingDirParam)
	suite.NotNil(action.Wrapped.ServicesParam)
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithComplexServicesResolution() {
	logger := command_mock.NewDiscardLogger()
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("build-action", map[string]interface{}{
		"workingDir": "/tmp/build",
	})
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"services": "frontend,backend,cache",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "build-action",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "services",
	}

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	// Set up mock expectation with the context that will actually be used
	suite.mockProcessor.On("RunCommandInDirWithContext", ctx, "/tmp/build", "docker", "compose", "up", "-d", "frontend", "backend", "cache").Return("Container up-test-frontend-1 Started...", nil)

	execErr := action.Wrapped.Execute(ctx)

	suite.NoError(execErr)
	suite.Equal("/tmp/build", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"frontend", "backend", "cache"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

// ===== BACKWARD COMPATIBILITY TESTS =====

func (suite *DockerComposeUpTestSuite) TestBackwardCompatibility_OriginalConstructor() {
	logger := command_mock.NewDiscardLogger()

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/original"},
		task_engine.StaticParameter{Value: []string{"web", "db"}},
	)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.NotNil(action.Wrapped.WorkingDirParam) // Parameters are always set in current implementation
	suite.NotNil(action.Wrapped.ServicesParam)
}

func (suite *DockerComposeUpTestSuite) TestBackwardCompatibility_ExecuteWithoutGlobalContext() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/static"}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}
	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/static", "docker", "compose", "up", "-d", "web").Return("Container up-test-web-1 Started...", nil)

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/static", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"web"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

// ===== EDGE CASES =====

func (suite *DockerComposeUpTestSuite) TestExecute_WithEmptyServicesParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: []string{}} // Empty slice

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "up", "-d").Return("Network up-test_default created...", nil)

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Empty(action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecute_WithSingleServiceStringParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: "single-service"} // Single service without comma or space

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "up", "-d", "single-service").Return("Container up-test-single-service-1 Started...", nil)

	action, err := docker.NewDockerComposeUpAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal([]string{"single-service"}, action.Wrapped.ResolvedServices)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestDockerComposeUpAction_GetOutput() {
	action := &docker.DockerComposeUpAction{
		ResolvedServices:   []string{"web", "db"},
		ResolvedWorkingDir: "/tmp/workdir",
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal([]string{"web", "db"}, m["services"])
	suite.Equal("/tmp/workdir", m["workingDir"])
	suite.Equal(true, m["success"])
}

func TestDockerComposeUpTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeUpTestSuite))
}
