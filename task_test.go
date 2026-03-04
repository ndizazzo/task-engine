package task_engine_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/utility"
	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TaskTestSuite tests the Task functionality
type TaskTestSuite struct {
	suite.Suite
}

// TestTaskTestSuite runs the Task test suite
func TestTaskTestSuite(t *testing.T) {
	suite.Run(t, new(TaskTestSuite))
}

// mockAction is a simple action for testing task execution flow.
type mockAction struct {
	engine.BaseAction
	Name         string
	ReturnError  error
	ExecuteDelay time.Duration
	Executed     *bool
}

// Execute implements the Action interface.
func (a *mockAction) Execute(ctx context.Context) error {
	if a.Executed != nil {
		*a.Executed = true
	}
	if a.ExecuteDelay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(a.ExecuteDelay):
		}
	}
	return a.ReturnError
}

func newMockAction(logger *slog.Logger, name string, returnError error, executed *bool) engine.ActionWrapper {
	return &engine.Action[*mockAction]{
		ID: name,
		Wrapped: &mockAction{
			BaseAction:  engine.BaseAction{Logger: logger},
			Name:        name,
			ReturnError: returnError,
			Executed:    executed,
		},
	}
}

// outputAction is a simple action that returns a fixed output map
type outputAction struct {
	engine.BaseAction
	Output map[string]interface{}
}

func (a *outputAction) Execute(ctx context.Context) error { return nil }
func (a *outputAction) GetOutput() interface{}            { return a.Output }

// overrideTaskResultAction sets a task-level ResultProvider during execution
type overrideTaskResultAction struct {
	engine.BaseAction
	TaskID   string
	Provider *mocks.ResultProviderMock
}

func (a *overrideTaskResultAction) Execute(ctx context.Context) error {
	if gc, ok := ctx.Value(engine.GlobalContextKey).(*engine.GlobalContext); ok {
		gc.StoreTaskResult(a.TaskID, a.Provider)
	}
	return nil
}
func (a *overrideTaskResultAction) GetOutput() interface{} { return nil }

type testResult struct {
	Value string
}

func (suite *TaskTestSuite) TestRun_Success() {
	logger := mocks.NewDiscardLogger()
	action1Executed := false
	action2Executed := false

	task := &engine.Task{
		ID:     "test-success-task",
		Name:   "Test Success Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			newMockAction(logger, "action1", nil, &action1Executed),
			newMockAction(logger, "action2", nil, &action2Executed),
		},
	}

	err := task.Run(context.Background())

	assert.NoError(suite.T(), err, "Task.Run should not return an error on success")
	assert.True(suite.T(), action1Executed, "Action 1 should have been executed")
	assert.True(suite.T(), action2Executed, "Action 2 should have been executed")
	assert.Equal(suite.T(), 2, task.CompletedTasks, "Completed tasks count should be 2")
}

func (suite *TaskTestSuite) TestRun_StopsOnFirstError() {
	logger := mocks.NewDiscardLogger()
	action1Executed := false
	action2Executed := false
	mockErr := errors.New("action 1 failed")

	task := &engine.Task{
		ID:     "test-fail-task",
		Name:   "Test Fail Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			newMockAction(logger, "action1", mockErr, &action1Executed),
			newMockAction(logger, "action2", nil, &action2Executed),
		},
	}

	err := task.Run(context.Background())

	assert.ErrorIs(suite.T(), err, mockErr, "Task.Run should return the error from the failed action")
	assert.True(suite.T(), action1Executed, "Action 1 should have been executed")
	assert.False(suite.T(), action2Executed, "Action 2 should NOT have been executed after Action 1 failed")
	assert.Equal(suite.T(), 0, task.CompletedTasks, "Completed tasks count should be 0")
}

func (suite *TaskTestSuite) TestRun_StopsOnPrerequisiteError() {
	logger := mocks.NewDiscardLogger()
	action1Executed := false
	action2Executed := false

	// Create a prerequisite check action that fails
	prereqAction := &utility.PrerequisiteCheckAction{
		BaseAction: engine.BaseAction{Logger: logger},
		Check: func(ctx context.Context, logger *slog.Logger) (abortTask bool, err error) {
			return true, errors.New("prerequisite check failed")
		},
	}

	task := &engine.Task{
		ID:     "test-prereq-fail-task",
		Name:   "Test Prereq Fail Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			&engine.Action[*utility.PrerequisiteCheckAction]{
				ID:      "prereq-check",
				Wrapped: prereqAction,
			},
			newMockAction(logger, "action1", nil, &action1Executed),
			newMockAction(logger, "action2", nil, &action2Executed),
		},
	}

	err := task.Run(context.Background())

	assert.Error(suite.T(), err, "Task.Run should return an error when prerequisite check fails")
	assert.Contains(suite.T(), err.Error(), "prerequisite check failed", "Error should contain prerequisite failure message")
	assert.False(suite.T(), action1Executed, "Action 1 should NOT have been executed after prerequisite failure")
	assert.False(suite.T(), action2Executed, "Action 2 should NOT have been executed after prerequisite failure")
	assert.Equal(suite.T(), 0, task.CompletedTasks, "Completed tasks count should be 0")
}

func (suite *TaskTestSuite) TestRun_ContextCancellation() {
	logger := mocks.NewDiscardLogger()
	action1Executed := false
	action2Executed := false

	task := &engine.Task{
		ID:     "test-context-cancel-task",
		Name:   "Test Context Cancel Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			newMockAction(logger, "action1", nil, &action1Executed),
			&engine.Action[*mockAction]{
				ID: "action2",
				Wrapped: &mockAction{
					BaseAction:   engine.BaseAction{Logger: logger},
					Name:         "action2",
					ReturnError:  nil,
					Executed:     &action2Executed,
					ExecuteDelay: 100 * time.Millisecond, // Add delay so context cancellation can be detected
				},
			},
		},
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start the task in a goroutine and wait for completion deterministically
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = task.Run(ctx)
	}()

	// Wait a bit for the first action to start and complete
	time.Sleep(10 * time.Millisecond)

	// Cancel the context
	cancel()

	// Wait for the task to finish
	wg.Wait()

	// The first action should have been executed, but the second might not
	assert.True(suite.T(), action1Executed, "Action 1 should have been executed before cancellation")
	// Note: action2 execution depends on timing, so we don't assert on it
	// During cancellation, we might have 0 or 1 completed tasks depending on timing
	assert.LessOrEqual(suite.T(), task.GetCompletedTasks(), 1, "Completed tasks should be 0 or 1 due to cancellation")
}

func (suite *TaskTestSuite) TestTask_ImplementsTaskWithResults_AndBuilderStoresCustomStruct() {
	logger := mocks.NewDiscardLogger()
	gc := engine.NewGlobalContext()

	// Action that produces an output used by the builder
	producer := &engine.Action[*outputAction]{
		ID: "produce",
		Wrapped: &outputAction{
			BaseAction: engine.BaseAction{Logger: logger},
			Output:     map[string]interface{}{"field": "value-from-action"},
		},
	}

	// Task with a ResultBuilder that pulls from action output
	task := &engine.Task{
		ID:     "builder-task",
		Name:   "Builder Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			producer,
		},
		ResultBuilder: func(ctx *engine.TaskContext) (interface{}, error) {
			v, err := engine.ActionOutputField("produce", "field").Resolve(context.Background(), ctx.GlobalContext)
			if err != nil {
				return nil, err
			}
			s, _ := v.(string)
			return &testResult{Value: s}, nil
		},
	}

	// Compile-time-ish and runtime checks for TaskWithResults
	if _, ok := any(task).(engine.TaskWithResults); !ok {
		suite.T().Fatal("Task should implement TaskWithResults")
	}

	// Run and verify
	assert.NoError(suite.T(), task.RunWithContext(context.Background(), gc))

	// Fetch provider stored on task
	rp, exists := gc.TaskResults[task.ID]
	assert.True(suite.T(), exists)
	res, ok := rp.GetResult().(*testResult)
	if assert.True(suite.T(), ok) {
		assert.Equal(suite.T(), "value-from-action", res.Value)
	}

	// Also verify TaskResult parameter resolves the entire struct
	val, err := engine.TaskResult(task.ID).Resolve(context.Background(), gc)
	assert.NoError(suite.T(), err)
	_, ok = val.(*testResult)
	assert.True(suite.T(), ok)
}

func (suite *TaskTestSuite) TestTask_ActionOverrideWinsOverBuilder() {
	logger := mocks.NewDiscardLogger()
	gc := engine.NewGlobalContext()

	// Provider to override task result
	p := mocks.NewResultProviderMock()
	p.SetResult("override-result")
	// Ensure mock has expectations to satisfy testify's Called()
	p.On("GetResult").Return("override-result")
	p.On("GetError").Return(nil)

	override := &engine.Action[*overrideTaskResultAction]{
		ID: "override",
		Wrapped: &overrideTaskResultAction{
			BaseAction: engine.BaseAction{Logger: logger},
			TaskID:     "override-task",
			Provider:   p,
		},
	}

	// Builder returns a different value, which should NOT replace override
	task := &engine.Task{
		ID:     "override-task",
		Name:   "Override Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			override,
		},
		ResultBuilder: func(ctx *engine.TaskContext) (interface{}, error) {
			return &testResult{Value: "builder-result"}, nil
		},
	}

	assert.NoError(suite.T(), task.RunWithContext(context.Background(), gc))

	rp, exists := gc.TaskResults[task.ID]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), "override-result", rp.GetResult())

	// Parameter-based fetch of entire result
	val, err := engine.TaskResult(task.ID).Resolve(context.Background(), gc)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "override-result", val)
}

// Minimal pattern test:
// - two actions produce a string and an int as outputs
// - the task aggregates into a single struct via ResultBuilder
// - test fetches the struct from task results and asserts values
func (suite *TaskTestSuite) TestTask_SimpleResultAggregation() {
	logger := mocks.NewDiscardLogger()
	gc := engine.NewGlobalContext()

	strAction := &engine.Action[*outputAction]{
		ID: "string-action",
		Wrapped: &outputAction{
			BaseAction: engine.BaseAction{Logger: logger},
			Output:     map[string]interface{}{"text": "hello"},
		},
	}

	intAction := &engine.Action[*outputAction]{
		ID: "int-action",
		Wrapped: &outputAction{
			BaseAction: engine.BaseAction{Logger: logger},
			Output:     map[string]interface{}{"num": 42},
		},
	}

	type simpleResult struct {
		Text string
		Num  int
	}

	task := &engine.Task{
		ID:     "simple-aggregate-task",
		Name:   "Simple Aggregate",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			strAction,
			intAction,
		},
		ResultBuilder: func(ctx *engine.TaskContext) (interface{}, error) {
			gc := ctx.GlobalContext
			res := &simpleResult{}

			if v, err := engine.ActionOutputField("string-action", "text").Resolve(context.Background(), gc); err == nil {
				if s, ok := v.(string); ok {
					res.Text = s
				}
			}
			if v, err := engine.ActionOutputField("int-action", "num").Resolve(context.Background(), gc); err == nil {
				// handle both int and numeric types gracefully for the test
				switch n := v.(type) {
				case int:
					res.Num = n
				case int32:
					res.Num = int(n)
				case int64:
					res.Num = int(n)
				case float64:
					res.Num = int(n)
				}
			}

			return res, nil
		},
	}

	assert.NoError(suite.T(), task.RunWithContext(context.Background(), gc))

	rp, exists := gc.TaskResults[task.ID]
	assert.True(suite.T(), exists)
	if out, ok := rp.GetResult().(*simpleResult); ok {
		assert.Equal(suite.T(), "hello", out.Text)
		assert.Equal(suite.T(), 42, out.Num)
	} else {
		suite.T().Fatal("unexpected result type")
	}
}

// TestValidateActionParameters_DuplicateActionIDs verifies that duplicate action IDs are detected
func (suite *TaskTestSuite) TestValidateActionParameters_DuplicateActionIDs() {
	logger := mocks.NewDiscardLogger()

	task := &engine.Task{
		ID:     "test-duplicate-ids-task",
		Name:   "Test Duplicate IDs Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			newMockAction(logger, "action-duplicate", nil, nil),
			newMockAction(logger, "action-duplicate", nil, nil),
		},
	}

	err := task.Run(context.Background())

	assert.Error(suite.T(), err, "Task.Run should return an error when duplicate action IDs are found")
	assert.Contains(suite.T(), err.Error(), "duplicate action ID", "Error should mention duplicate action ID")
	assert.Contains(suite.T(), err.Error(), "action-duplicate", "Error should contain the duplicate ID")
}

// TestValidateActionParameters_ValidUniqueActionIDs verifies that tasks with unique action IDs pass validation
func (suite *TaskTestSuite) TestValidateActionParameters_ValidUniqueActionIDs() {
	logger := mocks.NewDiscardLogger()
	action1Executed := false
	action2Executed := false
	action3Executed := false

	task := &engine.Task{
		ID:     "test-unique-ids-task",
		Name:   "Test Unique IDs Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			newMockAction(logger, "action-1", nil, &action1Executed),
			newMockAction(logger, "action-2", nil, &action2Executed),
			newMockAction(logger, "action-3", nil, &action3Executed),
		},
	}

	err := task.Run(context.Background())

	assert.NoError(suite.T(), err, "Task.Run should succeed with unique action IDs")
	assert.True(suite.T(), action1Executed, "Action 1 should have been executed")
	assert.True(suite.T(), action2Executed, "Action 2 should have been executed")
	assert.True(suite.T(), action3Executed, "Action 3 should have been executed")
	assert.Equal(suite.T(), 3, task.CompletedTasks, "All 3 actions should be completed")
}

// TestValidateActionParameters_EmptyActionID verifies that empty action IDs are rejected
func (suite *TaskTestSuite) TestValidateActionParameters_EmptyActionID() {
	logger := mocks.NewDiscardLogger()

	emptyIDAction := &engine.Action[*mockAction]{
		ID: "", // Empty ID
		Wrapped: &mockAction{
			BaseAction: engine.BaseAction{Logger: logger},
			Name:       "empty-id-action",
		},
	}

	task := &engine.Task{
		ID:      "test-empty-id-task",
		Name:    "Test Empty ID Task",
		Logger:  logger,
		Actions: []engine.ActionWrapper{emptyIDAction},
	}

	err := task.Run(context.Background())

	assert.Error(suite.T(), err, "Task.Run should return an error when an action has an empty ID")
	assert.Contains(suite.T(), err.Error(), "empty ID", "Error should mention empty ID")
}

// TestValidateActionParameters_MultipleActionsMultipleDuplicates verifies detection of duplicates across multiple actions
func (suite *TaskTestSuite) TestValidateActionParameters_MultipleActionsMultipleDuplicates() {
	logger := mocks.NewDiscardLogger()

	task := &engine.Task{
		ID:     "test-multi-duplicate-task",
		Name:   "Test Multiple Duplicates Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			newMockAction(logger, "action-a", nil, nil),
			newMockAction(logger, "action-b", nil, nil),
			newMockAction(logger, "action-a", nil, nil), // Duplicate of first
		},
	}

	err := task.Run(context.Background())

	assert.Error(suite.T(), err, "Task.Run should return an error when duplicate IDs are found")
	assert.Contains(suite.T(), err.Error(), "duplicate action ID", "Error should mention duplicate action ID")
	assert.Contains(suite.T(), err.Error(), "action-a", "Error should contain the duplicate ID")
	assert.Contains(suite.T(), err.Error(), "index 0", "Error should indicate first occurrence")
	assert.Contains(suite.T(), err.Error(), "and 2", "Error should indicate duplicate occurrence")
}

func TestTaskGetIDAndGetName(t *testing.T) {
	task := &engine.Task{
		ID:   "my-task-id",
		Name: "My Task Name",
	}
	if task.GetID() != "my-task-id" {
		t.Fatalf("expected 'my-task-id', got %q", task.GetID())
	}
	if task.GetName() != "My Task Name" {
		t.Fatalf("expected 'My Task Name', got %q", task.GetName())
	}
}

// Task cancellation should still store task output and task result
func TestTaskCancellationStoresOutputAndResult(t *testing.T) {
	logger := NewDiscardLogger()
	gc := engine.NewGlobalContext()

	task := &engine.Task{
		ID:   "cancel-task",
		Name: "Cancellation Test",
		Actions: []engine.ActionWrapper{
			&engine.Action[*DelayAction]{
				ID:      "quick",
				Wrapped: &DelayAction{BaseAction: engine.BaseAction{Logger: logger}, Delay: 1 * time.Millisecond},
				Logger:  logger,
			},
			&engine.Action[*CancelAwareAction]{
				ID:      "slow",
				Wrapped: &CancelAwareAction{BaseAction: engine.BaseAction{Logger: logger}, Delay: 2 * time.Second},
				Logger:  logger,
			},
		},
		Logger: logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	_ = task.RunWithContext(ctx, gc)

	if _, ok := gc.TaskOutputs[task.ID]; !ok {
		t.Fatalf("expected TaskOutputs to contain task output on cancellation")
	}
	if _, ok := gc.TaskResults[task.ID]; !ok {
		t.Fatalf("expected TaskResults to contain task result provider on cancellation")
	}
	out := gc.TaskOutputs[task.ID].(map[string]interface{})
	if out["success"].(bool) {
		t.Fatalf("expected success=false on cancellation")
	}
}

// ResultBuilder error should set task error and mark success=false in outputs
func TestTaskResultBuilderErrorPath(t *testing.T) {
	logger := NewDiscardLogger()
	gc := engine.NewGlobalContext()

	errSentinel := errors.New("builder failed")
	builderTask := &engine.Task{
		ID:   "builder-error",
		Name: "Builder Error",
		Actions: []engine.ActionWrapper{
			&engine.Action[*DelayAction]{ID: "noop", Wrapped: &DelayAction{}, Logger: logger},
		},
		Logger: logger,
		ResultBuilder: func(ctx *engine.TaskContext) (interface{}, error) {
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
	res, ok := engine.TaskResultAs[map[string]interface{}](gc, builderTask.ID)
	if !ok {
		t.Fatalf("expected typed task result from task provider")
	}
	if res["success"].(bool) {
		t.Fatalf("expected task result success=false when builder fails")
	}
}
