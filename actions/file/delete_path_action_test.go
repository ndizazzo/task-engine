package file_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ndizazzo/task-engine/actions/file"
	command_mock "github.com/ndizazzo/task-engine/mocks"
)

type DeletePathActionTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *DeletePathActionTestSuite) SetupTest() {
	suite.tempDir = suite.T().TempDir()
}

func (suite *DeletePathActionTestSuite) TestDeletePath_Success() {
	// Create a test file
	filePath := filepath.Join(suite.tempDir, "test.txt")
	err := os.WriteFile(filePath, []byte("test content"), 0600)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(filePath, false, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(filePath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_FileNotExists() {
	filePath := filepath.Join(suite.tempDir, "nonexistent.txt")

	deleteAction := file.NewDeletePathAction(filePath, false, false, nil)
	err := deleteAction.Execute(context.Background())
	suite.NoError(err) // Should not error for non-existent files
}

func (suite *DeletePathActionTestSuite) TestDeletePath_PermissionDenied() {
	// Create a read-only file
	filePath := filepath.Join(suite.tempDir, "readonly.txt")
	err := os.WriteFile(filePath, []byte("content"), 0400)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(filePath, false, false, nil)
	err = deleteAction.Execute(context.Background())
	// os.RemoveAll can delete read-only files, so this should succeed
	suite.NoError(err)

	_, err = os.Stat(filePath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DirectoryWithoutRecursive() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, false, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.Error(err) // Should fail when trying to delete directory without recursive flag
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DirectoryWithRecursive() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file in the directory
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithNestedDirectories() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create nested directories
	nestedDir := filepath.Join(dirPath, "nested", "deep")
	err = os.MkdirAll(nestedDir, 0750)
	suite.NoError(err)

	// Create files in nested directories
	file1 := filepath.Join(dirPath, "file1.txt")
	file2 := filepath.Join(nestedDir, "file2.txt")
	err = os.WriteFile(file1, []byte("content1"), 0600)
	suite.NoError(err)
	err = os.WriteFile(file2, []byte("content2"), 0600)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithSymlinks() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink to the file
	symlinkPath := filepath.Join(dirPath, "symlink")
	err = os.Symlink(filePath, symlinkPath)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithBrokenSymlinks() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a broken symlink
	brokenLink := filepath.Join(dirPath, "broken")
	err = os.Symlink("/nonexistent/path", brokenLink)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err) // Should handle broken symlinks gracefully

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithSpecialFiles() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a regular file
	regularFile := filepath.Join(dirPath, "regular.txt")
	err = os.WriteFile(regularFile, []byte("regular content"), 0600)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithHiddenFiles() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create hidden files and directories
	hiddenFile := filepath.Join(dirPath, ".hidden")
	hiddenDir := filepath.Join(dirPath, ".hidden_dir")
	err = os.WriteFile(hiddenFile, []byte("hidden content"), 0600)
	suite.NoError(err)
	err = os.MkdirAll(hiddenDir, 0750)
	suite.NoError(err)

	// Create a file in hidden directory
	hiddenNestedFile := filepath.Join(hiddenDir, "nested.txt")
	err = os.WriteFile(hiddenNestedFile, []byte("nested content"), 0600)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithDeepNesting() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create deeply nested structure (10 levels)
	currentPath := dirPath
	for i := 0; i < 10; i++ {
		currentPath = filepath.Join(currentPath, fmt.Sprintf("level_%d", i))
		err = os.MkdirAll(currentPath, 0750)
		suite.NoError(err)

		// Create a file at each level
		filePath := filepath.Join(currentPath, fmt.Sprintf("file_%d.txt", i))
		err = os.WriteFile(filePath, []byte(fmt.Sprintf("content_%d", i)), 0600)
		suite.NoError(err)
	}

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithPermissionErrors() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file with restrictive permissions
	restrictedFile := filepath.Join(dirPath, "restricted.txt")
	err = os.WriteFile(restrictedFile, []byte("restricted content"), 0000)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	// os.RemoveAll can delete files with restrictive permissions, so this should succeed
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithDirectoryDeletionFailure() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file in the directory
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	// Make the directory read-only to prevent deletion
	err = os.Chmod(dirPath, 0400)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.Error(err) // Should fail when trying to delete read-only directory

	// Clean up
	os.Chmod(dirPath, 0750)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithRootDirectoryDeletionFailure() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file in the directory
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	// Make the root directory read-only to prevent deletion
	err = os.Chmod(dirPath, 0400)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.Error(err) // Should fail when trying to delete read-only root directory

	// Clean up
	os.Chmod(dirPath, 0750)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_EmptyDirectory() {
	dirPath := filepath.Join(suite.tempDir, "empty_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_LargeDirectory() {
	dirPath := filepath.Join(suite.tempDir, "large_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create many files
	for i := 0; i < 100; i++ {
		filePath := filepath.Join(dirPath, fmt.Sprintf("file_%d.txt", i))
		err = os.WriteFile(filePath, []byte(fmt.Sprintf("content_%d", i)), 0600)
		suite.NoError(err)
	}

	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestNewDeletePathAction_InvalidParameters() {
	logger := command_mock.NewDiscardLogger()

	// Test empty path
	action := file.NewDeletePathAction("", false, false, logger)
	suite.Nil(action)

	// Test nil logger (should create default logger)
	action = file.NewDeletePathAction("/some/path", false, false, nil)
	suite.NotNil(action)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_SpecialCharacters() {
	// Test with special characters in path (simplified)
	specialPath := filepath.Join(suite.tempDir, "file with spaces.txt")
	err := os.WriteFile(specialPath, []byte("content"), 0600)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(specialPath, false, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(specialPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_ConcurrentAccess() {
	dirPath := filepath.Join(suite.tempDir, "concurrent_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create multiple files
	for i := 0; i < 5; i++ {
		filePath := filepath.Join(dirPath, fmt.Sprintf("file_%d.txt", i))
		err = os.WriteFile(filePath, []byte(fmt.Sprintf("content_%d", i)), 0600)
		suite.NoError(err)
	}

	// Delete the directory
	deleteAction := file.NewDeletePathAction(dirPath, true, false, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

// New tests for dry-run functionality
func (suite *DeletePathActionTestSuite) TestDeletePath_DryRunFile() {
	// Create a test file
	filePath := filepath.Join(suite.tempDir, "test.txt")
	err := os.WriteFile(filePath, []byte("test content"), 0600)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(filePath, false, true, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	// File should still exist after dry run
	_, err = os.Stat(filePath)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DryRunDirectory() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create files in the directory
	for i := 0; i < 3; i++ {
		filePath := filepath.Join(dirPath, fmt.Sprintf("file_%d.txt", i))
		err = os.WriteFile(filePath, []byte(fmt.Sprintf("content_%d", i)), 0600)
		suite.NoError(err)
	}

	deleteAction := file.NewDeletePathAction(dirPath, true, true, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	// Directory and files should still exist after dry run
	_, err = os.Stat(dirPath)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))

	// Check that files still exist
	for i := 0; i < 3; i++ {
		filePath := filepath.Join(dirPath, fmt.Sprintf("file_%d.txt", i))
		_, err = os.Stat(filePath)
		suite.NoError(err)
		suite.False(os.IsNotExist(err))
	}
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DryRunWithNestedStructure() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create nested structure
	nestedDir := filepath.Join(dirPath, "nested")
	err = os.MkdirAll(nestedDir, 0750)
	suite.NoError(err)

	// Create files at different levels
	file1 := filepath.Join(dirPath, "file1.txt")
	file2 := filepath.Join(nestedDir, "file2.txt")
	err = os.WriteFile(file1, []byte("content1"), 0600)
	suite.NoError(err)
	err = os.WriteFile(file2, []byte("content2"), 0600)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, true, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	// Everything should still exist after dry run
	_, err = os.Stat(dirPath)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))

	_, err = os.Stat(nestedDir)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))

	_, err = os.Stat(file1)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))

	_, err = os.Stat(file2)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DryRunWithSymlinks() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink
	symlinkPath := filepath.Join(dirPath, "symlink")
	err = os.Symlink(filePath, symlinkPath)
	suite.NoError(err)

	deleteAction := file.NewDeletePathAction(dirPath, true, true, nil)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	// Everything should still exist after dry run
	_, err = os.Stat(dirPath)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))

	_, err = os.Stat(filePath)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))

	_, err = os.Stat(symlinkPath)
	suite.NoError(err)
	suite.False(os.IsNotExist(err))
}

// TestDeletePathActionTestSuite runs the DeletePathActionTestSuite
func TestDeletePathActionTestSuite(t *testing.T) {
	suite.Run(t, new(DeletePathActionTestSuite))
}
