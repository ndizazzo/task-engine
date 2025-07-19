package file_test

import (
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/suite"
)

// DecompressFileTestSuite defines the test suite for DecompressFileAction
type DecompressFileTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *DecompressFileTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "decompress_file_test_*")
	suite.Require().NoError(err)
}

func (suite *DecompressFileTestSuite) TearDownTest() {
	_ = os.RemoveAll(suite.tempDir)
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessGzip() {
	// Create a compressed file first
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	originalContent := "This is the original content that was compressed."

	// Create gzip compressed file
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	_, err = gzipWriter.Write([]byte(originalContent))
	suite.Require().NoError(err, "Setup: Failed to write compressed content")
	gzipWriter.Close()
	compressedFile.Close()

	destFile := filepath.Join(suite.tempDir, "decompressed.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the decompressed file was created and contains the original content
	decompressedContent, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal(originalContent, string(decompressedContent))
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessGzipAutoDetect() {
	// Create a compressed file with .gz extension
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	originalContent := "This is the original content that was compressed."

	// Create gzip compressed file
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	_, err = gzipWriter.Write([]byte(originalContent))
	suite.Require().NoError(err, "Setup: Failed to write compressed content")
	gzipWriter.Close()
	compressedFile.Close()

	destFile := filepath.Join(suite.tempDir, "decompressed.txt")
	logger := command_mock.NewDiscardLogger()
	// Don't specify compression type - should auto-detect from .gz extension
	action := file.NewDecompressFileAction(sourceFile, destFile, "", logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the decompressed file was created and contains the original content
	decompressedContent, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal(originalContent, string(decompressedContent))
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessGzipLargeFile() {
	// Create a large compressed file
	sourceFile := filepath.Join(suite.tempDir, "large_compressed.gz")
	originalContent := "This is repeated content for a large file. "
	for i := 0; i < 1000; i++ {
		originalContent += "This is repeated content for a large file. "
	}

	// Create gzip compressed file
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create large compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	_, err = gzipWriter.Write([]byte(originalContent))
	suite.Require().NoError(err, "Setup: Failed to write large compressed content")
	gzipWriter.Close()
	compressedFile.Close()

	destFile := filepath.Join(suite.tempDir, "large_decompressed.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the decompressed file was created and contains the original content
	decompressedContent, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal(originalContent, string(decompressedContent))
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessGzipEmptyFile() {
	// Create an empty compressed file
	sourceFile := filepath.Join(suite.tempDir, "empty_compressed.gz")

	// Create gzip compressed file with empty content
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create empty compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	// Write nothing - empty content
	gzipWriter.Close()
	compressedFile.Close()

	destFile := filepath.Join(suite.tempDir, "empty_decompressed.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the decompressed file was created and is empty
	decompressedContent, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal("", string(decompressedContent))
}

func (suite *DecompressFileTestSuite) TestExecuteFailureSourceNotExists() {
	nonExistentFile := filepath.Join(suite.tempDir, "nonexistent.gz")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(nonExistentFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err := action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "does not exist")
}

func (suite *DecompressFileTestSuite) TestExecuteFailureSourceIsDirectory() {
	// Create a directory
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	err := os.Mkdir(sourceDir, 0755)
	suite.Require().NoError(err, "Setup: Failed to create source directory")

	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(sourceDir, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "is a directory, not a file")
}

func (suite *DecompressFileTestSuite) TestExecuteFailureInvalidGzipFile() {
	// Create a file that's not actually gzip compressed
	sourceFile := filepath.Join(suite.tempDir, "invalid.gz")
	err := os.WriteFile(sourceFile, []byte("This is not gzip compressed content"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create invalid gzip file")

	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to create gzip reader")
}

func (suite *DecompressFileTestSuite) TestExecuteFailureNoWritePermission() {
	// Create a compressed file
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	gzipWriter.Write([]byte("test content"))
	gzipWriter.Close()
	compressedFile.Close()

	// Create a read-only directory
	readOnlyDir := filepath.Join(suite.tempDir, "read_only")
	err = os.Mkdir(readOnlyDir, 0555)
	suite.Require().NoError(err, "Setup: Failed to create read-only directory")

	destFile := filepath.Join(readOnlyDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to create destination file")
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionNilLogger() {
	sourceFile := filepath.Join(suite.tempDir, "source.gz")
	destFile := filepath.Join(suite.tempDir, "output.txt")

	// Should not panic and should create a default logger
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, nil)
	suite.NotNil(action)
	suite.NotNil(action.Wrapped.Logger)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionEmptySourcePath() {
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()

	// Should return nil for empty source path
	action := file.NewDecompressFileAction("", destFile, file.GzipCompression, logger)
	suite.Nil(action)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionEmptyDestinationPath() {
	sourceFile := filepath.Join(suite.tempDir, "source.gz")
	logger := command_mock.NewDiscardLogger()

	// Should return nil for empty destination path
	action := file.NewDecompressFileAction(sourceFile, "", file.GzipCompression, logger)
	suite.Nil(action)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionInvalidCompressionType() {
	sourceFile := filepath.Join(suite.tempDir, "source.gz")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()

	// Should return nil for invalid compression type
	action := file.NewDecompressFileAction(sourceFile, destFile, "invalid", logger)
	suite.Nil(action)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionValidParameters() {
	sourceFile := filepath.Join(suite.tempDir, "source.gz")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()

	// Should return valid action for valid parameters
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, logger)
	suite.NotNil(action)
	suite.Equal("decompress-file-gzip-source.gz", action.ID)
	suite.Equal(sourceFile, action.Wrapped.SourcePath)
	suite.Equal(destFile, action.Wrapped.DestinationPath)
	suite.Equal(file.GzipCompression, action.Wrapped.CompressionType)
	suite.Equal(logger, action.Wrapped.Logger)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionAutoDetectFailure() {
	// Create a file with unknown extension
	sourceFile := filepath.Join(suite.tempDir, "source.unknown")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()

	// Should return nil when auto-detection fails
	action := file.NewDecompressFileAction(sourceFile, destFile, "", logger)
	suite.Nil(action)
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessCreatesDestinationDirectory() {
	// Create a compressed file
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	gzipWriter.Write([]byte("test content"))
	gzipWriter.Close()
	compressedFile.Close()

	// Try to decompress to a path with non-existent directory
	destFile := filepath.Join(suite.tempDir, "new_dir", "subdir", "output.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	// Verify the destination directory was created
	destDir := filepath.Dir(destFile)
	_, err = os.Stat(destDir)
	suite.NoError(err, "Destination directory should have been created")

	// Verify the decompressed file was created
	_, err = os.Stat(destFile)
	suite.NoError(err, "Decompressed file should have been created")
}

func (suite *DecompressFileTestSuite) TestDetectCompressionType() {
	// Test gzip extensions
	suite.Equal(file.GzipCompression, file.DetectCompressionType("file.gz"))
	suite.Equal(file.GzipCompression, file.DetectCompressionType("file.gzip"))
	suite.Equal(file.GzipCompression, file.DetectCompressionType("file.GZ"))
	suite.Equal(file.GzipCompression, file.DetectCompressionType("file.GZIP"))

	// Test unknown extensions
	suite.Equal(file.CompressionType(""), file.DetectCompressionType("file.txt"))
	suite.Equal(file.CompressionType(""), file.DetectCompressionType("file.unknown"))
	suite.Equal(file.CompressionType(""), file.DetectCompressionType("file"))
	suite.Equal(file.CompressionType(""), file.DetectCompressionType(""))
}

func (suite *DecompressFileTestSuite) TestExecuteFailureStatErrorNotIsNotExist() {
	// Create a compressed file
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	gzipWriter.Write([]byte("test content"))
	gzipWriter.Close()
	compressedFile.Close()

	// Remove read permissions to cause a stat error that's not IsNotExist
	err = os.Chmod(sourceFile, 0000)
	suite.Require().NoError(err, "Setup: Failed to change file permissions")

	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action := file.NewDecompressFileAction(sourceFile, destFile, file.GzipCompression, logger)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to open source file")
}

func (suite *DecompressFileTestSuite) TestExecuteFailureUnsupportedCompressionType() {
	// Create a test file
	sourceFile := filepath.Join(suite.tempDir, "source.unknown")
	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create source file")

	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()

	// Create action with invalid compression type (this would normally be prevented by constructor)
	// But we can test the Execute method directly
	action := &file.DecompressFileAction{
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

func TestDecompressFileTestSuite(t *testing.T) {
	suite.Run(t, new(DecompressFileTestSuite))
}
