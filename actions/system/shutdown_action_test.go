package system_test

import (
	"testing"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
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
	action, err := system.NewShutdownAction(nil).WithParameters(task_engine.StaticParameter{Value: "shutdown"}, task_engine.StaticParameter{Value: delay})
	suite.Require().NoError(err)

	suite.mockProcessor.On("RunCommand", "shutdown", "-h", "now").Return("", nil)

	action.Wrapped.CommandProcessor = suite.mockProcessor

	err = action.Execute(suite.T().Context())

	suite.NoError(err)
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "shutdown", "-h", "now")
}

func (suite *ShutdownActionTestSuite) TestRun_RestartWithNumericDelay() {
	delay := 5 * time.Second
	action, err := system.NewShutdownAction(nil).WithParameters(task_engine.StaticParameter{Value: "restart"}, task_engine.StaticParameter{Value: delay})
	suite.Require().NoError(err)

	suite.mockProcessor.On("RunCommand", "shutdown", "-r", "+5").Return("", nil)

	action.Wrapped.CommandProcessor = suite.mockProcessor

	err = action.Execute(suite.T().Context())

	suite.NoError(err)
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "shutdown", "-r", "+5")
}

func (suite *ShutdownActionTestSuite) TestRun_RestartWithZeroDelay() {
	delay := 0 * time.Second
	action, err := system.NewShutdownAction(nil).WithParameters(task_engine.StaticParameter{Value: "restart"}, task_engine.StaticParameter{Value: delay})
	suite.Require().NoError(err)

	suite.mockProcessor.On("RunCommand", "shutdown", "-r", "now").Return("", nil)

	action.Wrapped.CommandProcessor = suite.mockProcessor

	err = action.Execute(suite.T().Context())

	suite.NoError(err)
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "shutdown", "-r", "now")
}

func TestShutdownActionTestSuite(t *testing.T) {
	suite.Run(t, new(ShutdownActionTestSuite))
}

func (suite *ShutdownActionTestSuite) TestShutdownAction_SetCommandRunner() {
	delay := 0 * time.Second
	action, err := system.NewShutdownAction(nil).WithParameters(
		task_engine.StaticParameter{Value: "shutdown"},
		task_engine.StaticParameter{Value: delay},
	)
	suite.Require().NoError(err)

	// Use the setter to cover SetCommandRunner
	action.Wrapped.SetCommandRunner(suite.mockProcessor)
	suite.mockProcessor.On("RunCommand", "shutdown", "-h", "now").Return("", nil)

	err = action.Execute(suite.T().Context())
	suite.NoError(err)
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "shutdown", "-h", "now")
}

func (suite *ShutdownActionTestSuite) TestShutdownAction_GetOutput() {
	action := &system.ShutdownAction{}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(true, m["success"])
}
