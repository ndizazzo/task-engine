package tasks

import (
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/file"
)

// ExampleReadFileOperations demonstrates how to use the ReadFileAction
func ExampleReadFileOperations() {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a task manager
	taskManager := task_engine.NewTaskManager(logger)

	// Create a task that reads a file
	task := &task_engine.Task{
		ID:   "read-file-example",
		Name: "Read File Example",
		Actions: []task_engine.ActionWrapper{
			// Create a test file first
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(
					"/tmp/test_read_file.txt",
					[]byte("Hello, this is test content for reading!\nLine 2 with some data.\nLine 3 with more content."),
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

	// Add the task to the manager
	if err := taskManager.AddTask(task); err != nil {
		logger.Error("Failed to add task", "error", err)
		return
	}

	// Run the task
	err := taskManager.RunTask("read-file-example")
	if err != nil {
		logger.Error("Failed to run task", "error", err)
		return
	}

	logger.Info("Task completed successfully")
}

// ExampleReadFileWithErrorHandling demonstrates error handling with ReadFileAction
func ExampleReadFileWithErrorHandling() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	taskManager := task_engine.NewTaskManager(logger)

	// Try to read a non-existent file
	var content []byte
	task := &task_engine.Task{
		ID:   "read-file-error-handling",
		Name: "Read File Error Handling",
		Actions: []task_engine.ActionWrapper{
			func() task_engine.ActionWrapper {
				action, err := file.NewReadFileAction("/nonexistent/file.txt", &content, logger)
				if err != nil {
					logger.Error("Failed to create read file action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}

	if err := taskManager.AddTask(task); err != nil {
		logger.Error("Failed to add task", "error", err)
		return
	}

	err := taskManager.RunTask("read-file-error-handling")
	if err != nil {
		logger.Info("Expected error occurred", "error", err)
	} else {
		logger.Info("Unexpected success")
	}
}

// ExampleReadFileInWorkflow demonstrates using ReadFileAction in a complex workflow
func ExampleReadFileInWorkflow() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	taskManager := task_engine.NewTaskManager(logger)

	// Step 1: Create a source file
	sourceFile := "/tmp/source_data.txt"
	sourceContent := []byte("Source data line 1\nSource data line 2\nSource data line 3")

	// Step 2: Read the source file
	var sourceData []byte

	// Step 3: Process the data (in this case, just copy to a new file)
	processedFile := "/tmp/processed_data.txt"

	// Step 4: Read the processed file to verify
	var processedData []byte

	task := &task_engine.Task{
		ID:   "read-file-workflow",
		Name: "Read File Workflow",
		Actions: []task_engine.ActionWrapper{
			// Step 1: Create a source file
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(sourceFile, sourceContent, true, nil, logger)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 2: Read the source file
			func() task_engine.ActionWrapper {
				action, err := file.NewReadFileAction(sourceFile, &sourceData, logger)
				if err != nil {
					logger.Error("Failed to create read file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 3: Process the data (in this case, just copy to a new file)
			func() task_engine.ActionWrapper {
				action, err := file.NewWriteFileAction(processedFile, sourceData, true, nil, logger)
				if err != nil {
					logger.Error("Failed to create write file action", "error", err)
					return nil
				}
				return action
			}(),

			// Step 4: Read the processed file to verify
			func() task_engine.ActionWrapper {
				action, err := file.NewReadFileAction(processedFile, &processedData, logger)
				if err != nil {
					logger.Error("Failed to create read file action", "error", err)
					return nil
				}
				return action
			}(),
		},
		Logger: logger,
	}

	if err := taskManager.AddTask(task); err != nil {
		logger.Error("Failed to add task", "error", err)
		return
	}

	err := taskManager.RunTask("read-file-workflow")
	if err != nil {
		logger.Error("Workflow failed", "error", err)
		return
	}

	logger.Info("Workflow completed successfully",
		"source_size", len(sourceData),
		"processed_size", len(processedData),
		"source_content", string(sourceData),
		"processed_content", string(processedData))
}
