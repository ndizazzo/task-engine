package file_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/suite"
)

type WriteFileTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *WriteFileTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "write_file_test_*")
	suite.Require().NoError(err)
}

func (suite *WriteFileTestSuite) TearDownTest() {
	_ = os.RemoveAll(suite.tempDir)
}

func (suite *WriteFileTestSuite) TestExecuteSuccessNewFile() {
	targetFile := filepath.Join(suite.tempDir, "subdir", "output.txt")
	expectedContent := []byte("Hello, World!\nThis is a test.")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(targetFile, expectedContent, false, nil, logger)
	suite.Require().NoError(err)

	_, err = os.Stat(targetFile)
	suite.True(os.IsNotExist(err))

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	actualContent, readErr := os.ReadFile(targetFile)
	suite.NoError(readErr)
	suite.Equal(expectedContent, actualContent)
}

func (suite *WriteFileTestSuite) TestExecuteSuccessEmptyContent() {
	targetFile := filepath.Join(suite.tempDir, "empty_marker")
	logger := command_mock.NewDiscardLogger()
	// Use an empty buffer instead of nil content to satisfy validation
	var buffer bytes.Buffer
	action, err := file.NewWriteFileAction(targetFile, nil, false, &buffer, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)
	fileInfo, err := os.Stat(targetFile)
	suite.NoError(err)
	suite.Equal(int64(0), fileInfo.Size())
}

func (suite *WriteFileTestSuite) TestExecuteFailureAlreadyExistsNoOverwrite() {
	targetFile := filepath.Join(suite.tempDir, "existing.txt")
	initialContent := []byte("Initial Content")
	err := os.WriteFile(targetFile, initialContent, 0600)
	suite.Require().NoError(err, "Setup: Failed to create existing file")

	newContent := []byte("New Content")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(targetFile, newContent, false, nil, logger)
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
	suite.ErrorContains(execErr, "already exists and overwrite is set to false")

	actualContent, readErr := os.ReadFile(targetFile)
	suite.NoError(readErr)
	suite.Equal(initialContent, actualContent, "File content should not have changed")
}

func (suite *WriteFileTestSuite) TestExecuteSuccessAlreadyExistsOverwrite() {
	targetFile := filepath.Join(suite.tempDir, "existing_overwrite.txt")
	initialContent := []byte("Initial Content")
	err := os.WriteFile(targetFile, initialContent, 0600)
	suite.Require().NoError(err, "Setup: Failed to create existing file")

	newContent := []byte("New Content - Overwritten")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(targetFile, newContent, true, nil, logger) // overwrite = true
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())
	suite.NoError(execErr)

	actualContent, readErr := os.ReadFile(targetFile)
	suite.NoError(readErr)
	suite.Equal(newContent, actualContent, "File content should have been overwritten")
}

func (suite *WriteFileTestSuite) TestExecuteFailureNoPermissions() {
	readOnlyDir := filepath.Join(suite.tempDir, "read_only")
	err := os.Mkdir(readOnlyDir, 0555)
	suite.Require().NoError(err)

	targetFile := filepath.Join(readOnlyDir, "cant_write_here")
	content := []byte("some content")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(targetFile, content, false, nil, logger) // overwrite = false
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "permission denied")
}

func (suite *WriteFileTestSuite) TestExecuteSuccessWithBuffer() {
	targetFile := filepath.Join(suite.tempDir, "output_from_buffer.txt")
	expectedContent := "Content from buffer"
	logger := command_mock.NewDiscardLogger()
	var buffer bytes.Buffer
	_, err := buffer.WriteString(expectedContent)
	suite.Require().NoError(err)

	action, err := file.NewWriteFileAction(targetFile, nil, true, &buffer, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	actualContent, readErr := os.ReadFile(targetFile)
	suite.NoError(readErr)
	suite.Equal([]byte(expectedContent), actualContent)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionNilLogger() {
	targetFile := filepath.Join(suite.tempDir, "test.txt")
	content := []byte("test content")

	// Should not panic and should allow nil logger
	action, err := file.NewWriteFileAction(targetFile, content, true, nil, nil)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Nil(action.Wrapped.Logger)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionEmptyFilePath() {
	logger := command_mock.NewDiscardLogger()
	content := []byte("test content")

	// Should return error for empty file path
	action, err := file.NewWriteFileAction("", content, true, nil, logger)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionNoContentNoBuffer() {
	targetFile := filepath.Join(suite.tempDir, "test.txt")
	logger := command_mock.NewDiscardLogger()

	// Should return error when neither content nor buffer is provided
	action, err := file.NewWriteFileAction(targetFile, nil, true, nil, logger)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionValidParameters() {
	targetFile := filepath.Join(suite.tempDir, "test.txt")
	content := []byte("test content")
	logger := command_mock.NewDiscardLogger()

	// Should return valid action for valid parameters
	action, err := file.NewWriteFileAction(targetFile, content, true, nil, logger)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Equal("write-file-test.txt", action.ID)
	suite.Equal(targetFile, action.Wrapped.FilePath)
	suite.Equal(content, action.Wrapped.Content)
	suite.True(action.Wrapped.Overwrite)
	suite.Nil(action.Wrapped.InputBuffer)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionWithBuffer() {
	targetFile := filepath.Join(suite.tempDir, "test.txt")
	var buffer bytes.Buffer
	buffer.WriteString("buffer content")
	logger := command_mock.NewDiscardLogger()

	// Should return valid action when using buffer
	action, err := file.NewWriteFileAction(targetFile, nil, false, &buffer, logger)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Equal("write-file-test.txt", action.ID)
	suite.Equal(targetFile, action.Wrapped.FilePath)
	suite.Nil(action.Wrapped.Content)
	suite.False(action.Wrapped.Overwrite)
	suite.Equal(&buffer, action.Wrapped.InputBuffer)
}

func (suite *WriteFileTestSuite) TestExecuteFailureStatError() {
	// Create a path that will cause a stat error (e.g., a path with invalid characters on some systems)
	// This is a bit tricky to test reliably across platforms, so let's test a different scenario
	// where we can't create the parent directory due to permissions

	// Create a read-only directory
	readOnlyDir := filepath.Join(suite.tempDir, "read_only")
	err := os.Mkdir(readOnlyDir, 0555)
	suite.Require().NoError(err, "Setup: Failed to create read-only directory")

	// Try to write to a file in a subdirectory that can't be created
	targetFile := filepath.Join(readOnlyDir, "subdir", "test.txt")
	content := []byte("test content")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(targetFile, content, false, nil, logger)
	suite.Require().NoError(err)

	// Execute the action - should fail because we can't create the parent directory
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
	suite.ErrorContains(execErr, "failed to create directory")
}

func TestWriteFileTestSuite(t *testing.T) {
	suite.Run(t, new(WriteFileTestSuite))
}
