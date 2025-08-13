package tasks

import (
	"context"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
)

// NewDockerStatusTask creates a task that demonstrates docker status operations
func NewDockerStatusTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "container-state-example",
		Name: "Container State Operations Example",
		Actions: []task_engine.ActionWrapper{
			// Example 1: Get state of all containers (empty param -> all)
			func() task_engine.ActionWrapper {
				action, err := docker.NewGetContainerStateAction(logger).WithParameters(task_engine.StaticParameter{Value: ""})
				if err != nil {
					logger.Error("Failed to create docker status action", "error", err)
					return nil
				}
				return action
			}(),
			// Example 2: Get state of specific containers by name (run twice)
			func() task_engine.ActionWrapper {
				action, err := docker.NewGetContainerStateAction(logger).WithParameters(task_engine.StaticParameter{Value: "nginx"})
				if err != nil {
					logger.Error("Failed to create docker status action", "error", err)
					return nil
				}
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := docker.NewGetContainerStateAction(logger).WithParameters(task_engine.StaticParameter{Value: "redis"})
				if err != nil {
					logger.Error("Failed to create docker status action", "error", err)
					return nil
				}
				return action
			}(),
			// Example 3: Get state of a single container
			func() task_engine.ActionWrapper {
				action, err := docker.NewGetContainerStateAction(logger).WithParameters(task_engine.StaticParameter{Value: "my-app"})
				if err != nil {
					logger.Error("Failed to create docker status action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleDockerStatusOperations demonstrates how to use the docker status action
func ExampleDockerStatusOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerStatusTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker status operations", "error", err)
		return
	}

	// Note: In a real scenario, you would access the results from the actions
	// after they have been executed. The actions store their results in their
	// respective fields (e.g., ContainerStates slice).
	logger.Info("Docker status operations completed successfully")
}

// NewDockerStatusFilteringTask creates a task that demonstrates filtering containers by status
func NewDockerStatusFilteringTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-status-filtering-example",
		Name: "Docker Status Filtering Example",
		Actions: []task_engine.ActionWrapper{
			// Get all containers
			func() task_engine.ActionWrapper {
				action, err := docker.NewGetContainerStateAction(logger).WithParameters(task_engine.StaticParameter{Value: ""})
				if err != nil {
					logger.Error("Failed to create docker status action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleDockerStatusWithFiltering demonstrates filtering containers by status
func ExampleDockerStatusWithFiltering() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerStatusFilteringTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker status operations", "error", err)
		return
	}

	logger.Info("Docker status filtering operations completed successfully")
}

// NewDockerStatusMonitoringTask creates a task that demonstrates monitoring container health
func NewDockerStatusMonitoringTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "docker-status-monitoring-example",
		Name: "Docker Status Monitoring Example",
		Actions: []task_engine.ActionWrapper{
			// Get state of critical containers
			func() task_engine.ActionWrapper {
				action, err := docker.NewGetContainerStateAction(logger).WithParameters(task_engine.StaticParameter{Value: "web-server"})
				if err != nil {
					logger.Error("Failed to create docker status action", "error", err)
					return nil
				}
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := docker.NewGetContainerStateAction(logger).WithParameters(task_engine.StaticParameter{Value: "database"})
				if err != nil {
					logger.Error("Failed to create docker status action", "error", err)
					return nil
				}
				return action
			}(),
			func() task_engine.ActionWrapper {
				action, err := docker.NewGetContainerStateAction(logger).WithParameters(task_engine.StaticParameter{Value: "cache"})
				if err != nil {
					logger.Error("Failed to create docker status action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleDockerStatusMonitoring demonstrates monitoring container health
func ExampleDockerStatusMonitoring() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewDockerStatusMonitoringTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute docker monitoring", "error", err)
		return
	}

	logger.Info("Docker status monitoring operations completed successfully")
}
