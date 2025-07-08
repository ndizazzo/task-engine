package utility

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
)

// ErrNilCheckFunction is returned by NewPrerequisiteCheckAction if the provided check function is nil.
var ErrNilCheckFunction = errors.New("prerequisite check function cannot be nil")

// PrerequisiteCheckFunc defines the signature for a callback function that checks a prerequisite.
// It returns true if the parent task should be aborted (prerequisite not met),
// and an error if the check itself encounters an issue.
// If the prerequisite is met, it should return (false, nil).
type PrerequisiteCheckFunc func(ctx context.Context, logger *slog.Logger) (abortTask bool, err error)

// PrerequisiteCheckAction is an action that executes a callback function to check a prerequisite.
// If the callback indicates the prerequisite is not met, the action returns ErrPrerequisiteNotMet.
type PrerequisiteCheckAction struct {
	task_engine.BaseAction
	Check       PrerequisiteCheckFunc
	Description string // A human-readable description of what is being checked
}

// NewPrerequisiteCheckAction creates a new PrerequisiteCheckAction.
// logger: The logger to be used by the action.
// description: A human-readable string describing the prerequisite being checked.
// check: The callback function to execute for the prerequisite check.
// Returns an error if the check function is nil.
func NewPrerequisiteCheckAction(logger *slog.Logger, description string, check PrerequisiteCheckFunc) (*task_engine.Action[*PrerequisiteCheckAction], error) {
	if check == nil {
		return nil, ErrNilCheckFunction
	}

	id := fmt.Sprintf("prerequisite-check-%s-action", description)

	return &task_engine.Action[*PrerequisiteCheckAction]{
		ID: id,
		Wrapped: &PrerequisiteCheckAction{
			BaseAction:  task_engine.BaseAction{Logger: logger},
			Check:       check,
			Description: description,
		},
	}, nil
}

// Execute runs the prerequisite check callback.
// If the callback indicates the prerequisite is not met (returns abortTask=true),
// it returns ErrPrerequisiteNotMet.
// If the callback itself returns an error, that error is propagated.
func (a *PrerequisiteCheckAction) Execute(ctx context.Context) error {
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
