# Examples

Working examples for the Task Engine.

## Quick Examples

### File Operations

```go
task := tasks.NewFileOperationsTask(logger, "/tmp/project")
err := task.Run(context.Background())
```

### Docker Setup

```go
task := tasks.NewDockerSetupTask(logger, "/path/to/compose")
err := task.Run(context.Background())
```

### Package Management

```go
task := tasks.NewPackageManagementTask(logger, []string{"git", "curl"})
err := task.Run(context.Background())
```

## Parameter Passing

[parameter_passing_examples.md](parameter_passing_examples.md) - Complete examples of action parameter passing:

- File processing pipelines
- Docker workflows
- Multi-task workflows
- Testing and performance

## Key Concepts

- **Parameters**: `task_engine.ActionOutput()`, `task_engine.TaskOutput()`
- **Global Context**: Share data between tasks using `TaskManager`
- **Output Methods**: Implement `GetOutput()` in actions

See [README.md](../../README.md) for quick start and [ACTIONS.md](../../ACTIONS.md) for available actions.
