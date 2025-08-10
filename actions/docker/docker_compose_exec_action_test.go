package docker_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

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

	action := docker.NewDockerComposeExecAction(suite.logger, dummyWorkingDir, service, cmdArgs...)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", service, "echo", "hello").Return("hello\n", nil).Once()

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposeExecTestSuite) TestExecuteFailure() {
	service := "db"
	cmdArgs := []string{"invalid-command"}
	dummyWorkingDir := "/tmp/exec-test"

	action := docker.NewDockerComposeExecAction(suite.logger, dummyWorkingDir, service, cmdArgs...)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	failError := fmt.Errorf("command not found")
	suite.mockRunner.On("RunCommandInDirWithContext", context.Background(), dummyWorkingDir, "docker", "compose", "exec", service, "invalid-command").Return("", failError).Once()

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.ErrorContains(err, failError.Error())
	suite.ErrorContains(err, "failed to run docker compose exec")
	suite.mockRunner.AssertExpectations(suite.T())
}

func TestDockerComposeExecTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeExecTestSuite))
}
