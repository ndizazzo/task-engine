package tasks

import (
	"context"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
)

// NewDockerLoadTask creates a task that demonstrates docker load operations
func NewDockerLoadTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-load-example",
		Name: "Docker Load Operations Example",
		Actions: []task_engine.ActionWrapper{
			// Example 1: Basic image load from tar file
			docker.NewDockerLoadAction(logger, "/path/to/nginx.tar"),

			// Example 2: Load with platform specification
			docker.NewDockerLoadAction(logger, "/path/to/multi-platform.tar",
				docker.WithPlatform("linux/amd64")),

			// Example 3: Load with quiet mode
			docker.NewDockerLoadAction(logger, "/path/to/redis.tar",
				docker.WithQuiet()),

			// Example 4: Load with both platform and quiet options
			docker.NewDockerLoadAction(logger, "/path/to/postgres.tar",
				docker.WithPlatform("linux/arm64"),
				docker.WithQuiet()),
		},
		Logger: logger,
	}
}

// ExampleDockerLoadOperations demonstrates how to use the docker load action
func ExampleDockerLoadOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerLoadTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker load operations", "error", err)
		return
	}

	logger.Info("Docker load operations completed successfully")
}

// NewDockerLoadBatchTask creates a task that demonstrates batch loading of multiple images
func NewDockerLoadBatchTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-load-batch-example",
		Name: "Docker Load Batch Operations Example",
		Actions: []task_engine.ActionWrapper{
			// Load multiple images in sequence
			docker.NewDockerLoadAction(logger, "/images/nginx.tar"),
			docker.NewDockerLoadAction(logger, "/images/redis.tar"),
			docker.NewDockerLoadAction(logger, "/images/postgres.tar"),
			docker.NewDockerLoadAction(logger, "/images/node.tar"),
		},
		Logger: logger,
	}
}

// ExampleDockerLoadBatchOperations demonstrates batch loading of multiple images
func ExampleDockerLoadBatchOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerLoadBatchTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker load batch operations", "error", err)
		return
	}

	logger.Info("Docker load batch operations completed successfully")
}

// NewDockerLoadPlatformSpecificTask creates a task that demonstrates platform-specific loading
func NewDockerLoadPlatformSpecificTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-load-platform-example",
		Name: "Docker Load Platform-Specific Operations Example",
		Actions: []task_engine.ActionWrapper{
			// Load AMD64 images
			docker.NewDockerLoadAction(logger, "/images/amd64/nginx.tar",
				docker.WithPlatform("linux/amd64")),
			docker.NewDockerLoadAction(logger, "/images/amd64/redis.tar",
				docker.WithPlatform("linux/amd64")),

			// Load ARM64 images
			docker.NewDockerLoadAction(logger, "/images/arm64/nginx.tar",
				docker.WithPlatform("linux/arm64")),
			docker.NewDockerLoadAction(logger, "/images/arm64/redis.tar",
				docker.WithPlatform("linux/arm64")),
		},
		Logger: logger,
	}
}

// ExampleDockerLoadPlatformSpecificOperations demonstrates platform-specific loading
func ExampleDockerLoadPlatformSpecificOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerLoadPlatformSpecificTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker load platform operations", "error", err)
		return
	}

	logger.Info("Docker load platform operations completed successfully")
}
