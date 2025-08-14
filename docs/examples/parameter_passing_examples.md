# Parameter Passing Examples

## Basic Examples

### 1. File Processing Pipeline

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "regexp"

    "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions/file"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create a file processing pipeline
    task := &task_engine.Task{
        ID:   "file-pipeline",
        Name: "Process File Content",
        Actions: []task_engine.ActionWrapper{
            // Step 1: Read source file
            createReadAction("input.txt", logger),

            // Step 2: Process content using output from Step 1
            createProcessAction(logger),

            // Step 3: Write processed content using output from Step 2
            createWriteAction("output.txt", logger),
        },
        Logger: logger,
    }

    // Execute the task
    if err := task.Run(context.Background()); err != nil {
        logger.Error("Task failed", "error", err)
        os.Exit(1)
    }

    logger.Info("File processing completed successfully!")
}

func createReadAction(filePath string, logger *slog.Logger) *task_engine.Action[*file.ReadFileAction] {
    action, err := file.NewReadFileAction(filePath, nil, logger)
    if err != nil {
        panic(err)
    }
    action.ID = "read-source-file"
    return action
}

func createProcessAction(logger *slog.Logger) *task_engine.Action[*ContentProcessorAction] {
    return &task_engine.Action[*ContentProcessorAction]{
        ID: "process-content",
        Wrapped: &ContentProcessorAction{
            BaseAction: task_engine.NewBaseAction(logger),
        },
    }
}

func createWriteAction(destPath string, logger *slog.Logger) *task_engine.Action[*file.WriteFileAction] {
    action, err := file.NewWriteFileAction(
        destPath,
        task_engine.ActionOutput("process-content", "processedContent"),
        true, // overwrite
        logger,
    )
    if err != nil {
        panic(err)
    }
    action.ID = "write-output-file"
    return action
}

// Custom action that processes content
type ContentProcessorAction struct {
    task_engine.BaseAction
    processedContent []byte
}

func (a *ContentProcessorAction) Execute(ctx context.Context) error {
    // Resolve content directly using parameter helper instead of manual map parsing
    globalCtx, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext)
    if !ok {
        return fmt.Errorf("global context not found")
    }
    v, err := task_engine.ActionOutputField("read-source-file", "content").Resolve(ctx, globalCtx)
    if err != nil {
        return err
    }
    contentBytes, ok := v.([]byte)
    if !ok {
        return fmt.Errorf("content is not []byte")
    }
    a.processedContent = bytes.ToUpper(contentBytes)
    return nil
}

func (a *ContentProcessorAction) GetOutput() interface{} {
    return map[string]interface{}{
        "processedContent": a.processedContent,
        "originalSize":     len(a.processedContent),
        "success":          true,
    }
}
```

### 2. Docker Build and Deploy Pipeline

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions/docker"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create build task
    buildTask := &task_engine.Task{
        ID: "build-app",
        Name: "Build Docker Application",
        Actions: []task_engine.ActionWrapper{
            docker.NewDockerBuildAction("Dockerfile", ".", logger),
        },
        Logger: logger,
    }

    // Create deploy task that uses build output
    deployTask := &task_engine.Task{
        ID: "deploy-app",
        Name: "Deploy Application",
        Actions: []task_engine.ActionWrapper{
            docker.NewDockerRunAction(
                task_engine.TaskOutput("build-app", "imageID"),
                []string{"-p", "8080:8080", "-d"},
                logger,
            ),
        },
        Logger: logger,
    }

    // Use TaskManager for cross-task parameter passing
    manager := task_engine.NewTaskManager(logger)
    globalCtx := task_engine.NewGlobalContext()

    // Add and run tasks
    buildID := manager.AddTask(buildTask)
    deployID := manager.AddTask(deployTask)

    // Execute build first
    if err := manager.RunTaskWithContext(context.Background(), buildID, globalCtx); err != nil {
        logger.Error("Build failed", "error", err)
        os.Exit(1)
    }

    // Execute deploy using build output
    if err := manager.RunTaskWithContext(context.Background(), deployID, globalCtx); err != nil {
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
    "log/slog"
    "os"

    "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions/system"
    "github.com/ndizazzo/task-engine/actions/file"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    task := &task_engine.Task{
        ID: "conditional-processing",
        Name: "Conditional File Processing",
        Actions: []task_engine.ActionWrapper{
            // Check if service is running
            system.NewServiceStatusAction("nginx", logger),

            // Process based on service status
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

func createConditionalAction(logger *slog.Logger) *task_engine.Action[*ConditionalFileAction] {
    return &task_engine.Action[*ConditionalFileAction]{
        ID: "conditional-file-action",
        Wrapped: &ConditionalFileAction{
            BaseAction: task_engine.NewBaseAction(logger),
        },
    }
}

type ConditionalFileAction struct {
    task_engine.BaseAction
    result string
}

func (a *ConditionalFileAction) Execute(ctx context.Context) error {
    globalCtx, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext)
    if !ok {
        return fmt.Errorf("global context not found")
    }

    // Get service status
    statusOutput, exists := globalCtx.ActionOutputs["service-status"]
    if !exists {
        return fmt.Errorf("service status output not found")
    }

    statusMap, ok := statusOutput.(map[string]interface{})
    if !ok {
        return fmt.Errorf("status output is not a map")
    }

    isRunning, exists := statusMap["running"]
    if !exists {
        return fmt.Errorf("running field not found")
    }

    running, ok := isRunning.(bool)
    if !ok {
        return fmt.Errorf("running field is not boolean")
    }

    // Process based on status
    if running {
        a.result = "Service is running - processing enabled"
        // Perform processing logic here
    } else {
        a.result = "Service is stopped - processing disabled"
        // Skip processing logic here
    }

    return nil
}

func (a *ConditionalFileAction) GetOutput() interface{} {
    return map[string]interface{}{
        "result": a.result,
        "success": true,
    }
}
```

## Advanced Examples

### 4. Multi-Task Workflow with Data Flow

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/ndizazzo/task-engine"
    "github.com/ndizazzo/task-engine/actions/docker"
    "github.com/ndizazzo/task-engine/actions/file"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create a complex workflow: Build → Test → Deploy → Monitor
    workflow := createWorkflow(logger)

    // Execute with shared context
    manager := task_engine.NewTaskManager(logger)
    globalCtx := task_engine.NewGlobalContext()

    // Add all tasks
    taskIDs := make([]string, len(workflow))
    for i, task := range workflow {
        taskIDs[i] = manager.AddTask(task)
    }

    // Execute in sequence
    for _, taskID := range taskIDs {
        if err := manager.RunTaskWithContext(context.Background(), taskID, globalCtx); err != nil {
            logger.Error("Workflow failed", "taskID", taskID, "error", err)
            os.Exit(1)
        }
    }

    logger.Info("Complete workflow executed successfully!")
}

func createWorkflow(logger *slog.Logger) []*task_engine.Task {
    return []*task_engine.Task{
        // Task 1: Build
        &task_engine.Task{
            ID: "build",
            Name: "Build Application",
            Actions: []task_engine.ActionWrapper{
                docker.NewDockerBuildAction("Dockerfile", ".", logger),
            },
            Logger: logger,
        },

        // Task 2: Test (uses build output)
        &task_engine.Task{
            ID: "test",
            Name: "Run Tests",
            Actions: []task_engine.ActionWrapper{
                docker.NewDockerRunAction(
                    task_engine.TaskOutput("build", "imageID"),
                    []string{"go", "test", "./..."},
                    logger,
                ),
            },
            Logger: logger,
        },

        // Task 3: Deploy (uses build output)
        &task_engine.Task{
            ID: "deploy",
            Name: "Deploy to Production",
            Actions: []task_engine.ActionWrapper{
                docker.NewDockerRunAction(
                    task_engine.TaskOutput("build", "imageID"),
                    []string{"-p", "8080:8080", "-d", "--name", "production-app"},
                    logger,
                ),
            },
            Logger: logger,
        },

        // Task 4: Monitor (uses deploy output)
        &task_engine.Task{
            ID: "monitor",
            Name: "Monitor Deployment",
            Actions: []task_engine.ActionWrapper{
                createMonitorAction(logger),
            },
            Logger: logger,
        },
    }
}

func createMonitorAction(logger *slog.Logger) *task_engine.Action[*MonitorAction] {
    return &task_engine.Action[*MonitorAction]{
        ID: "monitor-deployment",
        Wrapped: &MonitorAction{
            BaseAction: task_engine.NewBaseAction(logger),
        },
    }
}

type MonitorAction struct {
    task_engine.BaseAction
    status string
}

func (a *MonitorAction) Execute(ctx context.Context) error {
    globalCtx, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext)
    if !ok {
        return fmt.Errorf("global context not found")
    }

    // Get deploy output
    deployOutput, exists := globalCtx.TaskOutputs["deploy"]
    if !exists {
        return fmt.Errorf("deploy task output not found")
    }

    // Monitor the deployment
    a.status = "Deployment monitored successfully"

    return nil
}

func (a *MonitorAction) GetOutput() interface{} {
    return map[string]interface{}{
        "status": a.status,
        "success": true,
    }
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
    "reflect"

    "github.com/ndizazzo/task-engine"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    task := &task_engine.Task{
        ID: "validation-example",
        Name: "Parameter Validation Example",
        Actions: []task_engine.ActionWrapper{
            createValidationAction(logger),
        },
        Logger: logger,
    }

    if err := task.Run(context.Background()); err != nil {
        logger.Error("Task failed", "error", err)
        os.Exit(1)
    }

    logger.Info("Validation example completed!")
}

func createValidationAction(logger *slog.Logger) *task_engine.Action[*ValidationAction] {
    return &task_engine.Action[*ValidationAction]{
        ID: "validation-action",
        Wrapped: &ValidationAction{
            BaseAction: task_engine.NewBaseAction(logger),
            // Use different parameter types
            StringParam: task_engine.StaticParameter{Value: "test-string"},
            IntParam:    task_engine.StaticParameter{Value: 42},
            FloatParam:  task_engine.StaticParameter{Value: 3.14},
        },
    }
}

type ValidationAction struct {
    task_engine.BaseAction
    StringParam task_engine.ActionParameter
    IntParam    task_engine.ActionParameter
    FloatParam  task_engine.ActionParameter
    results     map[string]interface{}
}

func (a *ValidationAction) Execute(ctx context.Context) error {
    a.results = make(map[string]interface{})

    // Validate and resolve string parameter
    if err := a.validateStringParam(ctx); err != nil {
        return fmt.Errorf("string parameter validation failed: %w", err)
    }

    // Validate and resolve int parameter
    if err := a.validateIntParam(ctx); err != nil {
        return fmt.Errorf("int parameter validation failed: %w", err)
    }

    // Validate and resolve float parameter
    if err := a.validateFloatParam(ctx); err != nil {
        return fmt.Errorf("float parameter validation failed: %w", err)
    }

    return nil
}

func (a *ValidationAction) validateStringParam(ctx context.Context) error {
    value, err := a.StringParam.Resolve(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to resolve string parameter: %w", err)
    }

    strValue, ok := value.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", value)
    }

    if len(strValue) == 0 {
        return fmt.Errorf("string parameter cannot be empty")
    }

    a.results["stringResult"] = strValue
    return nil
}

func (a *ValidationAction) validateIntParam(ctx context.Context) error {
    value, err := a.IntParam.Resolve(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to resolve int parameter: %w", err)
    }

    intValue, ok := value.(int)
    if !ok {
        return fmt.Errorf("expected int, got %T", value)
    }

    if intValue < 0 {
        return fmt.Errorf("int parameter must be non-negative, got %d", intValue)
    }

    a.results["intResult"] = intValue
    return nil
}

func (a *ValidationAction) validateFloatParam(ctx context.Context) error {
    value, err := a.FloatParam.Resolve(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to resolve float parameter: %w", err)
    }

    floatValue, ok := value.(float64)
    if !ok {
        return fmt.Errorf("expected float64, got %T", value)
    }

    if floatValue < 0 {
        return fmt.Errorf("float parameter must be non-negative, got %f", floatValue)
    }

    a.results["floatResult"] = floatValue
    return nil
}

func (a *ValidationAction) GetOutput() interface{} {
    return map[string]interface{}{
        "results": a.results,
        "success": true,
        "paramCount": len(a.results),
    }
}
```

## Testing Examples

### 6. Testing Parameter Resolution

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "testing"

    "github.com/ndizazzo/task-engine"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type ParameterResolutionTestSuite struct {
    suite.Suite
    globalCtx *task_engine.GlobalContext
    logger    *slog.Logger
}

func TestParameterResolutionSuite(t *testing.T) {
    suite.Run(t, new(ParameterResolutionTestSuite))
}

func (suite *ParameterResolutionTestSuite) SetupTest() {
    suite.globalCtx = task_engine.NewGlobalContext()
    suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *ParameterResolutionTestSuite) TestStaticParameterResolution() {
    param := task_engine.StaticParameter{Value: "test-value"}

    value, err := param.Resolve(context.Background(), suite.globalCtx)

    suite.NoError(err)
    suite.Equal("test-value", value)
}

func (suite *ParameterResolutionTestSuite) TestActionOutputParameterResolution() {
    // Store action output
    suite.globalCtx.StoreActionOutput("test-action", map[string]interface{}{
        "content": "action-content",
        "size":    1024,
    })

    // Test resolving entire output
    param := task_engine.ActionOutputParameter{ActionID: "test-action"}
    value, err := param.Resolve(context.Background(), suite.globalCtx)

    suite.NoError(err)
    outputMap, ok := value.(map[string]interface{})
    suite.True(ok)
    suite.Equal("action-content", outputMap["content"])
    suite.Equal(1024, outputMap["size"])

    // Test resolving specific field
    fieldParam := task_engine.ActionOutputParameter{
        ActionID:  "test-action",
        OutputKey: "content",
    }
    fieldValue, err := fieldParam.Resolve(context.Background(), suite.globalCtx)

    suite.NoError(err)
    suite.Equal("action-content", fieldValue)
}

func (suite *ParameterResolutionTestSuite) TestParameterResolutionErrors() {
    // Test non-existent action
    param := task_engine.ActionOutputParameter{ActionID: "non-existent"}
    _, err := param.Resolve(context.Background(), suite.globalCtx)

    suite.Error(err)
    suite.Contains(err.Error(), "action 'non-existent' not found")

    // Test non-existent field
    suite.globalCtx.StoreActionOutput("test-action", map[string]interface{}{
        "existing": "value",
    })

    fieldParam := task_engine.ActionOutputParameter{
        ActionID:  "test-action",
        OutputKey: "non-existent-field",
    }
    _, err = fieldParam.Resolve(context.Background(), suite.globalCtx)

    suite.Error(err)
    suite.Contains(err.Error(), "output key 'non-existent-field' not found")
}

func (suite *ParameterResolutionTestSuite) TestCrossTaskParameterResolution() {
    // Store task output
    suite.globalCtx.StoreTaskOutput("build-task", map[string]interface{}{
        "packagePath": "/tmp/build/package.tar",
        "buildTime":   "2024-01-01T00:00:00Z",
    })

    // Test resolving task output
    param := task_engine.TaskOutputParameter{
        TaskID:    "build-task",
        OutputKey: "packagePath",
    }

    value, err := param.Resolve(context.Background(), suite.globalCtx)

    suite.NoError(err)
    suite.Equal("/tmp/build/package.tar", value)
}
```

## Performance Examples

### 7. Bulk Parameter Resolution

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "sync"
    "time"

    "github.com/ndizazzo/task-engine"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    // Create bulk processing task
    task := createBulkProcessingTask(logger)

    start := time.Now()
    if err := task.Run(context.Background()); err != nil {
        logger.Error("Bulk processing failed", "error", err)
        os.Exit(1)
    }

    duration := time.Since(start)
    logger.Info("Bulk processing completed", "duration", duration)
}

func createBulkProcessingTask(logger *slog.Logger) *task_engine.Task {
    actions := make([]task_engine.ActionWrapper, 100)

    for i := 0; i < 100; i++ {
        actions[i] = createBulkAction(i, logger)
    }

    return &task_engine.Task{
        ID:      "bulk-processing",
        Name:    "Bulk Parameter Processing",
        Actions: actions,
        Logger:  logger,
    }
}

func createBulkAction(index int, logger *slog.Logger) *task_engine.Action[*BulkAction] {
    return &task_engine.Action[*BulkAction]{
        ID: fmt.Sprintf("bulk-action-%d", index),
        Wrapped: &BulkAction{
            BaseAction: task_engine.NewBaseAction(logger),
            Index:      index,
        },
    }
}

type BulkAction struct {
    task_engine.BaseAction
    Index  int
    Result string
}

func (a *BulkAction) Execute(ctx context.Context) error {
    // Simulate work
    time.Sleep(1 * time.Millisecond)

    a.Result = fmt.Sprintf("processed-%d", a.Index)
    return nil
}

func (a *BulkAction) GetOutput() interface{} {
    return map[string]interface{}{
        "index":  a.Index,
        "result": a.Result,
        "success": true,
    }
}
```

These examples demonstrate the full range of capabilities of the Action Parameter Passing system, from basic usage to advanced workflows and performance considerations.
