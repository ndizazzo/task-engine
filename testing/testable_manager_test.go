package testing

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	task_engine "github.com/ndizazzo/task-engine"
)

func TestTestableTaskManager(t *testing.T) {
	// Use a discard logger to prevent test output
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("NewTestableTaskManager", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)
		require.NotNil(t, tm)
		assert.NotNil(t, tm.TaskManager)
		assert.NotNil(t, tm.taskResults)
		assert.NotNil(t, tm.taskErrors)
		assert.NotNil(t, tm.taskTiming)
	})

	t.Run("Hooks and Callbacks", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		var taskAddedCalled bool
		var taskStartedCalled bool
		var taskCompletedCalled bool
		var taskStoppedCalled bool

		tm.SetTaskAddedHook(func(task *task_engine.Task) {
			taskAddedCalled = true
		})

		tm.SetTaskStartedHook(func(taskID string) {
			taskStartedCalled = true
		})

		tm.SetTaskCompletedHook(func(taskID string, err error) {
			taskCompletedCalled = true
		})

		tm.SetTaskStoppedHook(func(taskID string) {
			taskStoppedCalled = true
		})

		// Create a simple task
		task := &task_engine.Task{
			ID:      "test-task",
			Name:    "Test Task",
			Actions: []task_engine.ActionWrapper{},
		}

		// Test AddTask hook
		err := tm.AddTask(task)
		require.NoError(t, err)
		assert.True(t, taskAddedCalled)

		// Test RunTask hook
		err = tm.RunTask("test-task")
		require.NoError(t, err)
		assert.True(t, taskStartedCalled)

		// Test StopTask hook
		err = tm.StopTask("test-task")
		require.NoError(t, err)
		assert.True(t, taskStoppedCalled)

		// Test SimulateTaskCompletion hook
		tm.SimulateTaskCompletion("test-task", nil)
		assert.True(t, taskCompletedCalled)

		// Clean up: wait for all tasks to complete naturally
		err = tm.WaitForAllTasksToComplete(100 * time.Millisecond)
		require.NoError(t, err, "All tasks should complete within timeout")
	})

	t.Run("Result Override and Retrieval", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		// Test setting and getting task results
		expectedResult := "test result"
		tm.OverrideTaskResult("task1", expectedResult)

		result, exists := tm.GetTaskResult("task1")
		assert.True(t, exists)
		assert.Equal(t, expectedResult, result)

		// Test non-existent result
		_, exists = tm.GetTaskResult("nonexistent")
		assert.False(t, exists)
	})

	t.Run("Error Override and Retrieval", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		// Test setting and getting task errors
		expectedError := errors.New("test error")
		tm.OverrideTaskError("task1", expectedError)

		err, exists := tm.GetTaskError("task1")
		assert.True(t, exists)
		assert.Equal(t, expectedError, err)

		// Test non-existent error
		_, exists = tm.GetTaskError("nonexistent")
		assert.False(t, exists)
	})

	t.Run("Timing Override and Retrieval", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		// Test setting and getting task timing
		expectedDuration := 5 * time.Second
		tm.OverrideTaskTiming("task1", expectedDuration)

		duration, exists := tm.GetTaskTiming("task1")
		assert.True(t, exists)
		assert.Equal(t, expectedDuration, duration)

		// Test non-existent timing
		_, exists = tm.GetTaskTiming("nonexistent")
		assert.False(t, exists)
	})

	t.Run("Call Tracking", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		// Create tasks
		task1 := &task_engine.Task{ID: "task1", Name: "Task 1", Actions: []task_engine.ActionWrapper{}}
		task2 := &task_engine.Task{ID: "task2", Name: "Task 2", Actions: []task_engine.ActionWrapper{}}

		// Add tasks
		err := tm.AddTask(task1)
		require.NoError(t, err)
		err = tm.AddTask(task2)
		require.NoError(t, err)

		// Check added calls
		addedCalls := tm.GetTaskAddedCalls()
		assert.Len(t, addedCalls, 2)
		assert.Equal(t, "task1", addedCalls[0].ID)
		assert.Equal(t, "task2", addedCalls[1].ID)

		// Run tasks
		err = tm.RunTask("task1")
		require.NoError(t, err)
		err = tm.RunTask("task2")
		require.NoError(t, err)

		// Check started calls
		startedCalls := tm.GetTaskStartedCalls()
		assert.Len(t, startedCalls, 2)
		assert.Equal(t, "task1", startedCalls[0])
		assert.Equal(t, "task2", startedCalls[1])

		// Stop tasks
		err = tm.StopTask("task1")
		require.NoError(t, err)
		err = tm.StopTask("task2")
		require.NoError(t, err)

		// Check stopped calls
		stoppedCalls := tm.GetTaskStoppedCalls()
		assert.Len(t, stoppedCalls, 2)
		assert.Equal(t, "task1", stoppedCalls[0])
		assert.Equal(t, "task2", stoppedCalls[1])

		// Clean up: wait for all tasks to complete naturally
		err = tm.WaitForAllTasksToComplete(100 * time.Millisecond)
		require.NoError(t, err, "All tasks should complete within timeout")
	})

	t.Run("Test Metrics", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		// Set up some test data
		tm.OverrideTaskResult("task1", "result1")
		tm.OverrideTaskError("task2", errors.New("error2"))
		tm.OverrideTaskTiming("task3", 3*time.Second)

		metrics := tm.GetTestMetrics()

		assert.Equal(t, 0, metrics["total_tasks_added"])
		assert.Equal(t, 0, metrics["total_tasks_started"])
		assert.Equal(t, 0, metrics["total_tasks_stopped"])
		assert.Equal(t, 1, metrics["total_results_set"])
		assert.Equal(t, 1, metrics["total_errors_set"])
		assert.Equal(t, 1, metrics["total_timing_set"])
	})

	t.Run("Clear Test Data", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		// Set up test data
		tm.OverrideTaskResult("task1", "result1")
		tm.OverrideTaskError("task2", errors.New("error2"))
		tm.OverrideTaskTiming("task3", 3*time.Second)

		// Verify data exists
		_, exists := tm.GetTaskResult("task1")
		assert.True(t, exists)

		// Clear test data
		tm.ClearTestData()

		// Verify data is cleared
		_, exists = tm.GetTaskResult("task1")
		assert.False(t, exists)

		_, exists = tm.GetTaskError("task2")
		assert.False(t, exists)

		_, exists = tm.GetTaskTiming("task3")
		assert.False(t, exists)
	})

	t.Run("Reset To Clean State", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		// Set up test data and hooks
		tm.OverrideTaskResult("task1", "result1")
		tm.SetTaskAddedHook(func(task *task_engine.Task) {})

		// Add a task
		task := &task_engine.Task{ID: "test-task", Name: "Test", Actions: []task_engine.ActionWrapper{}}
		err := tm.AddTask(task)
		require.NoError(t, err)

		// Verify state exists
		assert.Len(t, tm.Tasks, 1)
		_, exists := tm.GetTaskResult("task1")
		assert.True(t, exists)

		// Reset to clean state
		tm.ResetToCleanState()

		// Verify state is reset
		assert.Len(t, tm.Tasks, 0)
		_, exists = tm.GetTaskResult("task1")
		assert.False(t, exists)

		// Clean up: wait for all tasks to complete naturally
		err = tm.WaitForAllTasksToComplete(100 * time.Millisecond)
		require.NoError(t, err, "All tasks should complete within timeout")
	})

	t.Run("Integration with Real Task Manager", func(t *testing.T) {
		tm := NewTestableTaskManager(logger)

		// Create a mock action
		mockAction := &MockAction{
			ID:       "mock-action",
			Duration: 100 * time.Millisecond,
			Logger:   logger,
		}

		// Create a task with the mock action
		task := &task_engine.Task{
			ID:      "integration-test",
			Name:    "Integration Test",
			Actions: []task_engine.ActionWrapper{mockAction},
		}

		// Add and run the task
		err := tm.AddTask(task)
		require.NoError(t, err)

		err = tm.RunTask("integration-test")
		require.NoError(t, err)

		// Wait for task to complete
		time.Sleep(200 * time.Millisecond)

		// Verify task was added and started
		addedCalls := tm.GetTaskAddedCalls()
		assert.Len(t, addedCalls, 1)
		assert.Equal(t, "integration-test", addedCalls[0].ID)

		startedCalls := tm.GetTaskStartedCalls()
		assert.Len(t, startedCalls, 1)
		assert.Equal(t, "integration-test", startedCalls[0])

		// Clean up: wait for all tasks to complete naturally
		err = tm.WaitForAllTasksToComplete(100 * time.Millisecond)
		require.NoError(t, err, "All tasks should complete within timeout")
	})
}

// MockAction implements ActionWrapper for testing
type MockAction struct {
	ID       string
	Duration time.Duration
	Logger   *slog.Logger
}

func (ma *MockAction) GetID() string {
	return ma.ID
}

func (ma *MockAction) GetDuration() time.Duration {
	return ma.Duration
}

func (ma *MockAction) GetLogger() *slog.Logger {
	return ma.Logger
}

func (ma *MockAction) Execute(ctx context.Context) error {
	// Use a more deterministic approach that executes immediately
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Execute immediately without delays
		return nil
	}
}
