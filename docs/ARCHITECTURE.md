# Architecture Overview

## Core Components

### Task

A `Task` is a collection of `Action`s that execute sequentially. Tasks manage execution flow, error handling, parameter validation, and output storage.

```go
type Task struct {
    ID             string
    RunID          string   // auto-generated per execution
    Name           string
    Actions        []ActionWrapper
    Logger         *slog.Logger
    TotalTime      time.Duration
    CompletedTasks int
    // Optional: build a structured result at the end of execution
    ResultBuilder func(ctx *TaskContext) (interface{}, error)
}
```

### TaskContext

A `TaskContext` is passed to a task's `ResultBuilder` function. It provides the task ID, access to the shared `GlobalContext`, and a logger.

```go
type TaskContext struct {
    TaskID        string
    GlobalContext *GlobalContext
    Logger        *slog.Logger
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

`ActionWrapper` is the interface tasks use to run actions, access metadata, and retrieve output.

```go
type ActionWrapper interface {
    Execute(ctx context.Context) error
    GetID() string
    GetName() string
    GetOutput() interface{}
    GetDuration() time.Duration
    GetLogger() *slog.Logger
    SetID(string)
}
```

### TaskManager

`TaskManager` orchestrates multiple tasks, manages shared `GlobalContext`, and provides task lifecycle control (run, stop, wait).

## Parameter System

### Static Parameters

Fixed values known at task creation time.

```go
engine.StaticParameter{Value: "/path/to/file"}
```

### Action Output Parameters

Reference outputs from previous actions within the same task.

```go
engine.ActionOutput("read-action")
engine.ActionOutputField("read-action", "content")
```

### Task Output Parameters

Reference outputs from other tasks using the global context.

```go
engine.TaskOutput("build-task")
engine.TaskOutputField("build-task", "imageID")
```

### Action Result Parameters

Use rich results from actions that implement `ResultProvider`.

```go
engine.ActionResult("download-artifact")
engine.ActionResultField("download-artifact", "checksum")
```

### Task Result Parameters

Use rich results from tasks that implement `ResultProvider` or define a `ResultBuilder`.

```go
engine.TaskResult("preflight")
engine.TaskResultField("preflight", "UpdateMode")
```

## Action Construction Pattern

All built-in actions use a builder pattern. Constructors return a builder (or the action itself) that accepts parameters via a `WithParameters` method, which returns a ready-to-use `*Action[T]`.

```go
// 1. Create the action builder
action, err := file.NewWriteFileAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to/file"},
    engine.StaticParameter{Value: []byte("content")},
    true, // overwrite
    nil,
)
if err != nil {
    return err
}

// 2. Optionally set a custom ID (default is generated from the action name)
action.ID = "write-config"

// 3. Add to a task
task.Actions = append(task.Actions, action)
```

## Common Package (`actions/common`)

The `common` package provides three reusable building blocks to eliminate boilerplate in custom actions:

| Type | Purpose |
|------|---------|
| `BaseConstructor[T]` | Provides `WrapAction()` to produce an `*Action[T]` without repeating struct literal setup |
| `ParameterResolver` | Embedded in action structs; handles `GlobalContext` extraction and typed parameter resolution |
| `OutputBuilder` | Embedded in action structs; provides consistent `map[string]interface{}` output construction |

Embed them in your action struct and initialize in the constructor:

```go
type MyAction struct {
    task_engine.BaseAction
    common.ParameterResolver
    common.OutputBuilder
    // your fields
}

func NewMyAction(logger *slog.Logger) *MyAction {
    return &MyAction{
        BaseAction:        task_engine.NewBaseAction(logger),
        ParameterResolver: *common.NewParameterResolver(logger),
        OutputBuilder:     *common.NewOutputBuilder(logger),
    }
}
```

## Execution Flow

1. **Task Creation**: Actions are created via builders and added to a task's `Actions` slice as `ActionWrapper` values
2. **Parameter Validation**: Before execution begins, `validateParameters()` checks for empty or duplicate action IDs
3. **Action Execution**: Each action runs `BeforeExecute → Execute → AfterExecute` hooks with a context that embeds the `GlobalContext`
4. **Output Storage**: After each action, its output and any `ResultProvider` result are stored in the `GlobalContext`
5. **Result Building**: After all actions complete, the optional `ResultBuilder` is called with a `TaskContext`
6. **Task Output Storage**: Task output and result are stored in the `GlobalContext` for cross-task reference
7. **Error Handling**: Tasks stop on first error; `ErrPrerequisiteNotMet` signals a graceful abort

## Context Management

The `GlobalContext` maintains:

- `ActionOutputs`: Results from completed actions (keyed by action ID)
- `ActionResults`: Rich results from actions implementing `ResultProvider`
- `TaskOutputs`: Results from completed tasks (keyed by task ID), auto-populated with execution metadata
- `TaskResults`: Rich results from tasks implementing `ResultProvider` (or using `ResultBuilder`)

Context is shared across tasks via the `TaskManager` and embedded into each action's execution context under `GlobalContextKey`.

## Error Handling

- **Prerequisites**: Return `ErrPrerequisiteNotMet` to gracefully abort tasks
- **Execution Errors**: Stop task execution and return error details
- **Context Cancellation**: Respect context cancellation for timeouts and graceful shutdown
- **Duplicate IDs**: Detected at validation time before execution begins

## Testing Support

- **Mocks**: Complete mock implementations for all interfaces (`testing/mocks/`)
- **Testable Manager**: Enhanced `TaskManager` with testing hooks (`testing/`)
- **Performance Testing**: Built-in benchmarking and load testing utilities (`testing/`)
