package task_engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var _ TaskManagerInterface = (*TaskManager)(nil)

// TaskManager implements TaskManagerInterface for managing task execution
type TaskManager struct {
	Tasks        map[string]*Task
	runningTasks map[string]context.CancelFunc
	Logger       *slog.Logger
	mu           sync.Mutex
	// Global context for cross-task parameter passing. This enables actions
	// in different tasks to reference outputs from other tasks.
	globalContext *GlobalContext
}

func NewTaskManager(logger *slog.Logger) *TaskManager {
	return &TaskManager{
		Tasks:         make(map[string]*Task),
		runningTasks:  make(map[string]context.CancelFunc),
		Logger:        logger,
		globalContext: NewGlobalContext(),
	}
}

func (tm *TaskManager) AddTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	task.Logger = tm.Logger.With("taskID", task.ID)
	tm.Tasks[task.ID] = task
	tm.Logger.Info("Task added", "taskID", task.ID)

	return nil
}

func (tm *TaskManager) RunTask(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, exists := tm.Tasks[taskID]
	if !exists {
		tm.Logger.Error("Task not found", "taskID", taskID)
		return fmt.Errorf("task %q not found", taskID)
	}

	// Create a context for every task
	ctx, cancel := context.WithCancel(context.Background())
	tm.runningTasks[taskID] = cancel

	// Start every task in a goroutine
	go func() {
		defer func() {
			tm.mu.Lock()
			delete(tm.runningTasks, taskID)
			tm.mu.Unlock()
		}()

		// Run task with the global context for parameter resolution
		err := task.RunWithContext(ctx, tm.globalContext)
		if err != nil {
			if ctx.Err() != nil {
				tm.Logger.Info("Task canceled", "taskID", taskID, "error", err)
			} else {
				tm.Logger.Error("Task execution failed", "taskID", taskID, "error", err)
			}
		} else {
			tm.Logger.Info("Task completed", "taskID", taskID)
		}
	}()

	return nil
}

func (tm *TaskManager) StopTask(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	cancel, exists := tm.runningTasks[taskID]
	if !exists {
		return fmt.Errorf("task %q is not running", taskID)
	}

	// Cancel the task's context
	cancel()
	tm.Logger.Info("Task stopped", "taskID", taskID)
	delete(tm.runningTasks, taskID)
	return nil
}

func (tm *TaskManager) StopAllTasks() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for taskID, cancel := range tm.runningTasks {
		cancel()
		tm.Logger.Info("Task stopped", "taskID", taskID)
		delete(tm.runningTasks, taskID)
	}
}

// GetRunningTasks returns a list of currently running task IDs
func (tm *TaskManager) GetRunningTasks() []string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	taskIDs := make([]string, 0, len(tm.runningTasks))
	for taskID := range tm.runningTasks {
		taskIDs = append(taskIDs, taskID)
	}
	return taskIDs
}

// IsTaskRunning checks if a specific task is currently running
func (tm *TaskManager) IsTaskRunning(taskID string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	_, exists := tm.runningTasks[taskID]
	return exists
}

// WaitForAllTasksToComplete waits for all running tasks to complete
func (tm *TaskManager) WaitForAllTasksToComplete(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		tm.mu.Lock()
		runningCount := len(tm.runningTasks)
		tm.mu.Unlock()

		if runningCount == 0 {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for %d tasks to complete", runningCount)
		}

		// Log the current state for debugging
		tm.Logger.Debug("Waiting for tasks to complete", "runningCount", runningCount, "timeout", timeout)
		time.Sleep(10 * time.Millisecond)
	}
}

// GetGlobalContext returns the global context for parameter resolution.
// Use this to access the shared context that stores outputs from all tasks
// and actions, enabling cross-entity parameter references.
func (tm *TaskManager) GetGlobalContext() *GlobalContext {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.globalContext
}

// ResetGlobalContext resets the global context, clearing all stored outputs and results.
// Use this when you want to start fresh with parameter passing, such as between
// different workflow executions or test runs.
func (tm *TaskManager) ResetGlobalContext() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.globalContext = NewGlobalContext()
	tm.Logger.Info("Global context reset")
}
