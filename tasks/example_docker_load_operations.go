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
			func() task_engine.ActionWrapper {
				action := docker.NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: "/path/to/nginx.tar"})
				return action
			}(),

			// Example 2: Load with platform specification
			func() task_engine.ActionWrapper {
				action := docker.NewDockerLoadAction(logger).WithOptions(docker.WithPlatform("linux/amd64")).WithParameters(task_engine.StaticParameter{Value: "/path/to/multi-platform.tar"})
				return action
			}(),

			// Example 3: Load with quiet mode
			func() task_engine.ActionWrapper {
				action := docker.NewDockerLoadAction(logger).WithOptions(docker.WithQuiet()).WithParameters(task_engine.StaticParameter{Value: "/path/to/redis.tar"})
				return action
			}(),

			// Example 4: Load with both platform and quiet options
			func() task_engine.ActionWrapper {
				action := docker.NewDockerLoadAction(logger).WithOptions(
					docker.WithPlatform("linux/arm64"),
					docker.WithQuiet(),
				).WithParameters(task_engine.StaticParameter{Value: "/path/to/postgres.tar"})
				return action
			}(),
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
			docker.NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: "/images/nginx.tar"}),
			docker.NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: "/images/redis.tar"}),
			docker.NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: "/images/postgres.tar"}),
			docker.NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: "/images/node.tar"}),
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
			docker.NewDockerLoadAction(logger).WithOptions(docker.WithPlatform("linux/amd64")).WithParameters(task_engine.StaticParameter{Value: "/images/amd64/nginx.tar"}),
			docker.NewDockerLoadAction(logger).WithOptions(docker.WithPlatform("linux/amd64")).WithParameters(task_engine.StaticParameter{Value: "/images/amd64/redis.tar"}),

			// Load ARM64 images
			docker.NewDockerLoadAction(logger).WithOptions(docker.WithPlatform("linux/arm64")).WithParameters(task_engine.StaticParameter{Value: "/images/arm64/nginx.tar"}),
			docker.NewDockerLoadAction(logger).WithOptions(docker.WithPlatform("linux/arm64")).WithParameters(task_engine.StaticParameter{Value: "/images/arm64/redis.tar"}),
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
