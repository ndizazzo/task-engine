package file_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/mocks"
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
	action := file.NewChangePermissionsAction(suite.tempFile, "755", false, logger)

	suite.NotNil(action)
	expectedID := "change-permissions-" + strings.ReplaceAll(suite.tempFile, "/", "-")
	suite.Equal(expectedID, action.ID)
}

func (suite *ChangePermissionsTestSuite) TestNewChangePermissionsAction_InvalidInputs() {
	logger := command_mock.NewDiscardLogger()

	suite.Nil(file.NewChangePermissionsAction("", "755", false, logger))
	suite.Nil(file.NewChangePermissionsAction(suite.tempFile, "", false, logger))
}

func (suite *ChangePermissionsTestSuite) TestExecute_OctalPermissions() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangePermissionsAction(suite.tempFile, "755", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chmod", "755", suite.tempFile).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangePermissionsTestSuite) TestExecute_SymbolicPermissions() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangePermissionsAction(suite.tempFile, "u+x", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chmod", "u+x", suite.tempFile).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangePermissionsTestSuite) TestExecute_Recursive() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangePermissionsAction(suite.tempFile, "644", true, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chmod", "-R", "644", suite.tempFile).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangePermissionsTestSuite) TestExecute_NonExistentPath() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangePermissionsAction("/nonexistent/path", "755", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "path does not exist")
}

func (suite *ChangePermissionsTestSuite) TestExecute_CommandFailure() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangePermissionsAction(suite.tempFile, "755", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chmod", "755", suite.tempFile).Return("invalid permissions", assert.AnError)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to change permissions")
	suite.mockRunner.AssertExpectations(suite.T())
}

func TestChangePermissionsTestSuite(t *testing.T) {
	suite.Run(t, new(ChangePermissionsTestSuite))
}
