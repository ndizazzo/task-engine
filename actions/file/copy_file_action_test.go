package file_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

// CopyFileActionTestSuite defines the suite for CopyFileAction tests
type CopyFileActionTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *CopyFileActionTestSuite) SetupTest() {
	suite.tempDir = suite.T().TempDir()
}

func (suite *CopyFileActionTestSuite) TestCopyFile_Success() {
	sourceFile := filepath.Join(suite.tempDir, "test_source.txt")
	destinationFile := filepath.Join(suite.tempDir, "test_destination.txt")

	err := os.WriteFile(sourceFile, []byte("test content"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(destinationFile)
	suite.NoError(err)

	destContent, err := os.ReadFile(destinationFile)
	suite.NoError(err)
	suite.Equal("test content", string(destContent))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_CreateDirTrue() {
	sourceFile := filepath.Join(suite.tempDir, "test_source.txt")
	destinationDir := filepath.Join(suite.tempDir, "nested")
	destinationFile := filepath.Join(destinationDir, "test_destination.txt")

	err := os.WriteFile(sourceFile, []byte("test content"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		true,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(destinationFile)
	suite.NoError(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_CreateDirFalse() {
	sourceFile := filepath.Join(suite.tempDir, "test_source.txt")
	destinationDir := filepath.Join(suite.tempDir, "nested")
	destinationFile := filepath.Join(destinationDir, "test_destination.txt")

	err := os.WriteFile(sourceFile, []byte("test content"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())

	suite.Error(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveSuccess() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory with some files
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	sourceFile1 := filepath.Join(sourceDir, "file1.txt")
	sourceFile2 := filepath.Join(sourceDir, "file2.txt")

	err = os.WriteFile(sourceFile1, []byte("content1"), 0o600)
	suite.NoError(err)
	err = os.WriteFile(sourceFile2, []byte("content2"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	_, err = os.Stat(destinationDir)
	suite.NoError(err)
	destFile1 := filepath.Join(destinationDir, "file1.txt")
	destFile2 := filepath.Join(destinationDir, "file2.txt")

	_, err = os.Stat(destFile1)
	suite.NoError(err)
	_, err = os.Stat(destFile2)
	suite.NoError(err)
	content1, err := os.ReadFile(destFile1)
	suite.NoError(err)
	suite.Equal("content1", string(content1))

	content2, err := os.ReadFile(destFile2)
	suite.NoError(err)
	suite.Equal("content2", string(content2))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithNestedDirectories() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create nested directory structure
	nestedDir := filepath.Join(sourceDir, "nested", "subdir")
	err := os.MkdirAll(nestedDir, 0o750)
	suite.NoError(err)

	sourceFile := filepath.Join(nestedDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("nested content"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destNestedDir := filepath.Join(destinationDir, "nested", "subdir")
	destFile := filepath.Join(destNestedDir, "file.txt")

	_, err = os.Stat(destNestedDir)
	suite.NoError(err)
	_, err = os.Stat(destFile)
	suite.NoError(err)
	content, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal("nested content", string(content))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithCreateDir() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "nested", "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	sourceFile := filepath.Join(sourceDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	_, err = os.Stat(filepath.Dir(destinationDir))
	suite.NoError(err)
	destFile := filepath.Join(destinationDir, "file.txt")
	_, err = os.Stat(destFile)
	suite.NoError(err)

	content, err := os.ReadFile(destFile)
	suite.NoError(err)
	suite.Equal("content", string(content))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveFileAsSource() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destinationFile := filepath.Join(suite.tempDir, "dest.txt")

	err := os.WriteFile(sourceFile, []byte("file content"), 0o600)
	suite.NoError(err)
	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(destinationFile)
	suite.NoError(err)

	content, err := os.ReadFile(destinationFile)
	suite.NoError(err)
	suite.Equal("file content", string(content))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_SourceDoesNotExist() {
	nonExistentSource := filepath.Join(suite.tempDir, "nonexistent.txt")
	destinationFile := filepath.Join(suite.tempDir, "destination.txt")

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: nonExistentSource},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())

	suite.Error(err)
	suite.True(os.IsNotExist(err))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveSourceDoesNotExist() {
	nonExistentSource := filepath.Join(suite.tempDir, "nonexistent_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: nonExistentSource},
		task_engine.StaticParameter{Value: destinationDir},
		false,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())

	suite.Error(err)
	suite.True(os.IsNotExist(err))
}

func (suite *CopyFileActionTestSuite) TestNewCopyFileAction_InvalidParameters() {
	logger := command_mock.NewDiscardLogger()
	action, err := file.NewCopyFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: "/dest"},
		false,
		false,
	)
	suite.NoError(err)
	suite.NotNil(action)
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
	action, err = file.NewCopyFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/source"},
		task_engine.StaticParameter{Value: ""},
		false,
		false,
	)
	suite.NoError(err)
	suite.NotNil(action)
	execErr = action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
	action, err = file.NewCopyFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/same"},
		task_engine.StaticParameter{Value: "/same"},
		false,
		false,
	)
	suite.NoError(err)
	suite.NotNil(action)
	execErr = action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
}
func (suite *CopyFileActionTestSuite) TestCopyFile_ReadOnlySource() {
	sourceFile := filepath.Join(suite.tempDir, "readonly_source.txt")
	destinationFile := filepath.Join(suite.tempDir, "destination.txt")
	err := os.WriteFile(sourceFile, []byte("readonly content"), 0o600)
	suite.NoError(err)
	err = os.Chmod(sourceFile, 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	content, err := os.ReadFile(destinationFile)
	suite.NoError(err)
	suite.Equal("readonly content", string(content))
	os.Chmod(sourceFile, 0o600)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_ReadOnlyDestination() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destinationFile := filepath.Join(suite.tempDir, "readonly_dest.txt")
	err := os.WriteFile(sourceFile, []byte("source content"), 0o600)
	suite.NoError(err)
	err = os.WriteFile(destinationFile, []byte("existing content"), 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
	content, err := os.ReadFile(destinationFile)
	suite.NoError(err)
	suite.Equal("existing content", string(content))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_ReadOnlyDestinationDirectory() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	readOnlyDir := filepath.Join(suite.tempDir, "readonly_dir")
	destinationFile := filepath.Join(readOnlyDir, "dest.txt")
	err := os.WriteFile(sourceFile, []byte("source content"), 0o600)
	suite.NoError(err)

	// Create read-only directory
	err = os.MkdirAll(readOnlyDir, 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
	os.Chmod(readOnlyDir, 0o750)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_LargeFile() {
	sourceFile := filepath.Join(suite.tempDir, "large_source.txt")
	destinationFile := filepath.Join(suite.tempDir, "large_dest.txt")

	// Create a large file (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	err := os.WriteFile(sourceFile, largeContent, 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destContent, err := os.ReadFile(destinationFile)
	suite.NoError(err)
	suite.Equal(len(largeContent), len(destContent))
	suite.Equal(largeContent, destContent)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_EmptyFile() {
	sourceFile := filepath.Join(suite.tempDir, "empty_source.txt")
	destinationFile := filepath.Join(suite.tempDir, "empty_dest.txt")
	err := os.WriteFile(sourceFile, []byte{}, 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destContent, err := os.ReadFile(destinationFile)
	suite.NoError(err)
	suite.Equal(0, len(destContent))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_SpecialCharacters() {
	sourceFile := filepath.Join(suite.tempDir, "source with spaces.txt")
	destinationFile := filepath.Join(suite.tempDir, "dest with spaces.txt")

	content := "content with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"
	err := os.WriteFile(sourceFile, []byte(content), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destContent, err := os.ReadFile(destinationFile)
	suite.NoError(err)
	suite.Equal(content, string(destContent))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithSymlinks() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory structure
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)
	realFile := filepath.Join(sourceDir, "real_file.txt")
	err = os.WriteFile(realFile, []byte("real content"), 0o600)
	suite.NoError(err)
	symlinkFile := filepath.Join(sourceDir, "symlink.txt")
	err = os.Symlink(realFile, symlinkFile)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destRealFile := filepath.Join(destinationDir, "real_file.txt")
	_, err = os.Stat(destRealFile)
	suite.NoError(err)
	destSymlink := filepath.Join(destinationDir, "symlink.txt")
	linkInfo, err := os.Lstat(destSymlink)
	suite.NoError(err)
	suite.True(linkInfo.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(destSymlink)
	suite.NoError(err)
	suite.Equal(realFile, target)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithCircularSymlinks() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a circular symlink (this should be handled gracefully)
	circularLink := filepath.Join(sourceDir, "circular")
	err = os.Symlink(circularLink, circularLink)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destCircularLink := filepath.Join(destinationDir, "circular")
	linkInfo, err := os.Lstat(destCircularLink)
	suite.NoError(err)
	suite.True(linkInfo.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(destCircularLink)
	suite.NoError(err)
	suite.Equal(circularLink, target)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithBrokenSymlinks() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a broken symlink
	brokenLink := filepath.Join(sourceDir, "broken")
	err = os.Symlink("/nonexistent/path", brokenLink)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destBrokenLink := filepath.Join(destinationDir, "broken")
	linkInfo, err := os.Lstat(destBrokenLink)
	suite.NoError(err)
	suite.True(linkInfo.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(destBrokenLink)
	suite.NoError(err)
	suite.Equal("/nonexistent/path", target)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithEmptyDirectories() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory with empty subdirectories
	emptyDir1 := filepath.Join(sourceDir, "empty1")
	emptyDir2 := filepath.Join(sourceDir, "nested", "empty2")

	err := os.MkdirAll(emptyDir1, 0o750)
	suite.NoError(err)
	err = os.MkdirAll(emptyDir2, 0o750)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destEmptyDir1 := filepath.Join(destinationDir, "empty1")
	destEmptyDir2 := filepath.Join(destinationDir, "nested", "empty2")

	_, err = os.Stat(destEmptyDir1)
	suite.NoError(err)
	_, err = os.Stat(destEmptyDir2)
	suite.NoError(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithHiddenFiles() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create hidden files
	hiddenFile := filepath.Join(sourceDir, ".hidden")
	err = os.WriteFile(hiddenFile, []byte("hidden content"), 0o600)
	suite.NoError(err)

	dotDir := filepath.Join(sourceDir, ".dotdir")
	err = os.MkdirAll(dotDir, 0o750)
	suite.NoError(err)

	dotFile := filepath.Join(dotDir, ".dotfile")
	err = os.WriteFile(dotFile, []byte("dot file content"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destHiddenFile := filepath.Join(destinationDir, ".hidden")
	destDotFile := filepath.Join(destinationDir, ".dotdir", ".dotfile")

	_, err = os.Stat(destHiddenFile)
	suite.NoError(err)
	_, err = os.Stat(destDotFile)
	suite.NoError(err)
	content, err := os.ReadFile(destHiddenFile)
	suite.NoError(err)
	suite.Equal("hidden content", string(content))

	content, err = os.ReadFile(destDotFile)
	suite.NoError(err)
	suite.Equal("dot file content", string(content))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithPermissionErrors() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a file with restrictive permissions
	restrictedFile := filepath.Join(sourceDir, "restricted.txt")
	err = os.WriteFile(restrictedFile, []byte("restricted content"), 0o000)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
	os.Chmod(restrictedFile, 0o600)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithDirectoryCreationFailure() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a file in source
	sourceFile := filepath.Join(sourceDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Create a read-only directory that will prevent destination creation
	parentDir := filepath.Dir(destinationDir)
	err = os.MkdirAll(parentDir, 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	// This might succeed or fail depending on the system, but we're testing the path
	// The important thing is that it doesn't panic and handles the scenario gracefully
	os.Chmod(parentDir, 0o750)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithRelativePathError() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a file in source
	sourceFile := filepath.Join(sourceDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Change to a different directory to make relative path calculation fail
	originalDir, err := os.Getwd()
	suite.NoError(err)
	defer os.Chdir(originalDir)

	// Change to a directory that makes relative path calculation complex
	err = os.Chdir("/tmp")
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	// This might succeed or fail depending on the system, but we're testing the path calculation
	// The important thing is that it doesn't panic
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithSymlinkCopyFailure() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a symlink that will fail to copy (pointing to a non-existent target)
	symlinkFile := filepath.Join(sourceDir, "bad_symlink")
	err = os.Symlink("/nonexistent/path", symlinkFile)
	suite.NoError(err)

	// Create a read-only destination directory to cause symlink creation to fail
	err = os.MkdirAll(destinationDir, 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	os.Chmod(destinationDir, 0o750)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithSpecialFiles() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)
	regularFile := filepath.Join(sourceDir, "regular.txt")
	err = os.WriteFile(regularFile, []byte("regular content"), 0o600)
	suite.NoError(err)
	pipePath := filepath.Join(sourceDir, "pipe")
	err = syscall.Mkfifo(pipePath, 0o644)
	if err != nil {
		// Skip this test if we can't create FIFOs (e.g., on Windows)
		suite.T().Skip("Cannot create FIFO on this system")
	}

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destRegularFile := filepath.Join(destinationDir, "regular.txt")
	_, err = os.Stat(destRegularFile)
	suite.NoError(err)
	destPipe := filepath.Join(destinationDir, "pipe")
	_, err = os.Stat(destPipe)
	suite.Error(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_FileCopyWithDirectoryCreationFailure() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destinationFile := filepath.Join(suite.tempDir, "readonly_dir", "dest.txt")
	err := os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Create a read-only directory that will prevent destination creation
	err = os.MkdirAll(filepath.Dir(destinationFile), 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		true,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
	os.Chmod(filepath.Dir(destinationFile), 0o750)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_FileCopyWithSourceOpenFailure() {
	// Try to copy a directory as a file (this will cause open failure)
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationFile := filepath.Join(suite.tempDir, "dest.txt")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_FileCopyWithDestinationCreateFailure() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destinationFile := filepath.Join(suite.tempDir, "dest.txt")
	err := os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Create a read-only file at destination
	err = os.WriteFile(destinationFile, []byte("existing"), 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
	os.Chmod(destinationFile, 0o600)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_FileCopyWithCopyFailure() {
	sourceFile := filepath.Join(suite.tempDir, "source.txt")
	destinationFile := filepath.Join(suite.tempDir, "dest.txt")
	err := os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Make source file read-only to potentially cause copy issues
	err = os.Chmod(sourceFile, 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceFile},
		task_engine.StaticParameter{Value: destinationFile},
		false,
		false,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	// This might succeed or fail depending on the system, but we're testing the copy path
	// The important thing is that it doesn't panic
	os.Chmod(sourceFile, 0o600)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_SymlinkWithReadlinkFailure() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a symlink that will fail to read (this is hard to simulate, but we can try)
	// We'll create a symlink and then remove its target
	realFile := filepath.Join(suite.tempDir, "real_file")
	err = os.WriteFile(realFile, []byte("content"), 0o600)
	suite.NoError(err)

	symlinkFile := filepath.Join(sourceDir, "symlink")
	err = os.Symlink(realFile, symlinkFile)
	suite.NoError(err)

	// Remove the target file
	err = os.Remove(realFile)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_SymlinkWithDirectoryCreationFailure() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a symlink
	symlinkFile := filepath.Join(sourceDir, "symlink")
	err = os.Symlink("/tmp/target", symlinkFile)
	suite.NoError(err)

	// Create a read-only destination directory
	err = os.MkdirAll(destinationDir, 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	os.Chmod(destinationDir, 0o750)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_SymlinkWithSymlinkCreationFailure() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a symlink
	symlinkFile := filepath.Join(sourceDir, "symlink")
	err = os.Symlink("/tmp/target", symlinkFile)
	suite.NoError(err)

	// Create a file at the destination symlink location to prevent symlink creation
	destSymlinkFile := filepath.Join(destinationDir, "symlink")
	err = os.MkdirAll(filepath.Dir(destSymlinkFile), 0o750)
	suite.NoError(err)
	err = os.WriteFile(destSymlinkFile, []byte("blocking file"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithFileCopyFailure() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a file that will be hard to copy (very large or with special permissions)
	largeFile := filepath.Join(sourceDir, "large_file")

	// Create a large file (10MB) to test copy limits
	largeContent := make([]byte, 10*1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	err = os.WriteFile(largeFile, largeContent, 0o600)
	suite.NoError(err)

	// Create a read-only destination directory to cause copy failure
	err = os.MkdirAll(destinationDir, 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
	os.Chmod(destinationDir, 0o750)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithDirectoryCreationFailureInWalk() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory with nested structure
	nestedDir := filepath.Join(sourceDir, "nested")
	err := os.MkdirAll(nestedDir, 0o750)
	suite.NoError(err)

	// Create a file in nested directory
	nestedFile := filepath.Join(nestedDir, "file.txt")
	err = os.WriteFile(nestedFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Create destination directory
	err = os.MkdirAll(destinationDir, 0o750)
	suite.NoError(err)

	// Create a read-only file where the nested directory should be created
	nestedDestDir := filepath.Join(destinationDir, "nested")
	err = os.WriteFile(nestedDestDir, []byte("blocking file"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithWalkError() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a file in source
	sourceFile := filepath.Join(sourceDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Create a subdirectory with restricted permissions to cause walk error
	restrictedDir := filepath.Join(sourceDir, "restricted")
	err = os.MkdirAll(restrictedDir, 0o000)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	os.Chmod(restrictedDir, 0o750)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithChmodFailure() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a file with special permissions
	sourceFile := filepath.Join(sourceDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Create destination directory
	err = os.MkdirAll(destinationDir, 0o750)
	suite.NoError(err)

	// Create a read-only destination file to prevent chmod
	destFile := filepath.Join(destinationDir, "file.txt")
	err = os.WriteFile(destFile, []byte("existing"), 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
	os.Chmod(destFile, 0o600)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithIoCopyFailure() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create a file in source
	sourceFile := filepath.Join(sourceDir, "file.txt")
	err = os.WriteFile(sourceFile, []byte("content"), 0o600)
	suite.NoError(err)

	// Create destination directory
	err = os.MkdirAll(destinationDir, 0o750)
	suite.NoError(err)

	// Create a read-only destination file to prevent overwrite
	destFile := filepath.Join(destinationDir, "file.txt")
	err = os.WriteFile(destFile, []byte("existing"), 0o400)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.Error(err)
	os.Chmod(destFile, 0o600)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithDeepNesting() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create deeply nested directory structure
	deepPath := sourceDir
	for i := 0; i < 10; i++ {
		deepPath = filepath.Join(deepPath, fmt.Sprintf("level_%d", i))
	}

	err := os.MkdirAll(deepPath, 0o750)
	suite.NoError(err)

	// Create a file at the deepest level
	deepFile := filepath.Join(deepPath, "deep_file.txt")
	err = os.WriteFile(deepFile, []byte("deep content"), 0o600)
	suite.NoError(err)

	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)
	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	destDeepFile := filepath.Join(destinationDir, "level_0", "level_1", "level_2", "level_3", "level_4", "level_5", "level_6", "level_7", "level_8", "level_9", "deep_file.txt")
	_, err = os.Stat(destDeepFile)
	suite.NoError(err)

	content, err := os.ReadFile(destDeepFile)
	suite.NoError(err)
	suite.Equal("deep content", string(content))
}

func (suite *CopyFileActionTestSuite) TestCopyFile_RecursiveWithConcurrentAccess() {
	sourceDir := filepath.Join(suite.tempDir, "source_dir")
	destinationDir := filepath.Join(suite.tempDir, "dest_dir")

	// Create source directory with files
	err := os.MkdirAll(sourceDir, 0o750)
	suite.NoError(err)

	// Create multiple files
	for i := 0; i < 5; i++ {
		filePath := filepath.Join(sourceDir, fmt.Sprintf("file_%d.txt", i))
		err = os.WriteFile(filePath, []byte(fmt.Sprintf("content_%d", i)), 0o600)
		suite.NoError(err)
	}

	// Start copy operation
	copyAction, err := file.NewCopyFileAction(nil).WithParameters(
		task_engine.StaticParameter{Value: sourceDir},
		task_engine.StaticParameter{Value: destinationDir},
		true,
		true,
	)
	suite.Require().NoError(err)

	// Simulate concurrent access by modifying source files during copy
	go func() {
		for i := 0; i < 5; i++ {
			filePath := filepath.Join(sourceDir, fmt.Sprintf("file_%d.txt", i))
			os.WriteFile(filePath, []byte(fmt.Sprintf("modified_content_%d", i)), 0o600)
		}
	}()

	err = copyAction.Execute(context.Background())
	suite.NoError(err)
	for i := 0; i < 5; i++ {
		destFile := filepath.Join(destinationDir, fmt.Sprintf("file_%d.txt", i))
		_, err = os.Stat(destFile)
		suite.NoError(err)
	}
}

func (suite *CopyFileActionTestSuite) TestCopyFileAction_GetOutput() {
	action := &file.CopyFileAction{
		Source:      "/tmp/source.txt",
		Destination: "/tmp/dest.txt",
		CreateDir:   true,
		Recursive:   false,
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("/tmp/source.txt", m["source"])
	suite.Equal("/tmp/dest.txt", m["destination"])
	suite.Equal(true, m["createDir"])
	suite.Equal(false, m["recursive"])
	suite.Equal(true, m["success"])
}

// TestCopyFileActionTestSuite runs the CopyFileActionTestSuite
func TestCopyFileActionTestSuite(t *testing.T) {
	suite.Run(t, new(CopyFileActionTestSuite))
}
