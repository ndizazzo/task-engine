package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

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
	testFile := filepath.Join(suite.tempDir, "test.txt")
	expectedContent := []byte("Hello, World!")
	err := os.WriteFile(testFile, expectedContent, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create test file")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	suite.Equal(expectedContent, outputBuffer)
}

func (suite *ReadFileTestSuite) TestExecuteSuccessEmptyFile() {
	testFile := filepath.Join(suite.tempDir, "empty.txt")
	err := os.WriteFile(testFile, []byte{}, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create empty test file")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	suite.Equal([]byte{}, outputBuffer)
	suite.Equal(0, len(outputBuffer))
}

func (suite *ReadFileTestSuite) TestExecuteSuccessLargeFile() {
	testFile := filepath.Join(suite.tempDir, "large.txt")
	expectedContent := make([]byte, 1024*1024) // 1MB file
	for i := range expectedContent {
		expectedContent[i] = byte(i % 256)
	}
	err := os.WriteFile(testFile, expectedContent, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create large test file")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	suite.Equal(expectedContent, outputBuffer)
	suite.Equal(1024*1024, len(outputBuffer))
}

func (suite *ReadFileTestSuite) TestExecuteFailureFileNotExists() {
	nonExistentFile := filepath.Join(suite.tempDir, "nonexistent.txt")
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: nonExistentFile}, &outputBuffer)
	suite.Require().NoError(err)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "does not exist")
}

func (suite *ReadFileTestSuite) TestExecuteFailurePathIsDirectory() {
	testDir := filepath.Join(suite.tempDir, "testdir")
	err := os.Mkdir(testDir, 0o755)
	suite.Require().NoError(err, "Setup: Failed to create test directory")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testDir}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "is a directory, not a file")
}

func (suite *ReadFileTestSuite) TestExecuteFailureNoReadPermission() {
	testFile := filepath.Join(suite.tempDir, "no_read.txt")
	content := []byte("some content")
	err := os.WriteFile(testFile, content, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create test file")

	err = os.Chmod(testFile, 0o200) // Write-only
	suite.Require().NoError(err, "Setup: Failed to change file permissions")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to read file")
}

func (suite *ReadFileTestSuite) TestNewReadFileActionNilLogger() {
	testFile := filepath.Join(suite.tempDir, "test.txt")
	var outputBuffer []byte

	action, err := file.NewReadFileAction(nil).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.NoError(err)
	suite.NotNil(action)
	suite.NotNil(action.Wrapped.Logger)
}

func (suite *ReadFileTestSuite) TestNewReadFileActionEmptyFilePath() {
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()

	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: ""}, &outputBuffer)
	suite.NoError(err)
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
}

func (suite *ReadFileTestSuite) TestNewReadFileActionNilOutputBuffer() {
	testFile := filepath.Join(suite.tempDir, "test.txt")
	logger := command_mock.NewDiscardLogger()

	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, nil)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *ReadFileTestSuite) TestExecuteWithSpecialCharacters() {
	testFile := filepath.Join(suite.tempDir, "special.txt")
	expectedContent := []byte("Hello\n\tWorld\r\nSpecial chars: !@#$%^&*()_+-=[]{}|;':\",./<>?")
	err := os.WriteFile(testFile, expectedContent, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create test file with special characters")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	suite.Equal(expectedContent, outputBuffer)
}

func (suite *ReadFileTestSuite) TestExecuteWithUnicodeContent() {
	testFile := filepath.Join(suite.tempDir, "unicode.txt")
	expectedContent := []byte("Hello 世界! 🌍 Привет мир! こんにちは世界!")
	err := os.WriteFile(testFile, expectedContent, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create test file with unicode content")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	suite.Equal(expectedContent, outputBuffer)
}

func (suite *ReadFileTestSuite) TestExecuteOverwritesExistingBuffer() {
	testFile := filepath.Join(suite.tempDir, "overwrite.txt")
	expectedContent := []byte("New content")
	err := os.WriteFile(testFile, expectedContent, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create test file")

	outputBuffer := []byte("Old content that should be overwritten")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	suite.Equal(expectedContent, outputBuffer)
}

func (suite *ReadFileTestSuite) TestNewReadFileActionValidParameters() {
	testFile := filepath.Join(suite.tempDir, "test.txt")
	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()

	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Equal("read-file-action", action.ID)
	suite.Equal(&outputBuffer, action.Wrapped.OutputBuffer)
	suite.Equal(logger, action.Wrapped.Logger)
}

func (suite *ReadFileTestSuite) TestExecuteFailureStatError() {
	testFile := filepath.Join(suite.tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0o600)
	suite.Require().NoError(err, "Setup: Failed to create test file")

	err = os.Chmod(testFile, 0o000)
	suite.Require().NoError(err, "Setup: Failed to change file permissions")

	var outputBuffer []byte
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewReadFileAction(logger).WithParameters(task_engine.StaticParameter{Value: testFile}, &outputBuffer)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to read file")

	os.Chmod(testFile, 0o600)
}

func (suite *ReadFileTestSuite) TestReadFileAction_GetOutput() {
	content := []byte("file content")
	action := &file.ReadFileAction{
		FilePath:     "/tmp/testfile.txt",
		OutputBuffer: &content,
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("/tmp/testfile.txt", m["filePath"])
	suite.Equal([]byte("file content"), m["content"])
	suite.Equal(12, m["fileSize"])
	suite.Equal(true, m["success"])
}

func TestReadFileTestSuite(t *testing.T) {
	suite.Run(t, new(ReadFileTestSuite))
}
