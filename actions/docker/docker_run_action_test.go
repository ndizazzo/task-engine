package docker_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ndizazzo/task-engine/actions/docker"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DockerRunTestSuite struct {
	suite.Suite
	mockProcessor *command_mock.MockCommandRunner
}

func (suite *DockerRunTestSuite) SetupTest() {
	suite.mockProcessor = new(command_mock.MockCommandRunner)
}

func (suite *DockerRunTestSuite) TestExecuteSuccess() {
	image := "hello-world:latest"
	runArgs := []string{"--rm", image}
	logger := command_mock.NewDiscardLogger()
	action := docker.NewDockerRunAction(logger, image, nil, runArgs...)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	expectedOutput := "Hello from Docker! ...some more output..."
	suite.mockProcessor.On("RunCommand", "docker", "run", "--rm", image).Return(expectedOutput+"\n  ", nil) // Simulate untrimmed output

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Equal(strings.TrimSpace(expectedOutput), action.Wrapped.Output)
}

func (suite *DockerRunTestSuite) TestExecuteSuccessWithCommand() {
	image := "busybox:latest"
	runArgs := []string{"--rm", image, "echo", "hello from busybox"}
	logger := command_mock.NewDiscardLogger()
	action := docker.NewDockerRunAction(logger, image, nil, runArgs...)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	expectedOutput := "hello from busybox"
	suite.mockProcessor.On("RunCommand", "docker", "run", "--rm", image, "echo", "hello from busybox").Return(expectedOutput+"\n", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Equal(expectedOutput, action.Wrapped.Output)
}

func (suite *DockerRunTestSuite) TestExecuteCommandFailure() {
	image := "nonexistent-image:latest"
	runArgs := []string{"--rm", image}
	logger := command_mock.NewDiscardLogger()
	action := docker.NewDockerRunAction(logger, image, nil, runArgs...)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	expectedOutput := "Error: image not found..."
	suite.mockProcessor.On("RunCommand", "docker", "run", "--rm", image).Return(expectedOutput+" ", assert.AnError)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to run docker container")
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Equal(strings.TrimSpace(expectedOutput), action.Wrapped.Output) // Check output stored even on error
}

func (suite *DockerRunTestSuite) TestExecuteSuccessWithBuffer() {
	image := "alpine"
	runArgs := []string{"--rm", image, "echo", "-n", "buffer test"} // Use -n to avoid trailing newline
	logger := command_mock.NewDiscardLogger()
	var buffer bytes.Buffer // Create buffer

	action := docker.NewDockerRunAction(logger, image, &buffer, runArgs...)
	action.Wrapped.SetCommandRunner(suite.mockProcessor)

	expectedOutput := "buffer test"
	suite.mockProcessor.On("RunCommand", "docker", "run", "--rm", image, "echo", "-n", "buffer test").Return(expectedOutput, nil)

	err := action.Wrapped.Execute(context.Background())
	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Equal(expectedOutput, action.Wrapped.Output) // Check internal field too
	suite.Equal(expectedOutput, buffer.String())       // Check buffer content
}

func TestDockerRunTestSuite(t *testing.T) {
	suite.Run(t, new(DockerRunTestSuite))
}
