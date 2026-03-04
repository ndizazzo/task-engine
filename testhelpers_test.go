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

// TestAction is a simple test action that records execution and optionally fails.
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

// DelayAction sleeps for a fixed duration.
type DelayAction struct {
	task_engine.BaseAction
	Delay time.Duration
}

func (a *DelayAction) Execute(ctx context.Context) error {
	time.Sleep(a.Delay)
	return nil
}

// BeforeExecuteFailingAction optionally fails in BeforeExecute.
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

func (a *BeforeExecuteFailingAction) Execute(ctx context.Context) error { return nil }

// AfterExecuteFailingAction optionally fails in AfterExecute.
type AfterExecuteFailingAction struct {
	task_engine.BaseAction
	ShouldFailAfter bool
}

func (a *AfterExecuteFailingAction) BeforeExecute(ctx context.Context) error { return nil }
func (a *AfterExecuteFailingAction) Execute(ctx context.Context) error       { return nil }

// CancelAwareAction returns context error if canceled, otherwise completes after Delay.
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

// testResultProvider is a minimal ResultProvider for tests.
type testResultProvider struct{ v interface{} }

func (p testResultProvider) GetResult() interface{} { return p.v }
func (p testResultProvider) GetError() error        { return nil }

// mockActionWithOutput implements ActionInterface and produces a fixed output value.
type mockActionWithOutput struct {
	task_engine.BaseAction
	output interface{}
}

func (a *mockActionWithOutput) Execute(ctx context.Context) error { return nil }
func (a *mockActionWithOutput) GetOutput() interface{}            { return a.output }

// NewDiscardLogger creates a new logger that discards all output.
func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

var (
	// DiscardLogger is a logger that discards all log output, useful for tests.
	DiscardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

	// noOpLogger is kept for backward compatibility with task_manager_test.go.
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
