package file_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/suite"
)

type DeleteFileActionTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (suite *DeleteFileActionTestSuite) SetupTest() {
	suite.logger = command_mock.NewDiscardLogger()
}

func (suite *DeleteFileActionTestSuite) TestExecuteSuccess() {
	tempDir := suite.T().TempDir()
	filePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(filePath, []byte("test content"), 0600)
	suite.Require().NoError(err, "Failed to create test file")

	action := file.NewDeleteFileAction(suite.logger, filePath)
	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	_, err = os.Stat(filePath)
	suite.True(os.IsNotExist(err), "File should be deleted")
}

func (suite *DeleteFileActionTestSuite) TestExecuteFileNotExists() {
	filePath := filepath.Join(suite.T().TempDir(), "nonexistent.txt")

	action := file.NewDeleteFileAction(suite.logger, filePath)
	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err, "Deleting non-existent file should not return error")
}

func (suite *DeleteFileActionTestSuite) TestExecutePermissionDenied() {
	tempDir := suite.T().TempDir()
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0755)
	suite.Require().NoError(err, "Failed to create directory")

	filePath := filepath.Join(restrictedDir, "test.txt")
	err = os.WriteFile(filePath, []byte("test content"), 0600)
	suite.Require().NoError(err, "Failed to create test file")

	err = os.Chmod(restrictedDir, 0000)
	suite.Require().NoError(err, "Failed to set restricted permissions")

	action := file.NewDeleteFileAction(suite.logger, filePath)
	err = action.Wrapped.Execute(context.Background())

	suite.Error(err, "Deleting file in restricted directory should return error")
	suite.Contains(err.Error(), "permission denied")

	_ = os.Chmod(restrictedDir, 0755)
}

func TestDeleteFileActionTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteFileActionTestSuite))
}
