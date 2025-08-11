package file_test

import (
	"context"
	"os"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChangeOwnershipTestSuite struct {
	suite.Suite
	mockRunner *command_mock.MockCommandRunner
	tempFile   string
}

func (suite *ChangeOwnershipTestSuite) SetupTest() {
	suite.mockRunner = new(command_mock.MockCommandRunner)

	file, err := os.CreateTemp("", "ownership_test_*.txt")
	suite.NoError(err)
	suite.tempFile = file.Name()
	file.Close()
}

func (suite *ChangeOwnershipTestSuite) TearDownTest() {
	os.Remove(suite.tempFile)
}

func (suite *ChangeOwnershipTestSuite) TestNewChangeOwnershipAction_ValidInputs() {
	logger := command_mock.NewDiscardLogger()
	ownershipAction := file.NewChangeOwnershipAction(logger)
	action, err := ownershipAction.WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "user"},
		task_engine.StaticParameter{Value: "group"},
		false,
	)

	suite.Require().NoError(err)
	suite.NotNil(action)
	suite.Equal("change-ownership-action", action.ID)
}

func (suite *ChangeOwnershipTestSuite) TestNewChangeOwnershipAction_InvalidInputs() {
	logger := command_mock.NewDiscardLogger()

	// Empty path should error on Execute
	ownershipAction1 := file.NewChangeOwnershipAction(logger)
	action1, err := ownershipAction1.WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: "user"},
		task_engine.StaticParameter{Value: "group"},
		false,
	)
	suite.Require().NoError(err)
	execErr := action1.Execute(context.Background())
	suite.Error(execErr)
	suite.Contains(execErr.Error(), "path cannot be empty")

	// Empty owner and group should error on Execute
	ownershipAction2 := file.NewChangeOwnershipAction(logger)
	action2, err := ownershipAction2.WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: ""},
		false,
	)
	suite.Require().NoError(err)
	execErr = action2.Execute(context.Background())
	suite.Error(execErr)
	suite.Contains(execErr.Error(), "at least one of owner or group must be specified")
}

func (suite *ChangeOwnershipTestSuite) TestExecute_OwnerAndGroup() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	ownershipAction := file.NewChangeOwnershipAction(logger)
	action, err := ownershipAction.WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "testuser"},
		task_engine.StaticParameter{Value: "testgroup"},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chown", "testuser:testgroup", suite.tempFile).Return("", nil)

	err = action.Wrapped.Execute(ctx)

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestExecute_OwnerOnly() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	ownershipAction := file.NewChangeOwnershipAction(logger)
	action, err := ownershipAction.WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "testuser"},
		task_engine.StaticParameter{Value: ""},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chown", "testuser", suite.tempFile).Return("", nil)

	err = action.Wrapped.Execute(ctx)

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestExecute_GroupOnly() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	ownershipAction := file.NewChangeOwnershipAction(logger)
	action, err := ownershipAction.WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: "testgroup"},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chown", ":testgroup", suite.tempFile).Return("", nil)

	err = action.Wrapped.Execute(ctx)

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestExecute_Recursive() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	ownershipAction := file.NewChangeOwnershipAction(logger)
	action, err := ownershipAction.WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "testuser"},
		task_engine.StaticParameter{Value: "testgroup"},
		true,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chown", "-R", "testuser:testgroup", suite.tempFile).Return("", nil)

	err = action.Wrapped.Execute(ctx)

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestExecute_NonExistentPath() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	ownershipAction := file.NewChangeOwnershipAction(logger)
	action, err := ownershipAction.WithParameters(
		task_engine.StaticParameter{Value: "/nonexistent/path"},
		task_engine.StaticParameter{Value: "testuser"},
		task_engine.StaticParameter{Value: "testgroup"},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	err = action.Wrapped.Execute(ctx)

	suite.Error(err)
	suite.Contains(err.Error(), "path does not exist")
}

func (suite *ChangeOwnershipTestSuite) TestExecute_CommandFailure() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	ownershipAction := file.NewChangeOwnershipAction(logger)
	action, err := ownershipAction.WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "testuser"},
		task_engine.StaticParameter{Value: "testgroup"},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chown", "testuser:testgroup", suite.tempFile).Return("permission denied", assert.AnError)

	err = action.Wrapped.Execute(ctx)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to change ownership")
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestChangeOwnershipAction_GetOutput() {
	action := &file.ChangeOwnershipAction{
		Path:      "/tmp/testfile",
		Owner:     "testuser",
		Group:     "testgroup",
		Recursive: true,
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("/tmp/testfile", m["path"])
	suite.Equal("testuser", m["owner"])
	suite.Equal("testgroup", m["group"])
	suite.Equal(true, m["recursive"])
	suite.Equal(true, m["success"])
}

func TestChangeOwnershipTestSuite(t *testing.T) {
	suite.Run(t, new(ChangeOwnershipTestSuite))
}
