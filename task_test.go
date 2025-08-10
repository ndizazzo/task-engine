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
