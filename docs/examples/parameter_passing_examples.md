# Parameter Passing Examples

## Basic Examples

### 1. File Processing Pipeline

Read a file, process its content, and write the result to a new file using action output references.

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "log/slog"
    "os"

    engine "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions/file"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Step 1: Read source file
    var contentBuffer []byte
    readAction, err := file.NewReadFileAction(logger).WithParameters(
        engine.StaticParameter{Value: "input.txt"},
        &contentBuffer,
    )
    if err != nil {
        logger.Error("Failed to create read action", "error", err)
        os.Exit(1)
    }
    readAction.ID = "read-source-file"

    // Step 2: Process content (custom action)
    processAction := &engine.Action[*ContentProcessorAction]{
        ID: "process-content",
        Wrapped: &ContentProcessorAction{
            BaseAction: engine.NewBaseAction(logger),
        },
    }

    // Step 3: Write processed content using output from step 2
    writeAction, err := file.NewWriteFileAction(logger).WithParameters(
        engine.StaticParameter{Value: "output.txt"},
        engine.ActionOutputField("process-content", "processedContent"), // resolved at runtime
        true,
        nil,
    )
    if err != nil {
        logger.Error("Failed to create write action", "error", err)
        os.Exit(1)
    }
    writeAction.ID = "write-output-file"

    task := &engine.Task{
        ID:      "file-pipeline",
        Name:    "Process File Content",
        Actions: []engine.ActionWrapper{readAction, processAction, writeAction},
        Logger:  logger,
    }

    if err := task.Run(context.Background()); err != nil {
        logger.Error("Task failed", "error", err)
        os.Exit(1)
    }

    logger.Info("File processing completed successfully!")
}

// Custom action that processes content
type ContentProcessorAction struct {
    engine.BaseAction
    processedContent []byte
}

func (a *ContentProcessorAction) Execute(ctx context.Context) error {
    globalCtx, ok := ctx.Value(engine.GlobalContextKey).(*engine.GlobalContext)
    if !ok {
        return fmt.Errorf("global context not found")
    }

    // Use the typed helper to extract the content field from the read action
    content, err := engine.ActionOutputFieldAs[[]byte](globalCtx, "read-source-file", "content")
    if err != nil {
        return fmt.Errorf("read action output not found: %w", err)
    }

    a.processedContent = bytes.ToUpper(content)
    return nil
}

func (a *ContentProcessorAction) GetOutput() interface{} {
    return map[string]interface{}{
        "processedContent": a.processedContent,
        "success":          true,
    }
}
```

### 2. Cross-Task Workflow (Build → Deploy)

Use `TaskManager` to share context between tasks. The deploy task references output produced by the build task.

```go
package main

import (
    "context"
    "log/slog"
    "os"

    engine "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions/docker"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    manager := engine.NewTaskManager(logger)

    // Build task: produces output including the built image ID
    buildTask := &engine.Task{
        ID:   "build-app",
        Name: "Build Docker Application",
        Actions: []engine.ActionWrapper{
            // ... your build actions ...
        },
        Logger: logger,
    }

    // Deploy task: consumes build task output
    deployAction, err := docker.NewDockerRunAction(logger).WithParameters(
        engine.TaskOutputField("build-app", "imageID"), // resolved from build task output
        nil,
        "-p", "8080:8080", "-d",
    )
    if err != nil {
        logger.Error("Failed to create deploy action", "error", err)
        os.Exit(1)
    }

    deployTask := &engine.Task{
        ID:      "deploy-app",
        Name:    "Deploy Application",
        Actions: []engine.ActionWrapper{deployAction},
        Logger:  logger,
    }

    if err := manager.AddTask(buildTask); err != nil {
        logger.Error("Failed to add build task", "error", err)
        os.Exit(1)
    }
    if err := manager.AddTask(deployTask); err != nil {
        logger.Error("Failed to add deploy task", "error", err)
        os.Exit(1)
    }

    // Execute build first, wait for completion
    buildHandle, err := manager.RunTask("build-app")
    if err != nil {
        logger.Error("Build failed to start", "error", err)
        os.Exit(1)
    }
    <-buildHandle.Done()
    if err := buildHandle.Err(); err != nil {
        logger.Error("Build failed", "error", err)
        os.Exit(1)
    }

    // Then execute deploy (shares the same global context via TaskManager)
    deployHandle, err := manager.RunTask("deploy-app")
    if err != nil {
        logger.Error("Deploy failed to start", "error", err)
        os.Exit(1)
    }
    <-deployHandle.Done()
    if err := deployHandle.Err(); err != nil {
        logger.Error("Deploy failed", "error", err)
        os.Exit(1)
    }

    logger.Info("Build and deploy completed successfully!")
}
```

### 3. Conditional Processing Based on Action Output

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    engine "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions/system"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Check service status
    statusAction, err := system.NewServiceStatusAction(logger).WithParameters(
        engine.StaticParameter{Value: []string{"nginx"}},
    )
    if err != nil {
        logger.Error("Failed to create status action", "error", err)
        os.Exit(1)
    }
    statusAction.ID = "service-status"

    task := &engine.Task{
        ID:   "conditional-processing",
        Name: "Conditional File Processing",
        Actions: []engine.ActionWrapper{
            statusAction,
            createConditionalAction(logger),
        },
        Logger: logger,
    }

    if err := task.Run(context.Background()); err != nil {
        logger.Error("Task failed", "error", err)
        os.Exit(1)
    }
    logger.Info("Conditional processing completed!")
}

func createConditionalAction(logger *slog.Logger) engine.ActionWrapper {
    return &engine.Action[*ConditionalAction]{
        ID:      "conditional-action",
        Wrapped: &ConditionalAction{BaseAction: engine.NewBaseAction(logger)},
    }
}

type ConditionalAction struct {
    engine.BaseAction
    result string
}

func (a *ConditionalAction) Execute(ctx context.Context) error {
    globalCtx, ok := ctx.Value(engine.GlobalContextKey).(*engine.GlobalContext)
    if !ok {
        return fmt.Errorf("global context not found")
    }

    // Use EntityValue to read the service status output
    statusOutput, err := engine.EntityValue(globalCtx, "action", "service-status", "")
    if err != nil {
        return fmt.Errorf("service status output not found: %w", err)
    }

    statusMap, ok := statusOutput.(map[string]interface{})
    if !ok {
        return fmt.Errorf("unexpected status output type: %T", statusOutput)
    }

    if success, _ := statusMap["success"].(bool); success {
        a.result = "Service available - processing enabled"
    } else {
        a.result = "Service unavailable - processing disabled"
    }
    return nil
}

func (a *ConditionalAction) GetOutput() interface{} {
    return map[string]interface{}{"result": a.result, "success": true}
}
```

## Advanced Examples

### 4. Multi-Task Workflow with Data Flow

Use `TaskManager` to orchestrate a Build → Test → Deploy pipeline where each stage consumes the prior stage's output.

```go
package main

import (
    "context"
    "log/slog"
    "os"

    engine "github.com/ndizazzo/task-engine"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    manager := engine.NewTaskManager(logger)

    workflow := []*engine.Task{
        {
            ID:   "build",
            Name: "Build Application",
            // ... build actions ...
            Logger: logger,
        },
        {
            ID:   "test",
            Name: "Run Tests",
            // Actions here can reference engine.TaskOutputField("build", "imageID")
            Logger: logger,
        },
        {
            ID:   "deploy",
            Name: "Deploy to Production",
            // Actions here can reference engine.TaskOutputField("build", "imageID")
            Logger: logger,
        },
    }

    for _, task := range workflow {
        if err := manager.AddTask(task); err != nil {
            logger.Error("Failed to add task", "taskID", task.ID, "error", err)
            os.Exit(1)
        }
    }

    // Execute each stage in sequence; TaskManager shares GlobalContext automatically
    for _, task := range workflow {
        handle, err := manager.RunTask(task.ID)
        if err != nil {
            logger.Error("Task failed to start", "taskID", task.ID, "error", err)
            os.Exit(1)
        }
        <-handle.Done()
        if err := handle.Err(); err != nil {
            logger.Error("Task failed", "taskID", task.ID, "error", err)
            os.Exit(1)
        }
    }

    logger.Info("Workflow completed successfully!")
}
```

### 5. Parameter Validation and Error Handling

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    engine "github.com/ndizazzo/task-engine"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    task := &engine.Task{
        ID:   "validation-example",
        Name: "Parameter Validation Example",
        Actions: []engine.ActionWrapper{
            &engine.Action[*ValidationAction]{
                ID: "validation-action",
                Wrapped: &ValidationAction{
                    BaseAction:  engine.NewBaseAction(logger),
                    StringParam: engine.StaticParameter{Value: "test-string"},
                    IntParam:    engine.StaticParameter{Value: 42},
                },
            },
        },
        Logger: logger,
    }

    if err := task.Run(context.Background()); err != nil {
        logger.Error("Task failed", "error", err)
        os.Exit(1)
    }

    logger.Info("Validation example completed!")
}

type ValidationAction struct {
    engine.BaseAction
    StringParam engine.ActionParameter
    IntParam    engine.ActionParameter
    results     map[string]interface{}
}

func (a *ValidationAction) Execute(ctx context.Context) error {
    gc, _ := ctx.Value(engine.GlobalContextKey).(*engine.GlobalContext)

    strVal, err := engine.ResolveString(ctx, a.StringParam, gc)
    if err != nil {
        return fmt.Errorf("string parameter invalid: %w", err)
    }
    if strVal == "" {
        return fmt.Errorf("string parameter cannot be empty")
    }

    intVal, err := engine.ResolveAs[int](ctx, a.IntParam, gc)
    if err != nil {
        return fmt.Errorf("int parameter invalid: %w", err)
    }
    if intVal < 0 {
        return fmt.Errorf("int parameter must be non-negative, got %d", intVal)
    }

    a.results = map[string]interface{}{
        "stringResult": strVal,
        "intResult":    intVal,
    }
    return nil
}

func (a *ValidationAction) GetOutput() interface{} {
    return map[string]interface{}{
        "results":    a.results,
        "success":    true,
        "paramCount": len(a.results),
    }
}
```

## Testing Examples

### 6. Testing Parameter Resolution

```go
package mypackage_test

import (
    "context"
    "log/slog"
    "os"
    "testing"

    engine "github.com/ndizazzo/task-engine"
    "github.com/stretchr/testify/suite"
)

type ParameterResolutionTestSuite struct {
    suite.Suite
    globalCtx *engine.GlobalContext
    logger    *slog.Logger
}

func TestParameterResolutionSuite(t *testing.T) {
    suite.Run(t, new(ParameterResolutionTestSuite))
}

func (suite *ParameterResolutionTestSuite) SetupTest() {
    suite.globalCtx = engine.NewGlobalContext()
    suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *ParameterResolutionTestSuite) TestStaticParameterResolution() {
    param := engine.StaticParameter{Value: "test-value"}
    value, err := param.Resolve(context.Background(), suite.globalCtx)
    suite.NoError(err)
    suite.Equal("test-value", value)
}

func (suite *ParameterResolutionTestSuite) TestActionOutputParameterResolution() {
    suite.globalCtx.StoreActionOutput("test-action", map[string]interface{}{
        "content": "action-content",
        "size":    1024,
    })

    // Resolve entire output
    param := engine.ActionOutputParameter{ActionID: "test-action"}
    value, err := param.Resolve(context.Background(), suite.globalCtx)
    suite.NoError(err)
    outputMap, ok := value.(map[string]interface{})
    suite.True(ok)
    suite.Equal("action-content", outputMap["content"])

    // Resolve specific field
    fieldParam := engine.ActionOutputParameter{ActionID: "test-action", OutputKey: "content"}
    fieldValue, err := fieldParam.Resolve(context.Background(), suite.globalCtx)
    suite.NoError(err)
    suite.Equal("action-content", fieldValue)
}

func (suite *ParameterResolutionTestSuite) TestCrossTaskParameterResolution() {
    suite.globalCtx.StoreTaskOutput("build-task", map[string]interface{}{
        "packagePath": "/tmp/build/package.tar",
    })

    param := engine.TaskOutputParameter{TaskID: "build-task", OutputKey: "packagePath"}
    value, err := param.Resolve(context.Background(), suite.globalCtx)
    suite.NoError(err)
    suite.Equal("/tmp/build/package.tar", value)
}

func (suite *ParameterResolutionTestSuite) TestParameterResolutionErrors() {
    param := engine.ActionOutputParameter{ActionID: "non-existent"}
    _, err := param.Resolve(context.Background(), suite.globalCtx)
    suite.Error(err)
    suite.Contains(err.Error(), "action 'non-existent' not found")
}
```

## Performance Examples

### 7. Bulk Parameter Resolution

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "time"

    engine "github.com/ndizazzo/task-engine"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    task := createBulkProcessingTask(logger)

    start := time.Now()
    if err := task.Run(context.Background()); err != nil {
        logger.Error("Bulk processing failed", "error", err)
        os.Exit(1)
    }
    logger.Info("Bulk processing completed", "duration", time.Since(start))
}

func createBulkProcessingTask(logger *slog.Logger) *engine.Task {
    actions := make([]engine.ActionWrapper, 100)
    for i := 0; i < 100; i++ {
        actions[i] = &engine.Action[*BulkAction]{
            ID:      fmt.Sprintf("bulk-action-%d", i),
            Wrapped: &BulkAction{BaseAction: engine.NewBaseAction(logger), Index: i},
        }
    }
    return &engine.Task{
        ID:      "bulk-processing",
        Name:    "Bulk Parameter Processing",
        Actions: actions,
        Logger:  logger,
    }
}

type BulkAction struct {
    engine.BaseAction
    Index  int
    Result string
}

func (a *BulkAction) Execute(ctx context.Context) error {
    time.Sleep(1 * time.Millisecond)
    a.Result = fmt.Sprintf("processed-%d", a.Index)
    return nil
}

func (a *BulkAction) GetOutput() interface{} {
    return map[string]interface{}{
        "index":   a.Index,
        "result":  a.Result,
        "success": true,
    }
}
```

These examples demonstrate the full range of parameter passing capabilities, from basic action-level references to complex cross-task workflows.
