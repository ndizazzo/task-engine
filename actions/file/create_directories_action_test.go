package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ndizazzo/task-engine/actions/file"
	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/suite"
)

type CreateDirectoriesActionTestSuite struct {
	suite.Suite
	tempDir  string
	rootPath string
}

func (suite *CreateDirectoriesActionTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "create_dirs_test_*")
	suite.Require().NoError(err)

	suite.rootPath = filepath.Join(suite.tempDir, "installation")
	err = os.MkdirAll(suite.rootPath, 0755)
	suite.Require().NoError(err)
}

func (suite *CreateDirectoriesActionTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_Success() {
	logger := mocks.NewDiscardLogger()
	directories := []string{
		"data/models",
		"config/backend",
		"logs",
		"scripts",
	}

	action := file.NewCreateDirectoriesAction(logger, suite.rootPath, directories)
	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(len(directories), action.Wrapped.CreatedDirsCount)

	// Verify all directories were created
	for _, dir := range directories {
		fullPath := filepath.Join(suite.rootPath, dir)
		suite.DirExists(fullPath, "Directory should exist: %s", fullPath)

		info, err := os.Stat(fullPath)
		suite.NoError(err)
		suite.True(info.IsDir())
		suite.Equal(os.FileMode(0755), info.Mode().Perm())
	}
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_EmptyInstallationPath() {
	logger := mocks.NewDiscardLogger()
	directories := []string{"data"}

	action := file.NewCreateDirectoriesAction(logger, "", directories)

	// With validation, action should be nil for invalid parameters
	suite.Nil(action, "Action should be nil when rootPath is empty")
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_EmptyDirectoriesList() {
	logger := mocks.NewDiscardLogger()
	directories := []string{}

	action := file.NewCreateDirectoriesAction(logger, suite.rootPath, directories)

	// With validation, action should be nil for empty directories list
	suite.Nil(action, "Action should be nil when directories list is empty")
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_WithEmptyDirNames() {
	logger := mocks.NewDiscardLogger()
	directories := []string{
		"data",
		"", // empty directory name
		"logs",
		"", // another empty directory name
		"config",
	}

	action := file.NewCreateDirectoriesAction(logger, suite.rootPath, directories)
	err := action.Execute(context.Background())

	suite.NoError(err)
	// Should create 3 directories (skipping the 2 empty ones)
	suite.Equal(3, action.Wrapped.CreatedDirsCount)

	// Verify only non-empty directories were created
	expectedDirs := []string{"data", "logs", "config"}
	for _, dir := range expectedDirs {
		fullPath := filepath.Join(suite.rootPath, dir)
		suite.DirExists(fullPath, "Directory should exist: %s", fullPath)
	}
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_NestedPaths() {
	logger := mocks.NewDiscardLogger()
	directories := []string{
		"data/models",
		"data/utility_models",
		"infrastructure/compose/app",
		"infrastructure/compose/display",
		"config/backend",
	}

	action := file.NewCreateDirectoriesAction(logger, suite.rootPath, directories)
	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(len(directories), action.Wrapped.CreatedDirsCount)

	// Verify all nested directories were created
	for _, dir := range directories {
		fullPath := filepath.Join(suite.rootPath, dir)
		suite.DirExists(fullPath, "Nested directory should exist: %s", fullPath)
	}
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_AlreadyExists() {
	logger := mocks.NewDiscardLogger()
	directories := []string{
		"existing_dir",
		"new_dir",
	}

	// Pre-create one of the directories
	existingPath := filepath.Join(suite.rootPath, "existing_dir")
	err := os.MkdirAll(existingPath, 0755)
	suite.Require().NoError(err)

	action := file.NewCreateDirectoriesAction(logger, suite.rootPath, directories)
	err = action.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(len(directories), action.Wrapped.CreatedDirsCount)

	// Verify both directories exist
	for _, dir := range directories {
		fullPath := filepath.Join(suite.rootPath, dir)
		suite.DirExists(fullPath, "Directory should exist: %s", fullPath)
	}
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_RelativePaths() {
	logger := mocks.NewDiscardLogger()
	directories := []string{
		"./data",
		"../test", // This should be relative to installation path
		"logs/./app",
	}

	action := file.NewCreateDirectoriesAction(logger, suite.rootPath, directories)
	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(len(directories), action.Wrapped.CreatedDirsCount)

	// Verify directories were created (filepath.Join handles relative paths)
	for _, dir := range directories {
		fullPath := filepath.Join(suite.rootPath, dir)
		cleanPath := filepath.Clean(fullPath)
		suite.DirExists(cleanPath, "Directory should exist: %s", cleanPath)
	}
}

func TestCreateDirectoriesActionTestSuite(t *testing.T) {
	suite.Run(t, new(CreateDirectoriesActionTestSuite))
}
