package task_engine

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type ActionInterface interface {
	BeforeExecute(ctx context.Context) error
	Execute(ctx context.Context) error
	AfterExecute(ctx context.Context) error
}

type ActionWrapper interface {
	Execute(ctx context.Context) error
	GetDuration() time.Duration
	GetLogger() *slog.Logger
	GetID() string
}

func (a *Action[T]) Execute(ctx context.Context) error {
	return a.InternalExecute(ctx)
}

func (a *Action[T]) GetDuration() time.Duration {
	return a.Duration
}

func (a *Action[T]) GetLogger() *slog.Logger {
	return a.Logger
}

func (a *Action[T]) GetID() string {
	return a.ID
}

// BaseAction is used as a composite struct for newly defined actions, to provide a default no-op implementation of the before/after
// hooks. It also has a logger passed from the action that wraps it.
type BaseAction struct {
	Logger *slog.Logger
}

func (ba *BaseAction) BeforeExecute(ctx context.Context) error {
	return nil
}

func (a *BaseAction) AfterExecute(ctx context.Context) error {
	return nil
}

// ---

type Action[T ActionInterface] struct {
	ID        string
	RunID     string
	Wrapped   T
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Logger    *slog.Logger
}

func (a *Action[T]) InternalExecute(ctx context.Context) error {
	a.RunID = uuid.New().String()
	a.log("Starting action", "actionID", a.ID, "runID", a.RunID)

	if a.StartTime.IsZero() {
		a.StartTime = time.Now()
	}

	if err := a.Wrapped.BeforeExecute(ctx); err != nil {
		a.log("BeforeExecute failed", "actionID", a.ID, "runID", a.RunID, "error", err)
		return err
	}

	if err := a.Wrapped.Execute(ctx); err != nil {
		a.log("Execute failed", "actionID", a.ID, "runID", a.RunID, "error", err)
		return err
	}

	if a.EndTime.IsZero() {
		a.EndTime = time.Now()
	}
	a.Duration = a.EndTime.Sub(a.StartTime)

	if err := a.Wrapped.AfterExecute(ctx); err != nil {
		a.log("AfterExecute failed", "actionID", a.ID, "runID", a.RunID, "error", err)
		return err
	}

	a.log("Action completed", "actionID", a.ID, "runID", a.RunID, "duration", a.Duration)
	return nil
}

func (a *Action[T]) log(message string, keyvals ...interface{}) {
	if a.Logger != nil {
		a.Logger.Info(message, keyvals...)
	}
}
