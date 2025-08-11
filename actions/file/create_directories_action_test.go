package file_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	"github.com/ndizazzo/task-engine/testing/mocks"
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
	err = os.MkdirAll(suite.rootPath, 0o750)
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

	action, err := file.NewCreateDirectoriesAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.rootPath},
		task_engine.StaticParameter{Value: directories},
	)
	suite.Require().NoError(err)
	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(len(directories), action.Wrapped.CreatedDirsCount)
	for _, dir := range directories {
		fullPath := filepath.Join(suite.rootPath, dir)
		suite.DirExists(fullPath, "Directory should exist: %s", fullPath)

		info, err := os.Stat(fullPath)
		suite.NoError(err)
		suite.True(info.IsDir())
		suite.Equal(os.FileMode(0o750), info.Mode().Perm())
	}
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_EmptyInstallationPath() {
	logger := mocks.NewDiscardLogger()
	directories := []string{"data"}

	action, err := file.NewCreateDirectoriesAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: directories},
	)
	suite.NoError(err)
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectories_EmptyDirectoriesList() {
	logger := mocks.NewDiscardLogger()
	directories := []string{}

	action, err := file.NewCreateDirectoriesAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.rootPath},
		task_engine.StaticParameter{Value: directories},
	)
	suite.NoError(err)
	execErr := action.Wrapped.Execute(context.Background())
	suite.Error(execErr)
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

	action, err := file.NewCreateDirectoriesAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.rootPath},
		task_engine.StaticParameter{Value: directories},
	)
	suite.Require().NoError(err)
	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(3, action.Wrapped.CreatedDirsCount)
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

	action, err := file.NewCreateDirectoriesAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.rootPath},
		task_engine.StaticParameter{Value: directories},
	)
	suite.Require().NoError(err)
	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(len(directories), action.Wrapped.CreatedDirsCount)
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
	err := os.MkdirAll(existingPath, 0o750)
	suite.Require().NoError(err)

	action, err := file.NewCreateDirectoriesAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.rootPath},
		task_engine.StaticParameter{Value: directories},
	)
	suite.Require().NoError(err)
	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(len(directories), action.Wrapped.CreatedDirsCount)
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

	action, err := file.NewCreateDirectoriesAction(logger).WithParameters(
		task_engine.StaticParameter{Value: suite.rootPath},
		task_engine.StaticParameter{Value: directories},
	)
	suite.Require().NoError(err)
	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(len(directories), action.Wrapped.CreatedDirsCount)
	for _, dir := range directories {
		fullPath := filepath.Join(suite.rootPath, dir)
		cleanPath := filepath.Clean(fullPath)
		suite.DirExists(cleanPath, "Directory should exist: %s", cleanPath)
	}
}

func TestCreateDirectoriesActionTestSuite(t *testing.T) {
	suite.Run(t, new(CreateDirectoriesActionTestSuite))
}

func (suite *CreateDirectoriesActionTestSuite) TestCreateDirectoriesAction_GetOutput() {
	action := &file.CreateDirectoriesAction{}
	action.RootPath = "/tmp/root"
	action.Directories = []string{"a", "b"}
	action.CreatedDirsCount = 2

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal("/tmp/root", m["rootPath"])
	suite.Len(m["directories"], 2)
	suite.Equal(2, m["created"])
	suite.Equal(2, m["total"])
	suite.Equal(true, m["success"])
}
