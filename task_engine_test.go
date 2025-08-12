package task_engine_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/tasks"
)

const (
	StaticActionTime = 10 * time.Millisecond
	LongActionTime   = 500 * time.Millisecond
)

// NewDiscardLogger creates a new logger that discards all output
// This is useful for tests to prevent log output from cluttering test results
func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type TestAction struct {
	task_engine.BaseAction
	Called     bool
	ShouldFail bool
}

func (a *TestAction) Execute(ctx context.Context) error {
	a.Called = true

	if a.ShouldFail {
		return errors.New("simulated failure")
	}

	return nil
}

type DelayAction struct {
	task_engine.BaseAction
	Delay time.Duration
}

func (a *DelayAction) Execute(ctx context.Context) error {
	time.Sleep(a.Delay)
	return nil
}

type BeforeExecuteFailingAction struct {
	task_engine.BaseAction
	ShouldFailBefore bool
}

func (a *BeforeExecuteFailingAction) BeforeExecute(ctx context.Context) error {
	if a.ShouldFailBefore {
		return errors.New("simulated BeforeExecute failure")
	}
	return nil
}

func (a *BeforeExecuteFailingAction) Execute(ctx context.Context) error {
	return nil
}

type AfterExecuteFailingAction struct {
	task_engine.BaseAction
	ShouldFailAfter bool
}

func (a *AfterExecuteFailingAction) BeforeExecute(ctx context.Context) error {
	return nil
}

func (a *AfterExecuteFailingAction) Execute(ctx context.Context) error {
	return nil
}

func (a *AfterExecuteFailingAction) AfterExecute(ctx context.Context) error {
	if a.ShouldFailAfter {
		return errors.New("simulated AfterExecute failure")
	}
	return nil
}

var (
	// DiscardLogger is a logger that discards all log output, useful for tests
	DiscardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

	// noOpLogger is kept for backward compatibility
	noOpLogger = DiscardLogger

	PassingTestAction = &task_engine.Action[*TestAction]{
		ID: "passing-action-1",
		Wrapped: &TestAction{
			BaseAction: task_engine.BaseAction{},
			Called:     false,
		},
	}

	FailingTestAction = &task_engine.Action[*TestAction]{
		ID: "failing-action-1",
		Wrapped: &TestAction{
			BaseAction: task_engine.BaseAction{},
			ShouldFail: true,
		},
	}

	LongRunningAction = &task_engine.Action[*DelayAction]{
		ID: "long-running-action",
		Wrapped: &DelayAction{
			BaseAction: task_engine.BaseAction{},
			Delay:      LongActionTime,
		},
	}

	BeforeExecuteFailingTestAction = &task_engine.Action[*BeforeExecuteFailingAction]{
		ID: "before-execute-failing-action",
		Wrapped: &BeforeExecuteFailingAction{
			BaseAction:       task_engine.BaseAction{},
			ShouldFailBefore: true,
		},
	}

	AfterExecuteFailingTestAction = &task_engine.Action[*AfterExecuteFailingAction]{
		ID: "after-execute-failing-action",
		Wrapped: &AfterExecuteFailingAction{
			BaseAction:      task_engine.BaseAction{},
			ShouldFailAfter: true,
		},
	}

	SingleAction = []task_engine.ActionWrapper{
		PassingTestAction,
	}

	MultipleActionsSuccess = []task_engine.ActionWrapper{
		PassingTestAction,
		PassingTestAction,
	}

	MultipleActionsFailure = []task_engine.ActionWrapper{
		PassingTestAction,
		FailingTestAction,
	}

	LongRunningActions = []task_engine.ActionWrapper{
		LongRunningAction,
	}

	ManyTasksForCancellation = []task_engine.ActionWrapper{
		LongRunningAction,
		PassingTestAction,
		PassingTestAction,
		LongRunningAction,
	}
)

func TestParameterPassingSystem(t *testing.T) {
	t.Run("StaticParameter", func(t *testing.T) {
		staticParam := task_engine.StaticParameter{Value: "test value"}
		globalContext := task_engine.NewGlobalContext()

		result, err := staticParam.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "test value" {
			t.Fatalf("Expected 'test value', got %v", result)
		}
	})
	t.Run("ActionOutputParameter", func(t *testing.T) {
		globalContext := task_engine.NewGlobalContext()
		globalContext.StoreActionOutput("test-action", map[string]interface{}{
			"content": "file content",
			"size":    12,
		})
		param := task_engine.ActionOutputParameter{ActionID: "test-action"}
		result, err := param.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		expected := map[string]interface{}{
			"content": "file content",
			"size":    12,
		}
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("Expected %v, got %v", expected, result)
		}
		paramWithKey := task_engine.ActionOutputParameter{ActionID: "test-action", OutputKey: "content"}
		result, err = paramWithKey.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "file content" {
			t.Fatalf("Expected 'file content', got %v", result)
		}
	})
	t.Run("TaskOutputParameter", func(t *testing.T) {
		globalContext := task_engine.NewGlobalContext()
		globalContext.StoreTaskOutput("test-task", map[string]interface{}{
			"result": "task result",
			"status": "completed",
		})

		param := task_engine.TaskOutputParameter{TaskID: "test-task", OutputKey: "result"}
		result, err := param.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "task result" {
			t.Fatalf("Expected 'task result', got %v", result)
		}
	})
	t.Run("EntityOutputParameter", func(t *testing.T) {
		globalContext := task_engine.NewGlobalContext()
		globalContext.StoreActionOutput("test-action", "action output")
		globalContext.StoreTaskOutput("test-task", "task output")
		actionParam := task_engine.EntityOutputParameter{EntityType: "action", EntityID: "test-action"}
		result, err := actionParam.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "action output" {
			t.Fatalf("Expected 'action output', got %v", result)
		}
		taskParam := task_engine.EntityOutputParameter{EntityType: "task", EntityID: "test-task"}
		result, err = taskParam.Resolve(context.Background(), globalContext)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result != "task output" {
			t.Fatalf("Expected 'task output', got %v", result)
		}
	})
	t.Run("HelperFunctions", func(t *testing.T) {
		param1 := task_engine.ActionOutput("test-action")
		if param1.ActionID != "test-action" {
			t.Fatalf("Expected ActionID 'test-action', got %s", param1.ActionID)
		}
		if param1.OutputKey != "" {
			t.Fatalf("Expected empty OutputKey, got %s", param1.OutputKey)
		}
		param2 := task_engine.ActionOutputField("test-action", "content")
		if param2.ActionID != "test-action" {
			t.Fatalf("Expected ActionID 'test-action', got %s", param2.ActionID)
		}
		if param2.OutputKey != "content" {
			t.Fatalf("Expected OutputKey 'content', got %s", param2.OutputKey)
		}
		param3 := task_engine.TaskOutput("test-task")
		if param3.TaskID != "test-task" {
			t.Fatalf("Expected TaskID 'test-task', got %s", param3.TaskID)
		}
		if param3.OutputKey != "" {
			t.Fatalf("Expected empty OutputKey, got %s", param3.OutputKey)
		}
	})
}

func TestGlobalContext(t *testing.T) {
	t.Run("GlobalContextOperations", func(t *testing.T) {
		gc := task_engine.NewGlobalContext()
		gc.StoreActionOutput("action1", "output1")
		if gc.ActionOutputs["action1"] != "output1" {
			t.Fatalf("Expected 'output1', got %v", gc.ActionOutputs["action1"])
		}
		gc.StoreTaskOutput("task1", "output1")
		if gc.TaskOutputs["task1"] != "output1" {
			t.Fatalf("Expected 'output1', got %v", gc.TaskOutputs["task1"])
		}
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(id int) {
				gc.StoreActionOutput(fmt.Sprintf("action%d", id), fmt.Sprintf("output%d", id))
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}
		for i := 0; i < 10; i++ {
			expected := fmt.Sprintf("output%d", i)
			actual := gc.ActionOutputs[fmt.Sprintf("action%d", i)]
			if actual != expected {
				t.Fatalf("Expected %s, got %v", expected, actual)
			}
		}
	})
}

func TestTaskWithParameterPassing(t *testing.T) {
	t.Run("TaskExecutionWithGlobalContext", func(t *testing.T) {
		logger := NewDiscardLogger()
		// Create a task manager with global context
		tm := task_engine.NewTaskManager(logger)

		// Create a simple task
		task := &task_engine.Task{
			ID:   "test-task",
			Name: "Test Task",
			Actions: []task_engine.ActionWrapper{
				&task_engine.Action[task_engine.ActionInterface]{
					ID: "test-action",
					Wrapped: &mockActionWithOutput{
						BaseAction: task_engine.BaseAction{Logger: logger},
						output:     "test output",
					},
					Logger: logger,
				},
			},
			Logger: logger,
		}

		// Add and run the task
		err := tm.AddTask(task)
		if err != nil {
			t.Fatalf("Expected no error adding task, got %v", err)
		}

		err = tm.RunTask("test-task")
		if err != nil {
			t.Fatalf("Expected no error running task, got %v", err)
		}

		// Wait for task to complete
		err = tm.WaitForAllTasksToComplete(5 * time.Second)
		if err != nil {
			t.Fatalf("Expected no error waiting for task completion, got %v", err)
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

// TestExampleParameterPassingTask tests the example task that demonstrates parameter passing
func TestExampleParameterPassingTask(t *testing.T) {
	t.Run("ExampleParameterPassingTask", func(t *testing.T) {
		logger := NewDiscardLogger()

		// Create a task manager
		tm := task_engine.NewTaskManager(logger)

		// Create the example parameter passing task
		config := tasks.ExampleParameterPassingConfig{
			SourcePath:      "testing/testdata/test.txt",
			DestinationPath: "testing/testdata/output.txt",
		}

		task := tasks.NewExampleParameterPassingTask(config, logger)

		// Debug: Print task structure
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

		// Add and run the task
		err := tm.AddTask(task)
		if err != nil {
			t.Fatalf("Expected no error adding task, got %v", err)
		}

		err = tm.RunTask("example-parameter-passing")
		if err != nil {
			t.Fatalf("Expected no error running task, got %v", err)
		}

		// Wait for task to complete
		err = tm.WaitForAllTasksToComplete(5 * time.Second)
		if err != nil {
			t.Fatalf("Expected no error waiting for task completion, got %v", err)
		}
		globalContext := tm.GetGlobalContext()

		// Debug: Print all action outputs
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

// Mock action that implements ActionInterface and produces output
type mockActionWithOutput struct {
	task_engine.BaseAction
	output interface{}
}

func (a *mockActionWithOutput) Execute(ctx context.Context) error {
	return nil
}

func (a *mockActionWithOutput) GetOutput() interface{} {
	return a.output
}
