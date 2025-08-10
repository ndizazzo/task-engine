# Testing Utilities

This directory contains comprehensive testing utilities and tools for the task-engine library.

## Contents

### Performance Testing Framework

The `performance_testing.go` file provides a comprehensive performance testing and benchmarking framework for the task-engine library.

#### Features

- **Performance Benchmarking**: Measure execution time and throughput for tasks
- **Load Testing**: Simulate high-load scenarios with controlled concurrency
- **Stress Testing**: Push the system to its limits to find breaking points
- **Comprehensive Metrics**: Track execution times, throughput, error rates, and more

#### Usage

```go
import "github.com/ndizazzo/task-engine/testing"

// Create a performance tester
tester := testing.NewPerformanceTester(taskManager, logger)

// Run benchmarks
metrics := tester.BenchmarkTaskExecution(ctx, task, iterations, concurrent)

// Run load tests
loadMetrics := tester.LoadTest(ctx, task, iterations, concurrency, duration)

// Run stress tests
stressMetrics := tester.StressTest(ctx, task, rounds, iterations, duration)
```

### Testable Task Manager

The `testable_manager.go` file provides an enhanced version of the TaskManager specifically designed for testing scenarios.

#### Features

- **Testing Hooks**: Set callbacks for task lifecycle events (added, started, completed, stopped)
- **Result Override**: Override expected results, errors, and timing for testing
- **Call Tracking**: Track all method calls for verification
- **State Management**: Enhanced state management and cleanup for tests
- **Integration Testing**: Seamless integration with the main TaskManager

#### Usage

```go
import "github.com/ndizazzo/task-engine/testing"

// Create a testable task manager
tm := testing.NewTestableTaskManager(logger)

// Set up testing hooks
tm.SetTaskAddedHook(func(task *task_engine.Task) {
    // Custom logic when tasks are added
})

tm.SetTaskCompletedHook(func(taskID string, err error) {
    // Custom logic when tasks complete
})

// Override expected results for testing
tm.OverrideTaskResult("task1", "expected result")
tm.OverrideTaskError("task2", errors.New("expected error"))

// Track method calls
addedCalls := tm.GetTaskAddedCalls()
startedCalls := tm.GetTaskStartedCalls()

// Clear test data between tests
tm.ClearTestData()
tm.ResetToCleanState()
```

### Mock Implementations

The `mocks/` directory contains comprehensive mock implementations for testing:

- **TaskManagerMock**: Mock implementation of TaskManagerInterface
- **CommandMock**: Mock implementation of CommandInterface
- **Enhanced Mock Tests**: Advanced mocking patterns and examples

#### Usage

```go
import "github.com/ndizazzo/task-engine/testing/mocks"

// Create mock implementations
mockTaskManager := &mocks.TaskManagerMock{}
mockCommand := &mocks.MockCommandRunner{}

// Set up expectations
mockTaskManager.On("AddTask", mock.AnythingOfType("*task_engine.Task")).Return(nil)
mockCommand.On("RunCommand", "echo", "hello").Return("hello", nil)
```

### Test Data

The `testdata/` directory contains test fixtures and sample data:

- **Compressed Files**: Sample compressed archives for testing extraction
- **Text Files**: Sample text files for testing file operations
- **Other Fixtures**: Various test data files for different test scenarios

## Directory Structure

```
testing/
├── README.md                     # This documentation
├── performance_testing.go        # Performance testing framework
├── testable_manager.go           # Enhanced testable task manager
├── testable_manager_test.go      # Tests for testable manager
├── mocks/                        # Mock implementations
│   ├── task_manager_mock.go     # TaskManager mock
│   ├── command_mock.go          # Command mock
│   └── enhanced_mock_test.go    # Advanced mocking examples
└── testdata/                     # Test fixtures and data
    ├── compressed.tar.gz         # Sample compressed file
    └── test.txt                  # Sample text file
```

## Best Practices

1. **Use TestableTaskManager** for integration tests that need real TaskManager behavior
2. **Use Mocks** for unit tests that need to isolate specific components
3. **Use Performance Testing** for benchmarking and load testing scenarios
4. **Clean Up** test data between tests using `ClearTestData()` or `ResetToCleanState()`
5. **Set Hooks** early in test setup to capture all relevant events

## Examples

See the individual test files for comprehensive examples of how to use each testing utility effectively.
