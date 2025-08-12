package file_test

import (
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

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
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	originalContent := "This is the original content that was compressed."

	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	_, err = gzipWriter.Write([]byte(originalContent))
	suite.Require().NoError(err, "Setup: Failed to write compressed content")
	gzipWriter.Close()
	compressedFile.Close()

	destFile := filepath.Join(suite.tempDir, "decompressed.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	decompressedContent, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal(originalContent, string(decompressedContent))
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessGzipAutoDetect() {
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	originalContent := "This is the original content that was compressed."

	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	_, err = gzipWriter.Write([]byte(originalContent))
	suite.Require().NoError(err, "Setup: Failed to write compressed content")
	gzipWriter.Close()
	compressedFile.Close()

	destFile := filepath.Join(suite.tempDir, "decompressed.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		"",
	)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	decompressedContent, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal(originalContent, string(decompressedContent))
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessGzipLargeFile() {
	sourceFile := filepath.Join(suite.tempDir, "large_compressed.gz")
	originalContent := "This is repeated content for a large file. "
	for i := 0; i < 1000; i++ {
		originalContent += "This is repeated content for a large file. "
	}

	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create large compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	_, err = gzipWriter.Write([]byte(originalContent))
	suite.Require().NoError(err, "Setup: Failed to write large compressed content")
	gzipWriter.Close()
	compressedFile.Close()

	destFile := filepath.Join(suite.tempDir, "large_decompressed.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	decompressedContent, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal(originalContent, string(decompressedContent))
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessGzipEmptyFile() {
	sourceFile := filepath.Join(suite.tempDir, "empty_compressed.gz")

	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create empty compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	gzipWriter.Close()
	compressedFile.Close()

	destFile := filepath.Join(suite.tempDir, "empty_decompressed.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	decompressedContent, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal("", string(decompressedContent))
}

func (suite *DecompressFileTestSuite) TestExecuteFailureSourceNotExists() {
	nonExistentFile := filepath.Join(suite.tempDir, "nonexistent.gz")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: nonExistentFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "does not exist")
}

func (suite *DecompressFileTestSuite) TestExecuteFailureSourceIsDirectory() {
	// Create a directory
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	err := os.Mkdir(sourceDir, 0o755)
	suite.Require().NoError(err, "Setup: Failed to create source directory")

	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "is a directory, not a file")
}

func (suite *DecompressFileTestSuite) TestExecuteFailureInvalidGzipFile() {
	// Create a file that's not actually gzip compressed
	sourceFile := filepath.Join(suite.tempDir, "invalid.gz")
	err := os.WriteFile(sourceFile, []byte("This is not gzip compressed content"), 0o600)
	suite.Require().NoError(err, "Setup: Failed to create invalid gzip file")

	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to create gzip reader")
}

func (suite *DecompressFileTestSuite) TestExecuteFailureNoWritePermission() {
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	gzipWriter.Write([]byte("test content"))
	gzipWriter.Close()
	compressedFile.Close()

	// Create a read-only directory
	readOnlyDir := filepath.Join(suite.tempDir, "read_only")
	err = os.Mkdir(readOnlyDir, 0o555)
	suite.Require().NoError(err, "Setup: Failed to create read-only directory")

	destFile := filepath.Join(readOnlyDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to create destination file")
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionNilLogger() {
	sourceFile := filepath.Join(suite.tempDir, "source.gz")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	action, err := file.NewDecompressFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.NoError(err)
	suite.NotNil(action)
	suite.NotNil(action.Wrapped.Logger)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionEmptySourcePath() {
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.NoError(err)
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionEmptyDestinationPath() {
	sourceFile := filepath.Join(suite.tempDir, "source.gz")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: ""},
		file.GzipCompression,
	)
	suite.NoError(err)
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionInvalidCompressionType() {
	sourceFile := filepath.Join(suite.tempDir, "source.gz")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		"invalid",
	)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionValidParameters() {
	sourceFile := filepath.Join(suite.tempDir, "source.gz")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Equal("decompress-file-action", action.ID)
	suite.NotNil(action.Wrapped.SourcePathParam)
	suite.NotNil(action.Wrapped.DestinationPathParam)
	suite.Equal(file.GzipCompression, action.Wrapped.CompressionType)
	suite.Equal(logger, action.Wrapped.Logger)
}

func (suite *DecompressFileTestSuite) TestNewDecompressFileActionAutoDetectFailure() {
	// Create a file with unknown extension
	sourceFile := filepath.Join(suite.tempDir, "unknown.xyz")
	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		"",
	)
	suite.NoError(err)
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
}

func (suite *DecompressFileTestSuite) TestExecuteSuccessCreatesDestinationDirectory() {
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
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)
	destDir := filepath.Dir(destFile)
	_, err = os.Stat(destDir)
	suite.NoError(err, "Destination directory should have been created")
	_, err = os.Stat(destFile)
	suite.NoError(err, "Decompressed file should have been created")
}

func (suite *DecompressFileTestSuite) TestDetectCompressionType() {
	testCases := []struct {
		filePath string
		expected file.CompressionType
	}{
		{"file.gz", file.GzipCompression},
		{"file.gzip", file.GzipCompression},
		{"file.xyz", ""}, // Unknown extension
		{"file", ""},     // No extension
	}

	for _, tc := range testCases {
		detected := file.DetectCompressionType(tc.filePath)
		suite.Equal(tc.expected, detected, "Failed for file: %s", tc.filePath)
	}
}

func (suite *DecompressFileTestSuite) TestExecuteFailureStatErrorNotIsNotExist() {
	sourceFile := filepath.Join(suite.tempDir, "compressed.gz")
	compressedFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create compressed file")

	gzipWriter := gzip.NewWriter(compressedFile)
	gzipWriter.Write([]byte("test content"))
	gzipWriter.Close()
	compressedFile.Close()

	// Remove read permissions to cause a stat error that's not IsNotExist
	err = os.Chmod(sourceFile, 0o000)
	suite.Require().NoError(err, "Setup: Failed to change file permissions")

	destFile := filepath.Join(suite.tempDir, "output.txt")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewDecompressFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destFile},
		file.GzipCompression,
	)
	suite.Require().NoError(err)

	// Execute the action
	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to open source file")
}

func (suite *DecompressFileTestSuite) TestExecuteFailureUnsupportedCompressionType() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0o600)
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

func (suite *DecompressFileTestSuite) TestDecompressFileAction_GetOutput() {
	action := &file.DecompressFileAction{
		SourcePath:      "/tmp/compressed.tar.gz",
		DestinationPath: "/tmp/extracted",
		CompressionType: file.GzipCompression,
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("/tmp/compressed.tar.gz", m["source"])
	suite.Equal("/tmp/extracted", m["destination"])
	suite.Equal(string(file.GzipCompression), m["compressionType"])
	suite.Equal(true, m["success"])
}

func TestDecompressFileTestSuite(t *testing.T) {
	suite.Run(t, new(DecompressFileTestSuite))
}
