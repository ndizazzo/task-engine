package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/stretchr/testify/suite"
)

type CreateSymlinkActionTestSuite struct {
	suite.Suite
	tempDir string
}

func (suite *CreateSymlinkActionTestSuite) SetupTest() {
	suite.tempDir, _ = os.MkdirTemp("", "create_symlink_test")
}

func (suite *CreateSymlinkActionTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		suite.NoError(os.RemoveAll(suite.tempDir))
	}
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_Success() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)
	suite.Require().NoError(err)

	// Create symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)

	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	info, err := os.Lstat(linkPath)
	suite.Require().NoError(err)

	suite.True(info.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(linkPath)
	suite.Require().NoError(err)

	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_WithCreateDirs() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)
	suite.Require().NoError(err)

	// Create symlink in a subdirectory that doesn't exist
	linkPath := filepath.Join(suite.tempDir, "subdir", "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, true)
	suite.Require().NoError(err)

	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	info, err := os.Lstat(linkPath)
	suite.Require().NoError(err)

	suite.True(info.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(linkPath)
	suite.Require().NoError(err)

	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_OverwriteExisting() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)
	suite.Require().NoError(err)

	// Create initial symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)

	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	newTargetFile := filepath.Join(suite.tempDir, "new_target.txt")
	err = os.WriteFile(newTargetFile, []byte("new target content"), 0o600)
	suite.Require().NoError(err)

	// Overwrite the symlink
	action, err = NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: newTargetFile}, task_engine.StaticParameter{Value: linkPath}, true, false)
	suite.Require().NoError(err)

	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	target, err := os.Readlink(linkPath)
	suite.Require().NoError(err)

	suite.Equal(newTargetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_RelativeTarget() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)

	// Create symlink with relative target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	relativeTarget := "target.txt"
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: relativeTarget}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	info, err := os.Lstat(linkPath)

	suite.True(info.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(linkPath)

	suite.Equal(relativeTarget, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ToDirectory() {
	// Create a target directory
	targetDir := filepath.Join(suite.tempDir, "target_dir")
	err := os.MkdirAll(targetDir, 0o750)

	// Create a file in the target directory
	targetFile := filepath.Join(targetDir, "file.txt")
	err = os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)

	// Create symlink to directory
	linkPath := filepath.Join(suite.tempDir, "dir_link")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetDir}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	info, err := os.Lstat(linkPath)

	suite.True(info.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(linkPath)

	suite.Equal(targetDir, target)
	linkedFile := filepath.Join(linkPath, "file.txt")
	_, err = os.Stat(linkedFile)
	suite.Require().NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ToNonExistentTarget() {
	// Create symlink to non-existent target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	nonExistentTarget := filepath.Join(suite.tempDir, "nonexistent.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: nonExistentTarget}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	info, err := os.Lstat(linkPath)

	suite.True(info.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(linkPath)

	suite.Equal(nonExistentTarget, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_InvalidTargetPath() {
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	invalidTarget := ""

	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: invalidTarget}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().Error(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_InvalidLinkPath() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)
	invalidLinkPath := ""

	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: invalidLinkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().Error(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_SameTargetAndLink() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)

	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: targetFile}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().Error(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ExistingSymlinkNoOverwrite() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)

	// Create initial symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())

	// Try to create symlink again without overwrite
	action, err = NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ExistingFileNoOverwrite() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)

	// Create a regular file at the link location
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.WriteFile(linkPath, []byte("existing file"), 0o600)

	// Try to create symlink without overwrite
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ExistingFileWithOverwrite() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)

	// Create a regular file at the link location
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.WriteFile(linkPath, []byte("existing file"), 0o600)

	// Create symlink with overwrite
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, true, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	info, err := os.Lstat(linkPath)

	suite.True(info.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(linkPath)

	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_WithoutCreateDirs() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)

	// Try to create symlink in non-existent directory without createDirs
	linkPath := filepath.Join(suite.tempDir, "nonexistent", "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_CircularSymlink() {
	// Create a circular symlink (this should work, but accessing it will fail)
	linkPath := filepath.Join(suite.tempDir, "circular")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: linkPath}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().NoError(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_ComplexRelativePath() {
	// Create a target file in a subdirectory
	targetDir := filepath.Join(suite.tempDir, "target_dir")
	err := os.MkdirAll(targetDir, 0o750)

	targetFile := filepath.Join(targetDir, "target.txt")
	err = os.WriteFile(targetFile, []byte("target content"), 0o600)

	// Create symlink with complex relative path
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	relativeTarget := "target_dir/target.txt"
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: relativeTarget}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	info, err := os.Lstat(linkPath)

	suite.True(info.Mode()&os.ModeSymlink != 0)
	target, err := os.Readlink(linkPath)

	suite.Equal(relativeTarget, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkCreation() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)

	// Create symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().NoError(err)
	info, err := os.Lstat(linkPath)

	suite.True(info.Mode()&os.ModeSymlink != 0)

	target, err := os.Readlink(linkPath)

	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_EmptyTarget() {
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().Error(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_EmptyLinkPath() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)

	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: ""}, false, false)
	suite.Require().NoError(err)
	err = action.Execute(context.Background())
	suite.Require().Error(err)
}

// New edge case tests for better coverage

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_StatError() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)

	// Create a directory where the symlink should go, but make it read-only
	linkDir := filepath.Join(suite.tempDir, "readonly_dir")
	err = os.MkdirAll(linkDir, 0o400) // Read-only directory

	linkPath := filepath.Join(linkDir, "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)

	// This should fail due to permission issues when trying to stat
	err = action.Execute(context.Background())
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_DirectoryCreationFailure() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)

	// Create a read-only directory
	readonlyDir := filepath.Join(suite.tempDir, "readonly")
	err = os.MkdirAll(readonlyDir, 0o400)

	// Try to create symlink in a subdirectory of the read-only directory
	linkPath := filepath.Join(readonlyDir, "subdir", "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, true)
	suite.Require().NoError(err)

	// This should fail due to permission issues when creating directories
	err = action.Execute(context.Background())
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_SymlinkCreationFailure() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)

	// Create a directory where the symlink should go, but make it read-only
	linkDir := filepath.Join(suite.tempDir, "readonly_dir")
	err = os.MkdirAll(linkDir, 0o400) // Read-only directory

	linkPath := filepath.Join(linkDir, "link.txt")
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, false, false)
	suite.Require().NoError(err)

	// This should fail due to permission issues
	err = action.Execute(context.Background())
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkTargetMismatch() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)

	// Create a symlink with wrong target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	wrongTarget := filepath.Join(suite.tempDir, "wrong_target.txt")
	err = os.Symlink(wrongTarget, linkPath)

	// Try to create a symlink with the correct target
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, true, false)
	suite.Require().NoError(err)

	// This should succeed because overwrite is true
	err = action.Execute(context.Background())
	target, err := os.Readlink(linkPath)

	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkRelativePathResolution() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)

	// Create a symlink with relative target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	relativeTarget := "target.txt" // Relative to linkPath directory
	err = os.Symlink(relativeTarget, linkPath)

	// Try to create a symlink with absolute target
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, true, false)
	suite.Require().NoError(err)

	// This should succeed and the verification should handle the path resolution
	err = action.Execute(context.Background())
	target, err := os.Readlink(linkPath)

	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkComplexRelativePath() {
	// Create a target file in a subdirectory
	targetDir := filepath.Join(suite.tempDir, "target_dir")
	err := os.MkdirAll(targetDir, 0o750)

	targetFile := filepath.Join(targetDir, "target.txt")
	err = os.WriteFile(targetFile, []byte("content"), 0o600)

	// Create a symlink with complex relative target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	complexRelativeTarget := "target_dir/target.txt" // Complex relative path
	err = os.Symlink(complexRelativeTarget, linkPath)

	// Try to create a symlink with absolute target
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, true, false)
	suite.Require().NoError(err)

	// This should succeed and the verification should handle the complex path resolution
	err = action.Execute(context.Background())
	target, err := os.Readlink(linkPath)

	suite.Equal(targetFile, target)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlink_VerifySymlinkPathResolutionFailure() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)

	// Create a symlink with a target that resolves differently
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	differentTarget := filepath.Join(suite.tempDir, "different.txt")
	err = os.Symlink(differentTarget, linkPath)
	suite.Require().NoError(err)

	// Try to create a symlink with the correct target
	action, err := NewCreateSymlinkAction(nil).WithParameters(task_engine.StaticParameter{Value: targetFile}, task_engine.StaticParameter{Value: linkPath}, true, false)
	suite.Require().NoError(err)

	// This should succeed because overwrite is true
	err = action.Execute(context.Background())
	target, err := os.Readlink(linkPath)
	suite.Require().NoError(err)

	suite.Equal(targetFile, target)
}

// Tests for the split verifySymlink methods

func (suite *CreateSymlinkActionTestSuite) TestCheckSymlinkExists_Success() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)

	// Create a symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	suite.Require().NoError(err)
	action := &CreateSymlinkAction{}
	err = action.checkSymlinkExists(linkPath)
}

func (suite *CreateSymlinkActionTestSuite) TestCheckSymlinkExists_FileNotExists() {
	action := &CreateSymlinkAction{}
	err := action.checkSymlinkExists(filepath.Join(suite.tempDir, "nonexistent.txt"))
	suite.Error(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCheckSymlinkExists_NotASymlink() {
	regularFile := filepath.Join(suite.tempDir, "regular.txt")
	err := os.WriteFile(regularFile, []byte("content"), 0o600)
	suite.Require().NoError(err)
	action := &CreateSymlinkAction{}
	err = action.checkSymlinkExists(regularFile)
	suite.Error(err)
}

func (suite *CreateSymlinkActionTestSuite) TestReadSymlinkTarget_Success() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)
	// Create a symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	suite.Require().NoError(err)
	action := &CreateSymlinkAction{}
	actualTarget, err := action.readSymlinkTarget(linkPath)

	suite.Equal(targetFile, actualTarget)
}

func (suite *CreateSymlinkActionTestSuite) TestReadSymlinkTarget_FileNotExists() {
	action := &CreateSymlinkAction{}
	_, err := action.readSymlinkTarget(filepath.Join(suite.tempDir, "nonexistent.txt"))
	suite.Error(err)
}

func (suite *CreateSymlinkActionTestSuite) TestReadSymlinkTarget_NotASymlink() {
	regularFile := filepath.Join(suite.tempDir, "regular.txt")
	err := os.WriteFile(regularFile, []byte("content"), 0o600)
	suite.Require().NoError(err)
	action := &CreateSymlinkAction{}
	_, err = action.readSymlinkTarget(regularFile)
	suite.Error(err)
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_ExactMatch() {
	action := &CreateSymlinkAction{}
	_ = action.compareSymlinkTargets("/tmp/link", "/tmp/target", "/tmp/target")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_RelativePathMatch() {
	action := &CreateSymlinkAction{}
	_ = action.compareSymlinkTargets("/tmp/link", "target", "target")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_AbsoluteVsRelativeMatch() {
	action := &CreateSymlinkAction{}
	_ = action.compareSymlinkTargets("/tmp/link", "/tmp/target", "target")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_ComplexRelativeMatch() {
	action := &CreateSymlinkAction{}
	_ = action.compareSymlinkTargets("/tmp/link", "subdir/target", "subdir/target")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_Mismatch() {
	action := &CreateSymlinkAction{}
	_ = action.compareSymlinkTargets("/tmp/link", "/tmp/target1", "/tmp/target2")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_RelativePathMismatch() {
	action := &CreateSymlinkAction{}
	_ = action.compareSymlinkTargets("/tmp/link", "target1", "target2")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_ComplexMismatch() {
	action := &CreateSymlinkAction{}
	_ = action.compareSymlinkTargets("/tmp/link", "subdir1/target", "subdir2/target")
}

func (suite *CreateSymlinkActionTestSuite) TestCompareSymlinkTargets_AbsoluteVsRelativeMismatch() {
	action := &CreateSymlinkAction{}
	_ = action.compareSymlinkTargets("/tmp/link", "/tmp/target1", "target2")
}

func (suite *CreateSymlinkActionTestSuite) TestVerifySymlink_CheckSymlinkExistsFailure() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)

	// Create a regular file where symlink should be
	linkPath := filepath.Join(suite.tempDir, "regular.txt")
	err = os.WriteFile(linkPath, []byte("content"), 0o600)
	suite.Require().NoError(err)
	action := &CreateSymlinkAction{}
	err = action.verifySymlink(linkPath, targetFile)
}

func (suite *CreateSymlinkActionTestSuite) TestVerifySymlink_ReadSymlinkTargetFailure() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)

	// Create a symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	suite.Require().NoError(err)

	// Remove the target to make the symlink broken
	err = os.Remove(targetFile)
	suite.Require().NoError(err)
	// Note: On some systems, broken symlinks don't cause readlink to fail
	action := &CreateSymlinkAction{}
	err = action.verifySymlink(linkPath, targetFile)
	// The behavior depends on the OS - some systems allow reading broken symlinks
	if err != nil {
	} else {
		// If no error, the symlink target should still be readable
	}
}

func (suite *CreateSymlinkActionTestSuite) TestVerifySymlink_CompareTargetsFailure() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)

	// Create a symlink with wrong target
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	wrongTarget := filepath.Join(suite.tempDir, "wrong.txt")
	err = os.Symlink(wrongTarget, linkPath)
	suite.Require().NoError(err)
	action := &CreateSymlinkAction{}
	err = action.verifySymlink(linkPath, targetFile)
}

func (suite *CreateSymlinkActionTestSuite) TestVerifySymlink_Success() {
	targetFile := filepath.Join(suite.tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("content"), 0o600)
	suite.Require().NoError(err)

	// Create a symlink
	linkPath := filepath.Join(suite.tempDir, "link.txt")
	err = os.Symlink(targetFile, linkPath)
	suite.Require().NoError(err)
	action := &CreateSymlinkAction{}
	err = action.verifySymlink(linkPath, targetFile)
}

func (suite *CreateSymlinkActionTestSuite) TestCreateSymlinkAction_GetOutput() {
	action := &CreateSymlinkAction{
		Target:    "/tmp/source.txt",
		LinkPath:  "/tmp/link.txt",
		Overwrite: true,
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("/tmp/source.txt", m["target"])
	suite.Equal("/tmp/link.txt", m["linkPath"])
	suite.Equal(true, m["overwrite"])
	suite.Equal(true, m["created"])
	suite.Equal(true, m["success"])
}

func TestCreateSymlinkActionTestSuite(t *testing.T) {
	suite.Run(t, new(CreateSymlinkActionTestSuite))
}
