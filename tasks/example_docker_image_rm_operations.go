package tasks

import (
	"context"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
)

// NewDockerImageRmTask creates a task that demonstrates docker image removal operations
func NewDockerImageRmTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-image-rm-example",
		Name: "Docker Image Removal Operations Example",
		Actions: []task_engine.ActionWrapper{
			// Example 1: Remove image by name/tag
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "nginx:latest"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create DockerImageRmAction", "error", err)
					return nil
				}
				return action
			}(),

			// Example 2: Remove image by ID
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: "sha256:abc123def456789"},
					task_engine.StaticParameter{Value: true},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create DockerImageRmAction", "error", err)
					return nil
				}
				return action
			}(),

			// Example 3: Force remove image by name
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "redis:alpine"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: true},
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create DockerImageRmAction", "error", err)
					return nil
				}
				return action
			}(),

			// Example 4: Remove image by ID with no-prune option
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: "sha256:def456ghi789012"},
					task_engine.StaticParameter{Value: true},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: true},
				)
				if err != nil {
					logger.Error("Failed to create DockerImageRmAction", "error", err)
					return nil
				}
				return action
			}(),

			// Example 5: Force remove with no-prune option
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "postgres:13"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: true},
					task_engine.StaticParameter{Value: true},
				)
				if err != nil {
					logger.Error("Failed to create DockerImageRmAction", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleDockerImageRmOperations demonstrates how to use the docker image removal action
func ExampleDockerImageRmOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerImageRmTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker image removal operations", "error", err)
		return
	}

	logger.Info("Docker image removal operations completed successfully")
}

// NewDockerImageRmBatchTask creates a task that demonstrates batch removal of multiple images
func NewDockerImageRmBatchTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-image-rm-batch-example",
		Name: "Docker Image Removal Batch Operations Example",
		Actions: []task_engine.ActionWrapper{
			// Example 1: Remove nginx image
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "nginx:latest"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create DockerImageRmAction for nginx", "error", err)
					return nil
				}
				return action
			}(),

			// Example 2: Remove redis image
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "redis:alpine"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create DockerImageRmAction for redis", "error", err)
					return nil
				}
				return action
			}(),

			// Example 3: Remove postgres image
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "postgres:13"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create DockerImageRmAction for postgres", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleDockerImageRmBatchOperations demonstrates batch removal of multiple images
func ExampleDockerImageRmBatchOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerImageRmBatchTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker image removal batch operations", "error", err)
		return
	}

	logger.Info("Docker image removal batch operations completed successfully")
}

// NewDockerImageRmForceTask creates a task that demonstrates force removal of images
func NewDockerImageRmForceTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-image-rm-force-example",
		Name: "Docker Image Force Removal Operations Example",
		Actions: []task_engine.ActionWrapper{
			// Example 1: Force remove nginx image
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "nginx:latest"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: true}, // force=true
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create force DockerImageRmAction for nginx", "error", err)
					return nil
				}
				return action
			}(),

			// Example 2: Force remove redis image
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "redis:alpine"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: true}, // force=true
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create force DockerImageRmAction for redis", "error", err)
					return nil
				}
				return action
			}(),

			// Example 3: Force remove image by ID
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: "sha256:force123def456789"},
					task_engine.StaticParameter{Value: true}, // removeByID=true
					task_engine.StaticParameter{Value: true}, // force=true
					task_engine.StaticParameter{Value: false},
				)
				if err != nil {
					logger.Error("Failed to create force DockerImageRmAction by ID", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleDockerImageRmForceOperations demonstrates force removal of images
func ExampleDockerImageRmForceOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerImageRmForceTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker image force removal operations", "error", err)
		return
	}

	logger.Info("Docker image force removal operations completed successfully")
}

// NewDockerImageRmCleanupTask creates a task that demonstrates cleanup operations with no-prune
func NewDockerImageRmCleanupTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-image-rm-cleanup-example",
		Name: "Docker Image Cleanup Operations Example",
		Actions: []task_engine.ActionWrapper{
			// Example 1: Remove nginx without pruning parent layers
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "nginx:latest"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: true}, // noPrune=true
				)
				if err != nil {
					logger.Error("Failed to create no-prune DockerImageRmAction for nginx", "error", err)
					return nil
				}
				return action
			}(),

			// Example 2: Remove image by ID with no-prune
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: "sha256:cleanup123def456"},
					task_engine.StaticParameter{Value: true}, // removeByID=true
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: true}, // noPrune=true
				)
				if err != nil {
					logger.Error("Failed to create no-prune DockerImageRmAction by ID", "error", err)
					return nil
				}
				return action
			}(),

			// Example 3: Force remove with no-prune
			func() task_engine.ActionWrapper {
				action, err := docker.NewDockerImageRmAction(logger).WithParameters(
					task_engine.StaticParameter{Value: "busybox:latest"},
					task_engine.StaticParameter{Value: ""},
					task_engine.StaticParameter{Value: false},
					task_engine.StaticParameter{Value: true}, // force=true
					task_engine.StaticParameter{Value: true}, // noPrune=true
				)
				if err != nil {
					logger.Error("Failed to create force no-prune DockerImageRmAction", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleDockerImageRmCleanupOperations demonstrates cleanup operations with no-prune
func ExampleDockerImageRmCleanupOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerImageRmCleanupTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker image cleanup operations", "error", err)
		return
	}

	logger.Info("Docker image cleanup operations completed successfully")
}
