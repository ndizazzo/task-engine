package utility

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

// WaitAction represents an action that waits for a specified duration
type WaitAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	// Parameter fields
	DurationParam task_engine.ActionParameter
}

// NewWaitAction creates a new WaitAction with the given logger
func NewWaitAction(logger *slog.Logger) *WaitAction {
	return &WaitAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for wait action and returns a wrapped Action
func (a *WaitAction) WithParameters(
	durationParam task_engine.ActionParameter,
) (*task_engine.Action[*WaitAction], error) {
	a.DurationParam = durationParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*WaitAction](a.Logger)
	return constructor.WrapAction(a, "Wait", "wait-action"), nil
}

func (a *WaitAction) Execute(ctx context.Context) error {
	// Use the new parameter resolver to handle duration parameter
	duration, err := a.ResolveDurationParameter(ctx, a.DurationParam, "duration")
	if err != nil {
		return err
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

// GetOutput returns the waited duration using the new output builder
func (a *WaitAction) GetOutput() interface{} {
	return a.BuildSimpleOutput(true, "")
}
