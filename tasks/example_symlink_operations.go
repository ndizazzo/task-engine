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
	createFileAction, err := file.NewWriteFileAction("/tmp/source.txt", []byte("Hello, World!"), false, nil, logger)
	if err != nil {
		logger.Error("Failed to create write file action", "error", err)
		return
	}

	createSymlinkAction, err := file.NewCreateSymlinkAction("/tmp/source.txt", "/tmp/link.txt", false, false, logger)
	if err != nil {
		logger.Error("Failed to create symlink action", "error", err)
		return
	}

	// Example 2: Create a symlink to a directory
	createDirAction, err := file.NewCreateDirectoriesAction(logger, "/tmp/source_dir", []string{"subdir1", "subdir2"})
	if err != nil {
		logger.Error("Failed to create directories action", "error", err)
		return
	}

	createDirSymlinkAction, err := file.NewCreateSymlinkAction("/tmp/source_dir", "/tmp/dir_link", false, false, logger)
	if err != nil {
		logger.Error("Failed to create directory symlink action", "error", err)
		return
	}

	// Example 3: Create a symlink with overwrite
	createOverwriteSymlinkAction, err := file.NewCreateSymlinkAction("/tmp/source.txt", "/tmp/overwrite_link.txt", true, false, logger)
	if err != nil {
		logger.Error("Failed to create overwrite symlink action", "error", err)
		return
	}

	// Example 4: Create a symlink with directory creation
	createSymlinkWithDirsAction, err := file.NewCreateSymlinkAction("/tmp/source.txt", "/tmp/nested/dirs/link.txt", false, true, logger)
	if err != nil {
		logger.Error("Failed to create symlink with dirs action", "error", err)
		return
	}

	// Example 5: Create a relative symlink
	createRelativeSymlinkAction, err := file.NewCreateSymlinkAction("source.txt", "/tmp/relative_link.txt", false, false, logger)
	if err != nil {
		logger.Error("Failed to create relative symlink action", "error", err)
		return
	}

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
	_, err := file.NewCreateSymlinkAction("", "/tmp/link.txt", false, false, logger)
	if err != nil {
		logger.Info("Expected error for empty target", "error", err)
	}

	// Example 2: Try to create symlink with empty link path (should fail)
	_, err = file.NewCreateSymlinkAction("/tmp/source.txt", "", false, false, logger)
	if err != nil {
		logger.Info("Expected error for empty link path", "error", err)
	}

	// Example 3: Try to create symlink with same target and link (should fail)
	_, err = file.NewCreateSymlinkAction("/tmp/source.txt", "/tmp/source.txt", false, false, logger)
	if err != nil {
		logger.Info("Expected error for same target and link", "error", err)
	}

	// Example 4: Try to create symlink without overwrite when target exists
	// First create a file
	_, err = file.NewWriteFileAction("/tmp/existing.txt", []byte("content"), false, nil, logger)
	if err == nil {
		// Then try to create a symlink at the same location without overwrite
		_, err = file.NewCreateSymlinkAction("/tmp/source.txt", "/tmp/existing.txt", false, false, logger)
		if err == nil {
			// This should fail during execution, not during creation
			logger.Info("Created action that will fail during execution")
		}
	}

	logger.Info("Completed symlink error handling examples")
}
