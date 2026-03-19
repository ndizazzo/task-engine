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
func (t *Task) SetResult(result interface{})
func (t *Task) GetResult() interface{}
func (t *Task) SetError(err error)
func (t *Task) GetError() error
```

### TaskContext

Passed to a task's `ResultBuilder` function. Provides access to the global context and task metadata during result construction.

```go
type TaskContext struct {
    TaskID        string
    GlobalContext *GlobalContext
    Logger        *slog.Logger
}

func NewTaskContext(taskID string, globalContext *GlobalContext, logger *slog.Logger) *TaskContext
```

### Action

```go
type Action[T ActionInterface] struct {
    ID        string
    Name      string
    RunID     string
    Wrapped   T
    StartTime time.Time
    EndTime   time.Time
    Duration  time.Duration
    Logger    *slog.Logger
}

func (a *Action[T]) Execute(ctx context.Context) error
func (a *Action[T]) GetOutput() interface{}
func (a *Action[T]) GetID() string
func (a *Action[T]) SetID(string)
func (a *Action[T]) GetName() string
func (a *Action[T]) GetDuration() time.Duration
func (a *Action[T]) GetLogger() *slog.Logger

// NewAction creates a wrapped action. ID is optional; if omitted it is generated from name.
func NewAction[T ActionInterface](wrapped T, name string, logger *slog.Logger, id ...string) *Action[T]
```

### BaseAction

Embed `BaseAction` in custom action structs for default no-op `BeforeExecute`/`AfterExecute` implementations and nil-safe logger support.

```go
type BaseAction struct {
    Logger *slog.Logger
}

func NewBaseAction(logger *slog.Logger) BaseAction
func (ba *BaseAction) BeforeExecute(ctx context.Context) error  // no-op
func (ba *BaseAction) AfterExecute(ctx context.Context) error   // no-op
func (ba *BaseAction) GetOutput() interface{}                   // returns nil
```

### ActionWrapper

```go
type ActionWrapper interface {
    Execute(ctx context.Context) error
    GetDuration() time.Duration
    GetLogger() *slog.Logger
    GetID() string
    SetID(string)
    GetName() string
    GetOutput() interface{}
}
```

### TaskHandle

```go
type TaskHandle struct { /* ... */ }

func (h *TaskHandle) Done() <-chan struct{}
func (h *TaskHandle) Err() error
func (h *TaskHandle) TaskID() string
```

### TaskManager

```go
type TaskManager struct {
    // ... internal fields
}

func NewTaskManager(logger *slog.Logger) *TaskManager
func (tm *TaskManager) AddTask(task *Task) error
func (tm *TaskManager) RunTask(taskID string) (*TaskHandle, error)
func (tm *TaskManager) StopTask(taskID string) error
func (tm *TaskManager) StopAllTasks()
func (tm *TaskManager) GetRunningTasks() []string
func (tm *TaskManager) IsTaskRunning(taskID string) bool
func (tm *TaskManager) WaitForAllTasksToComplete(timeout time.Duration) error
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

### TaskResultParameter

```go
type TaskResultParameter struct {
    TaskID    string
    ResultKey string
}

func (p TaskResultParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error)
```

### EntityOutputParameter

References output from any entity type (`"action"` or `"task"`). Falls back to Results if Outputs are not available.

```go
type EntityOutputParameter struct {
    EntityType string // "action" or "task"
    EntityID   string
    OutputKey  string
}

func (p EntityOutputParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error)
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

```go
func ActionResult(actionID string) ActionResultParameter
func ActionResultField(actionID, field string) ActionResultParameter
```

### TaskResult

```go
func TaskResult(taskID string) TaskResultParameter
func TaskResultField(taskID, field string) TaskResultParameter
```

### EntityOutput

```go
func EntityOutput(entityType, entityID string) EntityOutputParameter
func EntityOutputField(entityType, entityID, field string) EntityOutputParameter
```

### Typed Context Helpers (Recommended)

```go
// Typed access to action/task results
func ActionResultAs[T any](gc *GlobalContext, actionID string) (T, bool)
func TaskResultAs[T any](gc *GlobalContext, taskID string) (T, bool)

// Typed access to action/task output map fields
func ActionOutputFieldAs[T any](gc *GlobalContext, actionID, key string) (T, error)
func TaskOutputFieldAs[T any](gc *GlobalContext, taskID, key string) (T, error)

// Unified entity value lookup (tries Outputs first, then Results)
func EntityValue(gc *GlobalContext, entityType, id, key string) (interface{}, error)
func EntityValueAs[T any](gc *GlobalContext, entityType, id, key string) (T, error)

// Generic parameter resolver
func ResolveAs[T any](ctx context.Context, p ActionParameter, gc *GlobalContext) (T, error)
```

### Typed Parameter Resolution Helpers

These package-level functions resolve an `ActionParameter` directly to a concrete type without requiring `ParameterResolver` to be embedded:

```go
func ResolveString(ctx context.Context, p ActionParameter, gc *GlobalContext) (string, error)
func ResolveBool(ctx context.Context, p ActionParameter, gc *GlobalContext) (bool, error)
func ResolveStringSlice(ctx context.Context, p ActionParameter, gc *GlobalContext) ([]string, error)
```

### Action ID Utilities

```go
// SanitizeIDPart makes a value safe for inclusion in an action ID.
// Lowercases, trims, replaces spaces/slashes with '-', strips non-[a-z0-9_:\-.] characters.
func SanitizeIDPart(s string) string

// BuildActionID constructs a consistent action ID: "prefix-part1-part2-action".
func BuildActionID(prefix string, parts ...string) string
```

### TypedOutputKey

Provides compile-time-friendly output key validation for struct-based outputs.

```go
type TypedOutputKey[T any] struct {
    ActionID string
    Key      string
}

// Validate checks whether Key is a valid exported field on type T (when T is a struct).
func (k TypedOutputKey[T]) Validate() error
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

### ActionWithResults

Combines `ActionInterface` with `ResultProvider` for actions that produce rich typed results.

```go
type ActionWithResults interface {
    ActionInterface
    ResultProvider
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

### TaskWithResults

```go
type TaskWithResults interface {
    TaskInterface
    ResultProvider
}
```

### TaskManagerInterface

```go
type TaskManagerInterface interface {
    AddTask(task *Task) error
    RunTask(taskID string) (*TaskHandle, error)
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

## Common Package (`actions/common`)

The `common` package provides DRY helpers for building actions with consistent constructor, parameter resolution, and output patterns.

### BaseConstructor

```go
type BaseConstructor[T ActionInterface] struct { /* ... */ }

func NewBaseConstructor[T ActionInterface](logger *slog.Logger) *BaseConstructor[T]
func (c *BaseConstructor[T]) GetLogger() *slog.Logger
// WrapAction returns an Action[T] with the given name and optional explicit ID.
func (c *BaseConstructor[T]) WrapAction(action T, name string, id ...string) *Action[T]
```

### ParameterResolver

Embed `ParameterResolver` in action structs to get typed parameter resolution without manual GlobalContext extraction.

```go
type ParameterResolver struct { /* ... */ }

func NewParameterResolver(logger *slog.Logger) *ParameterResolver
func (pr *ParameterResolver) ResolveParameter(ctx context.Context, param ActionParameter, name string) (interface{}, error)
func (pr *ParameterResolver) ResolveStringParameter(ctx context.Context, param ActionParameter, name string) (string, error)
func (pr *ParameterResolver) ResolveBoolParameter(ctx context.Context, param ActionParameter, name string) (bool, error)
func (pr *ParameterResolver) ResolveIntParameter(ctx context.Context, param ActionParameter, name string) (int, error)
func (pr *ParameterResolver) ResolveStringSliceParameter(ctx context.Context, param ActionParameter, name string) ([]string, error)
func (pr *ParameterResolver) ResolveDurationParameter(ctx context.Context, param ActionParameter, name string) (time.Duration, error)
func (pr *ParameterResolver) ResolveMapParameter(ctx context.Context, param ActionParameter, name string) (map[string]interface{}, error)
func (pr *ParameterResolver) ResolveSliceParameter(ctx context.Context, param ActionParameter, name string) ([]interface{}, error)
```

### OutputBuilder

Embed `OutputBuilder` in action structs to produce consistent `map[string]interface{}` outputs.

```go
type OutputBuilder struct { /* ... */ }

func NewOutputBuilder(logger *slog.Logger) *OutputBuilder
func (ob *OutputBuilder) BuildStandardOutput(output interface{}, success bool, additionalFields map[string]interface{}) map[string]interface{}
func (ob *OutputBuilder) BuildOutputWithCount(items interface{}, success bool, additionalFields map[string]interface{}) map[string]interface{}
func (ob *OutputBuilder) BuildSimpleOutput(success bool, message string) map[string]interface{}
func (ob *OutputBuilder) BuildErrorOutput(error interface{}, additionalFields map[string]interface{}) map[string]interface{}
func (ob *OutputBuilder) BuildOutputFromStruct(action interface{}, success bool, excludeFields []string) map[string]interface{}
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
