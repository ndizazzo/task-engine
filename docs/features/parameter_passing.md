# Action Parameter Passing

Actions can reference outputs from previous actions and tasks using declarative parameters.

## Quick Examples

### Action-to-Action Parameter Passing

```go
func NewFileProcessingTask(logger *slog.Logger) *task_engine.Task {
    return &task_engine.Task{
        ID: "file-processing",
        Actions: []task_engine.ActionWrapper{
            file.NewReadFileAction("input.txt", nil, logger),
            file.NewReplaceLinesAction(
                "input.txt",
                map[*regexp.Regexp]string{
                    regexp.MustCompile("{{content}}"):
                        task_engine.ActionOutput("read-file", "content"),
                },
                logger,
            ),
        },
        Logger: logger,
    }
}
```

### Cross-Task Parameter Passing

```go
// Build task
buildTask := &task_engine.Task{
    ID: "build-app",
    Actions: []task_engine.ActionWrapper{
        docker.NewDockerBuildAction("Dockerfile", ".", logger),
    },
    Logger: logger,
}

// Deploy task using build output
deployTask := &task_engine.Task{
    ID: "deploy-app",
    Actions: []task_engine.ActionWrapper{
        docker.NewDockerRunAction(
            task_engine.TaskOutput("build-app", "imageID"),
            []string{"-p", "8080:8080"},
            logger,
        ),
    },
    Logger: logger,
}

// Execute with shared context
manager := task_engine.NewTaskManager(logger)
globalCtx := task_engine.NewGlobalContext()

buildID := manager.AddTask(buildTask)
deployID := manager.AddTask(deployTask)

manager.RunTaskWithContext(ctx, buildID, globalCtx)
manager.RunTaskWithContext(ctx, deployID, globalCtx)
```

## Parameter Types

| Type                    | Purpose          | Example                                                                                 |
| ----------------------- | ---------------- | --------------------------------------------------------------------------------------- |
| `StaticParameter`       | Fixed values     | `StaticParameter{Value: "/tmp/file"}`                                                   |
| `ActionOutputParameter` | Action outputs   | `ActionOutputParameter{ActionID: "read-file", OutputKey: "content"}`                    |
| `TaskOutputParameter`   | Task outputs     | `TaskOutputParameter{TaskID: "build", OutputKey: "packagePath"}`                        |
| `EntityOutputParameter` | Mixed references | `EntityOutputParameter{EntityType: "action", EntityID: "process", OutputKey: "result"}` |

## Helper Functions

```go
// Action output references
task_engine.ActionOutput("action-id")
task_engine.ActionOutputField("action-id", "field-name")

// Task output references
task_engine.TaskOutput("task-id")
task_engine.TaskOutputField("task-id", "field-name")
```

## Implementation

### Add GetOutput() to Actions

```go
func (a *MyAction) GetOutput() interface{} {
    return map[string]interface{}{
        "result": a.Result,
        "success": a.Error == nil,
    }
}
```

### Use Parameters in Actions

```go
func (a *MyAction) Execute(ctx context.Context) error {
    value, err := a.Param.Resolve(ctx, a.globalContext)
    if err != nil {
        return fmt.Errorf("parameter resolution failed: %w", err)
    }

    strValue, ok := value.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", value)
    }

    return a.process(strValue)
}
```

## Examples

See [examples/parameter_passing_examples.md](examples/parameter_passing_examples.md) for comprehensive examples.
