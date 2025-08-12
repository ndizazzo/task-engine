package file_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
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
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		engine.StaticParameter{Value: expectedContent},
		false,
		nil,
	)
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
	// Use an empty buffer to satisfy validation - this will create an empty file
	var buffer bytes.Buffer
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		nil,
		false,
		&buffer,
	)
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
	err := os.WriteFile(targetFile, initialContent, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create existing file")

	newContent := []byte("New Content")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		engine.StaticParameter{Value: newContent},
		false,
		nil,
	)
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
	err := os.WriteFile(targetFile, initialContent, 0o600)
	suite.Require().NoError(err, "Setup: Failed to create existing file")

	newContent := []byte("New Content - Overwritten")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		engine.StaticParameter{Value: newContent},
		true,
		nil,
	) // overwrite = true
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())
	suite.NoError(execErr)

	actualContent, readErr := os.ReadFile(targetFile)
	suite.NoError(readErr)
	suite.Equal(newContent, actualContent, "File content should have been overwritten")
}

func (suite *WriteFileTestSuite) TestExecuteFailureNoPermissions() {
	readOnlyDir := filepath.Join(suite.tempDir, "read_only")
	err := os.Mkdir(readOnlyDir, 0o555)
	suite.Require().NoError(err)

	targetFile := filepath.Join(readOnlyDir, "cant_write_here")
	content := []byte("some content")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		engine.StaticParameter{Value: content},
		false,
		nil,
	) // overwrite = false
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

	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		nil,
		true,
		&buffer,
	)
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
	action, err := file.NewWriteFileAction(nil).WithParameters(
		engine.StaticParameter{Value: targetFile},
		engine.StaticParameter{Value: content},
		true,
		nil,
	)
	suite.NoError(err)
	suite.NotNil(action)
	suite.NotNil(action.Wrapped.Logger)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionEmptyFilePath() {
	logger := command_mock.NewDiscardLogger()
	content := []byte("test content")
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: ""},
		engine.StaticParameter{Value: content},
		true,
		nil,
	)
	suite.NoError(err)
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
	suite.Nil(nil)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionNoContentNoBuffer() {
	targetFile := filepath.Join(suite.tempDir, "test.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		nil,
		true,
		nil,
	)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionValidParameters() {
	targetFile := filepath.Join(suite.tempDir, "test.txt")
	content := []byte("test content")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		engine.StaticParameter{Value: content},
		true,
		nil,
	)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Equal("write-file-action", action.ID)
	suite.Equal(engine.StaticParameter{Value: content}, action.Wrapped.Content)
	suite.True(action.Wrapped.Overwrite)
	suite.Nil(action.Wrapped.InputBuffer)
}

func (suite *WriteFileTestSuite) TestNewWriteFileActionWithBuffer() {
	targetFile := filepath.Join(suite.tempDir, "test.txt")
	var buffer bytes.Buffer
	buffer.WriteString("buffer content")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		nil,
		false,
		&buffer,
	)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Equal("write-file-action", action.ID)
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
	err := os.Mkdir(readOnlyDir, 0o555)
	suite.Require().NoError(err, "Setup: Failed to create read-only directory")

	// Try to write to a file in a subdirectory that can't be created
	targetFile := filepath.Join(readOnlyDir, "subdir", "test.txt")
	content := []byte("test content")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewWriteFileAction(logger).WithParameters(
		engine.StaticParameter{Value: targetFile},
		engine.StaticParameter{Value: content},
		false,
		nil,
	)
	suite.Require().NoError(err)

	// Execute the action - should fail because we can't create the parent directory
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
	suite.ErrorContains(execErr, "failed to create directory")
}

func (suite *WriteFileTestSuite) TestWriteFileAction_GetOutput() {
	action := &file.WriteFileAction{
		FilePath:  "/tmp/testfile.txt",
		Overwrite: true,
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("/tmp/testfile.txt", m["filePath"])
	suite.Equal(true, m["overwrite"])
	suite.Equal(true, m["success"]) // No writeError, so success is true
}

func TestWriteFileTestSuite(t *testing.T) {
	suite.Run(t, new(WriteFileTestSuite))
}
