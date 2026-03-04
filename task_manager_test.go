package task_engine_test

import (
	"testing"
	"time"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/tasks"
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

func (suite *TaskManagerTestSuite) TestAddTaskDuplicateID() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task1 := &engine.Task{
		ID:      "duplicate-task",
		Name:    "First Task",
		Actions: SingleAction,
	}

	task2 := &engine.Task{
		ID:      "duplicate-task",
		Name:    "Second Task",
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task1)
	require.NoError(suite.T(), err)

	err = taskManager.AddTask(task2)
	assert.Error(suite.T(), err, "Adding task with duplicate ID should return error")
	assert.Contains(suite.T(), err.Error(), "duplicate-task", "Error message should include the duplicate task ID")
	assert.Contains(suite.T(), err.Error(), "already exists", "Error message should indicate task already exists")

	assert.Equal(suite.T(), task1, taskManager.Tasks["duplicate-task"], "Original task should not be overwritten")
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

	handle, err := taskManager.RunTask("test-task")
	assert.NoError(suite.T(), err, "Task should start without errors")
	<-handle.Done()
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
	_, err = taskManager.RunTask("test-task")
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

	handle1, _ := taskManager.RunTask("task-1")
	handle2, _ := taskManager.RunTask("task-2")

	taskManager.StopAllTasks()

	<-handle1.Done()
	<-handle2.Done()

	assert.NotEqual(suite.T(), 100*time.Millisecond, task1.GetTotalTime(), "Task 1 should not complete fully")
	assert.NotEqual(suite.T(), 100*time.Millisecond, task2.GetTotalTime(), "Task 2 should not complete fully")
}

func (suite *TaskManagerTestSuite) TestTaskHandleDoneClosesWhenComplete() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "handle-test-task",
		Name:    "Handle Test Task",
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	handle, err := taskManager.RunTask("handle-test-task")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), handle, "TaskHandle should not be nil")

	// Wait for Done() channel to close
	select {
	case <-handle.Done():
		// Success - channel closed when task completed
	case <-time.After(1 * time.Second):
		suite.Fail("TaskHandle.Done() did not close after task completion")
	}
}

func (suite *TaskManagerTestSuite) TestTaskHandleErrReturnsTaskError() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "handle-fail-task",
		Name:    "Handle Fail Task",
		Actions: []engine.ActionWrapper{FailingTestAction},
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	handle, err := taskManager.RunTask("handle-fail-task")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), handle, "TaskHandle should not be nil")

	// Wait for task to complete
	<-handle.Done()

	// Check error from handle
	taskErr := handle.Err()
	assert.Error(suite.T(), taskErr, "TaskHandle.Err() should return error for failed task")
}

func (suite *TaskManagerTestSuite) TestTaskHandleTaskIDReturnsCorrectID() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "id-test-task",
		Name:    "ID Test Task",
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	handle, err := taskManager.RunTask("id-test-task")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), handle, "TaskHandle should not be nil")

	assert.Equal(suite.T(), "id-test-task", handle.TaskID(), "TaskHandle.TaskID() should return correct task ID")

	// Clean up
	<-handle.Done()
}

func (suite *TaskManagerTestSuite) TestGetRunningTasks() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "running-test",
		Name:    "Running Test",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	running := taskManager.GetRunningTasks()
	assert.Empty(suite.T(), running, "No tasks should be running before RunTask")

	handle, err := taskManager.RunTask("running-test")
	require.NoError(suite.T(), err)

	running = taskManager.GetRunningTasks()
	assert.Contains(suite.T(), running, "running-test", "Running tasks should include the started task")

	_ = taskManager.StopTask("running-test")
	<-handle.Done()

	running = taskManager.GetRunningTasks()
	assert.Empty(suite.T(), running, "No tasks should be running after stop")
}

func (suite *TaskManagerTestSuite) TestIsTaskRunning() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "is-running-test",
		Name:    "Is Running Test",
		Actions: LongRunningActions,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	assert.False(suite.T(), taskManager.IsTaskRunning("is-running-test"), "Task should not be running before RunTask")
	assert.False(suite.T(), taskManager.IsTaskRunning("nonexistent"), "Nonexistent task should not be running")

	handle, err := taskManager.RunTask("is-running-test")
	require.NoError(suite.T(), err)

	assert.True(suite.T(), taskManager.IsTaskRunning("is-running-test"), "Task should be running after RunTask")

	_ = taskManager.StopTask("is-running-test")
	<-handle.Done()

	assert.False(suite.T(), taskManager.IsTaskRunning("is-running-test"), "Task should not be running after stop")
}

func (suite *TaskManagerTestSuite) TestAddTaskNil() {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.AddTask(nil)
	assert.Error(suite.T(), err, "Adding nil task should return error")
	assert.Contains(suite.T(), err.Error(), "nil")
}

func (suite *TaskManagerTestSuite) TestRunTaskNotFound() {
	taskManager := engine.NewTaskManager(noOpLogger)

	handle, err := taskManager.RunTask("nonexistent")
	assert.Error(suite.T(), err, "Running nonexistent task should return error")
	assert.Nil(suite.T(), handle)
	assert.Contains(suite.T(), err.Error(), "nonexistent")
}

func (suite *TaskManagerTestSuite) TestStopTaskNotRunning() {
	taskManager := engine.NewTaskManager(noOpLogger)

	err := taskManager.StopTask("not-running")
	assert.Error(suite.T(), err, "Stopping non-running task should return error")
	assert.Contains(suite.T(), err.Error(), "not running")
}

func (suite *TaskManagerTestSuite) TestTaskHandleSuccessfulTaskNoError() {
	taskManager := engine.NewTaskManager(noOpLogger)

	task := &engine.Task{
		ID:      "success-task",
		Name:    "Success Task",
		Actions: SingleAction,
	}

	err := taskManager.AddTask(task)
	require.NoError(suite.T(), err)

	handle, err := taskManager.RunTask("success-task")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), handle, "TaskHandle should not be nil")

	// Wait for task to complete
	<-handle.Done()

	// Check no error for successful task
	taskErr := handle.Err()
	assert.NoError(suite.T(), taskErr, "TaskHandle.Err() should return nil for successful task")
}

func TestTaskManagerTimeoutAndResetGlobalContext(t *testing.T) {
	logger := NewDiscardLogger()
	tm := engine.NewTaskManager(logger)

	task := &engine.Task{
		ID:   "timeout-task",
		Name: "Timeout Task",
		Actions: []engine.ActionWrapper{
			&engine.Action[*DelayAction]{ID: "slow", Wrapped: &DelayAction{Delay: 2 * time.Second}, Logger: logger},
		},
		Logger: logger,
	}
	_ = tm.AddTask(task)
	_, _ = tm.RunTask("timeout-task")
	if err := tm.WaitForAllTasksToComplete(10 * time.Millisecond); err == nil {
		t.Fatalf("expected timeout error")
	}

	gc := tm.GetGlobalContext()
	gc.StoreActionOutput("a", "x")
	tm.ResetGlobalContext()
	gc2 := tm.GetGlobalContext()
	if gc2 == gc || len(gc2.ActionOutputs) != 0 || len(gc2.TaskOutputs) != 0 || len(gc2.ActionResults) != 0 || len(gc2.TaskResults) != 0 {
		t.Fatalf("expected a fresh global context after reset")
	}
	_ = tm.StopTask("timeout-task")
}

func TestTaskWithParameterPassing(t *testing.T) {
	t.Run("TaskExecutionWithGlobalContext", func(t *testing.T) {
		logger := NewDiscardLogger()
		tm := engine.NewTaskManager(logger)

		task := &engine.Task{
			ID:   "test-task",
			Name: "Test Task",
			Actions: []engine.ActionWrapper{
				&engine.Action[engine.ActionInterface]{
					ID: "test-action",
					Wrapped: &mockActionWithOutput{
						BaseAction: engine.BaseAction{Logger: logger},
						output:     "test output",
					},
					Logger: logger,
				},
			},
			Logger: logger,
		}

		err := tm.AddTask(task)
		if err != nil {
			t.Fatalf("Expected no error adding task, got %v", err)
		}

		handle, err := tm.RunTask("test-task")
		if err != nil {
			t.Fatalf("Expected no error running task, got %v", err)
		}
		<-handle.Done()
		if err := handle.Err(); err != nil {
			t.Fatalf("Task execution failed: %v", err)
		}
		globalContext := tm.GetGlobalContext()
		output, exists := globalContext.ActionOutputs["test-action"]
		if !exists {
			t.Fatal("Expected action output to exist in global context")
		}
		if output != "test output" {
			t.Fatalf("Expected 'test output', got %v", output)
		}
	})
}

func TestExampleParameterPassingTask(t *testing.T) {
	t.Run("ExampleParameterPassingTask", func(t *testing.T) {
		logger := NewDiscardLogger()
		tm := engine.NewTaskManager(logger)

		config := tasks.ExampleParameterPassingConfig{
			SourcePath:      "testing/testdata/test.txt",
			DestinationPath: "testing/testdata/output.txt",
		}

		task := tasks.NewExampleParameterPassingTask(config, logger)

		t.Logf("Task created with ID: %s", task.ID)
		t.Logf("Task has %d actions", len(task.Actions))
		for i, action := range task.Actions {
			t.Logf("Action %d: ID=%s, Type=%T", i, action.GetID(), action)
			if actionWithOutput, ok := action.(interface{ GetOutput() interface{} }); ok {
				t.Logf("Action %d implements GetOutput", i)
				output := actionWithOutput.GetOutput()
				t.Logf("Action %d GetOutput() returns: %+v", i, output)
			} else {
				t.Logf("Action %d does NOT implement GetOutput", i)
			}
		}

		err := tm.AddTask(task)
		if err != nil {
			t.Fatalf("Expected no error adding task, got %v", err)
		}

		handle, err := tm.RunTask("example-parameter-passing")
		if err != nil {
			t.Fatalf("Expected no error running task, got %v", err)
		}
		<-handle.Done()
		if err := handle.Err(); err != nil {
			t.Fatalf("Task execution failed: %v", err)
		}
		globalContext := tm.GetGlobalContext()

		t.Logf("All action outputs in global context: %+v", globalContext.ActionOutputs)
		t.Logf("All action results in global context: %+v", globalContext.ActionResults)
		readOutput, exists := globalContext.ActionOutputs["read-source-file"]
		if !exists {
			t.Fatal("Expected read action output to exist in global context")
		}
		writeOutput, exists := globalContext.ActionOutputs["write-destination-file"]
		if !exists {
			t.Fatal("Expected write action output to exist in global context")
		}
		if readOutput == nil {
			t.Fatal("Expected read action output to not be nil")
		}
		if writeOutput == nil {
			t.Fatal("Expected write action output to not be nil")
		}

		t.Logf("Read action output: %+v", readOutput)
		t.Logf("Write action output: %+v", writeOutput)
	})
}
