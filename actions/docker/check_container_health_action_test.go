package docker_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

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

	action := docker.NewCheckContainerHealthAction(dummyWorkingDir, serviceName, checkCommand, 3, 10*time.Millisecond, suite.logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("mysqld is alive", nil).Once()

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecuteSuccessAfterRetries() {
	serviceName := testServiceName
	checkCommand := []string{"mysqladmin", "ping", "-h", "localhost"}
	dummyWorkingDir := testWorkingDir

	action := docker.NewCheckContainerHealthAction(dummyWorkingDir, serviceName, checkCommand, 5, 10*time.Millisecond, suite.logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	// Mock failure twice, then success
	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("error: connection refused", assert.AnError).Times(2)
	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("mysqld is alive", nil).Once()

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecuteFailureAfterRetries() {
	serviceName := testServiceName
	checkCommand := []string{"mysqladmin", "ping", "-h", "localhost"}
	maxRetries := 3
	dummyWorkingDir := testWorkingDir

	action := docker.NewCheckContainerHealthAction(dummyWorkingDir, serviceName, checkCommand, maxRetries, 10*time.Millisecond, suite.logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	// Mock failure consistently
	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("error: connection refused", assert.AnError).Times(maxRetries)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), fmt.Sprintf("container %s failed health check after %d retries", serviceName, maxRetries))
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *CheckContainerHealthTestSuite) TestExecuteContextCancellation() {
	serviceName := testServiceName
	checkCommand := []string{"mysqladmin", "ping", "-h", "localhost"}
	dummyWorkingDir := testWorkingDir

	action := docker.NewCheckContainerHealthAction(dummyWorkingDir, serviceName, checkCommand, 5, 100*time.Millisecond, suite.logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond) // Timeout before retry delay
	defer cancel()

	// Mock failure once - use the actual context that will be passed
	suite.mockRunner.On("RunCommandInDirWithContext", ctx, dummyWorkingDir, "docker", "compose", "exec", serviceName, "mysqladmin", "ping", "-h", "localhost").Return("error: connection refused", assert.AnError).Once()

	err := action.Wrapped.Execute(ctx)

	suite.Error(err)
	suite.ErrorIs(err, context.DeadlineExceeded)   // Check for context error
	suite.mockRunner.AssertExpectations(suite.T()) // Should have been called once
}

func TestCheckContainerHealthTestSuite(t *testing.T) {
	suite.Run(t, new(CheckContainerHealthTestSuite))
}
