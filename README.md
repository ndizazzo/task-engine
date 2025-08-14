# Task Engine

[![CI/CD Pipeline](https://github.com/ndizazzo/task-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/ndizazzo/task-engine/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ndizazzo/task-engine/graph/badge.svg?token=O020V4C6TV)](https://codecov.io/gh/ndizazzo/task-engine/graph/badge.svg?token=O020V4C6TV)

A Go task execution framework with built-in actions for file operations, Docker management, and system administration.

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

    // Create and run a task
    task := tasks.NewFileOperationsTask(logger, "/tmp/myproject")
    if err := task.Run(context.Background()); err != nil {
        logger.Error("Task failed", "error", err)
        os.Exit(1)
    }
}
```

## Core Concepts

- **Tasks**: Collections of actions that execute sequentially
- **Actions**: Individual operations (file, Docker, system)
- **Parameters**: Pass data between actions using `ActionOutput()` and `TaskOutput()`, and fetch rich results using `ActionResult()` and `TaskResult()`
- **Context**: Share data across tasks with `TaskManager`

## Built-in Actions

24+ actions for common operations:

- **File**: Create, read, write, copy, compress, extract
- **Docker**: Run, compose, health checks, image management
- **System**: Service management, package installation
- **Utilities**: Network info, prerequisites, timing

See [ACTIONS.md](ACTIONS.md) for complete list.

## Parameter Passing

Actions can reference outputs from previous actions, results from actions, and results from tasks:

```go
// Reference action output
file.NewReplaceLinesAction(
    "config.txt",
    map[*regexp.Regexp]string{
        regexp.MustCompile("{{version}}"):
            task_engine.ActionOutput("read-file", "version"),
    },
    logger,
)

// Reference task output
docker.NewDockerRunAction(
    task_engine.TaskOutput("build", "imageID"),
    []string{"-p", "8080:8080"},
    logger,
)

// Reference action result (from an action implementing ResultProvider)
useChecksum := task_engine.ActionResultField("download-artifact", "checksum")

// Reference task result (from a task implementing ResultProvider or using ResultBuilder)
preflightMode := task_engine.TaskResultField("preflight", "UpdateMode")
```

## Task Management

```go
manager := task_engine.NewTaskManager(logger)

// Add and run tasks
taskID := manager.AddTask(task)
err := manager.RunTask(context.Background(), taskID)

// Stop tasks
manager.StopTask(taskID)
manager.StopAllTasks()
```

## Custom Actions

```go
type MyAction struct {
    task_engine.BaseAction
    Message string
}

func (a *MyAction) Execute(ctx context.Context) error {
    a.Logger.Info("Executing", "message", a.Message)
    return nil
}

func NewMyAction(message string, logger *slog.Logger) *task_engine.Action[*MyAction] {
    return &task_engine.Action[*MyAction]{
        ID: "my-action",
        Wrapped: &MyAction{
            BaseAction: task_engine.BaseAction{Logger: logger},
            Message:    message,
        },
    }
}
```

## Examples

See `tasks/` directory for complete examples:

- File operations workflow
- Docker setup and management
- Package installation
- Archive extraction

## Testing

```go
import "github.com/ndizazzo/task-engine/testing/mocks"

mockManager := mocks.NewEnhancedTaskManagerMock()
mockRunner := &mocks.MockCommandRunner{}
```

## Documentation

- [Quick Start](docs/QUICKSTART.md) - Get up and running in minutes
- [Architecture](docs/ARCHITECTURE.md) - Core concepts and structure
- [API Reference](docs/API.md) - Complete API documentation
- [Actions](ACTIONS.md) - Built-in actions reference
- [Examples](docs/examples/) - Usage examples and patterns
- [Testing](testing/README.md) - Testing utilities
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and solutions

## License

MIT License
