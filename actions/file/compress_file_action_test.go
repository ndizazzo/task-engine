package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/suite"
)

// CompressFileTestSuite defines the test suite for CompressFileAction
type CompressFileTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *CompressFileTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "compress_file_test_*")
	suite.Require().NoError(err)
}

func (suite *CompressFileTestSuite) TearDownTest() {
	_ = os.RemoveAll(suite.tempDir)
}

func (suite *CompressFileTestSuite) TestExecuteSuccessGzip() {
	// Create a test file with compressible content
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	content := "This is a test file with repeated content. " +
		"This is a test file with repeated content. " +
		"This is a test file with repeated content. " +
		"This is a test file with repeated content. " +
		"This is a test file with repeated content."
	err := os.WriteFile(sourceFile, []byte(content), 0600)
	suite.Require().NoError(err, "Setup: Failed to create source file")

	destFile := filepath.Join(suite.tempDir, "compressed.gz")
	logger := command_mock.NewDiscardLogger()
	action := file.NewCompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the compressed file was created and is smaller
	sourceInfo, err := os.Stat(sourceFile)
	suite.NoError(err)
	destInfo, err := os.Stat(destFile)
	suite.NoError(err)

	// Compressed file should be smaller than original
	suite.Less(destInfo.Size(), sourceInfo.Size(), "Compressed file should be smaller than original")
	suite.Greater(destInfo.Size(), int64(0), "Compressed file should not be empty")
}

func (suite *CompressFileTestSuite) TestExecuteSuccessGzipLargeFile() {
	// Create a large test file with compressible content
	sourceFile := filepath.Join(suite.tempDir, "large_source.txt")

	// Create content with lots of repetition for good compression
	baseContent := "This is repeated content that should compress well. "
	content := ""
	for i := 0; i < 1000; i++ {
		content += baseContent
	}

	err := os.WriteFile(sourceFile, []byte(content), 0600)
	suite.Require().NoError(err, "Setup: Failed to create large source file")

	destFile := filepath.Join(suite.tempDir, "large_compressed.gz")
	logger := command_mock.NewDiscardLogger()
	action := file.NewCompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the compressed file was created and is significantly smaller
	sourceInfo, err := os.Stat(sourceFile)
	suite.NoError(err)
	destInfo, err := os.Stat(destFile)
	suite.NoError(err)

	// Large file should compress significantly
	compressionRatio := float64(destInfo.Size()) / float64(sourceInfo.Size())
	suite.Less(compressionRatio, 0.5, "Large file should compress to less than 50% of original size")
}

func (suite *CompressFileTestSuite) TestExecuteSuccessGzipEmptyFile() {
	// Create an empty test file
	sourceFile := filepath.Join(suite.tempDir, "empty.txt")
	err := os.WriteFile(sourceFile, []byte{}, 0600)
	suite.Require().NoError(err, "Setup: Failed to create empty source file")

	destFile := filepath.Join(suite.tempDir, "empty_compressed.gz")
	logger := command_mock.NewDiscardLogger()
	action := file.NewCompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the compressed file was created
	destInfo, err := os.Stat(destFile)
	suite.NoError(err)
	suite.Greater(destInfo.Size(), int64(0), "Even empty files should have some gzip overhead")
}

func (suite *CompressFileTestSuite) TestExecuteFailureSourceNotExists() {
	nonExistentFile := filepath.Join(suite.tempDir, "nonexistent.txt")
	destFile := filepath.Join(suite.tempDir, "output.gz")
	logger := command_mock.NewDiscardLogger()
	action := file.NewCompressFileAction(nonExistentFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err := action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "does not exist")
}

func (suite *CompressFileTestSuite) TestExecuteFailureSourceIsDirectory() {
	// Create a directory
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	err := os.Mkdir(sourceDir, 0755)
	suite.Require().NoError(err, "Setup: Failed to create source directory")

	destFile := filepath.Join(suite.tempDir, "output.gz")
	logger := command_mock.NewDiscardLogger()
	action := file.NewCompressFileAction(sourceDir, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "is a directory, not a file")
}

func (suite *CompressFileTestSuite) TestExecuteFailureNoWritePermission() {
	// Create a test file
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create source file")

	// Create a read-only directory
	readOnlyDir := filepath.Join(suite.tempDir, "read_only")
	err = os.Mkdir(readOnlyDir, 0555)
	suite.Require().NoError(err, "Setup: Failed to create read-only directory")

	destFile := filepath.Join(readOnlyDir, "output.gz")
	logger := command_mock.NewDiscardLogger()
	action := file.NewCompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to create destination file")
}

func (suite *CompressFileTestSuite) TestNewCompressFileActionNilLogger() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destFile := filepath.Join(suite.tempDir, "output.gz")

	// Should not panic and should create a default logger
	action := file.NewCompressFileAction(sourceFile, destFile, file.GzipCompression, nil)
	suite.NotNil(action)
	suite.NotNil(action.Wrapped.Logger)
}

func (suite *CompressFileTestSuite) TestNewCompressFileActionEmptySourcePath() {
	destFile := filepath.Join(suite.tempDir, "output.gz")
	logger := command_mock.NewDiscardLogger()

	// Should return nil for empty source path
	action := file.NewCompressFileAction("", destFile, file.GzipCompression, logger)
	suite.Nil(action)
}

func (suite *CompressFileTestSuite) TestNewCompressFileActionEmptyDestinationPath() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	logger := command_mock.NewDiscardLogger()

	// Should return nil for empty destination path
	action := file.NewCompressFileAction(sourceFile, "", file.GzipCompression, logger)
	suite.Nil(action)
}

func (suite *CompressFileTestSuite) TestNewCompressFileActionEmptyCompressionType() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destFile := filepath.Join(suite.tempDir, "output.gz")
	logger := command_mock.NewDiscardLogger()

	// Should return nil for empty compression type
	action := file.NewCompressFileAction(sourceFile, destFile, "", logger)
	suite.Nil(action)
}

func (suite *CompressFileTestSuite) TestNewCompressFileActionInvalidCompressionType() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destFile := filepath.Join(suite.tempDir, "output.gz")
	logger := command_mock.NewDiscardLogger()

	// Should return nil for invalid compression type
	action := file.NewCompressFileAction(sourceFile, destFile, "invalid", logger)
	suite.Nil(action)
}

func (suite *CompressFileTestSuite) TestNewCompressFileActionValidParameters() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destFile := filepath.Join(suite.tempDir, "output.gz")
	logger := command_mock.NewDiscardLogger()

	// Should return valid action for valid parameters
	action := file.NewCompressFileAction(sourceFile, destFile, file.GzipCompression, logger)
	suite.NotNil(action)
	suite.Equal("compress-file-gzip-source.txt", action.ID)
	suite.Equal(sourceFile, action.Wrapped.SourcePath)
	suite.Equal(destFile, action.Wrapped.DestinationPath)
	suite.Equal(file.GzipCompression, action.Wrapped.CompressionType)
	suite.Equal(logger, action.Wrapped.Logger)
}

func (suite *CompressFileTestSuite) TestExecuteSuccessCreatesDestinationDirectory() {
	// Create a test file
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create source file")

	// Try to compress to a path with non-existent directory
	destFile := filepath.Join(suite.tempDir, "new_dir", "subdir", "output.gz")
	logger := command_mock.NewDiscardLogger()
	action := file.NewCompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the destination directory was created
	destDir := filepath.Dir(destFile)
	_, err = os.Stat(destDir)
	suite.NoError(err, "Destination directory should have been created")

	// Verify the compressed file was created
	_, err = os.Stat(destFile)
	suite.NoError(err, "Compressed file should have been created")
}

func (suite *CompressFileTestSuite) TestExecuteFailureStatErrorNotIsNotExist() {
	// Create a test file
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create source file")

	// Remove read permissions to cause a stat error that's not IsNotExist
	err = os.Chmod(sourceFile, 0000)
	suite.Require().NoError(err, "Setup: Failed to change file permissions")

	destFile := filepath.Join(suite.tempDir, "output.gz")
	logger := command_mock.NewDiscardLogger()
	action := file.NewCompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to open source file")
}

func (suite *CompressFileTestSuite) TestExecuteFailureUnsupportedCompressionType() {
	// Create a test file
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create source file")

	destFile := filepath.Join(suite.tempDir, "output.unknown")
	logger := command_mock.NewDiscardLogger()

	// Create action with invalid compression type (this would normally be prevented by constructor)
	// But we can test the Execute method directly
	action := &file.CompressFileAction{
		BaseAction:      task_engine.BaseAction{Logger: logger},
		SourcePath:      sourceFile,
		DestinationPath: destFile,
		CompressionType: "invalid",
	}

	// Execute the action
	err = action.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "unsupported compression type")
}

func TestCompressFileTestSuite(t *testing.T) {
	suite.Run(t, new(CompressFileTestSuite))
}
