package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MoveFileTestSuite struct {
	suite.Suite
	mockRunner *command_mock.MockCommandRunner
	tempDir    string
	tempFile   string
}

func (suite *MoveFileTestSuite) SetupTest() {
	suite.mockRunner = new(command_mock.MockCommandRunner)

	var err error
	suite.tempDir, err = os.MkdirTemp("", "move_test_*")
	suite.NoError(err)

	file, err := os.CreateTemp(suite.tempDir, "source_*.txt")
	suite.NoError(err)
	suite.tempFile = file.Name()
	file.Close()
}

func (suite *MoveFileTestSuite) TearDownTest() {
	os.RemoveAll(suite.tempDir)
}

func (suite *MoveFileTestSuite) TestNewMoveFileAction_ValidInputs() {
	logger := command_mock.NewDiscardLogger()
	destination := filepath.Join(suite.tempDir, "destination.txt")
	action := file.NewMoveFileAction(suite.tempFile, destination, false, logger)

	suite.NotNil(action)
	suite.Equal("move-file-"+filepath.Base(suite.tempFile), action.ID)
}

func (suite *MoveFileTestSuite) TestNewMoveFileAction_InvalidInputs() {
	logger := command_mock.NewDiscardLogger()
	destination := filepath.Join(suite.tempDir, "destination.txt")

	suite.Nil(file.NewMoveFileAction("", destination, false, logger))
	suite.Nil(file.NewMoveFileAction(suite.tempFile, "", false, logger))
	suite.Nil(file.NewMoveFileAction(suite.tempFile, suite.tempFile, false, logger))
}

func (suite *MoveFileTestSuite) TestExecute_SimpleMove() {
	logger := command_mock.NewDiscardLogger()
	destination := filepath.Join(suite.tempDir, "destination.txt")
	action := file.NewMoveFileAction(suite.tempFile, destination, false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "mv", suite.tempFile, destination).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *MoveFileTestSuite) TestExecute_WithCreateDirs() {
	logger := command_mock.NewDiscardLogger()
	destination := filepath.Join(suite.tempDir, "subdir", "destination.txt")
	action := file.NewMoveFileAction(suite.tempFile, destination, true, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "mv", suite.tempFile, destination).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())

	_, err = os.Stat(filepath.Dir(destination))
	suite.NoError(err)
}

func (suite *MoveFileTestSuite) TestExecute_NonExistentSource() {
	logger := command_mock.NewDiscardLogger()
	destination := filepath.Join(suite.tempDir, "destination.txt")
	action := file.NewMoveFileAction("/nonexistent/source.txt", destination, false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "source path does not exist")
}

func (suite *MoveFileTestSuite) TestExecute_CommandFailure() {
	logger := command_mock.NewDiscardLogger()
	destination := filepath.Join(suite.tempDir, "destination.txt")
	action := file.NewMoveFileAction(suite.tempFile, destination, false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "mv", suite.tempFile, destination).Return("permission denied", assert.AnError)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to move")
	suite.mockRunner.AssertExpectations(suite.T())
}

func (suite *MoveFileTestSuite) TestExecute_RenameFile() {
	logger := command_mock.NewDiscardLogger()
	destination := filepath.Join(suite.tempDir, "renamed.txt")
	action := file.NewMoveFileAction(suite.tempFile, destination, false, logger)
	action.Wrapped.SetCommandRunner(suite.mockRunner)

	suite.mockRunner.On("RunCommandWithContext", context.Background(), "mv", suite.tempFile, destination).Return("", nil)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.mockRunner.AssertExpectations(suite.T())
}

func TestMoveFileTestSuite(t *testing.T) {
	suite.Run(t, new(MoveFileTestSuite))
}
