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

// duplicate NewDiscardLogger removed (defined earlier in file)

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

// testResultProvider is a minimal ResultProvider for tests
type testResultProvider struct{ v interface{} }

func (p testResultProvider) GetResult() interface{} { return p.v }
func (p testResultProvider) GetError() error        { return nil }

// CancelAwareAction returns context error if canceled, otherwise completes after Delay
type CancelAwareAction struct {
	task_engine.BaseAction
	Delay time.Duration
}

func (a *CancelAwareAction) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(a.Delay):
		return nil
	}
}

// NewDiscardLogger creates a new logger that discards all output
// This is useful for tests to prevent log output from cluttering test results
func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
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

func TestTypedGlobalContextHelpers(t *testing.T) {
	gc := task_engine.NewGlobalContext()

	// Prepare action output and result
	gc.StoreActionOutput("act1", map[string]interface{}{"k": 123, "s": "abc"})

	// Simple ResultProviders
	gc.StoreActionResult("actRes", testResultProvider{v: map[string]interface{}{"sum": 7}})
	gc.StoreTaskOutput("task1", map[string]interface{}{"ok": true, "n": 9})
	gc.StoreTaskResult("taskRes", testResultProvider{v: "done"})

	// ActionOutputFieldAs
	vInt, err := task_engine.ActionOutputFieldAs[int](gc, "act1", "k")
	if err != nil || vInt != 123 {
		t.Fatalf("expected 123, got %v, err=%v", vInt, err)
	}
	vStr, err := task_engine.ActionOutputFieldAs[string](gc, "act1", "s")
	if err != nil || vStr != "abc" {
		t.Fatalf("expected 'abc', got %v, err=%v", vStr, err)
	}

	// TaskOutputFieldAs
	vBool, err := task_engine.TaskOutputFieldAs[bool](gc, "task1", "ok")
	if err != nil || vBool != true {
		t.Fatalf("expected true, got %v, err=%v", vBool, err)
	}
	vNum, err := task_engine.TaskOutputFieldAs[int](gc, "task1", "n")
	if err != nil || vNum != 9 {
		t.Fatalf("expected 9, got %v, err=%v", vNum, err)
	}

	// ActionResultAs / TaskResultAs
	rmap, ok := task_engine.ActionResultAs[map[string]interface{}](gc, "actRes")
	if !ok || rmap["sum"].(int) != 7 {
		t.Fatalf("expected action result sum=7, got %v", rmap)
	}
	rstr, ok := task_engine.TaskResultAs[string](gc, "taskRes")
	if !ok || rstr != "done" {
		t.Fatalf("expected task result 'done', got %v", rstr)
	}

	// EntityValue / EntityValueAs
	if v, err := task_engine.EntityValue(gc, "action", "act1", "k"); err != nil || v.(int) != 123 {
		t.Fatalf("EntityValue action k expected 123, got %v, err=%v", v, err)
	}
	if v, err := task_engine.EntityValue(gc, "task", "task1", "ok"); err != nil || v.(bool) != true {
		t.Fatalf("EntityValue task ok expected true, got %v, err=%v", v, err)
	}
	if v, err := task_engine.EntityValue(gc, "action", "actRes", ""); err != nil {
		t.Fatalf("EntityValue action result expected no error, got err=%v", err)
	} else {
		if vm, ok := v.(map[string]interface{}); !ok || vm["sum"].(int) != 7 {
			t.Fatalf("EntityValue action result expected map with sum=7, got %v", v)
		}
	}
	if s, err := task_engine.EntityValueAs[string](gc, "task", "taskRes", ""); err != nil || s != "done" {
		t.Fatalf("EntityValueAs task result expected 'done', got %v, err=%v", s, err)
	}
}

func TestResolveAsGeneric(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	gc.StoreActionOutput("act", map[string]interface{}{"name": "demo", "count": 5})

	name, err := task_engine.ResolveAs[string](context.Background(), task_engine.ActionOutputField("act", "name"), gc)
	if err != nil || name != "demo" {
		t.Fatalf("expected 'demo', got %v, err=%v", name, err)
	}
	count, err := task_engine.ResolveAs[int](context.Background(), task_engine.ActionOutputField("act", "count"), gc)
	if err != nil || count != 5 {
		t.Fatalf("expected 5, got %v, err=%v", count, err)
	}
}

func TestEntityValueNegativePaths(t *testing.T) {
	gc := task_engine.NewGlobalContext()

	if _, err := task_engine.EntityValue(gc, "invalid", "id", ""); err == nil {
		t.Fatalf("expected error for invalid entity type")
	}
	if _, err := task_engine.EntityValue(gc, "action", "missing", ""); err == nil {
		t.Fatalf("expected error for missing action")
	}
	gc.StoreActionOutput("a1", map[string]interface{}{"k": 1})
	if _, err := task_engine.ActionOutputFieldAs[string](gc, "a1", "k"); err == nil {
		t.Fatalf("expected type error for wrong cast")
	}
}

func TestResolveAsNegative(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	gc.StoreActionOutput("a", map[string]interface{}{"x": "str"})
	// wrong type
	if _, err := task_engine.ResolveAs[int](context.Background(), task_engine.ActionOutputField("a", "x"), gc); err == nil {
		t.Fatalf("expected type error for ResolveAs")
	}
}

func TestIDHelpers(t *testing.T) {
	if out := task_engine.SanitizeIDPart(" Hello/World _! "); out == "" {
		t.Fatalf("expected sanitized non-empty id")
	}
	id := task_engine.BuildActionID("prefix", " Part A ", "B/C")
	if id == "" || id == "action-action" {
		t.Fatalf("unexpected id: %s", id)
	}
}

// Task cancellation should still store task output and task result
func TestTaskCancellationStoresOutputAndResult(t *testing.T) {
	logger := NewDiscardLogger()
	gc := task_engine.NewGlobalContext()

	// Task with a quick action and a cancel-aware long-running action
	task := &task_engine.Task{
		ID:   "cancel-task",
		Name: "Cancellation Test",
		Actions: []task_engine.ActionWrapper{
			&task_engine.Action[*DelayAction]{
				ID:      "quick",
				Wrapped: &DelayAction{BaseAction: task_engine.BaseAction{Logger: logger}, Delay: 1 * time.Millisecond},
				Logger:  logger,
			},
			&task_engine.Action[*CancelAwareAction]{
				ID:      "slow",
				Wrapped: &CancelAwareAction{BaseAction: task_engine.BaseAction{Logger: logger}, Delay: 2 * time.Second},
				Logger:  logger,
			},
		},
		Logger: logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// cancel shortly after start
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	_ = task.RunWithContext(ctx, gc)

	// Verify task output and result stored
	if _, ok := gc.TaskOutputs[task.ID]; !ok {
		t.Fatalf("expected TaskOutputs to contain task output on cancellation")
	}
	if _, ok := gc.TaskResults[task.ID]; !ok {
		t.Fatalf("expected TaskResults to contain task result provider on cancellation")
	}
	// Check outputs map for success=false
	out := gc.TaskOutputs[task.ID].(map[string]interface{})
	if out["success"].(bool) {
		t.Fatalf("expected success=false on cancellation")
	}
}

// ResultBuilder error should set task error and mark success=false in outputs
func TestTaskResultBuilderErrorPath(t *testing.T) {
	logger := NewDiscardLogger()
	gc := task_engine.NewGlobalContext()

	errSentinel := errors.New("builder failed")
	builderTask := &task_engine.Task{
		ID:   "builder-error",
		Name: "Builder Error",
		Actions: []task_engine.ActionWrapper{
			&task_engine.Action[*DelayAction]{ID: "noop", Wrapped: &DelayAction{}, Logger: logger},
		},
		Logger: logger,
		ResultBuilder: func(ctx *task_engine.TaskContext) (interface{}, error) {
			return nil, errSentinel
		},
	}

	_ = builderTask.RunWithContext(context.Background(), gc)
	out, ok := gc.TaskOutputs[builderTask.ID]
	if !ok {
		t.Fatalf("expected TaskOutputs to contain output")
	}
	outMap := out.(map[string]interface{})
	if outMap["success"].(bool) {
		t.Fatalf("expected success=false when builder fails")
	}
	// Result should be from task provider with error surfaced in GetResult map
	res, ok := task_engine.TaskResultAs[map[string]interface{}](gc, builderTask.ID)
	if !ok {
		t.Fatalf("expected typed task result from task provider")
	}
	if res["success"].(bool) {
		t.Fatalf("expected task result success=false when builder fails")
	}
}

// Typed helper does not fallback from outputs to results for tasks; verify error
func TestTypedHelperNoFallbackForTaskOutputFieldAs(t *testing.T) {
	gc := task_engine.NewGlobalContext()
	// Only set task result, no task output
	gc.StoreTaskResult("t1", testResultProvider{v: map[string]interface{}{"v": 1}})
	if _, err := task_engine.TaskOutputFieldAs[int](gc, "t1", "v"); err == nil {
		t.Fatalf("expected error since TaskOutputFieldAs should not fallback to results")
	}
	// But EntityValue should fallback to results and succeed (full result)
	if v, err := task_engine.EntityValue(gc, "task", "t1", ""); err != nil {
		t.Fatalf("expected EntityValue to return fallback result, err=%v", err)
	} else {
		if m, ok := v.(map[string]interface{}); !ok || m["v"].(int) != 1 {
			t.Fatalf("unexpected result fallback: %v", v)
		}
	}
	// And with a key, EntityValue should read from result map
	if v, err := task_engine.EntityValue(gc, "task", "t1", "v"); err != nil || v.(int) != 1 {
		t.Fatalf("expected EntityValue with key to read from result map, got %v, err=%v", v, err)
	}
}

// TaskManager timeout and ResetGlobalContext behavior
func TestTaskManagerTimeoutAndResetGlobalContext(t *testing.T) {
	logger := NewDiscardLogger()
	tm := task_engine.NewTaskManager(logger)

	// Long-running task
	task := &task_engine.Task{
		ID:   "timeout-task",
		Name: "Timeout Task",
		Actions: []task_engine.ActionWrapper{
			&task_engine.Action[*DelayAction]{ID: "slow", Wrapped: &DelayAction{Delay: 2 * time.Second}, Logger: logger},
		},
		Logger: logger,
	}
	_ = tm.AddTask(task)
	_ = tm.RunTask("timeout-task")
	// Expect timeout quickly
	if err := tm.WaitForAllTasksToComplete(10 * time.Millisecond); err == nil {
		t.Fatalf("expected timeout error")
	}

	// Store something in current global context
	gc := tm.GetGlobalContext()
	gc.StoreActionOutput("a", "x")
	// Reset and verify cleared
	tm.ResetGlobalContext()
	gc2 := tm.GetGlobalContext()
	if gc2 == gc || len(gc2.ActionOutputs) != 0 || len(gc2.TaskOutputs) != 0 || len(gc2.ActionResults) != 0 || len(gc2.TaskResults) != 0 {
		t.Fatalf("expected a fresh global context after reset")
	}
	// Stop to clean up
	_ = tm.StopTask("timeout-task")
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
