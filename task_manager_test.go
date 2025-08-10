package task_engine_test

import (
	"testing"
	"time"

	engine "github.com/ndizazzo/task-engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TaskManagerTestSuite tests the TaskManager functionality
type TaskManagerTestSuite struct {
	suite.Suite
}

// TestTaskManagerTestSuite runs the TaskManager test suite
func TestTaskManagerTestSuite(t *testing.T) {
	suite.Run(t, new(TaskManagerTestSuite))
}

func (suite *TaskManagerTestSuite) TestAddTask() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "test-task",
		Name:    "Test Task",
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), taskManager.Tasks, "test-task", "TaskManager should contain the added task")
}

func (suite *TaskManagerTestSuite) TestRunTask() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "test-task",
		Name:    "Test Task",
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	err = taskManager.RunTask("test-task")
	assert.NoError(suite.T(), err, "Task should start without errors")
	assert.GreaterOrEqualf(suite.T(), task.GetTotalTime(), time.Duration(0), "Task duration should be greater than or equal to 0")
}

func (suite *TaskManagerTestSuite) TestStopTask() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "test-task",
		Name:    "Test Task",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)
	err = taskManager.RunTask("test-task")
	require.NoError(suite.T(), err)
	err = taskManager.StopTask("test-task")

	assert.NoError(suite.T(), err, "Task should be stopped without errors")
	assert.LessOrEqual(suite.T(), task.GetTotalTime(), LongActionTime, "Task should be stopped before the delay expires")
}

func (suite *TaskManagerTestSuite) TestStopAllTasks() {
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
	require.NoError(suite.T(), err)
	err = taskManager.AddTask(task2)
	require.NoError(suite.T(), err)

	_ = taskManager.RunTask("task-1")
	_ = taskManager.RunTask("task-2")

	time.Sleep(10 * time.Millisecond)
	taskManager.StopAllTasks()

	assert.NotEqual(suite.T(), 100*time.Millisecond, task1.GetTotalTime(), "Task 1 should not complete fully")
	assert.NotEqual(suite.T(), 100*time.Millisecond, task2.GetTotalTime(), "Task 2 should not complete fully")
}

func (suite *TaskManagerTestSuite) TestStopNonRunningTask() {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.StopTask("non-existent-task")
	assert.Error(suite.T(), err, "Stopping a non-existent task should return an error")
}

func (suite *TaskManagerTestSuite) TestAddNilTask() {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.AddTask(nil)
	assert.Error(suite.T(), err, "Adding a nil task should return an error")
}

func (suite *TaskManagerTestSuite) TestRunNonExistentTask() {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.RunTask("non-existent-task")
	assert.Error(suite.T(), err, "Running a non-existent task should return an error")
}

func (suite *TaskManagerTestSuite) TestRunTaskWithFailure() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "test-fail-task",
		Name:    "Test Fail Task",
		Actions: []engine.ActionWrapper{FailingTestAction},
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	// RunTask should not return an error for task execution failures
	// It only returns errors if the task is not found
	err = taskManager.RunTask("test-fail-task")
	assert.NoError(suite.T(), err, "RunTask should not return an error for task execution failures")

	// Wait a bit for the task to start and potentially fail
	time.Sleep(50 * time.Millisecond)

	// The task might complete quickly due to the failing action
	// Check if it's still running or has completed
	isRunning := taskManager.IsTaskRunning("test-fail-task")

	// If the task is still running, stop it
	if isRunning {
		err = taskManager.StopTask("test-fail-task")
		assert.NoError(suite.T(), err, "Task should be stopped without errors")
	} else {
		// Task completed (either successfully or with error), which is also valid
		// No need to stop it
	}
}

func (suite *TaskManagerTestSuite) TestGetRunningTasks() {
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
	require.NoError(suite.T(), err)
	err = taskManager.AddTask(task2)
	require.NoError(suite.T(), err)

	_ = taskManager.RunTask("task-1")
	_ = taskManager.RunTask("task-2")

	time.Sleep(10 * time.Millisecond)

	runningTasks := taskManager.GetRunningTasks()
	assert.Len(suite.T(), runningTasks, 2, "Should have 2 running tasks")
	assert.Contains(suite.T(), runningTasks, "task-1", "Task 1 should be running")
	assert.Contains(suite.T(), runningTasks, "task-2", "Task 2 should be running")

	taskManager.StopAllTasks()
}

func (suite *TaskManagerTestSuite) TestIsTaskRunning() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "test-task",
		Name:    "Test Task",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	// Task should not be running initially
	assert.False(suite.T(), taskManager.IsTaskRunning("test-task"), "Task should not be running initially")

	// Start the task
	err = taskManager.RunTask("test-task")
	require.NoError(suite.T(), err)

	// Task should be running now
	assert.True(suite.T(), taskManager.IsTaskRunning("test-task"), "Task should be running after start")

	// Stop the task
	err = taskManager.StopTask("test-task")
	require.NoError(suite.T(), err)

	// Task should not be running after stop
	assert.False(suite.T(), taskManager.IsTaskRunning("test-task"), "Task should not be running after stop")
}

func (suite *TaskManagerTestSuite) TestGetRunningTasksMultiple() {
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
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task1)
	require.NoError(suite.T(), err)
	err = taskManager.AddTask(task2)
	require.NoError(suite.T(), err)
	err = taskManager.AddTask(task3)
	require.NoError(suite.T(), err)

	_ = taskManager.RunTask("task-1")
	_ = taskManager.RunTask("task-2")
	_ = taskManager.RunTask("task-3")

	time.Sleep(10 * time.Millisecond)

	runningTasks := taskManager.GetRunningTasks()
	// task3 has a single action that completes quickly, so it might not be running
	// We should have at least 2 running tasks (task1 and task2)
	assert.GreaterOrEqual(suite.T(), len(runningTasks), 2, "Should have at least 2 running tasks initially")
	assert.Contains(suite.T(), runningTasks, "task-1", "Task 1 should be running")
	assert.Contains(suite.T(), runningTasks, "task-2", "Task 2 should be running")

	// Wait for task3 to complete (it's a single action)
	time.Sleep(50 * time.Millisecond)

	runningTasks = taskManager.GetRunningTasks()
	assert.Len(suite.T(), runningTasks, 2, "Should have 2 running tasks after task3 completes")
	assert.Contains(suite.T(), runningTasks, "task-1", "Task 1 should still be running")
	assert.Contains(suite.T(), runningTasks, "task-2", "Task 2 should still be running")

	taskManager.StopAllTasks()
}
