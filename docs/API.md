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
}

func (t *Task) Run(ctx context.Context) error
func (t *Task) RunWithContext(ctx context.Context, globalContext *GlobalContext) error
func (t *Task) GetID() string
func (t *Task) GetName() string
func (t *Task) GetCompletedTasks() int
func (t *Task) GetTotalTime() time.Duration
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
    TaskOutputs   map[string]interface{}
    ActionResults map[string]ResultProvider
    mu            sync.RWMutex
}

func NewGlobalContext() *GlobalContext
func (gc *GlobalContext) SetActionOutput(actionID string, output interface{})
func (gc *GlobalContext) SetTaskOutput(taskID string, output interface{})
func (gc *GlobalContext) GetActionOutput(actionID string) (interface{}, bool)
func (gc *GlobalContext) GetTaskOutput(taskID string) (interface{}, bool)
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
func ActionOutput(actionID string, outputKey string) ActionOutputParameter
```

### TaskOutput

```go
func TaskOutput(taskID string, outputKey string) TaskOutputParameter
```

### ActionResult

```go
func ActionResult(actionID string, resultKey string) ActionResultParameter
```

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

```go
type TaskInterface interface {
    GetID() string
    GetName() string
    Run(ctx context.Context) error
    RunWithContext(ctx context.Context, globalContext *GlobalContext) error
    GetCompletedTasks() int
    GetTotalTime() time.Duration
}
```

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
```

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
