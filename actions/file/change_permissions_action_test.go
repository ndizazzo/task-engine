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

type ChangePermissionsTestSuite struct {
	suite.Suite
	mockRunner *command_mock.MockCommandRunner
	tempFile   string
}

func (suite *ChangePermissionsTestSuite) SetupTest() {
	suite.mockRunner = new(command_mock.MockCommandRunner)

	file, err := os.CreateTemp("", "permissions_test_*.txt")
	suite.NoError(err)
	suite.tempFile = file.Name()
	file.Close()
}

func (suite *ChangePermissionsTestSuite) TearDownTest() {
	os.Remove(suite.tempFile)
}

func (suite *ChangePermissionsTestSuite) TestNewChangePermissionsAction_ValidInputs() {
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewChangePermissionsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "755"},
		false,
	)

	suite.Require().NoError(err)
	suite.NotNil(action)
	suite.Equal("change-permissions-action", action.ID)
}

func (suite *ChangePermissionsTestSuite) TestNewChangePermissionsAction_InvalidInputs() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})

	// Empty path should error on Execute
	action1, err := file.NewChangePermissionsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: "755"},
		false,
	)
	suite.Require().NoError(err)
	execErr := action1.Wrapped.Execute(ctx)
	suite.Error(execErr)
	suite.Contains(execErr.Error(), "path does not exist")

	// Empty permissions should error on Execute
	action2, err := file.NewChangePermissionsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: ""},
		false,
	)
	suite.Require().NoError(err)
	execErr = action2.Wrapped.Execute(ctx)
	suite.Error(execErr)
}

func (suite *ChangePermissionsTestSuite) TestExecute_OctalPermissions() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	action, err := file.NewChangePermissionsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "755"},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chmod", "755", suite.tempFile).Return("", nil)

	err = action.Wrapped.Execute(ctx)

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangePermissionsTestSuite) TestExecute_SymbolicPermissions() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	action, err := file.NewChangePermissionsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "u+x"},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chmod", "u+x", suite.tempFile).Return("", nil)

	err = action.Wrapped.Execute(ctx)

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangePermissionsTestSuite) TestExecute_Recursive() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	action, err := file.NewChangePermissionsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "644"},
		true,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chmod", "-R", "644", suite.tempFile).Return("", nil)

	err = action.Wrapped.Execute(ctx)

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangePermissionsTestSuite) TestExecute_NonExistentPath() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	action, err := file.NewChangePermissionsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/nonexistent/path"},
		task_engine.StaticParameter{Value: "755"},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	err = action.Wrapped.Execute(ctx)

	suite.Error(err)
	suite.Contains(err.Error(), "path does not exist")
}

func (suite *ChangePermissionsTestSuite) TestExecute_CommandFailure() {
	logger := command_mock.NewDiscardLogger()
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, &task_engine.GlobalContext{})
	action, err := file.NewChangePermissionsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.tempFile},
		task_engine.StaticParameter{Value: "755"},
		false,
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", ctx, "chmod", "755", suite.tempFile).Return("invalid permissions", assert.AnError)

	err = action.Wrapped.Execute(ctx)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to change permissions")
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangePermissionsTestSuite) TestChangePermissionsAction_GetOutput() {
	action := &file.ChangePermissionsAction{
		Path:        "/tmp/testfile",
		Permissions: "755",
		Recursive:   true,
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("/tmp/testfile", m["path"])
	suite.Equal("755", m["permissions"])
	suite.Equal(true, m["recursive"])
	suite.Equal(true, m["success"])
}

func TestChangePermissionsTestSuite(t *testing.T) {
	suite.Run(t, new(ChangePermissionsTestSuite))
}
