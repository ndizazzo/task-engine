# Task-Engine Mock Usage Examples

This directory contains examples of how downstream projects can use the enhanced mocks provided by task-engine to write better unit tests.

## Overview

The task-engine library provides enhanced mocks that make it easier to test components that depend on task management functionality. These mocks include:

- **EnhancedTaskManagerMock**: Comprehensive mocking of TaskManagerInterface
- **EnhancedTaskMock**: Mocking of individual TaskInterface implementations
- **ResultProviderMock**: Mocking of result-producing tasks

## Key Benefits

1. **Interface-Based Testing**: Use interfaces instead of concrete types for better testability
2. **State Tracking**: Mocks track internal state changes for comprehensive assertions
3. **Call Verification**: Verify exactly what methods were called with what arguments
4. **Result Override**: Set expected results and errors for testing different scenarios
5. **Test Isolation**: Each test runs in isolation without affecting others

## Example Usage

See `mock_usage_example_test.go` for a complete working example that demonstrates:

- How to create and configure enhanced mocks
- Setting up mock expectations
- Testing success and failure scenarios
- Verifying mock behavior and state changes
- Testing edge cases and error conditions

## Basic Pattern

```go
// 1. Create the enhanced mock
mockTaskManager := mocks.NewEnhancedTaskManagerMock()

// 2. Set up expectations
mockTaskManager.Mock.On("IsTaskRunning", "test-task").Return(false)
mockTaskManager.Mock.On("AddTask", mock.AnythingOfType("*task_engine.Task")).Return(nil)
mockTaskManager.Mock.On("RunTask", "test-task").Return(nil)

// 3. Use the mock in your component
processor := NewExampleTaskProcessor(mockTaskManager)
err := processor.ProcessTask("test-task")

// 4. Assertions
assert.NoError(t, err)
assert.Len(t, mockTaskManager.GetAddedTasks(), 1)
assert.Len(t, mockTaskManager.GetRunTaskCalls(), 1)

// 5. Verify all expectations were met
mockTaskManager.Mock.AssertExpectations(t)
```

## Available Mock Methods

### EnhancedTaskManagerMock

- `GetAddedTasks()` - Returns all tasks that were added
- `GetRunTaskCalls()` - Returns all RunTask calls made
- `GetTaskResult(taskID)` - Returns result set for a specific task
- `GetTaskError(taskID)` - Returns error set for a specific task
- `SetTaskResult(taskID, result)` - Sets expected result for a task
- `SetTaskError(taskID, error)` - Sets expected error for a task
- `ClearHistory()` - Clears all call history
- `ResetState()` - Resets internal state

### EnhancedTaskMock

- `SetResult(result)` - Sets expected result
- `SetError(error)` - Sets expected error
- `GetRunCallCount()` - Returns number of Run calls
- `ResetState()` - Resets internal state

### ResultProviderMock

- `SetResult(result)` - Sets expected result
- `SetError(error)` - Sets expected error
- `GetResultCallCount()` - Returns number of GetResult calls
- `GetErrorCallCount()` - Returns number of GetError calls
- `ResetState()` - Resets internal state

## Best Practices

1. **Use interfaces**: Design your components to accept interfaces rather than concrete types
2. **Set expectations**: Always set up mock expectations before calling the code under test
3. **Verify behavior**: Use the mock's state tracking methods to verify behavior
4. **Test isolation**: Reset mocks between tests to ensure clean state
5. **Assert expectations**: Use `AssertExpectations()` to ensure all expected calls were made

## Integration with testify/mock

The enhanced mocks are built on top of `testify/mock` and provide all the standard mock functionality:

- `On()` - Set expectations
- `Return()` - Set return values
- `Times()` - Set call count expectations
- `AssertExpectations()` - Verify all expectations were met
- `AssertCalled()` - Verify specific calls were made
- `AssertNotCalled()` - Verify specific calls were NOT made
