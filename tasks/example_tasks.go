package tasks

import (
	"log/slog"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
	"github.com/ndizazzo/task-engine/actions/system"
)

// NewDockerSetupTask creates a task that demonstrates Docker environment setup
func NewDockerSetupTask(logger *slog.Logger, projectPath string) *engine.Task {
	return &engine.Task{
		ID:   "docker-setup-example",
		Name: "Docker Environment Setup",
		Actions: []engine.ActionWrapper{
			// This would include Docker actions when they're available
			// For now, we'll create a placeholder task
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					projectPath+"/docker-setup.log",
					[]byte("Docker setup completed"),
					true,
					nil,
					logger,
				)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// NewPackageManagementTask creates a task that demonstrates package management
func NewPackageManagementTask(logger *slog.Logger, packages []string) *engine.Task {
	return &engine.Task{
		ID:   "package-management-example",
		Name: "Package Management Example",
		Actions: []engine.ActionWrapper{
			system.NewUpdatePackagesAction(packages, logger),
		},
		Logger: logger,
	}
}

// NewSystemManagementTask creates a task that demonstrates system management operations
func NewSystemManagementTask(logger *slog.Logger, serviceName string) *engine.Task {
	return &engine.Task{
		ID:   "system-management-example",
		Name: "System Management Example",
		Actions: []engine.ActionWrapper{
			// This would include system management actions
			// For now, we'll create a placeholder task
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					"/tmp/system-management.log",
					[]byte("System management operations completed"),
					true,
					nil,
					logger,
				)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// NewUtilityOperationsTask creates a task that demonstrates utility operations
func NewUtilityOperationsTask(logger *slog.Logger) *engine.Task {
	return &engine.Task{
		ID:   "utility-operations-example",
		Name: "Utility Operations Example",
		Actions: []engine.ActionWrapper{
			// This would include utility actions
			// For now, we'll create a placeholder task
			func() engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					"/tmp/utility-operations.log",
					[]byte("Utility operations completed"),
					true,
					nil,
					logger,
				)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}
