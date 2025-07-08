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

	// Wait for a short duration, then stop all tasks
	time.Sleep(10 * time.Millisecond)
	taskManager.StopAllTasks()

	// Ensure tasks were stopped
	assert.NotEqual(t, 100*time.Millisecond, task1.GetTotalTime(), "Task 1 should not complete fully")
	assert.NotEqual(t, 100*time.Millisecond, task2.GetTotalTime(), "Task 2 should not complete fully")
}

func TestTaskManager_StopNonRunningTask(t *testing.T) {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.StopTask("non-existent-task")
	assert.Error(t, err, "Stopping a non-existent task should return an error")
}
