package tasks

import (
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
)

// ExampleSymlinkOperations demonstrates various symlink creation scenarios
func ExampleSymlinkOperations() {
	logger := slog.Default()

	// Create a task manager
	taskManager := task_engine.NewTaskManager(logger)

	// Example 1: Create a simple symlink to a file
	createFileAction, err := file.NewWriteFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/source.txt"},
		task_engine.StaticParameter{Value: []byte("Hello, World!")},
		false,
		nil,
	)
	if err != nil {
		logger.Error("Failed to create write file action", "error", err)
		return
	}

	createSymlinkAction, err := file.NewCreateSymlinkAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/source.txt"},
		task_engine.StaticParameter{Value: "/tmp/link.txt"},
		false,
		false,
	)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}
	// builder returns action only; errors occur at execution

	// Example 2: Create a symlink to a directory
	createDirAction, err := file.NewCreateDirectoriesAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/source_dir"},
		task_engine.StaticParameter{Value: []string{"subdir1", "subdir2"}},
	)
	if err != nil {
		logger.Error("Failed to create directories action", "error", err)
		return
	}

	createDirSymlinkAction, err := file.NewCreateSymlinkAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/source_dir"},
		task_engine.StaticParameter{Value: "/tmp/dir_link"},
		false,
		false,
	)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}
	// builder returns action only; errors occur at execution

	// Example 3: Create a symlink with overwrite
	createOverwriteSymlinkAction, err := file.NewCreateSymlinkAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/source.txt"},
		task_engine.StaticParameter{Value: "/tmp/overwrite_link.txt"},
		true,
		false,
	)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}
	// builder returns action only; errors occur at execution

	// Example 4: Create a symlink with directory creation
	createSymlinkWithDirsAction, err := file.NewCreateSymlinkAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/source.txt"},
		task_engine.StaticParameter{Value: "/tmp/nested/dirs/link.txt"},
		false,
		true,
	)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}
	// builder returns action only; errors occur at execution

	// Example 5: Create a relative symlink
	createRelativeSymlinkAction, err := file.NewCreateSymlinkAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "source.txt"},
		task_engine.StaticParameter{Value: "/tmp/relative_link.txt"},
		false,
		false,
	)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}
	// builder returns action only; errors occur at execution

	// Create a task
	task := &task_engine.Task{
		ID:      "symlink-examples",
		Name:    "Symlink Examples",
		Logger:  logger,
		Actions: []task_engine.ActionWrapper{},
	}

	// Add all actions to the task
	task.Actions = append(task.Actions, createFileAction)
	task.Actions = append(task.Actions, createSymlinkAction)
	task.Actions = append(task.Actions, createDirAction)
	task.Actions = append(task.Actions, createDirSymlinkAction)
	task.Actions = append(task.Actions, createOverwriteSymlinkAction)
	task.Actions = append(task.Actions, createSymlinkWithDirsAction)
	task.Actions = append(task.Actions, createRelativeSymlinkAction)

	// Add the task to the task manager
	if err := taskManager.AddTask(task); err != nil {
		logger.Error("Failed to add task", "error", err)
		return
	}

	// Execute the task
	if err := taskManager.RunTask("symlink-examples"); err != nil {
		logger.Error("Failed to run symlink examples task", "error", err)
		return
	}

	logger.Info("Successfully started symlink examples task")
}

// ExampleSymlinkErrorHandling demonstrates error handling scenarios
func ExampleSymlinkErrorHandling() {
	logger := slog.Default()

	// Example 1: Try to create symlink with empty target (should fail)
	_, err := file.NewCreateSymlinkAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: "/tmp/link.txt"},
		false,
		false,
	)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}
	// execution-time error expected, not at build time

	// Example 2: Try to create symlink with empty link path (should fail)
	_, err = file.NewCreateSymlinkAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/source.txt"},
		task_engine.StaticParameter{Value: ""},
		false,
		false,
	)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}
	// execution-time error expected, not at build time

	// Example 3: Try to create symlink with same target and link (should fail)
	_, err = file.NewCreateSymlinkAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/source.txt"},
		task_engine.StaticParameter{Value: "/tmp/source.txt"},
		false,
		false,
	)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}
	// execution-time error expected, not at build time

	// Example 4: Try to create symlink without overwrite when target exists
	// First create a file
	_, err = file.NewWriteFileAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "/tmp/existing.txt"},
		task_engine.StaticParameter{Value: []byte("content")},
		false,
		nil,
	)
	if err == nil {
		// Then try to create a symlink at the same location without overwrite
		_, err = file.NewCreateSymlinkAction(logger).WithParameters(
			task_engine.StaticParameter{Value: "/tmp/source.txt"},
			task_engine.StaticParameter{Value: "/tmp/existing.txt"},
			false,
			false,
		)
		if err != nil {
			logger.Error("Failed to create symlink action", "error", err)
			return
		}
		// This will fail during execution if run
	}

	logger.Info("Completed symlink error handling examples")
}
