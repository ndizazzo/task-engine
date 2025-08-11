package system_test

import (
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/system"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ManageServiceTestSuite struct {
	suite.Suite
	mockProcessor *command_mock.MockCommandRunner
}

func (suite *ManageServiceTestSuite) SetupTest() {
	suite.mockProcessor = new(command_mock.MockCommandRunner)
}

func (suite *ManageServiceTestSuite) TestValidActions() {
	testActions := []struct {
		actionType  string
		serviceName string
		shouldError bool
	}{
		{"start", "mock-service", false},
		{"stop", "mock-service", false},
		{"restart", "mock-service", false},
		{"invalid", "mock-service", true},
	}

	for _, tt := range testActions {
		suite.Run(tt.actionType, func() {
			suite.runActionTest(tt.actionType, tt.serviceName, tt.shouldError)
		})
	}
}

func (suite *ManageServiceTestSuite) runActionTest(actionType, serviceName string, shouldError bool) {
	logger := command_mock.NewDiscardLogger()
	manageAction := system.NewManageServiceAction(logger)
	manageAction.CommandProcessor = suite.mockProcessor
	action, err := manageAction.WithParameters(task_engine.StaticParameter{Value: serviceName}, task_engine.StaticParameter{Value: actionType})
	suite.NoError(err)

	if shouldError {
		suite.mockProcessor.On("RunCommand", "systemctl", actionType, serviceName).Return("", assert.AnError)
	} else {
		suite.mockProcessor.On("RunCommand", "systemctl", actionType, serviceName).Return("success", nil)
	}

	err = action.Wrapped.Execute(suite.T().Context())

	if shouldError {
		suite.Error(err, "Expected an error for invalid action type")
		if actionType != "invalid" {
			suite.mockProcessor.AssertNotCalled(suite.T(), "RunCommand", "systemctl", actionType, serviceName)
		}
	} else {
		suite.NoError(err, "Expected no error for valid action type")
		suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "systemctl", actionType, serviceName)
	}
}

func (suite *ManageServiceTestSuite) TestCommandError() {
	logger := command_mock.NewDiscardLogger()
	manageAction := system.NewManageServiceAction(logger)
	manageAction.CommandProcessor = suite.mockProcessor
	action, err := manageAction.WithParameters(
		task_engine.StaticParameter{Value: "mock-service"},
		task_engine.StaticParameter{Value: "restart"},
	)
	suite.NoError(err)

	suite.mockProcessor.On("RunCommand", "systemctl", "restart", "mock-service").Return("", assert.AnError)

	err = action.Wrapped.Execute(suite.T().Context())

	suite.Error(err, "Expected an error due to simulated command failure")
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "systemctl", "restart", "mock-service")
}

func (suite *ManageServiceTestSuite) TestManageServiceAction_GetOutput() {
	action := &system.ManageServiceAction{
		ServiceName: "nginx",
		ActionType:  "start",
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("nginx", m["service"])
	suite.Equal("start", m["action"])
	suite.Equal(true, m["success"])
}

func TestManageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ManageServiceTestSuite))
}
