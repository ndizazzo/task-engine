# API Reference

## Core Types

### Task

```go
type Task struct {
    ID             string
    RunID          string
    Name           string
    Actions        []ActionWrapper
    Logger         *slog.Logger
    TotalTime      time.Duration
    CompletedTasks int
    // Optional builder to produce a structured task result at the end
    ResultBuilder  func(ctx *TaskContext) (interface{}, error)
}

func (t *Task) Run(ctx context.Context) error
func (t *Task) RunWithContext(ctx context.Context, globalContext *GlobalContext) error
func (t *Task) GetID() string
func (t *Task) GetName() string
func (t *Task) GetCompletedTasks() int
func (t *Task) GetTotalTime() time.Duration
// If the task provides results
func (t *Task) SetResult(result interface{})
func (t *Task) GetResult() interface{}
func (t *Task) GetError() error
```

### Action

```go
type Action[T ActionInterface] struct {
    ID      string
    Wrapped T
}

func (a *Action[T]) BeforeExecute(ctx context.Context) error
func (a *Action[T]) Execute(ctx context.Context) error
func (a *Action[T]) AfterExecute(ctx context.Context) error
func (a *Action[T]) GetOutput() interface{}
func (a *Action[T]) GetID() string
```

### ActionWrapper

```go
type ActionWrapper func() ActionInterface
```

### TaskManager

```go
type TaskManager struct {
    // ... internal fields
}

func NewTaskManager(logger *slog.Logger) *TaskManager
func (tm *TaskManager) AddTask(task *Task) error
func (tm *TaskManager) RunTask(ctx context.Context, taskID string) error
func (tm *TaskManager) StopTask(taskID string) error
func (tm *TaskManager) StopAllTasks()
func (tm *TaskManager) GetRunningTasks() []string
func (tm *TaskManager) IsTaskRunning(taskID string) bool
func (tm *TaskManager) GetGlobalContext() *GlobalContext
func (tm *TaskManager) ResetGlobalContext()
```

### GlobalContext

```go
type GlobalContext struct {
    ActionOutputs map[string]interface{}
    ActionResults map[string]ResultProvider
    TaskOutputs   map[string]interface{}
    TaskResults   map[string]ResultProvider
    mu            sync.RWMutex
}

func NewGlobalContext() *GlobalContext
func (gc *GlobalContext) StoreActionOutput(actionID string, output interface{})
func (gc *GlobalContext) StoreActionResult(actionID string, resultProvider ResultProvider)
func (gc *GlobalContext) StoreTaskOutput(taskID string, output interface{})
func (gc *GlobalContext) StoreTaskResult(taskID string, resultProvider ResultProvider)
```

## Parameter Types

### StaticParameter

```go
type StaticParameter struct {
    Value interface{}
}

func (p StaticParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error)
```

### ActionOutputParameter

```go
type ActionOutputParameter struct {
    ActionID  string
    OutputKey string
}

func (p ActionOutputParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error)
```

### TaskOutputParameter

```go
type TaskOutputParameter struct {
    TaskID    string
    OutputKey string
}

func (p TaskOutputParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error)
```

### ActionResultParameter

```go
type ActionResultParameter struct {
    ActionID  string
    ResultKey string
}

func (p ActionResultParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error)
```

## Helper Functions

### ActionOutput

```go
func ActionOutput(actionID string) ActionOutputParameter
func ActionOutputField(actionID, field string) ActionOutputParameter
```

### TaskOutput

```go
func TaskOutput(taskID string) TaskOutputParameter
func TaskOutputField(taskID, field string) TaskOutputParameter
```

### ActionResult

````go
func ActionResult(actionID string) ActionResultParameter
func ActionResultField(actionID, field string) ActionResultParameter

### TaskResult

```go
func TaskResult(taskID string) TaskResultParameter
func TaskResultField(taskID, field string) TaskResultParameter
```

### Simple Typed Helpers (Recommended)

```go
func ActionResultAs[T any](gc *GlobalContext, actionID string) (T, bool)
func TaskResultAs[T any](gc *GlobalContext, taskID string) (T, bool)
func ActionOutputFieldAs[T any](gc *GlobalContext, actionID, key string) (T, error)
func TaskOutputFieldAs[T any](gc *GlobalContext, taskID, key string) (T, error)
func EntityValue(gc *GlobalContext, entityType, id, key string) (interface{}, error)
func EntityValueAs[T any](gc *GlobalContext, entityType, id, key string) (T, error)

// Generic parameter resolver
func ResolveAs[T any](ctx context.Context, p ActionParameter, gc *GlobalContext) (T, error)
```

### EntityOutput

```go
func EntityOutput(entityType, entityID string) EntityOutputParameter
func EntityOutputField(entityType, entityID, field string) EntityOutputParameter
```

````

## Interfaces

### ActionInterface

```go
type ActionInterface interface {
    BeforeExecute(ctx context.Context) error
    Execute(ctx context.Context) error
    AfterExecute(ctx context.Context) error
    GetOutput() interface{}
}
```

### TaskInterface

````go
type TaskInterface interface {
    GetID() string
    GetName() string
    Run(ctx context.Context) error
    RunWithContext(ctx context.Context, globalContext *GlobalContext) error
    GetCompletedTasks() int
    GetTotalTime() time.Duration
}

### TaskWithResults

```go
type TaskWithResults interface {
    TaskInterface
    ResultProvider
}
````

````

### TaskManagerInterface

```go
type TaskManagerInterface interface {
    AddTask(task *Task) error
    RunTask(taskID string) error
    StopTask(taskID string) error
    StopAllTasks()
    GetRunningTasks() []string
    IsTaskRunning(taskID string) bool
    GetGlobalContext() *GlobalContext
    ResetGlobalContext()
}
````

### ResultProvider

```go
type ResultProvider interface {
    GetResult() interface{}
    GetError() error
}
```

## Constants

```go
const GlobalContextKey contextKey = "globalContext"
```

## Errors

```go
var ErrPrerequisiteNotMet = errors.New("task prerequisite not met")
```

## Context Keys

```go
type contextKey string
const GlobalContextKey contextKey = "globalContext"
```
