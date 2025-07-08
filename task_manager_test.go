package task_engine_test

import (
	"testing"
	"time"

	engine "github.com/ndizazzo/task-engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskManager_AddTask(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "test-task",
		Name:    "Test Task",
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task)
	require.NoError(t, err)
	assert.Contains(t, taskManager.Tasks, "test-task", "TaskManager should contain the added task")
}

func TestTaskManager_RunTask(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "test-task",
		Name:    "Test Task",
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task)
	require.NoError(t, err)

	err = taskManager.RunTask("test-task")
	assert.NoError(t, err, "Task should start without errors")
	assert.GreaterOrEqualf(t, task.GetTotalTime(), time.Duration(0), "Task duration should be greater than or equal to 0")
}

func TestTaskManager_StopTask(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "test-task",
		Name:    "Test Task",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task)
	require.NoError(t, err)
	err = taskManager.RunTask("test-task")
	require.NoError(t, err)
	err = taskManager.StopTask("test-task")

	assert.NoError(t, err, "Task should be stopped without errors")
	assert.LessOrEqual(t, task.GetTotalTime(), LongActionTime, "Task should be stopped before the delay expires")
}

func TestTaskManager_StopAllTasks(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	task1 := &engine.Task{
		ID:      "task-1",
		Name:    "Task 1",
		Actions: LongRunningActions,
	}

	task2 := &engine.Task{
		ID:      "task-2",
		Name:    "Task 2",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task1)
	require.NoError(t, err)
	err = taskManager.AddTask(task2)
	require.NoError(t, err)

	_ = taskManager.RunTask("task-1")
	_ = taskManager.RunTask("task-2")

	time.Sleep(10 * time.Millisecond)
	taskManager.StopAllTasks()

	assert.NotEqual(t, 100*time.Millisecond, task1.GetTotalTime(), "Task 1 should not complete fully")
	assert.NotEqual(t, 100*time.Millisecond, task2.GetTotalTime(), "Task 2 should not complete fully")
}

func TestTaskManager_StopNonRunningTask(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.StopTask("non-existent-task")
	assert.Error(t, err, "Stopping a non-existent task should return an error")
}

func TestTaskManager_AddNilTask(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.AddTask(nil)
	assert.Error(t, err, "Adding a nil task should return an error")
	assert.Contains(t, err.Error(), "task is nil", "Error message should indicate task is nil")
}

func TestTaskManager_RunNonExistentTask(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.RunTask("non-existent-task")
	assert.Error(t, err, "Running a non-existent task should return an error")
	assert.Contains(t, err.Error(), "not found", "Error message should indicate task was not found")
}

func TestTaskManager_RunTaskWithFailure(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "failing-task",
		Name:    "Failing Task",
		Actions: MultipleActionsFailure, // This contains a failing action
	}

	err := taskManager.AddTask(task)
	require.NoError(t, err)

	err = taskManager.RunTask("failing-task")
	assert.NoError(t, err, "RunTask should return no error (task runs in goroutine)")

	time.Sleep(50 * time.Millisecond)

	assert.Greater(t, task.GetTotalTime(), time.Duration(0), "Task should have some execution time even when failing")
}

func TestTaskManager_GetRunningTasks(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	runningTasks := taskManager.GetRunningTasks()
	assert.Empty(t, runningTasks, "Initially no tasks should be running")

	task := &engine.Task{
		ID:      "long-task",
		Name:    "Long Running Task",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task)
	require.NoError(t, err)

	err = taskManager.RunTask("long-task")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	runningTasks = taskManager.GetRunningTasks()
	assert.Len(t, runningTasks, 1, "Should have one running task")
	assert.Contains(t, runningTasks, "long-task", "Should contain the long-task")

	err = taskManager.StopTask("long-task")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	runningTasks = taskManager.GetRunningTasks()
	assert.Empty(t, runningTasks, "No tasks should be running after stopping")
}

func TestTaskManager_IsTaskRunning(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	isRunning := taskManager.IsTaskRunning("non-existent")
	assert.False(t, isRunning, "Non-existent task should not be running")

	task := &engine.Task{
		ID:      "test-task",
		Name:    "Test Task",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task)
	require.NoError(t, err)

	isRunning = taskManager.IsTaskRunning("test-task")
	assert.False(t, isRunning, "Task should not be running before starting")

	err = taskManager.RunTask("test-task")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	isRunning = taskManager.IsTaskRunning("test-task")
	assert.True(t, isRunning, "Task should be running after starting")

	err = taskManager.StopTask("test-task")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	isRunning = taskManager.IsTaskRunning("test-task")
	assert.False(t, isRunning, "Task should not be running after stopping")
}

func TestTaskManager_GetRunningTasksMultiple(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	task1 := &engine.Task{
		ID:      "task-1",
		Name:    "Task 1",
		Actions: LongRunningActions,
	}

	task2 := &engine.Task{
		ID:      "task-2",
		Name:    "Task 2",
		Actions: LongRunningActions,
	}

	task3 := &engine.Task{
		ID:      "task-3",
		Name:    "Task 3",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task1)
	require.NoError(t, err)
	err = taskManager.AddTask(task2)
	require.NoError(t, err)
	err = taskManager.AddTask(task3)
	require.NoError(t, err)

	err = taskManager.RunTask("task-1")
	require.NoError(t, err)
	err = taskManager.RunTask("task-2")
	require.NoError(t, err)
	err = taskManager.RunTask("task-3")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	runningTasks := taskManager.GetRunningTasks()
	assert.Len(t, runningTasks, 3, "Should have three running tasks")
	assert.Contains(t, runningTasks, "task-1", "Should contain task-1")
	assert.Contains(t, runningTasks, "task-2", "Should contain task-2")
	assert.Contains(t, runningTasks, "task-3", "Should contain task-3")

	taskManager.StopAllTasks()

	time.Sleep(10 * time.Millisecond)

	runningTasks = taskManager.GetRunningTasks()
	assert.Empty(t, runningTasks, "No tasks should be running after stopping all")
}
