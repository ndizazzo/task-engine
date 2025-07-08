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

var (
	noOpLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

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
