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

// Existing suite definition (ensure only one)
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
	action := file.NewWriteFileAction(targetFile, expectedContent, false, nil, logger)

	_, err := os.Stat(targetFile)
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
	action := file.NewWriteFileAction(targetFile, nil, false, &buffer, logger)

	err := action.Wrapped.Execute(context.Background())
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
	action := file.NewWriteFileAction(targetFile, newContent, false, nil, logger)

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
	action := file.NewWriteFileAction(targetFile, newContent, true, nil, logger) // overwrite = true

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
	action := file.NewWriteFileAction(targetFile, content, false, nil, logger) // overwrite = false

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

	action := file.NewWriteFileAction(targetFile, nil, true, &buffer, logger)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	actualContent, readErr := os.ReadFile(targetFile)
	suite.NoError(readErr)
	suite.Equal([]byte(expectedContent), actualContent)
}

func TestWriteFileTestSuite(t *testing.T) {
	suite.Run(t, new(WriteFileTestSuite))
}
