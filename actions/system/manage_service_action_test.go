package system_test

import (
	"context"
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
	action := system.NewManageServiceAction(serviceName, actionType, logger)
	action.Wrapped.CommandProcessor = suite.mockProcessor

	if shouldError {
		suite.mockProcessor.On("RunCommand", "systemctl", actionType, serviceName).Return("", assert.AnError)
	} else {
		suite.mockProcessor.On("RunCommand", "systemctl", actionType, serviceName).Return("success", nil)
	}

	err := action.Wrapped.Execute(context.Background())

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
	action := &task_engine.Action[*system.ManageServiceAction]{
		ID: "manage-service-command-error",
		Wrapped: &system.ManageServiceAction{
			ServiceName: "mock-service",
			ActionType:  "restart",
			BaseAction: task_engine.BaseAction{
				Logger: logger,
			},
		},
	}

	action.Wrapped.CommandProcessor = suite.mockProcessor

	suite.mockProcessor.On("RunCommand", "systemctl", "restart", "mock-service").Return("", assert.AnError)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err, "Expected an error due to simulated command failure")
	suite.mockProcessor.AssertCalled(suite.T(), "RunCommand", "systemctl", "restart", "mock-service")
}

func TestManageServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ManageServiceTestSuite))
}
