package docker_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

type DockerComposeExecTestSuite struct {
	suite.Suite
	mockRunner *command_mock.MockCommandRunner
	logger     *slog.Logger
}

func (suite *DockerComposeExecTestSuite) SetupTest() {
	suite.mockRunner = new(command_mock.MockCommandRunner)
	suite.logger = command_mock.NewDiscardLogger()
}

func (suite *DockerComposeExecTestSuite) TestExecuteSuccess() {
	service := "web"
	cmdArgs := []string{"echo", "hello"}
	dummyWorkingDir := "/tmp/exec-test"

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: service},
		task_engine.StaticParameter{Value: cmdArgs},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", service, "echo", "hello").Return("hello\n", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestExecuteFailure() {
	service := "db"
	cmdArgs := []string{"invalid-command"}
	dummyWorkingDir := "/tmp/exec-test"

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: service},
		task_engine.StaticParameter{Value: cmdArgs},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	failError := fmt.Errorf("command not found")
	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", service, "invalid-command").Return("", failError).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.ErrorContains(execErr, failError.Error())
	suite.ErrorContains(execErr, "failed to run docker compose exec")
	suite.mockRunner.AssertExpectations(suite.T())
}

// Parameter-aware constructor tests
func (suite *DockerComposeExecTestSuite) TestNewDockerComposeExecActionWithParams() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("Docker Compose Exec", action.Name)
	suite.NotNil(action.Wrapped)
	suite.NotNil(action.Wrapped.WorkingDirParam)
	suite.NotNil(action.Wrapped.ServiceParam)
	suite.NotNil(action.Wrapped.CommandArgsParam)
	// Resolved values are computed at execution time; no static fields to assert here
}

// Parameter resolution tests
func (suite *DockerComposeExecTestSuite) TestExecute_WithStaticParameters() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service", "echo", "hello").Return("hello\n", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"echo", "hello"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithStringCommandArgsParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: "echo hello world"}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service", "echo", "hello", "world").Return("hello world\n", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"echo", "hello", "world"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithActionOutputParameter() {
	// Create a mock global context with action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir":  "/tmp/from-action",
		"serviceName": "api-service",
		"commandArgs": []string{"mysql", "-u", "root", "-p"},
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	serviceParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "serviceName",
	}
	commandArgsParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "commandArgs",
	}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-action", "docker", "compose", "exec", "api-service", "mysql", "-u", "root", "-p").Return("mysql>", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/from-action", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("api-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"mysql", "-u", "root", "-p"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithTaskOutputParameter() {
	// Create a mock global context with task output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"workingDir":  "/tmp/from-task",
		"serviceName": "db-service",
		"commandArgs": []string{"psql", "-U", "postgres"},
	})

	workingDirParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "workingDir",
	}
	serviceParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "serviceName",
	}
	commandArgsParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "commandArgs",
	}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-task", "docker", "compose", "exec", "db-service", "psql", "-U", "postgres").Return("postgres=#", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/from-task", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("db-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"psql", "-U", "postgres"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithEntityOutputParameter() {
	// Create a mock global context with entity output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("docker-build", map[string]interface{}{
		"buildDir":     "/tmp/from-build",
		"buildService": "cache-service",
		"buildCommand": []string{"redis-cli", "ping"},
	})

	workingDirParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "buildDir",
	}
	serviceParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "buildService",
	}
	commandArgsParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "buildCommand",
	}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-build", "docker", "compose", "exec", "cache-service", "redis-cli", "ping").Return("PONG", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/from-build", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("cache-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"redis-cli", "ping"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

// Error handling tests
func (suite *DockerComposeExecTestSuite) TestExecute_WithInvalidActionOutputParameter() {
	// Create a mock global context without the referenced action
	globalContext := task_engine.NewGlobalContext()

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "non-existent-action",
		OutputKey: "workingDir",
	}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithInvalidOutputKey() {
	// Create a mock global context with action output but missing key
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"otherKey": "/tmp/other",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir", // This key doesn't exist
	}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithEmptyActionID() {
	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "", // Empty ActionID
		OutputKey: "workingDir",
	}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithNonMapOutput() {
	// Create a mock global context with non-map action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", "not-a-map")

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

// Parameter type validation tests
func (suite *DockerComposeExecTestSuite) TestExecute_WithNonStringWorkingDirParameter() {
	workingDirParam := task_engine.StaticParameter{Value: 123} // Not a string
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "working directory parameter resolved to non-string value")
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithNonStringServiceParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceParam := task_engine.StaticParameter{Value: 123} // Not a string
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "service parameter resolved to non-string value")
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithInvalidCommandArgsParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: 123} // Not a string or slice

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "command arguments parameter is not a string slice or string, got int")
}

// Complex scenario tests
func (suite *DockerComposeExecTestSuite) TestExecute_WithMixedParameterTypes() {
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/from-action",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	serviceParam := task_engine.StaticParameter{Value: "static-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"static", "command"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-action", "docker", "compose", "exec", "static-service", "static", "command").Return("success", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/from-action", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("static-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"static", "command"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithComplexCommandArgsResolution() {
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/build",
	})
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"serviceName": "frontend-service",
		"commandArgs": []string{"npm", "run", "build", "--production"},
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	serviceParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "serviceName",
	}
	commandArgsParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "commandArgs",
	}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/build", "docker", "compose", "exec", "frontend-service", "npm", "run", "build", "--production").Return("build completed", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/build", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("frontend-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"npm", "run", "build", "--production"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

// Edge case tests
func (suite *DockerComposeExecTestSuite) TestExecute_WithEmptyCommandArgsParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{}} // Empty slice

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service").Return("success", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedService)
	suite.Empty(action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestExecute_WithSingleCommandStringParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: "echo hello"} // Single command as string

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service", "echo", "hello").Return("hello\n", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"echo", "hello"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

// Backward compatibility tests
func (suite *DockerComposeExecTestSuite) TestBackwardCompatibility_OriginalConstructor() {
	workingDir := "/tmp/test-dir"
	service := "web-service"
	commandArgs := []string{"echo", "hello"}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: workingDir},
		task_engine.StaticParameter{Value: service},
		task_engine.StaticParameter{Value: commandArgs},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), workingDir, "docker", "compose", "exec", service, "echo", "hello").Return("hello\n", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(workingDir, action.Wrapped.ResolvedWorkingDir)
	suite.Equal(service, action.Wrapped.ResolvedService)
	suite.Equal(commandArgs, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestBackwardCompatibility_ExecuteWithoutGlobalContext() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceParam := task_engine.StaticParameter{Value: "web-service"}
	commandArgsParam := task_engine.StaticParameter{Value: []string{"echo", "hello"}}

	action, err := docker.NewDockerComposeExecAction(suite.logger).WithParameters(workingDirParam, serviceParam, commandArgsParam)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service", "echo", "hello").Return("hello\n", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedService)
	suite.Equal([]string{"echo", "hello"}, action.Wrapped.ResolvedCommandArgs)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestDockerComposeExecAction_GetOutput() {
	action := &docker.DockerComposeExecAction{
		ResolvedService:     "web",
		ResolvedWorkingDir:  "/tmp/workdir",
		ResolvedCommandArgs: []string{"echo", "hello"},
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("web", m["service"])
	suite.Equal("/tmp/workdir", m["workingDir"])
	suite.Equal([]string{"echo", "hello"}, m["command"])
	suite.Equal(true, m["success"])
}

func TestDockerComposeExecTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeExecTestSuite))
}
