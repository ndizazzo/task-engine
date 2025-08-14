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
            // Action 1: Create directory
            func() task_engine.ActionInterface {
                action, _ := file.NewCreateDirectoriesAction(logger).WithParameters(
                    task_engine.StaticParameter{Value: "/tmp/myproject"},
                    task_engine.StaticParameter{Value: []string{"src", "docs"}},
                )
                return action
            },

            // Action 2: Write file
            func() task_engine.ActionInterface {
                action, _ := file.NewWriteFileAction(logger).WithParameters(
                    task_engine.StaticParameter{Value: "/tmp/myproject/README.md"},
                    task_engine.StaticParameter{Value: []byte("# My Project\n\nCreated with Task Engine!")},
                    true, // overwrite
                    nil,  // inputBuffer
                )
                return action
            },
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
        // Read file
        func() task_engine.ActionInterface {
            var content []byte
            action, _ := file.NewReadFileAction("/tmp/input.txt", &content, logger)
            action.ID = "read-file"
            return action
        },

        // Process content (using output from read action)
        func() task_engine.ActionInterface {
            action := file.NewReplaceLinesAction(logger).WithParameters(
                task_engine.StaticParameter{Value: "/tmp/output.txt"},
                map[*regexp.Regexp]task_engine.ActionParameter{
                    regexp.MustCompile("old"): task_engine.ActionOutputField("read-file", "content"),
                },
            )
            return action
        },
    },
    Logger: logger,
}
```

## Task Manager

Manage multiple tasks with shared context:

```go
manager := task_engine.NewTaskManager(logger)

// Add tasks
task1ID := manager.AddTask(fileTask)
task2ID := manager.AddTask(dockerTask)

// Run tasks
if err := manager.RunTask(context.Background(), task1ID); err != nil {
    logger.Error("Task 1 failed", "error", err)
}

if err := manager.RunTask(context.Background(), task2ID); err != nil {
    logger.Error("Task 2 failed", "error", err)
}

// Stop tasks
manager.StopTask(task1ID)
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
task.Actions = append(task.Actions, func() task_engine.ActionInterface {
    return greetingAction
})
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
