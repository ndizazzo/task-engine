# Architecture Overview

## Core Components

### Task

A `Task` is a collection of `Action`s that execute sequentially. Tasks manage execution flow, error handling, and parameter resolution.

```go
type Task struct {
    ID      string
    Name    string
    Actions []ActionWrapper
    Logger  *slog.Logger
}
```

### Action

An `Action` represents a single operation (file I/O, Docker command, system call). Actions implement the `ActionInterface` with Before/Execute/After lifecycle hooks.

```go
type ActionInterface interface {
    BeforeExecute(ctx context.Context) error
    Execute(ctx context.Context) error
    AfterExecute(ctx context.Context) error
    GetOutput() interface{}
}
```

### ActionWrapper

`ActionWrapper` is a function type that returns an action. This enables lazy initialization and parameter injection.

```go
type ActionWrapper func() ActionInterface
```

### TaskManager

`TaskManager` orchestrates multiple tasks, manages shared context, and provides task lifecycle control.

## Parameter System

### Static Parameters

Fixed values known at task creation time.

```go
engine.StaticParameter{Value: "/path/to/file"}
```

### Action Output Parameters

Reference outputs from previous actions within the same task.

```go
engine.ActionOutput("read-action", "content")
```

### Task Output Parameters

Reference outputs from other tasks using the global context.

```go
engine.TaskOutput("build-task", "imageID")
```

## Execution Flow

1. **Task Creation**: Actions are wrapped in functions for lazy initialization
2. **Parameter Resolution**: Parameters are resolved at execution time using the global context
3. **Action Execution**: Each action runs Before → Execute → After hooks
4. **Output Storage**: Action outputs are stored in the global context for parameter passing
5. **Error Handling**: Tasks stop on first error, with special handling for prerequisites

## Context Management

The `GlobalContext` maintains:

- `ActionOutputs`: Results from completed actions
- `TaskOutputs`: Results from completed tasks
- `ActionResults`: Rich results from actions implementing `ResultProvider`

Context is shared across tasks via the `TaskManager` and embedded in the execution context.

## Error Handling

- **Prerequisites**: Return `ErrPrerequisiteNotMet` to gracefully abort tasks
- **Execution Errors**: Stop task execution and return error details
- **Context Cancellation**: Respect context cancellation for timeouts and graceful shutdown

## Testing Support

- **Mocks**: Complete mock implementations for all interfaces
- **Testable Manager**: Enhanced TaskManager with testing hooks
- **Performance Testing**: Built-in benchmarking and load testing utilities
