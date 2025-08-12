package docker_test

import (
	"context"
	"strings"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
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

// Tests for new constructor pattern with parameters
func (suite *DockerGenericActionTestSuite) TestNewDockerGenericActionConstructor_WithParameters() {
	logger := command_mock.NewDiscardLogger()
	dockerCmd := []string{"network", "ls", "-q", "--filter", "name=test_net"}

	// Create constructor and action with parameters
	constructor := docker.NewDockerGenericAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: dockerCmd},
	)

	suite.Require().NoError(err)
	suite.NotNil(action)
	suite.Equal("docker-generic-action", action.ID)
	suite.NotNil(action.Wrapped)
}

func (suite *DockerGenericActionTestSuite) TestNewDockerGenericActionConstructor_Execute_WithStringSliceParameter() {
	logger := command_mock.NewDiscardLogger()
	dockerCmd := []string{"network", "ls", "-q", "--filter", "name=test_net"}
	expectedOutput := "abcdef123456"

	// Create constructor and action with parameters
	constructor := docker.NewDockerGenericAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: dockerCmd},
	)

	suite.Require().NoError(err)
	action.Wrapped.CommandProcessor = suite.mockProcessor

	suite.mockProcessor.On("RunCommand", "docker", "network", "ls", "-q", "--filter", "name=test_net").Return(expectedOutput+"\n", nil)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Equal(expectedOutput, action.Wrapped.Output)
}

func (suite *DockerGenericActionTestSuite) TestNewDockerGenericActionConstructor_Execute_WithStringParameter() {
	logger := command_mock.NewDiscardLogger()
	dockerCmdStr := "network ls -q --filter name=test_net"
	expectedOutput := "abcdef123456"

	// Create constructor and action with string parameter
	constructor := docker.NewDockerGenericAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: dockerCmdStr},
	)

	suite.Require().NoError(err)
	action.Wrapped.CommandProcessor = suite.mockProcessor

	suite.mockProcessor.On("RunCommand", "docker", "network", "ls", "-q", "--filter", "name=test_net").Return(expectedOutput+"\n", nil)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Equal(expectedOutput, action.Wrapped.Output)
}

func (suite *DockerGenericActionTestSuite) TestNewDockerGenericActionConstructor_Execute_InvalidParameterType() {
	logger := command_mock.NewDiscardLogger()
	invalidParam := 123

	// Create constructor and action with invalid parameter type
	constructor := docker.NewDockerGenericAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: invalidParam},
	)

	suite.Require().NoError(err)
	action.Wrapped.CommandProcessor = suite.mockProcessor

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "docker command parameter is not a string slice or string")
}

func (suite *DockerGenericActionTestSuite) TestNewDockerGenericActionConstructor_Execute_CommandFailure() {
	logger := command_mock.NewDiscardLogger()
	dockerCmd := []string{"info", "--invalid-flag"}
	expectedOutput := "Error: unknown flag: --invalid-flag"

	// Create constructor and action with parameters
	constructor := docker.NewDockerGenericAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: dockerCmd},
	)

	suite.Require().NoError(err)
	action.Wrapped.CommandProcessor = suite.mockProcessor

	suite.mockProcessor.On("RunCommand", "docker", "info", "--invalid-flag").Return(expectedOutput, assert.AnError)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to run docker command")
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Equal(strings.TrimSpace(expectedOutput), action.Wrapped.Output)
}

func (suite *DockerGenericActionTestSuite) TestDockerGenericAction_GetOutput() {
	action := &docker.DockerGenericAction{
		DockerCmd: []string{"network", "ls"},
		Output:    "network1\nnetwork2",
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal([]string{"network", "ls"}, m["command"])
	suite.Equal("network1\nnetwork2", m["output"])
	suite.Equal(true, m["success"])
}

func TestDockerGenericTestSuite(t *testing.T) {
	suite.Run(t, new(DockerGenericActionTestSuite))
}
