# Task Engine

A powerful, type-safe Go task execution framework with built-in actions for file operations, Docker management, system administration, and more.

## Installation

Add the module to your Go project:

```bash
go get github.com/ndizazzo/task-engine
```

## Quick Start

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create a simple task
    task := &task_engine.Task{
        ID:   "setup-project",
        Name: "Project Setup",
        Actions: []task_engine.ActionWrapper{
            actions.NewCreateDirectoriesAction(
                "/tmp/myproject",
                []string{"src", "docs", "tests"},
                logger,
            ),
            actions.NewWriteFileAction(
                "/tmp/myproject/README.md",
                []byte("# My Project\n\nWelcome to my project!"),
                true, // overwrite if exists
                nil,  // no input buffer
                logger,
            ),
        },
        Logger: logger,
    }

    // Execute the task
    if err := task.Run(context.Background()); err != nil {
        logger.Error("Task failed", "error", err)
        os.Exit(1)
    }

    logger.Info("Task completed successfully!")
}
```

## Features

- **Type-Safe Actions**: Generic-based architecture with compile-time safety
- **19+ Built-in Actions**: File operations, Docker management, system commands
- **Lifecycle Management**: Before/Execute/After hooks with proper error handling
- **Context Support**: Cancellation and timeout handling
- **Comprehensive Logging**: Structured logging with configurable output
- **Easy Testing**: Built-in mocking support for all actions

## Built-in Actions

The Task Engine includes 19+ built-in actions covering file operations, Docker management, system administration, and utilities. For a complete inventory with detailed documentation, parameters, and examples, see [ACTIONS.md](ACTIONS.md).

### Docker Environment Setup

```go
import "github.com/ndizazzo/task-engine/tasks"

task := tasks.NewDockerSetupTask(logger, "/path/to/project")
```

Sets up a complete Docker environment with MySQL, Redis, Nginx, and monitoring.

### File Operations Workflow

```go
task := tasks.NewFileOperationsTask(logger, "/path/to/project")
```

Demonstrates file creation, copying, text replacement, and cleanup.

## Creating Custom Actions

```go
type MyAction struct {
    task_engine.BaseAction
    Message string
}

func (a *MyAction) Execute(ctx context.Context) error {
    a.Logger.Info("Executing custom action", "message", a.Message)
    return nil
}

func NewMyAction(message string, logger *slog.Logger) *task_engine.Action[*MyAction] {
    return &task_engine.Action[*MyAction]{
        ID: "my-custom-action",
        Wrapped: &MyAction{
            BaseAction: task_engine.BaseAction{Logger: logger},
            Message:    message,
        },
    }
}
```

## Task Management

Use the `TaskManager` for advanced task orchestration:

```go
manager := task_engine.NewTaskManager(logger)

// Add and run tasks
taskID := manager.AddTask(task)
err := manager.RunTask(context.Background(), taskID)

// Stop running tasks
manager.StopTask(taskID)
manager.StopAllTasks()
```

## Error Handling

Tasks automatically stop on first error and provide detailed error information:

```go
if err := task.Run(ctx); err != nil {
    if errors.Is(err, task_engine.ErrPrerequisiteNotMet) {
        logger.Warn("Prerequisites not met, skipping task")
    } else {
        logger.Error("Task execution failed", "error", err)
    }
}
```

## Testing

The module provides comprehensive mocking support:

```go
import "github.com/ndizazzo/task-engine/mocks"

// Create mock command runner for testing
mockRunner := &mocks.MockCommandRunner{}
mockRunner.On("RunCommand", "echo", "hello").Return("hello", nil)

// Use discard logger for tests
logger := mocks.NewDiscardLogger()
```

## License

This project is available under the MIT License.
