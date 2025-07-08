package docker_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ndizazzo/task-engine/actions/docker"
	command_mock "github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DockerGenericActionTestSuite struct {
	suite.Suite
	mockProcessor *command_mock.MockCommandRunner
}

func (suite *DockerGenericActionTestSuite) SetupTest() {
	suite.mockProcessor = new(command_mock.MockCommandRunner)
}

func (suite *DockerGenericActionTestSuite) TestExecuteSuccess() {
	dockerCmd := []string{"network", "ls", "-q", "--filter", "name=test_net"}
	expectedOutput := "abcdef123456"
	logger := command_mock.NewDiscardLogger()
	action := docker.NewDockerGenericAction(logger, dockerCmd...)
	action.Wrapped.CommandProcessor = suite.mockProcessor

	suite.mockProcessor.On("RunCommand", "docker", "network", "ls", "-q", "--filter", "name=test_net").Return(expectedOutput+"\n", nil) // Simulate trailing newline

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	// Check that the trimmed output was stored correctly
	suite.Equal(expectedOutput, action.Wrapped.Output)
}

func (suite *DockerGenericActionTestSuite) TestExecuteCommandFailure() {
	dockerCmd := []string{"info", "--invalid-flag"}
	expectedOutput := "Error: unknown flag: --invalid-flag"
	logger := command_mock.NewDiscardLogger()
	action := docker.NewDockerGenericAction(logger, dockerCmd...)
	action.Wrapped.CommandProcessor = suite.mockProcessor

	suite.mockProcessor.On("RunCommand", "docker", "info", "--invalid-flag").Return(expectedOutput, assert.AnError)

	err := action.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to run docker command")
	suite.mockProcessor.AssertExpectations(suite.T())

	// Check that the output was still stored, even on error
	suite.Equal(strings.TrimSpace(expectedOutput), action.Wrapped.Output)
}

func TestDockerGenericTestSuite(t *testing.T) {
	suite.Run(t, new(DockerGenericActionTestSuite))
}
