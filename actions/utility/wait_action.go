package utility

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
)

// NewWaitAction creates an action that waits for a specified duration.
func NewWaitAction(logger *slog.Logger, duration time.Duration) *task_engine.Action[*WaitAction] {
	return &task_engine.Action[*WaitAction]{
		ID: fmt.Sprintf("wait-%s-action", duration),
		Wrapped: &WaitAction{
			BaseAction: task_engine.BaseAction{Logger: logger},
			Duration:   duration,
		},
	}
}

type WaitAction struct {
	task_engine.BaseAction
	Duration time.Duration
}

func (a *WaitAction) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(a.Duration):
		return nil
	}
}
