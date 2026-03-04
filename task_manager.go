package task_engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var _ TaskManagerInterface = (*TaskManager)(nil)

// TaskHandle provides access to a running task's completion status and result
type TaskHandle struct {
	taskID string
	done   chan struct{}
	err    error
	mu     sync.Mutex
}

func (h *TaskHandle) Done() <-chan struct{} { return h.done }
func (h *TaskHandle) Err() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.err
}
func (h *TaskHandle) TaskID() string { return h.taskID }

// TaskManager implements TaskManagerInterface for managing task execution
type TaskManager struct {
	Tasks        map[string]*Task
	runningTasks map[string]context.CancelFunc
	Logger       *slog.Logger
	mu           sync.Mutex
	wg           sync.WaitGroup
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

	// Check for duplicate task IDs
	if _, exists := tm.Tasks[task.ID]; exists {
		return fmt.Errorf("task ID '%s' already exists", task.ID)
	}

	task.Logger = tm.Logger.With("taskID", task.ID)
	tm.Tasks[task.ID] = task
	tm.Logger.Info("Task added", "taskID", task.ID)

	return nil
}

func (tm *TaskManager) RunTask(taskID string) (*TaskHandle, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, exists := tm.Tasks[taskID]
	if !exists {
		tm.Logger.Error("Task not found", "taskID", taskID)
		return nil, fmt.Errorf("task %q not found", taskID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	tm.runningTasks[taskID] = cancel

	gc := tm.globalContext

	handle := &TaskHandle{taskID: taskID, done: make(chan struct{})}
	tm.wg.Add(1)

	go func(gcSnapshot *GlobalContext) {
		defer tm.wg.Done()
		defer close(handle.done)
		defer func() {
			tm.mu.Lock()
			delete(tm.runningTasks, taskID)
			tm.mu.Unlock()
		}()

		err := task.RunWithContext(ctx, gcSnapshot)

		handle.mu.Lock()
		handle.err = err
		handle.mu.Unlock()

		if err != nil {
			if ctx.Err() != nil {
				tm.Logger.Info("Task canceled", "taskID", taskID, "error", err)
			} else {
				tm.Logger.Error("Task execution failed", "taskID", taskID, "error", err)
			}
		} else {
			tm.Logger.Info("Task completed", "taskID", taskID)
		}
	}(gc)

	return handle, nil
}

func (tm *TaskManager) StopTask(taskID string) error {
	tm.mu.Lock()
	cancel, exists := tm.runningTasks[taskID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("task %q is not running", taskID)
	}
	delete(tm.runningTasks, taskID)
	tm.mu.Unlock()

	cancel()
	tm.Logger.Info("Task stopped", "taskID", taskID)
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

func (tm *TaskManager) WaitForAllTasksToComplete(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		tm.wg.Wait()
		close(done)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		tm.mu.Lock()
		runningCount := len(tm.runningTasks)
		tm.mu.Unlock()
		return fmt.Errorf("timeout waiting for %d tasks to complete", runningCount)
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
