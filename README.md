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
- **Action Parameter Passing**: Seamless data flow between actions and tasks using declarative parameter references
- **Lifecycle Management**: Before/Execute/After hooks with proper error handling
- **Context Support**: Cancellation and timeout handling
- **Comprehensive Logging**: Structured logging with configurable output
- **Easy Testing**: Built-in mocking support for all actions

## Built-in Actions

The Task Engine includes 19+ built-in actions for file operations, Docker management, system administration, and utilities. See [ACTIONS.md](ACTIONS.md) for complete documentation.

```go
import "github.com/ndizazzo/task-engine/tasks"

// Common task examples
dockerTask := tasks.NewDockerSetupTask(logger, "/path/to/project")
fileTask := tasks.NewFileOperationsTask(logger, "/path/to/project")
packageTask := tasks.NewPackageManagementTask(logger, []string{"git", "curl"})
```

## Action Parameter Passing

Actions can reference outputs from previous actions and tasks using declarative parameters:

```go
// Action-to-action parameter passing
file.NewReplaceLinesAction(
    "input.txt",
    map[*regexp.Regexp]string{
        regexp.MustCompile("{{content}}"):
            task_engine.ActionOutput("read-file", "content"),
    },
    logger,
)

// Cross-task parameter passing
docker.NewDockerRunAction(
    task_engine.TaskOutput("build-app", "imageID"),
    []string{"-p", "8080:8080"},
    logger,
)
```

See [docs/parameter_passing.md](docs/parameter_passing.md) for complete documentation.

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

The module provides comprehensive testing support with mocks and performance testing:

```go
import "github.com/ndizazzo/task-engine/testing/mocks"

// Mock implementations
mockManager := mocks.NewEnhancedTaskManagerMock()
mockRunner := &mocks.MockCommandRunner{}
```

See [testing/README.md](testing/README.md) for complete testing documentation.

## License

This project is available under the MIT License.
