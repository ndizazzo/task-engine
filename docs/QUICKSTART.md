# Quick Start Guide

Get up and running with Task Engine in minutes.

## Installation

```bash
go get github.com/ndizazzo/task-engine
```

## Basic Task

Create a simple task that creates a directory and writes a file:

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions/file"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create task with two actions
    task := &task_engine.Task{
        ID:   "my-first-task",
        Name: "Create Project Structure",
        Actions: []task_engine.ActionWrapper{
            file.NewCreateDirectoriesAction([]string{"src", "docs"}, logger),
            file.NewWriteFileAction(
                "/tmp/myproject/README.md",
                []byte("# My Project\n\nCreated with Task Engine!"),
                true,
                nil,
                logger,
            ),
        },
        Logger: logger,
    }

    // Run the task
    if err := task.Run(context.Background()); err != nil {
        logger.Error("Task failed", "error", err)
        os.Exit(1)
    }

    logger.Info("Task completed successfully!")
}
```

## Using Built-in Examples

Task Engine provides ready-to-use examples:

```go
import "github.com/ndizazzo/task-engine/tasks"

// File operations workflow
fileTask := tasks.NewFileOperationsTask(logger, "/tmp/project")

// Docker setup
dockerTask := tasks.NewDockerSetupTask(logger, "/path/to/compose")

// Package installation
packageTask := tasks.NewPackageManagementTask(logger, []string{"git", "curl"})

// Run any task
if err := fileTask.Run(context.Background()); err != nil {
    logger.Error("Task failed", "error", err)
}
```

## Parameter Passing

Pass data between actions:

```go
    task := &task_engine.Task{
        ID:   "file-pipeline",
        Name: "Process File",
        Actions: []task_engine.ActionWrapper{
            file.NewReadFileAction("read-file", "/tmp/input.txt", logger),
            file.NewReplaceLinesAction(
                "replace-lines",
                "/tmp/output.txt",
                map[*regexp.Regexp]task_engine.ActionParameter{
                    regexp.MustCompile("old"): task_engine.ActionOutputField("read-file", "content"),
                },
                logger,
            ),
        },
        Logger: logger,
    }
```

## Task Manager

Manage multiple tasks with shared context:

```go
manager := task_engine.NewTaskManager(logger)

// Add tasks
if err := manager.AddTask(fileTask); err != nil {
    logger.Error("Failed to add task", "error", err)
}
if err := manager.AddTask(dockerTask); err != nil {
    logger.Error("Failed to add task", "error", err)
}

// Run tasks — returns a TaskHandle for async tracking
handle1, err := manager.RunTask("file-operations")
if err != nil {
    logger.Error("Task 1 failed to start", "error", err)
}
<-handle1.Done()

handle2, err := manager.RunTask("docker-setup")
if err != nil {
    logger.Error("Task 2 failed to start", "error", err)
}
<-handle2.Done()

// Stop tasks
manager.StopAllTasks()
```

## Custom Actions

Create your own actions:

```go
type GreetingAction struct {
    task_engine.BaseAction
    Name string
}

func (a *GreetingAction) Execute(ctx context.Context) error {
    a.Logger.Info("Hello", "name", a.Name)
    return nil
}

func NewGreetingAction(name string, logger *slog.Logger) *task_engine.Action[*GreetingAction] {
    return &task_engine.Action[*GreetingAction]{
        ID: "greeting",
        Wrapped: &GreetingAction{
            BaseAction: task_engine.BaseAction{Logger: logger},
            Name:       name,
        },
    }
}

// Use in task
greetingAction := NewGreetingAction("World", logger)
task.Actions = append(task.Actions, greetingAction)
```

## Error Handling

Handle errors gracefully:

```go
if err := task.Run(ctx); err != nil {
    if errors.Is(err, task_engine.ErrPrerequisiteNotMet) {
        logger.Warn("Prerequisites not met, skipping task")
        return nil
    }
    logger.Error("Task execution failed", "error", err)
    return err
}
```

## Testing

Test your tasks with built-in utilities:

```go
import "github.com/ndizazzo/task-engine/testing/mocks"

func TestMyTask(t *testing.T) {
    // Create mock logger
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))

    // Create mock task manager
    mockManager := mocks.NewEnhancedTaskManagerMock()

    // Test your task
    task := createMyTask(logger)
    err := task.Run(context.Background())

    assert.NoError(t, err)
}
```

## Next Steps

- Explore [built-in actions](ACTIONS.md)
- Check [examples](docs/examples/) for patterns
- Read [architecture overview](docs/ARCHITECTURE.md)
- Review [API reference](docs/API.md)
- See [troubleshooting](docs/TROUBLESHOOTING.md) for common issues
