package tasks

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Additional smoke tests for constructors in the tasks package
// These cover constructors that are not covered by existing tests.

func TestNewDockerStatusTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerStatusTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "container-state-example", task.ID)
	assert.Equal(t, "Container State Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerStatusFilteringTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerStatusFilteringTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-status-filtering-example", task.ID)
	assert.Equal(t, "Docker Status Filtering Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerStatusMonitoringTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerStatusMonitoringTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-status-monitoring-example", task.ID)
	assert.Equal(t, "Docker Status Monitoring Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerLoadTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerLoadTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-load-example", task.ID)
	assert.Equal(t, "Docker Load Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerLoadBatchTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerLoadBatchTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-load-batch-example", task.ID)
	assert.Equal(t, "Docker Load Batch Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerLoadPlatformSpecificTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerLoadPlatformSpecificTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-load-platform-example", task.ID)
	assert.Equal(t, "Docker Load Platform-Specific Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewExtractOperationsTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewExtractOperationsTask(logger)
	assert.NotNil(t, task)
	// Some tasks do not set an explicit ID
	assert.Equal(t, "", task.ID)
	assert.Equal(t, "extract-operations", task.Name)
	// Some constructors do not set a Logger internally; allow nil
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewExtractWithDirectoriesTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewExtractWithDirectoriesTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "", task.ID)
	assert.Equal(t, "extract-with-directories", task.Name)
	// Logger may not be set by this constructor
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewExtractCompressedArchivesTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewExtractCompressedArchivesTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "", task.ID)
	assert.Equal(t, "extract-compressed-archives", task.Name)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewServiceStatusTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewServiceStatusTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "service-status-example", task.ID)
	assert.Equal(t, "Service Status Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewServiceHealthCheckTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewServiceHealthCheckTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "service-health-check-example", task.ID)
	assert.Equal(t, "Service Health Check Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewServiceMonitoringTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewServiceMonitoringTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "service-monitoring-example", task.ID)
	assert.Equal(t, "Service Monitoring Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerImageRmTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerImageRmTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-image-rm-example", task.ID)
	assert.Equal(t, "Docker Image Removal Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerImageRmBatchTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerImageRmBatchTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-image-rm-batch-example", task.ID)
	assert.Equal(t, "Docker Image Removal Batch Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerImageRmForceTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerImageRmForceTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-image-rm-force-example", task.ID)
	assert.Equal(t, "Docker Image Force Removal Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerImageRmCleanupTaskSmoke(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	task := NewDockerImageRmCleanupTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-image-rm-cleanup-example", task.ID)
	assert.Equal(t, "Docker Image Cleanup Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerSetupTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	projectPath := "/tmp/test-project"

	task := NewDockerSetupTask(logger, projectPath)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-setup-example", task.ID)
	assert.Equal(t, "Docker Environment Setup", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Len(t, task.Actions, 1)
}

func TestNewPackageManagementTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	packages := []string{"git", "curl", "wget"}

	task := NewPackageManagementTask(logger, packages)
	assert.NotNil(t, task)
	assert.Equal(t, "package-management-example", task.ID)
	assert.Equal(t, "Package Management Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Len(t, task.Actions, 1)
}

func TestNewSystemManagementTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	serviceName := "nginx"

	task := NewSystemManagementTask(logger, serviceName)
	assert.NotNil(t, task)
	assert.Equal(t, "system-management-example", task.ID)
	assert.Equal(t, "System Management Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Len(t, task.Actions, 1)
}

func TestNewUtilityOperationsTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewUtilityOperationsTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "utility-operations-example", task.ID)
	assert.Equal(t, "Utility Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Len(t, task.Actions, 1)
}

func TestTaskCreationWithNilLogger(t *testing.T) {
	// Test that tasks can be created with nil logger
	task := NewDockerSetupTask(nil, "/tmp/test")
	assert.NotNil(t, task)
	assert.Nil(t, task.Logger)
}

func TestTaskCreationWithEmptyProjectPath(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerSetupTask(logger, "")
	assert.NotNil(t, task)
	assert.Equal(t, "docker-setup-example", task.ID)
}

func TestTaskCreationWithEmptyPackages(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewPackageManagementTask(logger, []string{})
	assert.NotNil(t, task)
	assert.Equal(t, "package-management-example", task.ID)
}

func TestTaskCreationWithEmptyServiceName(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewSystemManagementTask(logger, "")
	assert.NotNil(t, task)
	assert.Equal(t, "system-management-example", task.ID)
}

// Compression Operations Tests
func TestNewCompressionOperationsTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	workingDir := "/tmp/test"

	task := NewCompressionOperationsTask(logger, workingDir)
	assert.NotNil(t, task)
	assert.Equal(t, "compression-example", task.ID)
	assert.Equal(t, "Compression Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewCompressionWithAutoDetectTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	workingDir := "/tmp/test"

	task := NewCompressionWithAutoDetectTask(logger, workingDir)
	assert.NotNil(t, task)
	assert.Equal(t, "compression-auto-detect", task.ID)
	assert.Equal(t, "Compression Auto-Detection Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewCompressionWorkflowTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	workingDir := "/tmp/test"

	task := NewCompressionWorkflowTask(logger, workingDir)
	assert.NotNil(t, task)
	assert.Equal(t, "compression-workflow", task.ID)
	assert.Equal(t, "Compression Workflow Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

// Docker Image RM Operations Tests
func TestNewDockerImageRmTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerImageRmTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-image-rm-example", task.ID)
	assert.Equal(t, "Docker Image Removal Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerImageRmBatchTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerImageRmBatchTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-image-rm-batch-example", task.ID)
	assert.Equal(t, "Docker Image Removal Batch Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerImageRmForceTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerImageRmForceTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-image-rm-force-example", task.ID)
	assert.Equal(t, "Docker Image Force Removal Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerImageRmCleanupTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerImageRmCleanupTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-image-rm-cleanup-example", task.ID)
	assert.Equal(t, "Docker Image Cleanup Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

// Docker Load Operations Tests
func TestNewDockerLoadTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerLoadTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-load-example", task.ID)
	assert.Equal(t, "Docker Load Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerLoadBatchTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerLoadBatchTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-load-batch-example", task.ID)
	assert.Equal(t, "Docker Load Batch Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerLoadPlatformSpecificTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerLoadPlatformSpecificTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-load-platform-example", task.ID)
	assert.Equal(t, "Docker Load Platform-Specific Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

// Docker Status Operations Tests
func TestNewDockerStatusTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerStatusTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "container-state-example", task.ID)
	assert.Equal(t, "Container State Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerStatusFilteringTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerStatusFilteringTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-status-filtering-example", task.ID)
	assert.Equal(t, "Docker Status Filtering Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewDockerStatusMonitoringTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewDockerStatusMonitoringTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "docker-status-monitoring-example", task.ID)
	assert.Equal(t, "Docker Status Monitoring Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

// Extract Operations Tests
func TestNewExtractOperationsTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewExtractOperationsTask(logger)
	assert.NotNil(t, task)
	// Logger is optional for this constructor in current implementation
	assert.Greater(t, len(task.Actions), 0)
}

// (Removed duplicate test for TestNewExtractWithDirectoriesTaskSmoke to avoid conflicts)

func TestNewExtractCompressedArchivesTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewExtractCompressedArchivesTask(logger)
	assert.NotNil(t, task)
	assert.Greater(t, len(task.Actions), 0)
}

// File Operations Tests
func TestNewFileOperationsTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	workingDir := "/tmp/test"

	task := NewFileOperationsTask(logger, workingDir)
	assert.NotNil(t, task)
	assert.Equal(t, "file-operations-example", task.ID)
	assert.Equal(t, "File Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

// Parameter Passing Tests
func TestNewExampleParameterPassingTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	config := ExampleParameterPassingConfig{
		SourcePath:      "/tmp/source.txt",
		DestinationPath: "/tmp/dest.txt",
	}

	task := NewExampleParameterPassingTask(config, logger)
	assert.NotNil(t, task)
	assert.Equal(t, "example-parameter-passing", task.ID)
	assert.Equal(t, "Example Parameter Passing Between Actions", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewExampleCrossTaskParameterPassing(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	config := CrossTaskConfig{
		SourceTaskID:      "task-1",
		DestinationTaskID: "task-2",
	}

	task := NewExampleCrossTaskParameterPassing(config, logger)
	assert.NotNil(t, task)
	assert.Equal(t, "example-cross-task-parameter-passing", task.ID)
	assert.Equal(t, "Example Cross-Task Parameter Passing", task.Name)
	assert.NotNil(t, task.Logger)
}

// Service Status Operations Tests
func TestNewServiceStatusTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewServiceStatusTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "service-status-example", task.ID)
	assert.Equal(t, "Service Status Operations Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewServiceHealthCheckTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewServiceHealthCheckTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "service-health-check-example", task.ID)
	assert.Equal(t, "Service Health Check Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}

func TestNewServiceMonitoringTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	task := NewServiceMonitoringTask(logger)
	assert.NotNil(t, task)
	assert.Equal(t, "service-monitoring-example", task.ID)
	assert.Equal(t, "Service Monitoring Example", task.Name)
	assert.NotNil(t, task.Logger)
	assert.Greater(t, len(task.Actions), 0)
}
