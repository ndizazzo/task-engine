package file_test

import (
	"archive/tar"
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/suite"
)

type ExtractFileTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *ExtractFileTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "extract_file_test_*")
	suite.Require().NoError(err)
}

func (suite *ExtractFileTestSuite) TearDownTest() {
	_ = os.RemoveAll(suite.tempDir)
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessTar() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	destDir := filepath.Join(suite.tempDir, "extracted")

	tarFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create tar file")

	tarWriter := tar.NewWriter(tarFile)

	content := "This is test content for tar extraction"
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(header)
	suite.Require().NoError(err, "Setup: Failed to write tar header")

	_, err = tarWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write tar content")

	tarWriter.Close()
	tarFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.TarArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	extractedFile := filepath.Join(destDir, "test.txt")
	extractedContent, err := os.ReadFile(extractedFile)
	suite.NoError(err)
	suite.Equal(content, string(extractedContent))
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessTarGz() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar.gz")
	destDir := filepath.Join(suite.tempDir, "extracted")

	tarFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create tar.gz file")

	tarWriter := tar.NewWriter(tarFile)

	content := "This is test content for tar.gz extraction"
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(header)
	suite.Require().NoError(err, "Setup: Failed to write tar header")

	_, err = tarWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write tar content")

	tarWriter.Close()
	tarFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.TarGzArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	extractedFile := filepath.Join(destDir, "test.txt")
	extractedContent, err := os.ReadFile(extractedFile)
	suite.NoError(err)
	suite.Equal(content, string(extractedContent))
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessZip() {
	sourceFile := filepath.Join(suite.tempDir, "test.zip")
	destDir := filepath.Join(suite.tempDir, "extracted")

	zipFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create zip file")

	zipWriter := zip.NewWriter(zipFile)

	content := "This is test content for zip extraction"
	fileWriter, err := zipWriter.Create("test.txt")
	suite.Require().NoError(err, "Setup: Failed to create zip file entry")

	_, err = fileWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write zip content")

	zipWriter.Close()
	zipFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.ZipArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	extractedFile := filepath.Join(destDir, "test.txt")
	extractedContent, err := os.ReadFile(extractedFile)
	suite.NoError(err)
	suite.Equal(content, string(extractedContent))
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessTarWithDirectories() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	destDir := filepath.Join(suite.tempDir, "extracted")

	tarFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create tar file")

	tarWriter := tar.NewWriter(tarFile)

	dirHeader := &tar.Header{
		Name:     "testdir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	err = tarWriter.WriteHeader(dirHeader)
	suite.Require().NoError(err, "Setup: Failed to write tar directory header")

	content := "This is test content in a subdirectory"
	fileHeader := &tar.Header{
		Name: "testdir/test.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(fileHeader)
	suite.Require().NoError(err, "Setup: Failed to write tar file header")

	_, err = tarWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write tar content")

	tarWriter.Close()
	tarFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.TarArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	extractedFile := filepath.Join(destDir, "testdir", "test.txt")
	extractedContent, err := os.ReadFile(extractedFile)
	suite.NoError(err)
	suite.Equal(content, string(extractedContent))

	extractedDir := filepath.Join(destDir, "testdir")
	dirInfo, err := os.Stat(extractedDir)
	suite.NoError(err)
	suite.True(dirInfo.IsDir())
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessZipWithDirectories() {
	sourceFile := filepath.Join(suite.tempDir, "test.zip")
	destDir := filepath.Join(suite.tempDir, "extracted")

	zipFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create zip file")

	zipWriter := zip.NewWriter(zipFile)

	content := "This is test content in a zip subdirectory"
	fileWriter, err := zipWriter.Create("testdir/test.txt")
	suite.Require().NoError(err, "Setup: Failed to create zip file entry")

	_, err = fileWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write zip content")

	zipWriter.Close()
	zipFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.ZipArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	extractedFile := filepath.Join(destDir, "testdir", "test.txt")
	extractedContent, err := os.ReadFile(extractedFile)
	suite.NoError(err)
	suite.Equal(content, string(extractedContent))

	extractedDir := filepath.Join(destDir, "testdir")
	dirInfo, err := os.Stat(extractedDir)
	suite.NoError(err)
	suite.True(dirInfo.IsDir())
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessAutoDetectTar() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	destDir := filepath.Join(suite.tempDir, "extracted")

	tarFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create tar file")

	tarWriter := tar.NewWriter(tarFile)
	content := "Auto-detected tar content"
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(header)
	suite.Require().NoError(err, "Setup: Failed to write tar header")

	_, err = tarWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write tar content")

	tarWriter.Close()
	tarFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, "", logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	extractedFile := filepath.Join(destDir, "test.txt")
	extractedContent, err := os.ReadFile(extractedFile)
	suite.NoError(err)
	suite.Equal(content, string(extractedContent))
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessAutoDetectTarGz() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar.gz")
	destDir := filepath.Join(suite.tempDir, "extracted")

	tarFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create tar.gz file")

	tarWriter := tar.NewWriter(tarFile)
	content := "Auto-detected tar.gz content"
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(header)
	suite.Require().NoError(err, "Setup: Failed to write tar header")

	_, err = tarWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write tar content")

	tarWriter.Close()
	tarFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, "", logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	extractedFile := filepath.Join(destDir, "test.txt")
	extractedContent, err := os.ReadFile(extractedFile)
	suite.NoError(err)
	suite.Equal(content, string(extractedContent))
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessAutoDetectZip() {
	sourceFile := filepath.Join(suite.tempDir, "test.zip")
	destDir := filepath.Join(suite.tempDir, "extracted")

	zipFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create zip file")

	zipWriter := zip.NewWriter(zipFile)
	content := "Auto-detected zip content"
	fileWriter, err := zipWriter.Create("test.txt")
	suite.Require().NoError(err, "Setup: Failed to create zip file entry")

	_, err = fileWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write zip content")

	zipWriter.Close()
	zipFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, "", logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	extractedFile := filepath.Join(destDir, "test.txt")
	extractedContent, err := os.ReadFile(extractedFile)
	suite.NoError(err)
	suite.Equal(content, string(extractedContent))
}

func (suite *ExtractFileTestSuite) TestExecuteFailureSourceNotExists() {
	nonExistentFile := filepath.Join(suite.tempDir, "nonexistent.tar")
	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(nonExistentFile, destDir, file.TarArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "does not exist")
}

func (suite *ExtractFileTestSuite) TestExecuteFailureSourceIsDirectory() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	err := os.Mkdir(sourceDir, 0755)
	suite.Require().NoError(err, "Setup: Failed to create source directory")

	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceDir, destDir, file.TarArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "is a directory, not a file")
}

func (suite *ExtractFileTestSuite) TestExecuteFailureInvalidTarFile() {
	sourceFile := filepath.Join(suite.tempDir, "invalid.tar")
	err := os.WriteFile(sourceFile, []byte("This is not a tar archive"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create invalid tar file")

	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.TarArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to read tar header")
}

func (suite *ExtractFileTestSuite) TestExecuteFailureInvalidZipFile() {
	sourceFile := filepath.Join(suite.tempDir, "invalid.zip")
	err := os.WriteFile(sourceFile, []byte("This is not a zip archive"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create invalid zip file")

	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.ZipArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to create zip reader")
}

func (suite *ExtractFileTestSuite) TestExecuteFailureZipSlipVulnerability() {
	sourceFile := filepath.Join(suite.tempDir, "malicious.zip")
	destDir := filepath.Join(suite.tempDir, "extracted")

	zipFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create malicious zip file")

	zipWriter := zip.NewWriter(zipFile)

	content := "Malicious content"
	fileWriter, err := zipWriter.Create("../../../malicious.txt")
	suite.Require().NoError(err, "Setup: Failed to create malicious zip file entry")

	_, err = fileWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write malicious zip content")

	zipWriter.Close()
	zipFile.Close()

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.ZipArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "illegal file path")
}

func (suite *ExtractFileTestSuite) TestExecuteFailureNoWritePermission() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	tarFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create tar file")

	tarWriter := tar.NewWriter(tarFile)
	content := "Test content"
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(header)
	suite.Require().NoError(err, "Setup: Failed to write tar header")

	_, err = tarWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write tar content")

	tarWriter.Close()
	tarFile.Close()

	readOnlyDir := filepath.Join(suite.tempDir, "read_only")
	err = os.Mkdir(readOnlyDir, 0555)
	suite.Require().NoError(err, "Setup: Failed to create read-only directory")

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, readOnlyDir, file.TarArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "failed to create file")
}

func (suite *ExtractFileTestSuite) TestNewExtractFileActionNilLogger() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	destDir := filepath.Join(suite.tempDir, "extracted")

	action, err := file.NewExtractFileAction(sourceFile, destDir, file.TarArchive, nil)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Nil(action.Wrapped.Logger)
}

func (suite *ExtractFileTestSuite) TestNewExtractFileActionEmptySourcePath() {
	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()

	action, err := file.NewExtractFileAction("", destDir, file.TarArchive, logger)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *ExtractFileTestSuite) TestNewExtractFileActionEmptyDestinationPath() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	logger := command_mock.NewDiscardLogger()

	action, err := file.NewExtractFileAction(sourceFile, "", file.TarArchive, logger)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *ExtractFileTestSuite) TestNewExtractFileActionInvalidArchiveType() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()

	action, err := file.NewExtractFileAction(sourceFile, destDir, "invalid", logger)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *ExtractFileTestSuite) TestNewExtractFileActionValidParameters() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()

	action, err := file.NewExtractFileAction(sourceFile, destDir, file.TarArchive, logger)
	suite.NoError(err)
	suite.NotNil(action)
	suite.Equal("extract-file-tar-test.tar", action.ID)
	suite.Equal(sourceFile, action.Wrapped.SourcePath)
	suite.Equal(destDir, action.Wrapped.DestinationPath)
	suite.Equal(file.TarArchive, action.Wrapped.ArchiveType)
	suite.Equal(logger, action.Wrapped.Logger)
}

func (suite *ExtractFileTestSuite) TestNewExtractFileActionAutoDetectFailure() {
	sourceFile := filepath.Join(suite.tempDir, "unknown.xyz")
	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()

	action, err := file.NewExtractFileAction(sourceFile, destDir, "", logger)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *ExtractFileTestSuite) TestDetectArchiveType() {
	testCases := []struct {
		filePath string
		expected file.ArchiveType
	}{
		{"file.tar", file.TarArchive},
		{"file.tar.gz", file.TarGzArchive},
		{"file.zip", file.ZipArchive},
		{"file.xyz", ""}, // Unknown extension
		{"file", ""},     // No extension
	}

	for _, tc := range testCases {
		detected := file.DetectArchiveType(tc.filePath)
		suite.Equal(tc.expected, detected, "Failed for file: %s", tc.filePath)
	}
}

func (suite *ExtractFileTestSuite) TestExecuteFailureUnsupportedArchiveType() {
	sourceFile := filepath.Join(suite.tempDir, "test.txt")
	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.Require().NoError(err, "Setup: Failed to create source file")

	destDir := filepath.Join(suite.tempDir, "extracted")
	logger := command_mock.NewDiscardLogger()

	action := &file.ExtractFileAction{
		BaseAction:      task_engine.BaseAction{Logger: logger},
		SourcePath:      sourceFile,
		DestinationPath: destDir,
		ArchiveType:     "invalid",
	}

	err = action.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "unsupported archive type")
}

func (suite *ExtractFileTestSuite) TestExecuteFailureCompressedTarGz() {
	cwd, err := os.Getwd()
	suite.Require().NoError(err, "Failed to get current working directory")
	projectRoot := filepath.Join(cwd, "..", "..")
	fixturePath := filepath.Join(projectRoot, "testdata", "compressed.tar.gz")
	sourceFile := filepath.Join(suite.tempDir, "compressed.tar.gz")

	data, err := os.ReadFile(fixturePath)
	suite.Require().NoError(err, "Failed to read fixture file")
	err = os.WriteFile(sourceFile, data, 0644)
	suite.Require().NoError(err, "Failed to copy fixture file")

	destDir := filepath.Join(suite.tempDir, "extracted")

	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.TarGzArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.ErrorContains(err, "is compressed with gzip")
	suite.ErrorContains(err, "Please decompress it first using DecompressFileAction")
}

func (suite *ExtractFileTestSuite) TestExecuteSuccessCreatesDestinationDirectory() {
	sourceFile := filepath.Join(suite.tempDir, "test.tar")
	tarFile, err := os.Create(sourceFile)
	suite.Require().NoError(err, "Setup: Failed to create tar file")

	tarWriter := tar.NewWriter(tarFile)
	content := "Test content"
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}
	err = tarWriter.WriteHeader(header)
	suite.Require().NoError(err, "Setup: Failed to write tar header")

	_, err = tarWriter.Write([]byte(content))
	suite.Require().NoError(err, "Setup: Failed to write tar content")

	tarWriter.Close()
	tarFile.Close()

	destDir := filepath.Join(suite.tempDir, "new_dir", "subdir", "extracted")
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewExtractFileAction(sourceFile, destDir, file.TarArchive, logger)
	suite.Require().NoError(err)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(destDir)
	suite.NoError(err, "Destination directory should have been created")

	extractedFile := filepath.Join(destDir, "test.txt")
	_, err = os.Stat(extractedFile)
	suite.NoError(err, "Extracted file should have been created")
}

func TestExtractFileTestSuite(t *testing.T) {
	suite.Run(t, new(ExtractFileTestSuite))
}
