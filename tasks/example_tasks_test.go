package tasks

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
