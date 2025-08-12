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

const testDownWorkingDir = "/tmp/down-test"

type DockerComposeDownTestSuite struct {
	suite.Suite
	mockProcessor *command_mock.MockCommandRunner
}

func (suite *DockerComposeDownTestSuite) SetupTest() {
	suite.mockProcessor = new(command_mock.MockCommandRunner)
}

func (suite *DockerComposeDownTestSuite) TestExecuteSuccessNoServices() {
	logger := command_mock.NewDiscardLogger()
	dummyWorkingDir := testDownWorkingDir

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: []string{}},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "down").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecuteSuccessWithServices() {
	logger := command_mock.NewDiscardLogger()
	services := []string{"web", "db"}
	dummyWorkingDir := testDownWorkingDir

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: services},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "down", "web", "db").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecuteCommandFailureNoServices() {
	logger := command_mock.NewDiscardLogger()
	dummyWorkingDir := testDownWorkingDir

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: []string{}},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "down").Return("error output", assert.AnError)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to run docker compose down")
	suite.Contains(execErr.Error(), "error output")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecuteCommandFailureWithServices() {
	logger := command_mock.NewDiscardLogger()
	services := []string{"web"}
	dummyWorkingDir := testDownWorkingDir

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: services},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "down", "web").Return("error output", assert.AnError)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to run docker compose down")
	suite.Contains(err.Error(), "error output")
	suite.mockProcessor.AssertExpectations(suite.T())
}

// Parameter-aware constructor tests
func (suite *DockerComposeDownTestSuite) TestNewDockerComposeDownActionWithParams() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: []string{"web", "db"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("Docker Compose Down", action.Name)
	suite.NotNil(action.Wrapped)
	suite.NotNil(action.Wrapped.WorkingDirParam)
	suite.NotNil(action.Wrapped.ServicesParam)
}

// Parameter resolution tests
func (suite *DockerComposeDownTestSuite) TestExecute_WithStaticParameters() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: []string{"web", "db"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "down", "web", "db").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithStringServicesParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: "web,db"}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "down", "web", "db").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithSpaceSeparatedServicesParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: "web db"}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "down", "web", "db").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithActionOutputParameter() {
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

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-action", "docker", "compose", "down", "api", "database").Return("Container down-test-api-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithTaskOutputParameter() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context with task output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"workingDir": "/tmp/from-task",
		"services":   []string{"frontend", "backend"},
	})

	workingDirParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "services",
	}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-task", "docker", "compose", "down", "frontend", "backend").Return("Container down-test-frontend-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithEntityOutputParameter() {
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

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-build", "docker", "compose", "down", "builder", "cache").Return("Container down-test-builder-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

// Error handling tests
func (suite *DockerComposeDownTestSuite) TestExecute_WithInvalidActionOutputParameter() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context without the referenced action
	globalContext := task_engine.NewGlobalContext()

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "non-existent-action",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve working directory parameter")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithInvalidOutputKey() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context with action output but missing key
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"otherKey": "/tmp/other",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir", // This key doesn't exist
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve working directory parameter")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithEmptyActionID() {
	logger := command_mock.NewDiscardLogger()

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "", // Empty ActionID
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to resolve working directory parameter")
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithNonMapOutput() {
	logger := command_mock.NewDiscardLogger()

	// Create a mock global context with non-map action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", "not-a-map")

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

// Parameter type validation tests
func (suite *DockerComposeDownTestSuite) TestExecute_WithNonStringWorkingDirParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: 123} // Not a string
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "working directory parameter is not a string, got int")
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithInvalidServicesParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: 123} // Not a string or slice

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "services parameter is not a string slice or string, got int")
}

// Complex scenario tests
func (suite *DockerComposeDownTestSuite) TestExecute_WithMixedParameterTypes() {
	logger := command_mock.NewDiscardLogger()
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/from-action",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.StaticParameter{Value: []string{"static-service"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-action", "docker", "compose", "down", "static-service").Return("Container down-test-static-service-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithComplexServicesResolution() {
	logger := command_mock.NewDiscardLogger()
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/build",
	})
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"services": []string{"frontend", "backend", "cache"},
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	servicesParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "services",
	}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/build", "docker", "compose", "down", "frontend", "backend", "cache").Return("Container down-test-frontend-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

// Edge case tests
func (suite *DockerComposeDownTestSuite) TestExecute_WithEmptyServicesParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: []string{}} // Empty slice

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "down").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecute_WithSingleServiceStringParameter() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: "web"} // Single service as string

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "down", "web").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

// Backward compatibility tests
func (suite *DockerComposeDownTestSuite) TestBackwardCompatibility_OriginalConstructor() {
	logger := command_mock.NewDiscardLogger()
	workingDir := "/tmp/test-dir"
	services := []string{"web", "db"}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(
		task_engine.StaticParameter{Value: workingDir},
		task_engine.StaticParameter{Value: services},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), workingDir, "docker", "compose", "down", "web", "db").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestBackwardCompatibility_ExecuteWithoutGlobalContext() {
	logger := command_mock.NewDiscardLogger()
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	servicesParam := task_engine.StaticParameter{Value: []string{"web"}}

	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(workingDirParam, servicesParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "down", "web").Return("Container down-test-web-1 Stopped...", nil)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func TestDockerComposeDownTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeDownTestSuite))
}

func (suite *DockerComposeDownTestSuite) TestDockerComposeDownAction_GetOutput() {
	logger := command_mock.NewDiscardLogger()
	action, err := docker.NewDockerComposeDownAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/test"},
		task_engine.StaticParameter{Value: []string{"web"}},
	)
	suite.Require().NoError(err)

	out := action.Wrapped.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(true, m["success"])
}
