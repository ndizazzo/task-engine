package testing

import (
	"log/slog"
	"sync"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
)

// TestableTaskManager provides enhanced testing capabilities
type TestableTaskManager struct {
	*task_engine.TaskManager
	mu sync.RWMutex

	// Testing hooks
	onTaskAdded     func(*task_engine.Task)
	onTaskStarted   func(string)
	onTaskCompleted func(string, error)
	onTaskStopped   func(string)

	// Result storage for testing
	taskResults map[string]interface{}
	taskErrors  map[string]error
	taskTiming  map[string]time.Duration

	// Call tracking for testing
	taskAddedCalls   []*task_engine.Task
	taskStartedCalls []string
	taskStoppedCalls []string
}

// NewTestableTaskManager creates a testable task manager
func NewTestableTaskManager(logger *slog.Logger) *TestableTaskManager {
	return &TestableTaskManager{
		TaskManager: task_engine.NewTaskManager(logger),
		taskResults: make(map[string]interface{}),
		taskErrors:  make(map[string]error),
		taskTiming:  make(map[string]time.Duration),
	}
}

// SetTaskAddedHook sets a callback for when tasks are added
func (tm *TestableTaskManager) SetTaskAddedHook(hook func(*task_engine.Task)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTaskAdded = hook
}

// SetTaskStartedHook sets a callback for when tasks start
func (tm *TestableTaskManager) SetTaskStartedHook(hook func(string)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTaskStarted = hook
}

// SetTaskCompletedHook sets a callback for when tasks complete
func (tm *TestableTaskManager) SetTaskCompletedHook(hook func(string, error)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTaskCompleted = hook
}

// SetTaskStoppedHook sets a callback for when tasks are stopped
func (tm *TestableTaskManager) SetTaskStoppedHook(hook func(string)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTaskStopped = hook
}

// OverrideTaskResult allows tests to set expected results
func (tm *TestableTaskManager) OverrideTaskResult(taskID string, result interface{}) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.taskResults[taskID] = result
}

// OverrideTaskError allows tests to set expected errors
func (tm *TestableTaskManager) OverrideTaskError(taskID string, err error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.taskErrors[taskID] = err
}

// OverrideTaskTiming allows tests to set expected timing
func (tm *TestableTaskManager) OverrideTaskTiming(taskID string, duration time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.taskTiming[taskID] = duration
}

// GetTaskResult retrieves the result for a specific task
func (tm *TestableTaskManager) GetTaskResult(taskID string) (interface{}, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result, exists := tm.taskResults[taskID]
	return result, exists
}

// GetTaskError retrieves the error for a specific task
func (tm *TestableTaskManager) GetTaskError(taskID string) (error, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	err, exists := tm.taskErrors[taskID]
	return err, exists
}

// GetTaskTiming retrieves the timing for a specific task
func (tm *TestableTaskManager) GetTaskTiming(taskID string) (time.Duration, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	timing, exists := tm.taskTiming[taskID]
	return timing, exists
}

// GetTaskAddedCalls returns all tasks that were added
func (tm *TestableTaskManager) GetTaskAddedCalls() []*task_engine.Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return append([]*task_engine.Task{}, tm.taskAddedCalls...)
}

// GetTaskStartedCalls returns all started task IDs
func (tm *TestableTaskManager) GetTaskStartedCalls() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return append([]string{}, tm.taskStartedCalls...)
}

// GetTaskStoppedCalls returns all stopped task IDs
func (tm *TestableTaskManager) GetTaskStoppedCalls() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return append([]string{}, tm.taskStoppedCalls...)
}

// ClearTestData clears all test-related data
func (tm *TestableTaskManager) ClearTestData() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.taskResults = make(map[string]interface{})
	tm.taskErrors = make(map[string]error)
	tm.taskTiming = make(map[string]time.Duration)
	tm.taskAddedCalls = nil
	tm.taskStartedCalls = nil
	tm.taskStoppedCalls = nil
}

// Override AddTask to include hooks and call tracking
func (tm *TestableTaskManager) AddTask(task *task_engine.Task) error {
	// Call the original implementation first
	err := tm.TaskManager.AddTask(task)
	if err != nil {
		return err
	}

	// Track the call and execute hook (protected by our lock)
	tm.mu.Lock()
	tm.taskAddedCalls = append(tm.taskAddedCalls, task)
	hook := tm.onTaskAdded
	tm.mu.Unlock()

	// Execute hook if set (outside of lock to avoid deadlocks)
	if hook != nil {
		hook(task)
	}

	return nil
}

// Override RunTask to include hooks and call tracking
func (tm *TestableTaskManager) RunTask(taskID string) error {
	// Track the call and execute hook (protected by our lock)
	tm.mu.Lock()
	tm.taskStartedCalls = append(tm.taskStartedCalls, taskID)
	hook := tm.onTaskStarted
	tm.mu.Unlock()

	// Execute hook if set (outside of lock to avoid deadlocks)
	if hook != nil {
		hook(taskID)
	}

	// Call the original implementation
	return tm.TaskManager.RunTask(taskID)
}

// Override StopTask to include hooks and call tracking
func (tm *TestableTaskManager) StopTask(taskID string) error {
	// Track the call and execute hook (protected by our lock)
	tm.mu.Lock()
	tm.taskStoppedCalls = append(tm.taskStoppedCalls, taskID)
	hook := tm.onTaskStopped
	tm.mu.Unlock()

	// Execute hook if set (outside of lock to avoid deadlocks)
	if hook != nil {
		hook(taskID)
	}

	// Call the original implementation
	return tm.TaskManager.StopTask(taskID)
}

// SimulateTaskCompletion allows tests to simulate task completion
func (tm *TestableTaskManager) SimulateTaskCompletion(taskID string, err error) {
	// Get hook and execute it (protected by our lock)
	tm.mu.Lock()
	hook := tm.onTaskCompleted
	tm.mu.Unlock()

	// Execute hook if set (outside of lock to avoid deadlocks)
	if hook != nil {
		hook(taskID, err)
	}
}

// GetTestMetrics returns comprehensive test metrics
func (tm *TestableTaskManager) GetTestMetrics() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	metrics := map[string]interface{}{
		"total_tasks_added":   len(tm.taskAddedCalls),
		"total_tasks_started": len(tm.taskStartedCalls),
		"total_tasks_stopped": len(tm.taskStoppedCalls),
		"total_results_set":   len(tm.taskResults),
		"total_errors_set":    len(tm.taskErrors),
		"total_timing_set":    len(tm.taskTiming),
	}

	return metrics
}

// ResetToCleanState resets the manager to a clean state for testing
func (tm *TestableTaskManager) ResetToCleanState() {
	// Clear all test data first (handles its own locking)
	tm.ClearTestData()

	// Now reset hooks and task maps under a single lock to avoid nested locking
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Reset hooks
	tm.onTaskAdded = nil
	tm.onTaskStarted = nil
	tm.onTaskCompleted = nil
	tm.onTaskStopped = nil

	// Clear all tasks and running tasks
	tm.Tasks = make(map[string]*task_engine.Task)
	// Note: runningTasks is unexported, so we can't directly clear it
	// The TaskManager will handle this internally
}
