# Task Engine

[![CI/CD Pipeline](https://github.com/ndizazzo/task-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/ndizazzo/task-engine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ndizazzo/task-engine/graph/badge.svg?token=O020V4C6TV)](https://codecov.io/gh/ndizazzo/task-engine)

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
    "github.com/ndizazzo/task-engine/tasks"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create a simple task using the file operations example
    task := tasks.NewFileOperationsTask(logger, "/tmp/myproject")

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
- **24+ Built-in Actions**: File operations, Docker management, system commands, package management
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

Sets up a complete Docker environment with service management and health checks.

### File Operations Workflow

```go
task := tasks.NewFileOperationsTask(logger, "/path/to/project")
```

Demonstrates file creation, copying, text replacement, and cleanup.

### Package Management

```go
task := tasks.NewPackageManagementTask(logger, []string{"git", "curl", "wget", "htop"})
```

Cross-platform package installation supporting Debian-based Linux (apt) and macOS (Homebrew).

### Compression Operations

```go
task := tasks.NewCompressionOperationsTask(logger, "/path/to/project")
```

Shows file compression and decompression workflows with auto-detection.

### System Management

```go
task := tasks.NewSystemManagementTask(logger, "nginx")
```

Demonstrates system service management and administrative operations.

### Utility Operations

```go
task := tasks.NewUtilityOperationsTask(logger)
```

Shows utility operations including timing, prerequisites, and system information.

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
