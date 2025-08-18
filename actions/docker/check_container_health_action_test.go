package docker_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testWorkingDir  = "/tmp/test"
	testServiceName = "db-test"
)

type CheckContainerHealthTestSuite struct {
	suite.Suite
	mockRunner *command_mock.MockCommandRunner
	logger     *slog.Logger
}

func (suite *CheckContainerHealthTestSuite) SetupTest() {
	suite.mockRunner = new(command_mock.MockCommandRunner)
	suite.logger = command_mock.NewDiscardLogger()
}

func (suite *CheckContainerHealthTestSuite) TestExecuteSuccessFirstTry() {
	serviceName := testServiceName
	checkCommand := []string{"mysqladmin", "ping", "-h", "localhost"}
	dummyWorkingDir := testWorkingDir

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: serviceName},
		task_engine.StaticParameter{Value: checkCommand},
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("mysqld is alive", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecuteSuccessAfterRetries() {
	serviceName := testServiceName
	checkCommand := []string{"mysqladmin", "ping", "-h", "localhost"}
	dummyWorkingDir := testWorkingDir

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: serviceName},
		task_engine.StaticParameter{Value: checkCommand},
		task_engine.StaticParameter{Value: 5},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	// Mock failure twice, then success
	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("error: connection refused", assert.AnError).Times(2)
	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("mysqld is alive", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecuteFailureAfterRetries() {
	serviceName := testServiceName
	checkCommand := []string{"mysqladmin", "ping", "-h", "localhost"}
	maxRetries := 3
	dummyWorkingDir := testWorkingDir

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: serviceName},
		task_engine.StaticParameter{Value: checkCommand},
		task_engine.StaticParameter{Value: maxRetries},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	// Mock failure consistently
	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("error: connection refused", assert.AnError).Times(maxRetries)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), fmt.Sprintf("container %s failed health check after %d retries", serviceName, maxRetries))
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecuteContextCancellation() {
	serviceName := testServiceName
	checkCommand := []string{"mysqladmin", "ping", "-h", "localhost"}
	dummyWorkingDir := testWorkingDir

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: dummyWorkingDir},
		task_engine.StaticParameter{Value: serviceName},
		task_engine.StaticParameter{Value: checkCommand},
		task_engine.StaticParameter{Value: 5},
		task_engine.StaticParameter{Value: 100 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond) // Timeout before retry delay
	defer cancel()

	// Mock failure once - use the actual context that will be passed
	suite.mockRunner.On("RunCommandInDirWithContext", ctx, dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("error: connection refused", assert.AnError).Once()

	execErr := action.Wrapped.Execute(ctx)

	suite.Error(execErr)
	suite.ErrorIs(execErr, context.DeadlineExceeded)
	suite.mockRunner.AssertExpectations(suite.T())
}

// Parameter-aware constructor tests
func (suite *CheckContainerHealthTestSuite) TestNewCheckContainerHealthActionWithParams() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)

	suite.Require().NoError(err)
	suite.NotNil(action)
	suite.Equal("Check Container Health", action.Name)
	suite.NotNil(action.Wrapped)
	suite.NotNil(action.Wrapped.WorkingDirParam)
	suite.NotNil(action.Wrapped.ServiceNameParam)
	suite.NotNil(action.Wrapped.CheckCommandParam)
}

// Parameter resolution tests
func (suite *CheckContainerHealthTestSuite) TestExecute_WithStaticParameters() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service", "curl", "-f", "http://localhost:8080/health").Return("HTTP/1.1 200 OK", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedServiceName)
	suite.Equal([]string{"curl", "-f", "http://localhost:8080/health"}, action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithStringCheckCommandParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: "curl -f http://localhost:8080/health"}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service", "curl", "-f", "http://localhost:8080/health").Return("HTTP/1.1 200 OK", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedServiceName)
	suite.Equal([]string{"curl", "-f", "http://localhost:8080/health"}, action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithActionOutputParameter() {
	// Create a mock global context with action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir":   "/tmp/from-action",
		"serviceName":  "api-service",
		"checkCommand": []string{"mysqladmin", "ping", "-h", "localhost"},
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	serviceNameParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "serviceName",
	}
	checkCommandParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "checkCommand",
	}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-action", "docker", "compose", "exec", "api-service", "mysqladmin", "ping", "-h", "localhost").Return("mysqld is alive", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/from-action", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("api-service", action.Wrapped.ResolvedServiceName)
	suite.Equal([]string{"mysqladmin", "ping", "-h", "localhost"}, action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithTaskOutputParameter() {
	// Create a mock global context with task output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"workingDir":   "/tmp/from-task",
		"serviceName":  "db-service",
		"checkCommand": []string{"pg_isready", "-h", "localhost"},
	})

	workingDirParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "workingDir",
	}
	serviceNameParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "serviceName",
	}
	checkCommandParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "checkCommand",
	}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-task", "docker", "compose", "exec", "db-service", "pg_isready", "-h", "localhost").Return("accepting connections", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/from-task", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("db-service", action.Wrapped.ResolvedServiceName)
	suite.Equal([]string{"pg_isready", "-h", "localhost"}, action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithEntityOutputParameter() {
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
	serviceNameParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "buildService",
	}
	checkCommandParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "buildCommand",
	}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-build", "docker", "compose", "exec", "cache-service", "redis-cli", "ping").Return("PONG", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/from-build", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("cache-service", action.Wrapped.ResolvedServiceName)
	suite.Equal([]string{"redis-cli", "ping"}, action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

// Error handling tests
func (suite *CheckContainerHealthTestSuite) TestExecute_WithInvalidActionOutputParameter() {
	// Create a mock global context without the referenced action
	globalContext := task_engine.NewGlobalContext()

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "non-existent-action",
		OutputKey: "workingDir",
	}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(workingDirParam, serviceNameParam, checkCommandParam, task_engine.StaticParameter{Value: 3}, task_engine.StaticParameter{Value: 10 * time.Millisecond})
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithInvalidOutputKey() {
	// Create a mock global context with action output but missing key
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"otherKey": "/tmp/other",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir", // This key doesn't exist
	}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(workingDirParam, serviceNameParam, checkCommandParam, task_engine.StaticParameter{Value: 3}, task_engine.StaticParameter{Value: 10 * time.Millisecond})
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithEmptyActionID() {
	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "", // Empty ActionID
		OutputKey: "workingDir",
	}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(workingDirParam, serviceNameParam, checkCommandParam, task_engine.StaticParameter{Value: 3}, task_engine.StaticParameter{Value: 10 * time.Millisecond})
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithNonMapOutput() {
	// Create a mock global context with non-map action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", "not-a-map")

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(workingDirParam, serviceNameParam, checkCommandParam, task_engine.StaticParameter{Value: 3}, task_engine.StaticParameter{Value: 10 * time.Millisecond})
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "failed to resolve working directory parameter")
}

// Parameter type validation tests
func (suite *CheckContainerHealthTestSuite) TestExecute_WithNonStringWorkingDirParameter() {
	workingDirParam := task_engine.StaticParameter{Value: 123} // Not a string
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(workingDirParam, serviceNameParam, checkCommandParam, task_engine.StaticParameter{Value: 3}, task_engine.StaticParameter{Value: 10 * time.Millisecond})
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "working directory parameter resolved to non-string value")
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithNonStringServiceNameParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceNameParam := task_engine.StaticParameter{Value: 123} // Not a string
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(workingDirParam, serviceNameParam, checkCommandParam, task_engine.StaticParameter{Value: 3}, task_engine.StaticParameter{Value: 10 * time.Millisecond})
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "service name parameter resolved to non-string value")
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithInvalidCheckCommandParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: 123} // Not a string or slice

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(workingDirParam, serviceNameParam, checkCommandParam, task_engine.StaticParameter{Value: 3}, task_engine.StaticParameter{Value: 10 * time.Millisecond})
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "check command parameter is not a string slice or string, got int")
}

// Complex scenario tests
func (suite *CheckContainerHealthTestSuite) TestExecute_WithMixedParameterTypes() {
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/from-action",
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	serviceNameParam := task_engine.StaticParameter{Value: "static-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"static", "command"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/from-action", "docker", "compose", "exec", "static-service", "static", "command").Return("success", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/from-action", action.Wrapped.ResolvedWorkingDir)              // From action output
	suite.Equal("static-service", action.Wrapped.ResolvedServiceName)               // From static parameter
	suite.Equal([]string{"static", "command"}, action.Wrapped.ResolvedCheckCommand) // From static parameter
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithComplexCheckCommandResolution() {
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("config-action", map[string]interface{}{
		"workingDir": "/tmp/build",
	})
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"serviceName":  "frontend-service",
		"checkCommand": []string{"curl", "-f", "http://localhost:3000/health", "--max-time", "5"},
	})

	workingDirParam := task_engine.ActionOutputParameter{
		ActionID:  "config-action",
		OutputKey: "workingDir",
	}
	serviceNameParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "serviceName",
	}
	checkCommandParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "checkCommand",
	}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext), "/tmp/build", "docker", "compose", "exec", "frontend-service", "curl", "-f", "http://localhost:3000/health", "--max-time", "5").Return("HTTP/1.1 200 OK", nil).Once()

	execErr := action.Wrapped.Execute(context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext))

	suite.NoError(execErr)
	suite.Equal("/tmp/build", action.Wrapped.ResolvedWorkingDir)                                                                // From action output
	suite.Equal("frontend-service", action.Wrapped.ResolvedServiceName)                                                         // From task output
	suite.Equal([]string{"curl", "-f", "http://localhost:3000/health", "--max-time", "5"}, action.Wrapped.ResolvedCheckCommand) // From task output
	suite.mockRunner.AssertExpectations(suite.T())
}

// Edge case tests
func (suite *CheckContainerHealthTestSuite) TestExecute_WithEmptyCheckCommandParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{}} // Empty slice

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service").Return("success", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedServiceName)
	suite.Empty(action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecute_WithSingleCommandStringParameter() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: "echo hello"} // Single command as string

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		workingDirParam,
		serviceNameParam,
		checkCommandParam,
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service", "echo", "hello").Return("hello", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedServiceName)
	suite.Equal([]string{"echo", "hello"}, action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

// Backward compatibility tests
func (suite *CheckContainerHealthTestSuite) TestBackwardCompatibility_OriginalConstructor() {
	workingDir := "/tmp/test-dir"
	serviceName := "web-service"
	checkCommand := []string{"curl", "-f", "http://localhost:8080/health"}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(
		task_engine.StaticParameter{Value: workingDir},
		task_engine.StaticParameter{Value: serviceName},
		task_engine.StaticParameter{Value: checkCommand},
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 10 * time.Millisecond},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), workingDir, "docker", "compose", "exec", serviceName, "curl", "-f", "http://localhost:8080/health").Return("HTTP/1.1 200 OK", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(workingDir, action.Wrapped.ResolvedWorkingDir)
	suite.Equal(serviceName, action.Wrapped.ResolvedServiceName)
	suite.Equal(checkCommand, action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestBackwardCompatibility_ExecuteWithoutGlobalContext() {
	workingDirParam := task_engine.StaticParameter{Value: "/tmp/test-dir"}
	serviceNameParam := task_engine.StaticParameter{Value: "web-service"}
	checkCommandParam := task_engine.StaticParameter{Value: []string{"curl", "-f", "http://localhost:8080/health"}}

	action, err := docker.NewCheckContainerHealthAction(suite.logger).WithParameters(workingDirParam, serviceNameParam, checkCommandParam, task_engine.StaticParameter{Value: 3}, task_engine.StaticParameter{Value: 10 * time.Millisecond})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), "/tmp/test-dir", "docker", "compose", "exec", "web-service", "curl", "-f", "http://localhost:8080/health").Return("HTTP/1.1 200 OK", nil).Once()

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal("/tmp/test-dir", action.Wrapped.ResolvedWorkingDir)
	suite.Equal("web-service", action.Wrapped.ResolvedServiceName)
	suite.Equal([]string{"curl", "-f", "http://localhost:8080/health"}, action.Wrapped.ResolvedCheckCommand)
	suite.mockRunner.AssertExpectations(suite.T())
}

func TestCheckContainerHealthTestSuite(t *testing.T) {
	suite.Run(t, new(CheckContainerHealthTestSuite))
}

func (suite *CheckContainerHealthTestSuite) TestGetOutput() {
	logger := command_mock.NewDiscardLogger()
	action, err := docker.NewCheckContainerHealthAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/workdir"},
		task_engine.StaticParameter{Value: "db"},
		task_engine.StaticParameter{Value: []string{"true"}},
		task_engine.StaticParameter{Value: 3},
		task_engine.StaticParameter{Value: 100 * time.Millisecond},
	)
	suite.Require().NoError(err)

	// Manually set resolved fields to simulate post-execute state
	action.Wrapped.ResolvedWorkingDir = "/tmp/workdir"
	action.Wrapped.ResolvedServiceName = "db"
	action.Wrapped.ResolvedCheckCommand = []string{"true"}
	action.Wrapped.ResolvedMaxRetries = 3
	action.Wrapped.ResolvedRetryDelay = 100 * time.Millisecond

	out := action.Wrapped.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("db", m["service"])
	suite.Equal([]string{"true"}, m["command"])
	suite.Equal(3, m["maxRetries"])
	suite.Equal("100ms", m["retryDelay"])
	suite.Equal("/tmp/workdir", m["workingDir"])
	suite.Equal(true, m["success"])
}
