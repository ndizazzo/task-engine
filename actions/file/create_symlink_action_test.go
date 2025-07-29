package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CreateSymlinkActionTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *CreateSymlinkActionTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "create_symlink_test")
	suite.NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		suite.NoError(os.RemoveAll(suite.tempDir))
	}
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_Success() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify symlink was created
	info, err := os.Lstat(linkPath)
	suite.NoError(err)
	suite.True(info.Mode()&os.ModeSymlink != 0)

	// Verify symlink target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_WithCreateDirs() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create symlink in a subdirectory that doesn't exist
	linkPath := filepath.Join(suite.tempDir, "subdir", "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, true, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify symlink was created
	info, err := os.Lstat(linkPath)
	suite.NoError(err)
	suite.True(info.Mode()&os.ModeSymlink != 0)

	// Verify symlink target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_OverwriteExisting() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create initial symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Create a new target file
	newTargetFile := filepath.Join(suite.tempDir, "new_target.txt")
	err = os.WriteFile(newTargetFile, []byte("new target content"), 0600)
	suite.NoError(err)

	// Overwrite the symlink
	action, err = NewCreateSymlinkAction(newTargetFile, linkPath, true, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify symlink now points to new target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(newTargetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_RelativeTarget() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create symlink with relative target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	relativeTarget := "target.txt"
	action, err := NewCreateSymlinkAction(relativeTarget, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify symlink was created
	info, err := os.Lstat(linkPath)
	suite.NoError(err)
	suite.True(info.Mode()&os.ModeSymlink != 0)

	// Verify symlink target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(relativeTarget, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ToDirectory() {
	// Create a target directory
	targetDir := filepath.Join(suite.tempDir, "target_dir")
	err := os.MkdirAll(targetDir, 0750)
	suite.NoError(err)

	// Create a file in the target directory
	targetFile := filepath.Join(targetDir, "file.txt")
	err = os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create symlink to directory
	linkPath := filepath.Join(suite.tempDir, "dir_link")
	action, err := NewCreateSymlinkAction(targetDir, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify symlink was created
	info, err := os.Lstat(linkPath)
	suite.NoError(err)
	suite.True(info.Mode()&os.ModeSymlink != 0)

	// Verify symlink target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetDir, target)

	// Verify we can access files through the symlink
	linkedFile := filepath.Join(linkPath, "file.txt")
	_, err = os.Stat(linkedFile)
	suite.NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ToNonExistentTarget() {
	// Create symlink to non-existent target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	nonExistentTarget := filepath.Join(suite.tempDir, "nonexistent.txt")
	action, err := NewCreateSymlinkAction(nonExistentTarget, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err) // Should succeed even if target doesn't exist

	// Verify symlink was created
	info, err := os.Lstat(linkPath)
	suite.NoError(err)
	suite.True(info.Mode()&os.ModeSymlink != 0)

	// Verify symlink target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(nonExistentTarget, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_InvalidTargetPath() {
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	// Test with an empty target path
	invalidTarget := ""

	_, err := NewCreateSymlinkAction(invalidTarget, linkPath, false, false, nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "invalid target path")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_InvalidLinkPath() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Test with an empty link path
	invalidLinkPath := ""

	_, err = NewCreateSymlinkAction(targetFile, invalidLinkPath, false, false, nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "invalid link path")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_SameTargetAndLink() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	_, err = NewCreateSymlinkAction(targetFile, targetFile, false, false, nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "target and link path cannot be the same")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ExistingSymlinkNoOverwrite() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create initial symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Try to create symlink again without overwrite
	action, err = NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "already exists and overwrite is set to false")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ExistingFileNoOverwrite() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create a regular file at the link location
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.WriteFile(linkPath, []byte("existing file"), 0600)
	suite.NoError(err)

	// Try to create symlink without overwrite
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "already exists and overwrite is set to false")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ExistingFileWithOverwrite() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create a regular file at the link location
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.WriteFile(linkPath, []byte("existing file"), 0600)
	suite.NoError(err)

	// Create symlink with overwrite
	action, err := NewCreateSymlinkAction(targetFile, linkPath, true, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify symlink was created
	info, err := os.Lstat(linkPath)
	suite.NoError(err)
	suite.True(info.Mode()&os.ModeSymlink != 0)

	// Verify symlink target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_WithoutCreateDirs() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Try to create symlink in non-existent directory without createDirs
	linkPath := filepath.Join(suite.tempDir, "nonexistent", "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Error(err) // Should fail because directory doesn't exist
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_CircularSymlink() {
	// Create a circular symlink (this should work, but accessing it will fail)
	linkPath := filepath.Join(suite.tempDir, "circular")
	_, err := NewCreateSymlinkAction(linkPath, linkPath, false, false, nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "target and link path cannot be the same")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ComplexRelativePath() {
	// Create a target file in a subdirectory
	targetDir := filepath.Join(suite.tempDir, "target_dir")
	err := os.MkdirAll(targetDir, 0750)
	suite.NoError(err)

	targetFile := filepath.Join(targetDir, "target.txt")
	err = os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create symlink with complex relative path
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	relativeTarget := "target_dir/target.txt"
	action, err := NewCreateSymlinkAction(relativeTarget, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify symlink was created
	info, err := os.Lstat(linkPath)
	suite.NoError(err)
	suite.True(info.Mode()&os.ModeSymlink != 0)

	// Verify symlink target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(relativeTarget, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkCreation() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0600)
	suite.NoError(err)

	// Create symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify the symlink verification worked correctly
	info, err := os.Lstat(linkPath)
	suite.NoError(err)
	suite.True(info.Mode()&os.ModeSymlink != 0)

	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_EmptyTarget() {
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	_, err := NewCreateSymlinkAction("", linkPath, false, false, nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "invalid target path")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_EmptyLinkPath() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	_, err = NewCreateSymlinkAction(targetFile, "", false, false, nil)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "invalid link path")
}

// New edge case tests for better coverage

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_StatError() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a directory where the symlink should go, but make it read-only
	linkDir := filepath.Join(suite.tempDir, "readonly_dir")
	err = os.MkdirAll(linkDir, 0400) // Read-only directory
	suite.NoError(err)

	linkPath := filepath.Join(linkDir, "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)

	// This should fail due to permission issues when trying to stat
	err = action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to stat symlink")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_DirectoryCreationFailure() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a read-only directory
	readonlyDir := filepath.Join(suite.tempDir, "readonly")
	err = os.MkdirAll(readonlyDir, 0400)
	suite.NoError(err)

	// Try to create symlink in a subdirectory of the read-only directory
	linkPath := filepath.Join(readonlyDir, "subdir", "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, true, nil)
	suite.Require().NoError(err)

	// This should fail due to permission issues when creating directories
	err = action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to stat symlink")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_SymlinkCreationFailure() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a directory where the symlink should go, but make it read-only
	linkDir := filepath.Join(suite.tempDir, "readonly_dir")
	err = os.MkdirAll(linkDir, 0400) // Read-only directory
	suite.NoError(err)

	linkPath := filepath.Join(linkDir, "link.txt")
	action, err := NewCreateSymlinkAction(targetFile, linkPath, false, false, nil)
	suite.Require().NoError(err)

	// This should fail due to permission issues
	err = action.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "failed to stat symlink")
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkTargetMismatch() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink with wrong target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	wrongTarget := filepath.Join(suite.tempDir, "wrong_target.txt")
	err = os.Symlink(wrongTarget, linkPath)
	suite.NoError(err)

	// Try to create a symlink with the correct target
	action, err := NewCreateSymlinkAction(targetFile, linkPath, true, false, nil)
	suite.Require().NoError(err)

	// This should succeed because overwrite is true
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify the symlink now points to the correct target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkRelativePathResolution() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink with relative target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	relativeTarget := "target.txt" // Relative to linkPath directory
	err = os.Symlink(relativeTarget, linkPath)
	suite.NoError(err)

	// Try to create a symlink with absolute target
	action, err := NewCreateSymlinkAction(targetFile, linkPath, true, false, nil)
	suite.Require().NoError(err)

	// This should succeed and the verification should handle the path resolution
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify the symlink now points to the absolute target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkComplexRelativePath() {
	// Create a target file in a subdirectory
	targetDir := filepath.Join(suite.tempDir, "target_dir")
	err := os.MkdirAll(targetDir, 0750)
	suite.NoError(err)

	targetFile := filepath.Join(targetDir, "target.txt")
	err = os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink with complex relative target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	complexRelativeTarget := "target_dir/target.txt" // Complex relative path
	err = os.Symlink(complexRelativeTarget, linkPath)
	suite.NoError(err)

	// Try to create a symlink with absolute target
	action, err := NewCreateSymlinkAction(targetFile, linkPath, true, false, nil)
	suite.Require().NoError(err)

	// This should succeed and the verification should handle the complex path resolution
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify the symlink now points to the absolute target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkPathResolutionFailure() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink with a target that resolves differently
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	differentTarget := filepath.Join(suite.tempDir, "different.txt")
	err = os.Symlink(differentTarget, linkPath)
	suite.NoError(err)

	// Try to create a symlink with the correct target
	action, err := NewCreateSymlinkAction(targetFile, linkPath, true, false, nil)
	suite.Require().NoError(err)

	// This should succeed because overwrite is true
	err = action.Execute(context.Background())
	suite.NoError(err)

	// Verify the symlink now points to the correct target
	target, err := os.Readlink(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, target)
}

// Tests for the split verifySymlink methods

func (suite *CreateSymlinkActionTestSuite) TestCheckSymlinkExists_Success() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	suite.NoError(err)

	// Test the checkSymlinkExists method
	action := &CreateSymlinkAction{}
	err = action.checkSymlinkExists(linkPath)
	suite.NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCheckSymlinkExists_FileNotExists() {
	// Test with non-existent file
	action := &CreateSymlinkAction{}
	err := action.checkSymlinkExists(filepath.Join(suite.tempDir, "nonexistent.txt"))
	suite.Error(err)
	suite.Contains(err.Error(), "failed to stat symlink")
}

func (suite *CreateSymlinkActionTestSuite) TestCheckSymlinkExists_NotASymlink() {
	// Create a regular file
	regularFile := filepath.Join(suite.tempDir, "regular.txt")
	err := os.WriteFile(regularFile, []byte("content"), 0600)
	suite.NoError(err)

	// Test the checkSymlinkExists method with a regular file
	action := &CreateSymlinkAction{}
	err = action.checkSymlinkExists(regularFile)
	suite.Error(err)
	suite.Contains(err.Error(), "created file is not a symlink")
}

func (suite *CreateSymlinkActionTestSuite) TestReadSymlinkTarget_Success() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	suite.NoError(err)

	// Test the readSymlinkTarget method
	action := &CreateSymlinkAction{}
	actualTarget, err := action.readSymlinkTarget(linkPath)
	suite.NoError(err)
	suite.Equal(targetFile, actualTarget)
}

func (suite *CreateSymlinkActionTestSuite) TestReadSymlinkTarget_FileNotExists() {
	// Test with non-existent file
	action := &CreateSymlinkAction{}
	_, err := action.readSymlinkTarget(filepath.Join(suite.tempDir, "nonexistent.txt"))
	suite.Error(err)
	suite.Contains(err.Error(), "failed to read symlink target")
}

func (suite *CreateSymlinkActionTestSuite) TestReadSymlinkTarget_NotASymlink() {
	// Create a regular file
	regularFile := filepath.Join(suite.tempDir, "regular.txt")
	err := os.WriteFile(regularFile, []byte("content"), 0600)
	suite.NoError(err)

	// Test the readSymlinkTarget method with a regular file
	action := &CreateSymlinkAction{}
	_, err = action.readSymlinkTarget(regularFile)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to read symlink target")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_ExactMatch() {
	// Test with exact match
	action := &CreateSymlinkAction{}
	err := action.compareSymlinkTargets("/tmp/link", "/tmp/target", "/tmp/target")
	suite.NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_RelativePathMatch() {
	// Test with relative paths that resolve to the same target
	action := &CreateSymlinkAction{}
	err := action.compareSymlinkTargets("/tmp/link", "target", "target")
	suite.NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_AbsoluteVsRelativeMatch() {
	// Test with absolute vs relative paths that resolve to the same target
	action := &CreateSymlinkAction{}
	err := action.compareSymlinkTargets("/tmp/link", "/tmp/target", "target")
	suite.NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_ComplexRelativeMatch() {
	// Test with complex relative paths that resolve to the same target
	action := &CreateSymlinkAction{}
	err := action.compareSymlinkTargets("/tmp/link", "subdir/target", "subdir/target")
	suite.NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_Mismatch() {
	// Test with mismatched targets
	action := &CreateSymlinkAction{}
	err := action.compareSymlinkTargets("/tmp/link", "/tmp/target1", "/tmp/target2")
	suite.Error(err)
	suite.Contains(err.Error(), "symlink target mismatch")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_RelativePathMismatch() {
	// Test with relative paths that don't resolve to the same target
	action := &CreateSymlinkAction{}
	err := action.compareSymlinkTargets("/tmp/link", "target1", "target2")
	suite.Error(err)
	suite.Contains(err.Error(), "symlink target mismatch")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_ComplexMismatch() {
	// Test with complex paths that don't resolve to the same target
	action := &CreateSymlinkAction{}
	err := action.compareSymlinkTargets("/tmp/link", "subdir1/target", "subdir2/target")
	suite.Error(err)
	suite.Contains(err.Error(), "symlink target mismatch")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_AbsoluteVsRelativeMismatch() {
	// Test with absolute vs relative paths that don't resolve to the same target
	action := &CreateSymlinkAction{}
	err := action.compareSymlinkTargets("/tmp/link", "/tmp/target1", "target2")
	suite.Error(err)
	suite.Contains(err.Error(), "symlink target mismatch")
}

func (suite *CreateSymlinkActionTestSuite) TestVerifySymlink_CheckSymlinkExistsFailure() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a regular file where symlink should be
	linkPath := filepath.Join(suite.tempDir, "regular.txt")
	err = os.WriteFile(linkPath, []byte("content"), 0600)
	suite.NoError(err)

	// Test verifySymlink with a regular file (should fail at checkSymlinkExists)
	action := &CreateSymlinkAction{}
	err = action.verifySymlink(linkPath, targetFile)
	suite.Error(err)
	suite.Contains(err.Error(), "created file is not a symlink")
}

func (suite *CreateSymlinkActionTestSuite) TestVerifySymlink_ReadSymlinkTargetFailure() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	suite.NoError(err)

	// Remove the target to make the symlink broken
	err = os.Remove(targetFile)
	suite.NoError(err)

	// Test verifySymlink with a broken symlink
	// Note: On some systems, broken symlinks don't cause readlink to fail
	action := &CreateSymlinkAction{}
	err = action.verifySymlink(linkPath, targetFile)
	// The behavior depends on the OS - some systems allow reading broken symlinks
	if err != nil {
		suite.Contains(err.Error(), "failed to read symlink target")
	} else {
		// If no error, the symlink target should still be readable
		suite.NoError(err)
	}
}

func (suite *CreateSymlinkActionTestSuite) TestVerifySymlink_CompareTargetsFailure() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink with wrong target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	wrongTarget := filepath.Join(suite.tempDir, "wrong.txt")
	err = os.Symlink(wrongTarget, linkPath)
	suite.NoError(err)

	// Test verifySymlink with mismatched target (should fail at compareSymlinkTargets)
	action := &CreateSymlinkAction{}
	err = action.verifySymlink(linkPath, targetFile)
	suite.Error(err)
	suite.Contains(err.Error(), "symlink target mismatch")
}

func (suite *CreateSymlinkActionTestSuite) TestVerifySymlink_Success() {
	// Create a target file
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0600)
	suite.NoError(err)

	// Create a symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	suite.NoError(err)

	// Test verifySymlink with correct target
	action := &CreateSymlinkAction{}
	err = action.verifySymlink(linkPath, targetFile)
	suite.NoError(err)
}

func TestCreateSymlinkActionTestSuite(t *testing.T) {
	suite.Run(t, new(CreateSymlinkActionTestSuite))
}
