# Troubleshooting

Common issues and solutions for the Task Engine.

## Parameter Resolution Errors

### "action not found in context"

**Problem**: Action output referenced before action executes.

```go
// ❌ Wrong - action doesn't exist yet
engine.ActionOutput("read-file", "content")

// ✅ Correct - reference existing action
engine.ActionOutput("read-source", "content")
```

**Solution**: Ensure action IDs match exactly and actions execute before being referenced.

### "output key not found"

**Problem**: Referenced output key doesn't exist in action output.

```go
// ❌ Wrong - key doesn't exist
engine.ActionOutput("read-file", "nonexistent")

// ✅ Correct - use existing key or omit for entire output
engine.ActionOutput("read-file", "content")
engine.ActionOutput("read-file", "") // entire output
```

**Solution**: Check action output structure or omit key for entire output.

## Task Execution Issues

### Task stops on first error

**Problem**: Task halts when any action fails.

```go
// ❌ Task stops here
action1.Execute(ctx) // fails
action2.Execute(ctx) // never runs

// ✅ Handle errors gracefully
if err := action1.Execute(ctx); err != nil {
    if errors.Is(err, engine.ErrPrerequisiteNotMet) {
        // Skip task gracefully
        return nil
    }
    return err
}
```

**Solution**: Use `ErrPrerequisiteNotMet` for graceful task abortion.

### Context cancellation

**Problem**: Task doesn't respect context cancellation.

```go
// ✅ Always check context in long-running actions
func (a *MyAction) Execute(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Do work
        }
    }
}
```

**Solution**: Check `ctx.Done()` in long-running operations.

## Docker Issues

### Container not found

**Problem**: Container name/ID doesn't exist.

```go
// ❌ Container might not exist
docker.NewGetContainerStateAction(logger, "my-container")

// ✅ Check if container exists first
containers := docker.NewGetAllContainersStateAction(logger)
// Filter by name/ID
```

**Solution**: Use `GetAllContainersStateAction` to verify container existence.

### Permission denied

**Problem**: Docker commands fail due to permissions.

```go
// ❌ Might fail without proper permissions
docker.NewDockerRunAction(logger, "nginx", nil, nil, nil)

// ✅ Ensure user is in docker group
// Run: sudo usermod -aG docker $USER
```

**Solution**: Add user to docker group or use sudo.

## File Operation Issues

### Path not found

**Problem**: File/directory doesn't exist.

```go
// ❌ Path might not exist
file.NewReadFileAction("/nonexistent/file", &content, logger)

// ✅ Create directories first
file.NewCreateDirectoriesAction(logger).WithParameters(
    engine.StaticParameter{Value: "/path/to"},
    engine.StaticParameter{Value: []string{"parent", "child"}},
)
```

**Solution**: Create parent directories before file operations.

### Permission denied

**Problem**: Insufficient file permissions.

```go
// ❌ Might fail without proper permissions
file.NewWriteFileAction("/etc/config", content, true, nil, logger)

// ✅ Check permissions or use appropriate paths
file.NewWriteFileAction("/tmp/config", content, true, nil, logger)
```

**Solution**: Use appropriate paths or check file permissions.

## System Service Issues

### Service not found

**Problem**: Systemd service doesn't exist.

```go
// ❌ Service might not exist
system.NewGetServiceStatusAction(logger, "nonexistent-service")

// ✅ Check service existence first
// Use systemctl list-units --type=service
```

**Solution**: Verify service name with `systemctl list-units`.

### Service management fails

**Problem**: Insufficient privileges for service control.

```go
// ❌ Might fail without sudo
system.NewManageServiceAction(logger).WithParameters(
    engine.StaticParameter{Value: "nginx"},
    engine.StaticParameter{Value: "restart"},
)

// ✅ Ensure proper privileges or use sudo
```

**Solution**: Run with appropriate privileges or use sudo.

## Common Patterns

### Error Handling

```go
// ✅ Proper error handling
if err := task.Run(ctx); err != nil {
    if errors.Is(err, engine.ErrPrerequisiteNotMet) {
        logger.Warn("Prerequisites not met", "error", err)
        return nil
    }
    logger.Error("Task failed", "error", err)
    return err
}
```

### Parameter Validation

```go
// ✅ Validate parameters before use
if actionID == "" {
    return fmt.Errorf("action ID cannot be empty")
}
```

### Context Usage

```go
// ✅ Always pass context through
func (a *MyAction) Execute(ctx context.Context) error {
    // Use ctx for cancellation, timeouts, etc.
    return nil
}
```

## Debugging

### Enable Debug Logging

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

### Check Action Outputs

```go
// Print all action outputs
for id, output := range globalContext.ActionOutputs {
    logger.Debug("Action output", "id", id, "output", output)
}
```

### Verify Task State

```go
// Check task execution state
logger.Info("Task state",
    "completed", task.GetCompletedTasks(),
    "totalTime", task.GetTotalTime(),
)
```
