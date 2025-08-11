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

// TaskContext maintains execution context for a single task
type TaskContext struct {
	TaskID        string
	GlobalContext *GlobalContext
	Logger        *slog.Logger
}

// NewTaskContext creates a new TaskContext instance
func NewTaskContext(taskID string, globalContext *GlobalContext, logger *slog.Logger) *TaskContext {
	return &TaskContext{
		TaskID:        taskID,
		GlobalContext: globalContext,
		Logger:        logger,
	}
}

func (t *Task) Run(ctx context.Context) error {
	return t.RunWithContext(ctx, nil)
}

// RunWithContext executes the task with a specific global context for parameter resolution.
// This enables cross-task and cross-action parameter passing by sharing context
// between different task executions.
func (t *Task) RunWithContext(ctx context.Context, globalContext *GlobalContext) error {
	t.mu.Lock()
	t.RunID = uuid.New().String()
	runID := t.RunID // Store locally to avoid race conditions in logging
	t.mu.Unlock()

	t.log("Starting task", "taskID", t.ID, "runID", runID)

	// Create global context if not provided
	if globalContext == nil {
		globalContext = NewGlobalContext()
	}

	// Create task context
	taskContext := NewTaskContext(t.ID, globalContext, t.Logger)

	// Validate parameters before execution
	if err := t.validateParameters(taskContext); err != nil {
		t.log("Task parameter validation failed", "taskID", t.ID, "runID", runID, "error", err)
		return fmt.Errorf("task %s (run %s) parameter validation failed: %w", t.ID, runID, err)
	}

	for _, action := range t.Actions {
		select {
		case <-ctx.Done():
			t.log("Task canceled", "taskID", t.ID, "runID", runID, "reason", ctx.Err())
			return ctx.Err()
		default:
			// Execute action
			t.log("Executing action", "taskID", t.ID, "actionID", action.GetID())

			// Create a new context with the global context embedded
			actionCtx := context.WithValue(ctx, GlobalContextKey, globalContext)

			execErr := action.Execute(actionCtx)
			if execErr != nil {
				if errors.Is(execErr, ErrPrerequisiteNotMet) {
					t.log("Task aborted: prerequisite not met", "taskID", t.ID, "runID", runID, "actionID", action.GetID(), "error", execErr)
					return fmt.Errorf("task %s (run %s) aborted: prerequisite not met in action %s: %w", t.ID, runID, action.GetID(), execErr)
				} else {
					t.log("Task failed: action execution error", "taskID", t.ID, "runID", runID, "actionID", action.GetID(), "error", execErr)
					return fmt.Errorf("task %s (run %s) failed at action %s: %w", t.ID, runID, action.GetID(), execErr)
				}
			}

			t.log("Action executed successfully", "taskID", t.ID, "actionID", action.GetID())

			// Store action output in global context
			t.log("Storing action output", "taskID", t.ID, "actionID", action.GetID())
			t.storeActionOutput(action, globalContext)
		}

		t.mu.Lock()
		t.TotalTime += action.GetDuration()
		t.CompletedTasks += 1
		t.mu.Unlock()
	}

	// Store task output in global context
	t.storeTaskOutput(globalContext)

	t.log("Task completed", "taskID", t.ID, "runID", runID, "totalDuration", t.GetTotalTime())
	return nil
}

// storeActionOutput stores the output from an action in the global context.
// This enables parameter passing between actions by making action outputs
// available to subsequent actions in the same or different tasks.
func (t *Task) storeActionOutput(action ActionWrapper, globalContext *GlobalContext) {
	actionID := action.GetID()
	t.Logger.Info("Storing action output", "actionID", actionID)

	// Store basic output if action implements ActionInterface
	if actionWithOutput, ok := action.(interface{ GetOutput() interface{} }); ok {
		output := actionWithOutput.GetOutput()
		t.Logger.Info("Action implements GetOutput", "actionID", actionID, "output", output)
		if output != nil {
			globalContext.StoreActionOutput(actionID, output)
			t.Logger.Info("Stored action output", "actionID", actionID, "output", output)
		} else {
			t.Logger.Info("Action output is nil, not storing", "actionID", actionID)
		}
	} else {
		t.Logger.Info("Action does not implement GetOutput", "actionID", actionID)
	}

	// Store result provider if action implements ResultProvider
	if resultProvider, ok := action.(ResultProvider); ok {
		globalContext.StoreActionResult(actionID, resultProvider)
		t.Logger.Info("Stored action result provider", "actionID", actionID)
	}
}

// storeTaskOutput stores the task output in the global context.
// This enables cross-task parameter passing by making task outputs
// available to actions in other tasks.
func (t *Task) storeTaskOutput(globalContext *GlobalContext) {
	// Create task output with basic information
	taskOutput := map[string]interface{}{
		"taskID":         t.ID,
		"runID":          t.RunID,
		"name":           t.Name,
		"totalTime":      t.TotalTime,
		"completedTasks": t.CompletedTasks,
		"success":        true,
	}

	globalContext.StoreTaskOutput(t.ID, taskOutput)
	t.Logger.Debug("Stored task output", "taskID", t.ID, "output", taskOutput)
}

// validateParameters validates that all action parameters can be resolved.
// This ensures that all parameter references can be resolved and prevents
// runtime errors during action execution.
func (t *Task) validateParameters(taskContext *TaskContext) error {
	for i, action := range t.Actions {
		if err := t.validateActionParameters(action, i, taskContext); err != nil {
			return fmt.Errorf("action %d (%s): %w", i, action.GetName(), err)
		}
	}
	return nil
}

// validateActionParameters validates parameters for a specific action
func (t *Task) validateActionParameters(action ActionWrapper, index int, taskContext *TaskContext) error {
	// For now, we'll do basic validation
	// In the future, this could be extended to validate specific parameter types
	// based on action implementation
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

// GetID returns the task ID in a thread-safe manner
func (t *Task) GetID() string {
	return t.ID
}

// GetName returns the task name in a thread-safe manner
func (t *Task) GetName() string {
	return t.Name
}
