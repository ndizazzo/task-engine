package file_test

import (
	"context"
	"os"
	"strings"
	"testing"

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
	action := file.NewChangeOwnershipAction(suite.tempFile, "user", "group", false, logger)

	suite.NotNil(action)
	expectedID := "change-ownership-" + strings.ReplaceAll(suite.tempFile, "/", "-")
	suite.Equal(expectedID, action.ID)
}

func (suite *ChangeOwnershipTestSuite) TestNewChangeOwnershipAction_InvalidInputs() {
	logger := command_mock.NewDiscardLogger()

	suite.Nil(file.NewChangeOwnershipAction("", "user", "group", false, logger))
	suite.Nil(file.NewChangeOwnershipAction(suite.tempFile, "", "", false, logger))
}

func (suite *ChangeOwnershipTestSuite) TestExecute_OwnerAndGroup() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangeOwnershipAction(suite.tempFile, "testuser", "testgroup", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chown", "testuser:testgroup", suite.tempFile).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestExecute_OwnerOnly() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangeOwnershipAction(suite.tempFile, "testuser", "", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chown", "testuser", suite.tempFile).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestExecute_GroupOnly() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangeOwnershipAction(suite.tempFile, "", "testgroup", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chown", ":testgroup", suite.tempFile).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestExecute_Recursive() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangeOwnershipAction(suite.tempFile, "testuser", "testgroup", true, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chown", "-R", "testuser:testgroup", suite.tempFile).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *ChangeOwnershipTestSuite) TestExecute_NonExistentPath() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangeOwnershipAction("/nonexistent/path", "testuser", "testgroup", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "path does not exist")
}

func (suite *ChangeOwnershipTestSuite) TestExecute_CommandFailure() {
	logger := command_mock.NewDiscardLogger()
	action := file.NewChangeOwnershipAction(suite.tempFile, "testuser", "testgroup", false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "chown", "testuser:testgroup", suite.tempFile).Return("permission denied", assert.AnError)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to change ownership")
	suite.mockRunner.AssertExpectations(suite.T())
}

func TestChangeOwnershipTestSuite(t *testing.T) {
	suite.Run(t, new(ChangeOwnershipTestSuite))
}
