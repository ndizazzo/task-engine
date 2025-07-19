package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/suite"
)

// ReadFileTestSuite defines the test suite for ReadFileAction
type ReadFileTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *ReadFileTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "read_file_test_*")
	suite.Require().NoError(err)
}

func (suite *ReadFileTestSuite) TearDownTest() {
	_ = os.RemoveAll(suite.tempDir)
}

func (suite *ReadFileTestSuite) TestExecuteSuccess() {
	// Create a test file with content
	testFile := filepath.Join(suite.tempDir, "test.txt")
	expectedContent := []byte("Hello, World!\nThis is a test file with multiple lines.\nAnd some more content.")
	err := os.WriteFile(testFile, expectedContent, 0600)
	suite.Require().NoError(err, "Setup: Failed to create test file")

	// Create output buffer
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the content was read correctly
	suite.Equal(expectedContent, outputBuffer)
}

func (suite *ReadFileTestSuite) TestExecuteSuccessEmptyFile() {
	// Create an empty test file
	testFile := filepath.Join(suite.tempDir, "empty.txt")
	err := os.WriteFile(testFile, []byte{}, 0600)
	suite.Require().NoError(err, "Setup: Failed to create empty test file")

	// Create output buffer
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the content was read correctly (should be empty)
	suite.Equal([]byte{}, outputBuffer)
	suite.Equal(0, len(outputBuffer))
}

func (suite *ReadFileTestSuite) TestExecuteSuccessLargeFile() {
	// Create a test file with larger content
	testFile := filepath.Join(suite.tempDir, "large.txt")
	expectedContent := make([]byte, 1024*1024) // 1MB file
	for i := range expectedContent {
		expectedContent[i] = byte(i % 256)
	}
	err := os.WriteFile(testFile, expectedContent, 0600)
	suite.Require().NoError(err, "Setup: Failed to create large test file")

	// Create output buffer
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the content was read correctly
	suite.Equal(expectedContent, outputBuffer)
	suite.Equal(1024*1024, len(outputBuffer))
}

func (suite *ReadFileTestSuite) TestExecuteFailureFileNotExists() {
	nonExistentFile := filepath.Join(suite.tempDir, "nonexistent.txt")
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(nonExistentFile, &outputBuffer, logger)

	// Execute the action
	err := action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "does not exist")
}

func (suite *ReadFileTestSuite) TestExecuteFailurePathIsDirectory() {
	// Create a directory
	testDir := filepath.Join(suite.tempDir, "testdir")
	err := os.Mkdir(testDir, 0755)
	suite.Require().NoError(err, "Setup: Failed to create test directory")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testDir, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "is a directory, not a file")
}

func (suite *ReadFileTestSuite) TestExecuteFailureNoReadPermission() {
	// Create a test file
	testFile := filepath.Join(suite.tempDir, "no_read.txt")
	content := []byte("some content")
	err := os.WriteFile(testFile, content, 0600)
	suite.Require().NoError(err, "Setup: Failed to create test file")

	// Remove read permissions
	err = os.Chmod(testFile, 0200) // Write-only
	suite.Require().NoError(err, "Setup: Failed to change file permissions")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to read file")
}

func (suite *ReadFileTestSuite) TestNewReadFileActionNilLogger() {
	testFile := filepath.Join(suite.tempDir, "test.txt")
	var outputBuffer []byte

	// Should not panic and should create a default logger
	action := file.NewReadFileAction(testFile, &outputBuffer, nil)
	suite.NotNil(action)
	suite.NotNil(action.Wrapped.Logger)
}

func (suite *ReadFileTestSuite) TestNewReadFileActionEmptyFilePath() {
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()

	// Should return nil for empty file path
	action := file.NewReadFileAction("", &outputBuffer, logger)
	suite.Nil(action)
}

func (suite *ReadFileTestSuite) TestNewReadFileActionNilOutputBuffer() {
	testFile := filepath.Join(suite.tempDir, "test.txt")
	logger := command_mock.NewDiscardLogger()

	// Should return nil for nil output buffer
	action := file.NewReadFileAction(testFile, nil, logger)
	suite.Nil(action)
}

func (suite *ReadFileTestSuite) TestExecuteWithSpecialCharacters() {
	// Create a test file with special characters in content
	testFile := filepath.Join(suite.tempDir, "special.txt")
	expectedContent := []byte("Hello\n\tWorld\r\nSpecial chars: !@#$%^&*()_+-=[]{}|;':\",./<>?")
	err := os.WriteFile(testFile, expectedContent, 0600)
	suite.Require().NoError(err, "Setup: Failed to create test file with special characters")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the content was read correctly
	suite.Equal(expectedContent, outputBuffer)
}

func (suite *ReadFileTestSuite) TestExecuteWithUnicodeContent() {
	// Create a test file with unicode content
	testFile := filepath.Join(suite.tempDir, "unicode.txt")
	expectedContent := []byte("Hello 世界! 🌍 Привет мир! こんにちは世界!")
	err := os.WriteFile(testFile, expectedContent, 0600)
	suite.Require().NoError(err, "Setup: Failed to create test file with unicode content")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the content was read correctly
	suite.Equal(expectedContent, outputBuffer)
}

func (suite *ReadFileTestSuite) TestExecuteOverwritesExistingBuffer() {
	// Create a test file
	testFile := filepath.Join(suite.tempDir, "overwrite.txt")
	expectedContent := []byte("New content")
	err := os.WriteFile(testFile, expectedContent, 0600)
	suite.Require().NoError(err, "Setup: Failed to create test file")

	// Create output buffer with existing content
	outputBuffer := []byte("Old content that should be overwritten")
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the content was overwritten correctly
	suite.Equal(expectedContent, outputBuffer)
}

func (suite *ReadFileTestSuite) TestNewReadFileActionValidParameters() {
	testFile := filepath.Join(suite.tempDir, "test.txt")
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()

	// Should return valid action for valid parameters
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)
	suite.NotNil(action)
	suite.Equal("read-file-test.txt", action.ID)
	suite.Equal(testFile, action.Wrapped.FilePath)
	suite.Equal(&outputBuffer, action.Wrapped.OutputBuffer)
	suite.Equal(logger, action.Wrapped.Logger)
}

func (suite *ReadFileTestSuite) TestExecuteFailureStatError() {
	// Create a test file
	testFile := filepath.Join(suite.tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create test file")

	// Remove read permissions to cause stat error
	err = os.Chmod(testFile, 0000)
	suite.Require().NoError(err, "Setup: Failed to change file permissions")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action := file.NewReadFileAction(testFile, &outputBuffer, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to read file")
}

func TestReadFileTestSuite(t *testing.T) {
	suite.Run(t, new(ReadFileTestSuite))
}
