package docker_test

import (
	"context"
	"testing"

	"github.com/ndizazzo/task-engine/actions/docker"
	command_mock "github.com/ndizazzo/task-engine/mocks"
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

	action := docker.NewDockerComposeDownAction(logger, dummyWorkingDir)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "down").Return("Container down-test-web-1 Stopped...", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecuteSuccessWithServices() {
	logger := command_mock.NewDiscardLogger()
	services := []string{"web", "db"}
	dummyWorkingDir := testDownWorkingDir

	action := docker.NewDockerComposeDownAction(logger, dummyWorkingDir, services...)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "down", "web", "db").Return("Container down-test-web-1 Stopped...", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecuteCommandFailureNoServices() {
	logger := command_mock.NewDiscardLogger()
	dummyWorkingDir := testDownWorkingDir

	action := docker.NewDockerComposeDownAction(logger, dummyWorkingDir)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "down").Return("error output", assert.AnError)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to run docker compose down")
	suite.Contains(err.Error(), "error output")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeDownTestSuite) TestExecuteCommandFailureWithServices() {
	logger := command_mock.NewDiscardLogger()
	services := []string{"web"}
	dummyWorkingDir := testDownWorkingDir

	action := docker.NewDockerComposeDownAction(logger, dummyWorkingDir, services...)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "down", "web").Return("error output", assert.AnError)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to run docker compose down")
	suite.Contains(err.Error(), "error output")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func TestDockerComposeDownTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeDownTestSuite))
}
