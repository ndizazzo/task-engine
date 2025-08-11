package utility

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
)

// WaitAction represents an action that waits for a specified duration
type WaitAction struct {
	task_engine.BaseAction
	// Parameter fields
	DurationParam task_engine.ActionParameter
}

// NewWaitAction creates a new WaitAction with the provided logger
func NewWaitAction(logger *slog.Logger) *WaitAction {
	return &WaitAction{
		BaseAction: task_engine.BaseAction{Logger: logger},
	}
}

// WithParameters sets the duration parameter and returns the wrapped action
func (a *WaitAction) WithParameters(durationParam task_engine.ActionParameter) (*task_engine.Action[*WaitAction], error) {
	if durationParam == nil {
		return nil, fmt.Errorf("duration parameter cannot be nil")
	}

	a.DurationParam = durationParam

	return &task_engine.Action[*WaitAction]{
		ID:      "wait-action",
		Name:    "Wait",
		Wrapped: a,
	}, nil
}

func (a *WaitAction) Execute(ctx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve duration parameter
	durationValue, err := a.DurationParam.Resolve(ctx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve duration parameter: %w", err)
	}

	var duration time.Duration
	if durationStr, ok := durationValue.(string); ok {
		// Parse duration string (e.g., "5s", "1m", "2h")
		parsedDuration, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("failed to parse duration string '%s': %w", durationStr, err)
		}
		duration = parsedDuration
	} else if durationInt, ok := durationValue.(int); ok {
		// Treat as seconds
		duration = time.Duration(durationInt) * time.Second
	} else if durationDirect, ok := durationValue.(time.Duration); ok {
		// Direct time.Duration value
		duration = durationDirect
	} else {
		return fmt.Errorf("duration parameter is not a string, int, or time.Duration, got %T", durationValue)
	}

	if duration <= 0 {
		return fmt.Errorf("invalid duration: must be positive")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}

// GetOutput returns the waited duration
func (a *WaitAction) GetOutput() interface{} {
	return map[string]interface{}{
		"success": true,
	}
}
