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
			docker.NewDockerImageRmByNameAction(logger, "nginx:latest"),

			// Example 2: Remove image by ID
			docker.NewDockerImageRmByIDAction(logger, "sha256:abc123def456789"),

			// Example 3: Force remove image by name
			docker.NewDockerImageRmByNameAction(logger, "redis:alpine",
				docker.WithForce()),

			// Example 4: Remove image by ID with no-prune option
			docker.NewDockerImageRmByIDAction(logger, "sha256:def456ghi789012",
				docker.WithNoPrune()),

			// Example 5: Force remove with no-prune option
			docker.NewDockerImageRmByNameAction(logger, "postgres:13",
				docker.WithForce(),
				docker.WithNoPrune()),
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
			// Remove multiple images by name
			docker.NewDockerImageRmByNameAction(logger, "nginx:latest"),
			docker.NewDockerImageRmByNameAction(logger, "redis:alpine"),
			docker.NewDockerImageRmByNameAction(logger, "postgres:13"),
			docker.NewDockerImageRmByNameAction(logger, "node:16"),

			// Remove multiple images by ID
			docker.NewDockerImageRmByIDAction(logger, "sha256:abc123def456789"),
			docker.NewDockerImageRmByIDAction(logger, "sha256:def456ghi789012"),
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
			// Force remove images that might be in use
			docker.NewDockerImageRmByNameAction(logger, "nginx:latest",
				docker.WithForce()),
			docker.NewDockerImageRmByNameAction(logger, "redis:alpine",
				docker.WithForce()),
			docker.NewDockerImageRmByIDAction(logger, "sha256:abc123def456789",
				docker.WithForce()),
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
			// Remove specific images without pruning parent layers
			docker.NewDockerImageRmByNameAction(logger, "nginx:latest",
				docker.WithNoPrune()),
			docker.NewDockerImageRmByNameAction(logger, "redis:alpine",
				docker.WithNoPrune()),
			docker.NewDockerImageRmByIDAction(logger, "sha256:abc123def456789",
				docker.WithNoPrune()),
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
