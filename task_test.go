package task_engine_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	engine "github.com/ndizazzo/task-engine"
	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/utility"
	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
)

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

func newMockAction(logger *slog.Logger, name string, returnError error, executed *bool) task_engine.ActionWrapper {
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

func TestTask_Run_Success(t *testing.T) {
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

	assert.NoError(t, err, "Task.Run should not return an error on success")
	assert.True(t, action1Executed, "Action 1 should have been executed")
	assert.True(t, action2Executed, "Action 2 should have been executed")
	assert.Equal(t, 2, task.CompletedTasks, "Completed tasks count should be 2")
}

func TestTask_Run_StopsOnFirstError(t *testing.T) {
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

	assert.ErrorIs(t, err, mockErr, "Task.Run should return the error from the failed action")
	assert.True(t, action1Executed, "Action 1 should have been executed")
	assert.False(t, action2Executed, "Action 2 should NOT have been executed after Action 1 failed")
	assert.Equal(t, 0, task.CompletedTasks, "Completed tasks count should be 0")
}

func TestTask_Run_StopsOnPrerequisiteError(t *testing.T) {
	logger := mocks.NewDiscardLogger()
	prereqExecuted := false
	nextActionExecuted := false

	prereqCheckFunc := func(ctx context.Context, logger *slog.Logger) (abortTask bool, err error) {
		prereqExecuted = true
		return true, nil // Signal to abort task
	}
	prereqAction, err := utility.NewPrerequisiteCheckAction(logger, "Test Prereq Fail", prereqCheckFunc)
	assert.NoError(t, err)

	task := &engine.Task{
		ID:     "test-prereq-fail-task",
		Name:   "Test Prerequisite Fail Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			prereqAction,
			newMockAction(logger, "nextAction", nil, &nextActionExecuted),
		},
	}

	runErr := task.Run(context.Background())

	assert.ErrorIs(t, runErr, engine.ErrPrerequisiteNotMet, "Task.Run should return ErrPrerequisiteNotMet from engine")
	assert.True(t, prereqExecuted, "Prerequisite action should have been executed")
	assert.False(t, nextActionExecuted, "The next action should NOT have been executed after prerequisite check failed")
	assert.Equal(t, 0, task.CompletedTasks, "Completed tasks count should be 0 when prerequisite fails")
}

func TestTask_Run_ContextCancellation(t *testing.T) {
	logger := mocks.NewDiscardLogger()
	action1Executed := false
	action2Executed := false

	task := &engine.Task{
		ID:     "test-cancel-task",
		Name:   "Test Cancel Task",
		Logger: logger,
		Actions: []engine.ActionWrapper{
			// Use a mock action with a delay to allow cancellation
			&engine.Action[*mockAction]{
				ID: "action1-cancel",
				Wrapped: &mockAction{
					BaseAction:   engine.BaseAction{Logger: logger},
					Name:         "action1-cancel",
					ExecuteDelay: time.Second,
					Executed:     &action1Executed,
				},
			},
			newMockAction(logger, "action2-cancel", nil, &action2Executed),
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel the context shortly after starting the task
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := task.Run(ctx)

	assert.ErrorIs(t, err, context.Canceled, "Task.Run should return context.Canceled error")
	// Depending on timing, action1 might start but not finish, Execute flag might be true or false.
	// The crucial part is that action2 should not run.
	assert.False(t, action2Executed, "Action 2 should NOT have been executed after context cancellation")
	assert.Equal(t, 0, task.CompletedTasks, "Completed tasks count should be 0 on cancellation")
}
