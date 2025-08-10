package docker_test

import (
	"context"
	"testing"

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

	action := docker.NewDockerComposeUpAction(logger, dummyWorkingDir)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "up", "-d").Return("Network up-test_default created...", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecuteSuccessWithServices() {
	logger := command_mock.NewDiscardLogger()
	services := []string{"web", "db"}
	dummyWorkingDir := testUpWorkingDir

	action := docker.NewDockerComposeUpAction(logger, dummyWorkingDir, services...)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "up", "-d", "web", "db").Return("Container up-test-web-1 Started...", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecuteCommandFailure() {
	logger := command_mock.NewDiscardLogger()
	services := []string{"web"}
	dummyWorkingDir := testUpWorkingDir

	action := docker.NewDockerComposeUpAction(logger, dummyWorkingDir, services...)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "up", "-d", "web").Return("error output", assert.AnError)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to run docker compose up")
	suite.Contains(err.Error(), "error output")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerComposeUpTestSuite) TestExecuteWithEmptyWorkingDir() {
	logger := command_mock.NewDiscardLogger()
	emptyWorkingDir := ""

	action := docker.NewDockerComposeUpAction(logger, emptyWorkingDir)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "compose", "up", "-d").Return("Network default created...", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
}

func TestDockerComposeUpTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeUpTestSuite))
}
