# Testing Utilities

Testing utilities for the task-engine library.

## Quick Start

### Mocks

```go
import "github.com/ndizazzo/task-engine/testing/mocks"

mockManager := mocks.NewEnhancedTaskManagerMock()
mockRunner := &mocks.MockCommandRunner{}
```

### Testable Manager

```go
import "github.com/ndizazzo/task-engine/testing"

tm := testing.NewTestableTaskManager(logger)

// Set hooks
tm.SetTaskAddedHook(func(task *task_engine.Task) {
    // Custom logic
})

// Override results
tm.OverrideTaskResult("task1", "expected result")
```

### Performance Testing

```go
tester := testing.NewPerformanceTester(taskManager, logger)

// Run benchmarks
metrics := tester.BenchmarkTaskExecution(ctx, task, 100, 5)

// Load test
loadMetrics := tester.LoadTest(ctx, task, 1000, 10, time.Minute)
```

## Components

- **Mocks**: Complete mock implementations for all interfaces
- **Testable Manager**: Enhanced TaskManager with testing hooks
- **Performance Testing**: Built-in benchmarking and load testing
- **Test Data**: Sample files and fixtures

## Best Practices

1. Use `TestableTaskManager` for integration tests
2. Use mocks for unit tests
3. Clean up test data between tests
4. Set hooks early in test setup

See individual test files for complete examples.
