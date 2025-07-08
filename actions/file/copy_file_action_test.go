package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
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

	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.NoError(err)

	copyAction := &task_engine.Action[*file.CopyFileAction]{
		ID: "copy-file-success",
		Wrapped: &file.CopyFileAction{
			Source:      sourceFile,
			Destination: destinationFile,
			CreateDir:   false,
		},
	}

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

	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.NoError(err)

	copyAction := &task_engine.Action[*file.CopyFileAction]{
		ID: "copy-file-create-dir",
		Wrapped: &file.CopyFileAction{
			Source:      sourceFile,
			Destination: destinationFile,
			CreateDir:   true,
		},
	}

	err = copyAction.Execute(context.Background())
	suite.NoError(err)

	_, err = os.Stat(destinationFile)
	suite.NoError(err)
}

func (suite *CopyFileActionTestSuite) TestCopyFile_CreateDirFalse() {
	sourceFile := filepath.Join(suite.tempDir, "test_source.txt")
	destinationDir := filepath.Join(suite.tempDir, "nested")
	destinationFile := filepath.Join(destinationDir, "test_destination.txt")

	err := os.WriteFile(sourceFile, []byte("test content"), 0600)
	suite.NoError(err)

	copyAction := file.NewCopyFileAction(sourceFile, destinationFile, false, nil)
	err = copyAction.Execute(context.Background())

	suite.Error(err)
}

// TestCopyFileActionTestSuite runs the CopyFileActionTestSuite
func TestCopyFileActionTestSuite(t *testing.T) {
	suite.Run(t, new(CopyFileActionTestSuite))
}
