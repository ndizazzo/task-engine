package system_test

import (
	"context"
	"testing"
	"time"

	"github.com/ndizazzo/task-engine/actions/system"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

type ShutdownActionTestSuite struct {
	suite.Suite
	mockProcessor *command_mock.MockCommandRunner
}

func (suite *ShutdownActionTestSuite) SetupTest() {
	suite.mockProcessor = new(command_mock.MockCommandRunner)
}

func (suite *ShutdownActionTestSuite) TestRun_DefaultShutdownCommand() {
	delay := 0 * time.Second
	action := system.NewShutdownAction(delay, system.ShutdownOperation_Shutdown, nil)

	suite.mockProcessor.On("RunCommand", "shutdown", "-h", "now").Return("", nil)

	action.Wrapped.CommandProcessor = suite.mockProcessor

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "shutdown", "-h", "now")
}

func (suite *ShutdownActionTestSuite) TestRun_RestartWithNumericDelay() {
	delay := 5 * time.Second
	action := system.NewShutdownAction(delay, system.ShutdownOperation_Restart, nil)

	suite.mockProcessor.On("RunCommand", "shutdown", "-r", "+5").Return("", nil)

	action.Wrapped.CommandProcessor = suite.mockProcessor

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "shutdown", "-r", "+5")
}

func (suite *ShutdownActionTestSuite) TestRun_RestartWithZeroDelay() {
	delay := 0 * time.Second
	action := system.NewShutdownAction(delay, system.ShutdownOperation_Restart, nil)

	suite.mockProcessor.On("RunCommand", "shutdown", "-r", "now").Return("", nil)

	action.Wrapped.CommandProcessor = suite.mockProcessor

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "shutdown", "-r", "now")
}

func TestShutdownActionTestSuite(t *testing.T) {
	suite.Run(t, new(ShutdownActionTestSuite))
}
