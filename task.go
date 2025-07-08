package task_engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
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
	mu             sync.Mutex // protects concurrent access to TotalTime and CompletedTasks
}

func (t *Task) Run(ctx context.Context) error {
	t.mu.Lock()
	t.RunID = uuid.New().String()
	runID := t.RunID // Store locally to avoid race conditions in logging
	t.mu.Unlock()

	t.log("Starting task", "taskID", t.ID, "runID", runID)

	for _, action := range t.Actions {
		select {
		case <-ctx.Done():
			t.log("Task canceled", "taskID", t.ID, "runID", runID, "reason", ctx.Err())
			return ctx.Err()
		default:
			execErr := action.Execute(ctx)
			if execErr != nil {
				if errors.Is(execErr, ErrPrerequisiteNotMet) {
					t.log("Task aborted: prerequisite not met", "taskID", t.ID, "runID", runID, "actionID", action.GetID(), "error", execErr)
					return fmt.Errorf("task %s (run %s) aborted: prerequisite not met in action %s: %w", t.ID, runID, action.GetID(), execErr)
				} else {
					t.log("Task failed: action execution error", "taskID", t.ID, "runID", runID, "actionID", action.GetID(), "error", execErr)
					return fmt.Errorf("task %s (run %s) failed at action %s: %w", t.ID, runID, action.GetID(), execErr)
				}
			}
		}

		t.mu.Lock()
		t.TotalTime += action.GetDuration()
		t.CompletedTasks += 1
		t.mu.Unlock()
	}

	t.log("Task completed", "taskID", t.ID, "runID", runID, "totalDuration", t.GetTotalTime())
	return nil
}

func (t *Task) log(message string, keyvals ...interface{}) {
	if t.Logger != nil {
		t.Logger.Info(message, keyvals...)
	}
}

// GetTotalTime returns the total time in a thread-safe manner
func (t *Task) GetTotalTime() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.TotalTime
}

// GetCompletedTasks returns the completed tasks count in a thread-safe manner
func (t *Task) GetCompletedTasks() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.CompletedTasks
}
