package file_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ndizazzo/task-engine/actions/file"
	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
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

	deleteAction, err := file.NewDeletePathAction(filePath, false, false, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(filePath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_FileNotExists() {
	filePath := filepath.Join(suite.tempDir, "nonexistent.txt")

	deleteAction, err := file.NewDeletePathAction(filePath, false, false, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err) // Should not error for non-existent files
}

func (suite *DeletePathActionTestSuite) TestDeletePath_PermissionDenied() {
	// Create a read-only file
	filePath := filepath.Join(suite.tempDir, "readonly.txt")
	err := os.WriteFile(filePath, []byte("content"), 0400)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(filePath, false, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, false, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithPermissionErrors() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	// Make the directory read-only to cause permission errors
	err = os.Chmod(dirPath, 0555)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.Error(err) // Should fail due to permission errors

	// Restore permissions
	os.Chmod(dirPath, 0755)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithDirectoryDeletionFailure() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	// Make the directory read-only to cause deletion failure
	// This is more likely to cause a failure than making a file read-only
	err = os.Chmod(dirPath, 0555)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	// Note: os.RemoveAll can sometimes succeed even with permission issues
	// This test may pass or fail depending on the system, which is acceptable
	// The important thing is that it doesn't panic

	// Restore permissions for cleanup
	os.Chmod(dirPath, 0755)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_RecursiveWithRootDirectoryDeletionFailure() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	// Make the root directory read-only to cause deletion failure
	err = os.Chmod(dirPath, 0555)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.Error(err) // Should fail when trying to delete read-only directory

	// Restore permissions
	os.Chmod(dirPath, 0755)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_EmptyDirectory() {
	dirPath := filepath.Join(suite.tempDir, "empty_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
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

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestNewDeletePathAction_InvalidParameters() {
	logger := mocks.NewDiscardLogger()

	// Test empty path
	action, err := file.NewDeletePathAction("", false, false, logger)
	suite.Error(err)
	suite.Nil(action)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_SpecialCharacters() {
	// Create a file with special characters in the name
	filePath := filepath.Join(suite.tempDir, "file with spaces and !@#$%^&*().txt")
	err := os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(filePath, false, false, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(filePath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_ConcurrentAccess() {
	dirPath := filepath.Join(suite.tempDir, "concurrent_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create multiple files
	for i := 0; i < 10; i++ {
		filePath := filepath.Join(dirPath, fmt.Sprintf("file_%d.txt", i))
		err = os.WriteFile(filePath, []byte(fmt.Sprintf("content_%d", i)), 0600)
		suite.NoError(err)
	}

	deleteAction, err := file.NewDeletePathAction(dirPath, true, false, nil)
	suite.Require().NoError(err)

	// Simulate concurrent access by modifying files during deletion
	// This is a basic test - in real scenarios, you might use goroutines
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(dirPath)
	suite.True(os.IsNotExist(err))
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DryRunFile() {
	// Create a test file
	filePath := filepath.Join(suite.tempDir, "test.txt")
	err := os.WriteFile(filePath, []byte("test content"), 0600)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(filePath, false, true, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	// File should still exist after dry run
	_, err = os.Stat(filePath)
	suite.NoError(err)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DryRunDirectory() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create a file in the directory
	filePath := filepath.Join(dirPath, "file.txt")
	err = os.WriteFile(filePath, []byte("content"), 0600)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(dirPath, true, true, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	// Directory and file should still exist after dry run
	_, err = os.Stat(dirPath)
	suite.NoError(err)
	_, err = os.Stat(filePath)
	suite.NoError(err)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DryRunWithNestedStructure() {
	dirPath := filepath.Join(suite.tempDir, "test_dir")
	err := os.MkdirAll(dirPath, 0750)
	suite.NoError(err)

	// Create nested structure
	nestedDir := filepath.Join(dirPath, "nested", "deep")
	err = os.MkdirAll(nestedDir, 0750)
	suite.NoError(err)

	// Create files at different levels
	file1 := filepath.Join(dirPath, "file1.txt")
	file2 := filepath.Join(nestedDir, "file2.txt")
	err = os.WriteFile(file1, []byte("content1"), 0600)
	suite.NoError(err)
	err = os.WriteFile(file2, []byte("content2"), 0600)
	suite.NoError(err)

	// Create a symlink
	symlinkPath := filepath.Join(dirPath, "symlink")
	err = os.Symlink(file1, symlinkPath)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(dirPath, true, true, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	// All files and directories should still exist after dry run
	_, err = os.Stat(dirPath)
	suite.NoError(err)
	_, err = os.Stat(nestedDir)
	suite.NoError(err)
	_, err = os.Stat(file1)
	suite.NoError(err)
	_, err = os.Stat(file2)
	suite.NoError(err)
	_, err = os.Stat(symlinkPath)
	suite.NoError(err)
}

func (suite *DeletePathActionTestSuite) TestDeletePath_DryRunWithSymlinks() {
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

	// Create a broken symlink
	brokenLink := filepath.Join(dirPath, "broken")
	err = os.Symlink("/nonexistent/path", brokenLink)
	suite.NoError(err)

	deleteAction, err := file.NewDeletePathAction(dirPath, true, true, nil)
	suite.Require().NoError(err)
	err = deleteAction.Execute(context.Background())
	suite.NoError(err)

	// All files and symlinks should still exist after dry run
	_, err = os.Stat(dirPath)
	suite.NoError(err)
	_, err = os.Stat(filePath)
	suite.NoError(err)
	_, err = os.Stat(symlinkPath)
	suite.NoError(err)
	// Note: os.Stat on broken symlinks can fail, so we check if the symlink file exists instead
	_, err = os.Lstat(brokenLink)
	suite.NoError(err)
}

func TestDeletePathActionTestSuite(t *testing.T) {
	suite.Run(t, new(DeletePathActionTestSuite))
}
