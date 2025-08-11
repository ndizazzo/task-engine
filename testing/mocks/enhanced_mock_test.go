package mocks

import (
	"errors"
	"testing"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEnhancedTaskManagerMock(t *testing.T) {
	t.Run("NewEnhancedTaskManagerMock", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()
		require.NotNil(t, mockTM)
		assert.NotNil(t, mockTM.tasks)
		assert.NotNil(t, mockTM.runningTasks)
		assert.NotNil(t, mockTM.taskResults)
		assert.NotNil(t, mockTM.taskErrors)
		assert.NotNil(t, mockTM.taskTiming)
		assert.NotNil(t, mockTM.isRunningCalls)
	})

	t.Run("AddTask with State Tracking", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectation
		mockTM.On("AddTask", mock.AnythingOfType("*task_engine.Task")).Return(nil)

		task := &task_engine.Task{ID: "test-task", Name: "Test Task"}

		err := mockTM.AddTask(task)
		require.NoError(t, err)
		addedTasks := mockTM.GetAddedTasks()
		assert.Len(t, addedTasks, 1)
		assert.Equal(t, "test-task", addedTasks[0].ID)
		mockTM.AssertExpectations(t)
	})

	t.Run("RunTask with State Tracking", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("RunTask", "test-task").Return(nil)
		mockTM.On("IsTaskRunning", "test-task").Return(true)

		err := mockTM.RunTask("test-task")
		require.NoError(t, err)
		runCalls := mockTM.GetRunTaskCalls()
		assert.Len(t, runCalls, 1)
		assert.Equal(t, "test-task", runCalls[0])
		assert.True(t, mockTM.IsTaskRunning("test-task"))

		mockTM.AssertExpectations(t)
	})

	t.Run("StopTask with State Tracking", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("RunTask", "test-task").Return(nil)
		mockTM.On("StopTask", "test-task").Return(nil)
		mockTM.On("IsTaskRunning", "test-task").Return(false)

		// Start task first
		err := mockTM.RunTask("test-task")
		require.NoError(t, err)

		// Stop task
		err = mockTM.StopTask("test-task")
		require.NoError(t, err)
		stopCalls := mockTM.GetStopTaskCalls()
		assert.Len(t, stopCalls, 1)
		assert.Equal(t, "test-task", stopCalls[0])
		assert.False(t, mockTM.IsTaskRunning("test-task"))

		mockTM.AssertExpectations(t)
	})

	t.Run("StopAllTasks with Call Tracking", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("RunTask", "task1").Return(nil)
		mockTM.On("RunTask", "task2").Return(nil)
		mockTM.On("StopAllTasks").Return()

		// Start multiple tasks
		err := mockTM.RunTask("task1")
		require.NoError(t, err)
		err = mockTM.RunTask("task2")
		require.NoError(t, err)

		// Stop all tasks
		mockTM.StopAllTasks()
		stopAllCalls := mockTM.GetStopAllCalls()
		assert.Equal(t, 1, stopAllCalls)
		mockTM.On("IsTaskRunning", "task1").Return(false)
		mockTM.On("IsTaskRunning", "task2").Return(false)
		assert.False(t, mockTM.IsTaskRunning("task1"))
		assert.False(t, mockTM.IsTaskRunning("task2"))

		mockTM.AssertExpectations(t)
	})

	t.Run("GetRunningTasks with Call Tracking", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("RunTask", "task1").Return(nil)
		mockTM.On("GetRunningTasks").Return([]string{"task1"})

		// Start task
		err := mockTM.RunTask("task1")
		require.NoError(t, err)

		// Get running tasks
		running := mockTM.GetRunningTasks()
		assert.Len(t, running, 1)
		assert.Equal(t, "task1", running[0])
		getRunningCalls := mockTM.GetGetRunningCalls()
		assert.Equal(t, 1, getRunningCalls)

		mockTM.AssertExpectations(t)
	})

	t.Run("IsTaskRunning with Call Tracking", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("RunTask", "task1").Return(nil)
		mockTM.On("IsTaskRunning", "task1").Return(true)

		// Start task
		err := mockTM.RunTask("task1")
		require.NoError(t, err)
		isRunning := mockTM.IsTaskRunning("task1")
		assert.True(t, isRunning)
		isRunningCalls := mockTM.GetIsRunningCalls("task1")
		assert.Equal(t, 1, isRunningCalls)

		mockTM.AssertExpectations(t)
	})

	t.Run("Task Results Management", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set task result
		expectedResult := "test result"
		mockTM.SetTaskResult("task1", expectedResult)

		// Get task result
		result := mockTM.GetTaskResult("task1")
		assert.Equal(t, expectedResult, result)
		result = mockTM.GetTaskResult("nonexistent")
		assert.Nil(t, result)
	})

	t.Run("Task Errors Management", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set task error
		expectedError := errors.New("test error")
		mockTM.SetTaskError("task1", expectedError)

		// Get task error
		err := mockTM.GetTaskError("task1")
		assert.Equal(t, expectedError, err)
		err = mockTM.GetTaskError("nonexistent")
		assert.Nil(t, err)
	})

	t.Run("Task Timing Management", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set task timing
		expectedDuration := 5 * time.Second
		mockTM.SetTaskTiming("task1", expectedDuration)

		// Get task timing
		duration, exists := mockTM.GetTaskTiming("task1")
		assert.True(t, exists)
		assert.Equal(t, expectedDuration, duration)
		_, exists = mockTM.GetTaskTiming("nonexistent")
		assert.False(t, exists)
	})

	t.Run("GetCurrentState", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up some state
		mockTM.SetTaskResult("task1", "result1")
		mockTM.SetTaskError("task2", errors.New("error2"))
		mockTM.SetTaskTiming("task3", 3*time.Second)

		// Get current state
		state := mockTM.GetCurrentState()

		assert.Equal(t, 0, state["total_tasks"])
		assert.Equal(t, 0, state["running_tasks"])
		assert.Equal(t, 1, state["total_results"])
		assert.Equal(t, 1, state["total_errors"])
		assert.Equal(t, 1, state["total_timing"])
		assert.Equal(t, 0, state["add_task_calls"])
		assert.Equal(t, 0, state["run_task_calls"])
		assert.Equal(t, 0, state["stop_task_calls"])
		assert.Equal(t, 0, state["stop_all_calls"])
		assert.Equal(t, 0, state["get_running_calls"])
	})

	t.Run("SimulateTaskCompletion", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("RunTask", "task1").Return(nil)
		mockTM.On("IsTaskRunning", "task1").Return(true).Once()
		mockTM.On("IsTaskRunning", "task1").Return(false).Once()

		// Start task
		err := mockTM.RunTask("task1")
		require.NoError(t, err)
		assert.True(t, mockTM.IsTaskRunning("task1"))

		// Simulate completion
		mockTM.SimulateTaskCompletion("task1")
		assert.False(t, mockTM.IsTaskRunning("task1"))

		mockTM.AssertExpectations(t)
	})

	t.Run("SimulateTaskFailure", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("RunTask", "task1").Return(nil)
		mockTM.On("IsTaskRunning", "task1").Return(false).Once()

		// Start task
		err := mockTM.RunTask("task1")
		require.NoError(t, err)

		// Simulate failure
		expectedError := errors.New("task failed")
		mockTM.SimulateTaskFailure("task1", expectedError)
		assert.False(t, mockTM.IsTaskRunning("task1"))
		err = mockTM.GetTaskError("task1")
		assert.Equal(t, expectedError, err)

		mockTM.AssertExpectations(t)
	})

	t.Run("SetExpectedBehavior", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set expected behavior
		mockTM.SetExpectedBehavior()
		task := &task_engine.Task{ID: "test-task", Name: "Test"}

		err := mockTM.AddTask(task)
		assert.NoError(t, err)

		err = mockTM.RunTask("test-task")
		assert.NoError(t, err)

		err = mockTM.StopTask("test-task")
		assert.NoError(t, err)

		mockTM.StopAllTasks()

		running := mockTM.GetRunningTasks()
		assert.Empty(t, running)

		mockTM.On("IsTaskRunning", "test-task").Return(false)
		isRunning := mockTM.IsTaskRunning("test-task")
		assert.False(t, isRunning)

		mockTM.AssertExpectations(t)
	})

	t.Run("VerifyAllExpectations", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("AddTask", mock.Anything).Return(nil).Once()
		assert.Len(t, mockTM.ExpectedCalls, 1)

		// Fulfill expectations
		task := &task_engine.Task{ID: "test-task", Name: "Test"}
		err := mockTM.AddTask(task)
		require.NoError(t, err)
		mockTM.AssertExpectations(t)
		results := mockTM.VerifyAllExpectations()
		assert.True(t, results["expectations_met"]) // Now met
		assert.True(t, results["state_consistent"])
	})

	t.Run("ClearHistory", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up some history
		mockTM.SetTaskResult("task1", "result1")
		mockTM.SetTaskError("task2", errors.New("error2"))
		assert.Len(t, mockTM.GetAddedTasks(), 0)

		// Clear history
		mockTM.ClearHistory()
		assert.Len(t, mockTM.GetAddedTasks(), 0)
		assert.Len(t, mockTM.GetRunTaskCalls(), 0)
		assert.Len(t, mockTM.GetStopTaskCalls(), 0)
		assert.Equal(t, 0, mockTM.GetStopAllCalls())
		assert.Equal(t, 0, mockTM.GetGetRunningCalls())
	})

	t.Run("ClearState", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up some state
		mockTM.SetTaskResult("task1", "result1")
		mockTM.SetTaskError("task2", errors.New("error2"))
		mockTM.SetTaskTiming("task3", 3*time.Second)
		result := mockTM.GetTaskResult("task1")
		assert.Equal(t, "result1", result)

		// Clear state
		mockTM.ClearState()
		result = mockTM.GetTaskResult("task1")
		assert.Nil(t, result)

		err := mockTM.GetTaskError("task2")
		assert.Nil(t, err)

		_, exists := mockTM.GetTaskTiming("task3")
		assert.False(t, exists)
	})

	t.Run("ResetToCleanState", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up some state and history
		mockTM.SetTaskResult("task1", "result1")
		mockTM.SetExpectedBehavior()
		result := mockTM.GetTaskResult("task1")
		assert.Equal(t, "result1", result)

		// Reset to clean state
		mockTM.ResetToCleanState()
		result = mockTM.GetTaskResult("task1")
		assert.Nil(t, result)

		assert.Len(t, mockTM.GetAddedTasks(), 0)
		assert.Len(t, mockTM.GetRunTaskCalls(), 0)
		assert.Len(t, mockTM.GetStopTaskCalls(), 0)
		assert.Equal(t, 0, mockTM.GetStopAllCalls())
		assert.Equal(t, 0, mockTM.GetGetRunningCalls())
	})

	t.Run("GetAllIsRunningCalls", func(t *testing.T) {
		mockTM := NewEnhancedTaskManagerMock()

		// Set up expectations
		mockTM.On("IsTaskRunning", "task1").Return(true)
		mockTM.On("IsTaskRunning", "task2").Return(false)

		// Call IsTaskRunning multiple times
		mockTM.IsTaskRunning("task1")
		mockTM.IsTaskRunning("task1")
		mockTM.IsTaskRunning("task2")

		// Get all call counts
		allCalls := mockTM.GetAllIsRunningCalls()

		assert.Equal(t, 2, allCalls["task1"])
		assert.Equal(t, 1, allCalls["task2"])

		mockTM.AssertExpectations(t)
	})
}
