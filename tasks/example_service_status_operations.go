package tasks

import (
	"context"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/system"
)

// NewServiceStatusTask creates a task that demonstrates service status operations
func NewServiceStatusTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "service-status-example",
		Name: "Service Status Operations Example",
		Actions: []task_engine.ActionWrapper{
			func() task_engine.ActionWrapper {
				action, err := system.NewServiceStatusAction(logger).WithParameters(
					task_engine.StaticParameter{Value: []string{"sshd.service", "docker.service", "nginx.service"}},
				)
				if err != nil {
					logger.Error("Failed to create service status action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleServiceStatusOperations demonstrates how to use the service status action
func ExampleServiceStatusOperations() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewServiceStatusTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute service status operations", "error", err)
		return
	}

	// Note: In a real scenario, you would access the results from the actions
	// after they have been executed. The actions store their results in their
	// respective fields (e.g., ServiceStatuses slice).
	logger.Info("Service status operations completed successfully")
}

// NewServiceHealthCheckTask creates a task that demonstrates service health checking
func NewServiceHealthCheckTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "service-health-check-example",
		Name: "Service Health Check Example",
		Actions: []task_engine.ActionWrapper{
			func() task_engine.ActionWrapper {
				action, err := system.NewServiceStatusAction(logger).WithParameters(
					task_engine.StaticParameter{Value: []string{"sshd.service", "systemd-networkd.service", "systemd-resolved.service"}},
				)
				if err != nil {
					logger.Error("Failed to create service status action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleServiceHealthCheck demonstrates checking service health
func ExampleServiceHealthCheck() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewServiceHealthCheckTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute service health check", "error", err)
		return
	}

	logger.Info("Service health check completed successfully")
}

// NewServiceMonitoringTask creates a task that demonstrates service monitoring
func NewServiceMonitoringTask(logger *slog.Logger) *task_engine.Task {
	return &task_engine.Task{
		ID:   "service-monitoring-example",
		Name: "Service Monitoring Example",
		Actions: []task_engine.ActionWrapper{
			func() task_engine.ActionWrapper {
				action, err := system.NewServiceStatusAction(logger).WithParameters(
					task_engine.StaticParameter{Value: []string{"docker.service", "kubelet.service", "etcd.service"}},
				)
				if err != nil {
					logger.Error("Failed to create service status action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}
}

// ExampleServiceMonitoring demonstrates service monitoring
func ExampleServiceMonitoring() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a task
	task := NewServiceMonitoringTask(logger)

	// Execute the task
	ctx := context.Background()
	err := task.Run(ctx)
	if err != nil {
		logger.Error("Failed to execute service monitoring", "error", err)
		return
	}

	logger.Info("Service monitoring completed successfully")
}
