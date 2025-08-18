package utility

import (
	"context"
	"fmt"
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

// PrerequisiteCheckFunc defines the signature for a callback function that checks a prerequisite.
// It returns true if the parent task should be aborted (prerequisite not met),
// and an error if the check itself encounters an issue.
// If the prerequisite is met, it should return (false, nil).
type PrerequisiteCheckFunc func(ctx context.Context, logger *slog.Logger) (abortTask bool, err error)

// PrerequisiteCheckAction is an action that executes a callback function to check a prerequisite.
// If the callback indicates the prerequisite is not met, the action returns ErrPrerequisiteNotMet.
type PrerequisiteCheckAction struct {
	task_engine.BaseAction
	// Parameter fields
	DescriptionParam task_engine.ActionParameter
	CheckParam       task_engine.ActionParameter
	CommandsParam    task_engine.ActionParameter
	// Result fields (resolved from parameters during execution)
	Check       PrerequisiteCheckFunc
	Description string // A human-readable description of what is being checked
}

// NewPrerequisiteCheckAction creates a new PrerequisiteCheckAction with the given logger
func NewPrerequisiteCheckAction(logger *slog.Logger) *PrerequisiteCheckAction {
	return &PrerequisiteCheckAction{
		BaseAction: task_engine.NewBaseAction(logger),
	}
}

// WithParameters sets the parameters for prerequisite checking and returns a wrapped Action
func (a *PrerequisiteCheckAction) WithParameters(
	descriptionParam task_engine.ActionParameter,
	checkParam task_engine.ActionParameter,
) (*task_engine.Action[*PrerequisiteCheckAction], error) {
	a.DescriptionParam = descriptionParam
	a.CheckParam = checkParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*PrerequisiteCheckAction](a.Logger)
	return constructor.WrapAction(a, "Prerequisite Check", "prerequisite-check-action"), nil
}

// Execute runs the prerequisite check callback.
// If the callback indicates the prerequisite is not met (returns abortTask=true),
// it returns ErrPrerequisiteNotMet.
// If the callback itself returns an error, that error is propagated.
func (a *PrerequisiteCheckAction) Execute(ctx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve description and check from parameters if provided
	if a.DescriptionParam != nil {
		v, err := a.DescriptionParam.Resolve(ctx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve description parameter: %w", err)
		}
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("description parameter is not a string, got %T", v)
		}
		a.Description = s
	}
	if a.CheckParam != nil {
		v, err := a.CheckParam.Resolve(ctx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve check parameter: %w", err)
		}
		fn, ok := v.(PrerequisiteCheckFunc)
		if !ok {
			return fmt.Errorf("check parameter is not a PrerequisiteCheckFunc, got %T", v)
		}
		a.Check = fn
	}

	if a.Check == nil {
		return fmt.Errorf("critical internal error: prerequisite check function for '%s' is not defined", a.Description)
	}

	a.Logger.Info("Performing prerequisite check", "description", a.Description)
	abortTask, err := a.Check(ctx, a.Logger)
	if err != nil {
		a.Logger.Error("Prerequisite check callback failed", "description", a.Description, "error", err)
		return fmt.Errorf("prerequisite check '%s' encountered an error: %w", a.Description, err)
	}

	if abortTask {
		a.Logger.Warn("Prerequisite not met, signaling task abort", "description", a.Description)
		return task_engine.ErrPrerequisiteNotMet
	}

	a.Logger.Info("Prerequisite check passed", "description", a.Description)
	return nil
}

// GetOutput returns the description of the check; success is true if no abort occurred
func (a *PrerequisiteCheckAction) GetOutput() interface{} {
	return map[string]interface{}{
		"description": a.Description,
		"success":     true,
	}
}
