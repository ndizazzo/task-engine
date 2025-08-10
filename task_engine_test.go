package task_engine_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
)

const (
	StaticActionTime = 10 * time.Millisecond
	LongActionTime   = 500 * time.Millisecond
)

// Explicit reusable context
func testContext() context.Context {
	return context.Background()
}

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
