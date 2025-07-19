package task_engine_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
)

// isDockerAvailable checks if Docker is available and accessible
func isDockerAvailable() bool {
	// On macOS CI runners, Docker is typically not available
	if runtime.GOOS == "darwin" {
		// Check if we're in a CI environment (GitHub Actions sets CI=true)
		if os.Getenv("CI") == "true" {
			return false
		}
	}

	// Try to run 'docker version' command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "version")
	err := cmd.Run()
	return err == nil
}

type TaskEngineE2ETestSuite struct {
	suite.Suite
	ctx           context.Context
	container     testcontainers.Container
	containerPath string
	hostTempDir   string // Host directory that's mounted to container
	logger        *slog.Logger
}

func (suite *TaskEngineE2ETestSuite) SetupSuite() {
	// Skip integration tests if explicitly disabled
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		suite.T().Skip("Integration tests disabled via SKIP_INTEGRATION_TESTS environment variable")
		return
	}

	// Skip the entire suite if Docker is not available (e.g., on macOS CI runners)
	if !isDockerAvailable() {
		suite.T().Skip("Docker is not available, skipping integration tests")
		return
	}

	suite.ctx = context.Background()
	suite.containerPath = "/tmp/test-workspace"
	suite.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	hostTempDir, err := os.MkdirTemp("", "task-engine-e2e-*")
	require.NoError(suite.T(), err, "Failed to create temp directory")
	suite.hostTempDir = hostTempDir

	req := testcontainers.ContainerRequest{
		Image:      "alpine:latest",
		Cmd:        []string{"tail", "-f", "/dev/null"},
		WaitingFor: wait.ForLog("").WithStartupTimeout(30 * time.Second),
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.Binds = []string{fmt.Sprintf("%s:%s", suite.hostTempDir, suite.containerPath)}
		},
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err, "Failed to start container")
	suite.container = container

	suite.logger.Info("Container started successfully", "containerID", container.GetContainerID())
}

func (suite *TaskEngineE2ETestSuite) TearDownSuite() {
	if suite.container != nil {
		err := suite.container.Terminate(suite.ctx)
		if err != nil {
			suite.logger.Error("Failed to terminate container", "error", err)
		}
	}

	if suite.hostTempDir != "" {
		err := os.RemoveAll(suite.hostTempDir)
		if err != nil {
			suite.logger.Error("Failed to cleanup temp directory", "error", err)
		}
	}
}

func (suite *TaskEngineE2ETestSuite) TestTaskEngineFileOperations() {
	testDirs := []string{"src", "docs", "tests", "configs"}
	testFileContent := []byte("# Test Project\n\nThis is a test project created by task engine.\n\nFeatures:\n- Directory creation\n- File writing\n- Container testing")
	testFilePath := filepath.Join(suite.hostTempDir, "README.md")
	task := &task_engine.Task{
		ID:   "e2e-file-operations",
		Name: "End-to-End File Operations Test",
		Actions: []task_engine.ActionWrapper{
			func() task_engine.ActionWrapper {
				action, err := file.NewCreateDirectoriesAction(
					suite.logger,
					suite.hostTempDir,
					testDirs,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					testFilePath,
					testFileContent,
					true,
					nil,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
		},
		Logger: suite.logger,
	}

	err := task.Run(suite.ctx)
	require.NoError(suite.T(), err, "Task execution should succeed")

	assert.Equal(suite.T(), 2, task.CompletedTasks, "All actions should be completed")
	assert.Greater(suite.T(), task.TotalTime, time.Duration(0), "Task should have taken some time")

	for _, dir := range testDirs {
		dirPath := filepath.Join(suite.containerPath, dir)
		suite.verifyDirectoryExists(dirPath)
	}

	containerFilePath := filepath.Join(suite.containerPath, "README.md")
	suite.verifyFileExists(containerFilePath)
	suite.verifyFileContent(containerFilePath, string(testFileContent))
}

func (suite *TaskEngineE2ETestSuite) TestTaskEngineWithCopyOperations() {
	sourceDir := filepath.Join(suite.hostTempDir, "source")
	destDir := filepath.Join(suite.hostTempDir, "destination")
	sourceFile := filepath.Join(sourceDir, "config.txt")
	destFile := filepath.Join(destDir, "config.txt")

	configContent := []byte("# Configuration File\nversion=1.0\ndebug=true\n")
	task := &task_engine.Task{
		ID:   "e2e-copy-operations",
		Name: "End-to-End Copy Operations Test",
		Actions: []task_engine.ActionWrapper{
			func() task_engine.ActionWrapper {
				action, err := file.NewCreateDirectoriesAction(
					suite.logger,
					suite.hostTempDir,
					[]string{"source"},
				)
				require.NoError(suite.T(), err)
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					sourceFile,
					configContent,
					true,
					nil,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := file.NewCopyFileAction(
					sourceFile,
					destFile,
					true,
					false,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
		},
		Logger: suite.logger,
	}

	err := task.Run(suite.ctx)
	require.NoError(suite.T(), err, "Copy task execution should succeed")

	containerSourceFile := filepath.Join(suite.containerPath, "source", "config.txt")
	containerDestFile := filepath.Join(suite.containerPath, "destination", "config.txt")
	suite.verifyFileExists(containerSourceFile)
	suite.verifyFileExists(containerDestFile)

	suite.verifyFileContent(containerSourceFile, string(configContent))
	suite.verifyFileContent(containerDestFile, string(configContent))
}

func (suite *TaskEngineE2ETestSuite) TestTaskEngineErrorHandling() {
	readOnlyDir := filepath.Join(suite.hostTempDir, "readonly-dir")
	err := os.MkdirAll(readOnlyDir, 0755)
	require.NoError(suite.T(), err, "Failed to create read-only directory")

	err = os.Chmod(readOnlyDir, 0444)
	require.NoError(suite.T(), err, "Failed to make directory read-only")

	invalidPath := filepath.Join(readOnlyDir, "cannot-write.txt")

	task := &task_engine.Task{
		ID:   "e2e-error-handling",
		Name: "End-to-End Error Handling Test",
		Actions: []task_engine.ActionWrapper{
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					invalidPath,
					[]byte("This should fail"),
					true,
					nil,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
		},
		Logger: suite.logger,
	}

	err = task.Run(suite.ctx)
	assert.Error(suite.T(), err, "Task should fail when trying to write to read-only directory")
	assert.Equal(suite.T(), 0, task.CompletedTasks, "No tasks should be completed on failure")

	containerInvalidPath := filepath.Join(suite.containerPath, "readonly-dir", "cannot-write.txt")
	suite.verifyFileDoesNotExist(containerInvalidPath)

	err = os.Chmod(readOnlyDir, 0755)
	require.NoError(suite.T(), err)
}

func (suite *TaskEngineE2ETestSuite) verifyDirectoryExists(path string) {
	exitCode, outputReader, err := suite.container.Exec(suite.ctx, []string{"ls", "-la", path})
	require.NoError(suite.T(), err, "Failed to execute ls command in container")
	assert.Equal(suite.T(), 0, exitCode, "Directory should exist: %s", path)

	_, err = io.ReadAll(outputReader)
	require.NoError(suite.T(), err, "Failed to read output")
}

func (suite *TaskEngineE2ETestSuite) verifyFileExists(path string) {
	exitCode, outputReader, err := suite.container.Exec(suite.ctx, []string{"ls", "-la", path})
	require.NoError(suite.T(), err, "Failed to execute ls command in container")
	assert.Equal(suite.T(), 0, exitCode, "File should exist: %s", path)

	_, err = io.ReadAll(outputReader)
	require.NoError(suite.T(), err, "Failed to read output")
}

func (suite *TaskEngineE2ETestSuite) verifyFileContent(path string, expectedContent string) {
	exitCode, outputReader, err := suite.container.Exec(suite.ctx, []string{"cat", path})
	require.NoError(suite.T(), err, "Failed to execute cat command in container")
	assert.Equal(suite.T(), 0, exitCode, "Should be able to read file: %s", path)

	outputBytes, err := io.ReadAll(outputReader)
	require.NoError(suite.T(), err, "Failed to read output")

	actualContent := string(outputBytes)

	if len(outputBytes) > 8 {
		actualContent = string(outputBytes[8:])
	}

	actualContent = strings.TrimSpace(actualContent)
	expectedContent = strings.TrimSpace(expectedContent)
	assert.Equal(suite.T(), expectedContent, actualContent, "File content should match")
}

func (suite *TaskEngineE2ETestSuite) verifyFileDoesNotExist(path string) {
	exitCode, _, err := suite.container.Exec(suite.ctx, []string{"ls", "-la", path})
	require.NoError(suite.T(), err, "Failed to execute ls command in container")
	assert.NotEqual(suite.T(), 0, exitCode, "File should not exist: %s", path)
}

func (suite *TaskEngineE2ETestSuite) TestTaskEngineComplexWorkflow() {
	projectPath := filepath.Join(suite.hostTempDir, "complex-project")

	task := &task_engine.Task{
		ID:   "e2e-complex-workflow",
		Name: "End-to-End Complex Workflow Test",
		Actions: []task_engine.ActionWrapper{
			func() task_engine.ActionWrapper {
				action, err := file.NewCreateDirectoriesAction(
					suite.logger,
					projectPath,
					[]string{"src", "docs", "tests", "configs", "scripts"},
				)
				require.NoError(suite.T(), err)
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					filepath.Join(projectPath, "src", "main.go"),
					[]byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}"),
					true,
					nil,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					filepath.Join(projectPath, "docs", "README.md"),
					[]byte("# Complex Project\n\nThis is a complex project structure created by task engine."),
					true,
					nil,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					filepath.Join(projectPath, "configs", "app.yaml"),
					[]byte("app:\n  name: complex-project\n  version: 1.0.0\n  debug: true"),
					true,
					nil,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					filepath.Join(projectPath, "tests", "main_test.go"),
					[]byte("package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {\n\t// Test implementation\n}"),
					true,
					nil,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := file.NewCopyFileAction(
					filepath.Join(projectPath, "src", "main.go"),
					filepath.Join(projectPath, "scripts", "main.go.backup"),
					false,
					false,
					suite.logger,
				)
				require.NoError(suite.T(), err)
				return action
			}(),
		},
		Logger: suite.logger,
	}

	err := task.Run(suite.ctx)
	require.NoError(suite.T(), err, "Complex workflow should succeed")

	assert.Equal(suite.T(), 6, task.CompletedTasks, "All 6 actions should be completed")

	containerProjectPath := filepath.Join(suite.containerPath, "complex-project")
	expectedDirs := []string{"src", "docs", "tests", "configs", "scripts"}
	for _, dir := range expectedDirs {
		suite.verifyDirectoryExists(filepath.Join(containerProjectPath, dir))
	}

	expectedFiles := []string{
		"src/main.go",
		"docs/README.md",
		"configs/app.yaml",
		"tests/main_test.go",
		"scripts/main.go.backup",
	}
	for _, file := range expectedFiles {
		suite.verifyFileExists(filepath.Join(containerProjectPath, file))
	}

	suite.verifyFileContent(
		filepath.Join(containerProjectPath, "src", "main.go"),
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
	)
	suite.verifyFileContent(
		filepath.Join(containerProjectPath, "scripts", "main.go.backup"),
		"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
	)
}

func TestTaskEngineE2ETestSuite(t *testing.T) {
	suite.Run(t, new(TaskEngineE2ETestSuite))
}
