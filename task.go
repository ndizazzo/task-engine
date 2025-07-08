package task_engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// ErrPrerequisiteNotMet is returned by an action when a prerequisite for task execution
// is not met, signaling that the task should be gracefully aborted.
var ErrPrerequisiteNotMet = errors.New("task prerequisite not met")

// Task represents a collection of actions to execute in sequential order
type Task struct {
	ID             string
	RunID          string
	Name           string
	Actions        []ActionWrapper
	Logger         *slog.Logger
	TotalTime      time.Duration
	CompletedTasks int
}

func (t *Task) Run(ctx context.Context) error {
	t.RunID = uuid.New().String()
	t.log("Starting task", "taskID", t.ID, "runID", t.RunID)

	for _, action := range t.Actions {
		select {
		case <-ctx.Done():
			t.log("Task canceled", "taskID", t.ID, "runID", t.RunID, "reason", ctx.Err())
			return ctx.Err()
		default:
			execErr := action.Execute(ctx)
			if execErr != nil {
				if errors.Is(execErr, ErrPrerequisiteNotMet) {
					t.log("Task aborted: prerequisite not met", "taskID", t.ID, "runID", t.RunID, "actionID", action.GetID(), "error", execErr)
					return fmt.Errorf("task %s (run %s) aborted: prerequisite not met in action %s: %w", t.ID, t.RunID, action.GetID(), execErr)
				} else {
					t.log("Task failed: action execution error", "taskID", t.ID, "runID", t.RunID, "actionID", action.GetID(), "error", execErr)
					return fmt.Errorf("task %s (run %s) failed at action %s: %w", t.ID, t.RunID, action.GetID(), execErr)
				}
			}
		}

		t.TotalTime += action.GetDuration()
		t.CompletedTasks += 1
	}

	t.log("Task completed", "taskID", t.ID, "runID", t.RunID, "totalDuration", t.TotalTime)
	return nil
}

func (t *Task) log(message string, keyvals ...interface{}) {
	if t.Logger != nil {
		t.Logger.Info(message, keyvals...)
	}
}
