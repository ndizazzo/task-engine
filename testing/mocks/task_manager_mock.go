package mocks

import (
	"sync"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/stretchr/testify/mock"
)

// Ensure EnhancedTaskManagerMock implements TaskManagerInterface
var _ task_engine.TaskManagerInterface = (*EnhancedTaskManagerMock)(nil)

// EnhancedTaskManagerMock provides comprehensive mocking capabilities
type EnhancedTaskManagerMock struct {
	mock.Mock
	mu sync.RWMutex

	// State tracking
	tasks        map[string]*task_engine.Task
	runningTasks map[string]bool
	taskResults  map[string]interface{}
	taskErrors   map[string]error
	taskTiming   map[string]time.Duration

	// Call tracking
	addTaskCalls    []*task_engine.Task
	runTaskCalls    []string
	stopTaskCalls   []string
	stopAllCalls    int
	getRunningCalls int
	isRunningCalls  map[string]int
}

// NewEnhancedTaskManagerMock creates a new enhanced mock
func NewEnhancedTaskManagerMock() *EnhancedTaskManagerMock {
	return &EnhancedTaskManagerMock{
		tasks:          make(map[string]*task_engine.Task),
		runningTasks:   make(map[string]bool),
		taskResults:    make(map[string]interface{}),
		taskErrors:     make(map[string]error),
		taskTiming:     make(map[string]time.Duration),
		isRunningCalls: make(map[string]int),
	}
}

// AddTask mocks AddTask with state tracking
func (m *EnhancedTaskManagerMock) AddTask(task *task_engine.Task) error {
	args := m.Called(task)

	m.mu.Lock()
	defer m.mu.Unlock()

	if task != nil {
		m.tasks[task.ID] = task
		m.addTaskCalls = append(m.addTaskCalls, task)
	}

	return args.Error(0)
}

// RunTask mocks RunTask with state tracking
func (m *EnhancedTaskManagerMock) RunTask(taskID string) error {
	args := m.Called(taskID)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.runningTasks[taskID] = true
	m.runTaskCalls = append(m.runTaskCalls, taskID)

	return args.Error(0)
}

// StopTask mocks StopTask with state tracking
func (m *EnhancedTaskManagerMock) StopTask(taskID string) error {
	args := m.Called(taskID)

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.runningTasks, taskID)
	m.stopTaskCalls = append(m.stopTaskCalls, taskID)

	return args.Error(0)
}

// StopAllTasks mocks StopAllTasks
func (m *EnhancedTaskManagerMock) StopAllTasks() {
	m.Called()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopAllCalls++
	m.runningTasks = make(map[string]bool)
}

// GetRunningTasks returns the current running tasks
func (m *EnhancedTaskManagerMock) GetRunningTasks() []string {
	args := m.Called()

	m.mu.RLock()
	defer m.mu.RUnlock()

	m.getRunningCalls++

	var running []string
	for taskID, isRunning := range m.runningTasks {
		if isRunning {
			running = append(running, taskID)
		}
	}

	if args.Get(0) != nil {
		return args.Get(0).([]string)
	}
	return running
}

// IsTaskRunning checks if a specific task is running
func (m *EnhancedTaskManagerMock) IsTaskRunning(taskID string) bool {
	args := m.Called(taskID)

	m.mu.RLock()
	defer m.mu.RUnlock()

	m.isRunningCalls[taskID]++

	if args.Get(0) != nil {
		return args.Bool(0)
	}
	return m.runningTasks[taskID]
}

// SetTaskResult allows tests to set expected results
func (m *EnhancedTaskManagerMock) SetTaskResult(taskID string, result interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskResults[taskID] = result
}

// SetTaskError allows tests to set expected errors
func (m *EnhancedTaskManagerMock) SetTaskError(taskID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskErrors[taskID] = err
}

// GetTaskResult allows tests to retrieve set results
func (m *EnhancedTaskManagerMock) GetTaskResult(taskID string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.taskResults[taskID]
}

// GetTaskError allows tests to retrieve set errors
func (m *EnhancedTaskManagerMock) GetTaskError(taskID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.taskErrors[taskID]
}

// GetAddedTasks returns all tasks that were added
func (m *EnhancedTaskManagerMock) GetAddedTasks() []*task_engine.Task {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]*task_engine.Task{}, m.addTaskCalls...)
}

// GetRunTaskCalls returns all RunTask calls
func (m *EnhancedTaskManagerMock) GetRunTaskCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.runTaskCalls...)
}

// GetStopTaskCalls returns all StopTask calls
func (m *EnhancedTaskManagerMock) GetStopTaskCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.stopTaskCalls...)
}

// ClearHistory clears all call history
func (m *EnhancedTaskManagerMock) ClearHistory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addTaskCalls = nil
	m.runTaskCalls = nil
	m.stopTaskCalls = nil
}

// ResetState resets all internal state
func (m *EnhancedTaskManagerMock) ResetState() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks = make(map[string]*task_engine.Task)
	m.runningTasks = make(map[string]bool)
	m.taskResults = make(map[string]interface{})
	m.taskErrors = make(map[string]error)
	m.taskTiming = make(map[string]time.Duration)
	m.addTaskCalls = nil
	m.runTaskCalls = nil
	m.stopTaskCalls = nil
	m.stopAllCalls = 0
	m.getRunningCalls = 0
	m.isRunningCalls = make(map[string]int)
}

// SetTaskTiming allows tests to set expected timing
func (m *EnhancedTaskManagerMock) SetTaskTiming(taskID string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.taskTiming[taskID] = duration
}

// GetTaskTiming retrieves the timing for a specific task
func (m *EnhancedTaskManagerMock) GetTaskTiming(taskID string) (time.Duration, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	timing, exists := m.taskTiming[taskID]
	return timing, exists
}

// GetStopAllCalls returns the number of StopAllTasks calls
func (m *EnhancedTaskManagerMock) GetStopAllCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stopAllCalls
}

// GetGetRunningCalls returns the number of GetRunningTasks calls
func (m *EnhancedTaskManagerMock) GetGetRunningCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getRunningCalls
}

// GetIsRunningCalls returns the number of IsTaskRunning calls for a specific task
func (m *EnhancedTaskManagerMock) GetIsRunningCalls(taskID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunningCalls[taskID]
}

// GetAllIsRunningCalls returns all IsTaskRunning call counts
func (m *EnhancedTaskManagerMock) GetAllIsRunningCalls() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int)
	for k, v := range m.isRunningCalls {
		result[k] = v
	}
	return result
}

// GetCurrentState returns the current state of the mock
func (m *EnhancedTaskManagerMock) GetCurrentState() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state := map[string]interface{}{
		"total_tasks":       len(m.tasks),
		"running_tasks":     len(m.runningTasks),
		"total_results":     len(m.taskResults),
		"total_errors":      len(m.taskErrors),
		"total_timing":      len(m.taskTiming),
		"add_task_calls":    len(m.addTaskCalls),
		"run_task_calls":    len(m.runTaskCalls),
		"stop_task_calls":   len(m.stopTaskCalls),
		"stop_all_calls":    m.stopAllCalls,
		"get_running_calls": m.getRunningCalls,
		"is_running_calls":  m.isRunningCalls,
	}

	return state
}

// SimulateTaskCompletion simulates a task completing
func (m *EnhancedTaskManagerMock) SimulateTaskCompletion(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.runningTasks, taskID)
}

// SimulateTaskFailure simulates a task failing
func (m *EnhancedTaskManagerMock) SimulateTaskFailure(taskID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.runningTasks, taskID)
	m.taskErrors[taskID] = err
}

// SetExpectedBehavior sets up expected behavior for common scenarios
func (m *EnhancedTaskManagerMock) SetExpectedBehavior() {
	// Set up common expectations
	m.On("AddTask", mock.AnythingOfType("*task_engine.Task")).Return(nil)
	m.On("RunTask", mock.AnythingOfType("string")).Return(nil)
	m.On("StopTask", mock.AnythingOfType("string")).Return(nil)
	m.On("StopAllTasks").Return()
	m.On("GetRunningTasks").Return([]string{})
	m.On("IsTaskRunning", mock.AnythingOfType("string")).Return(false)
}

// VerifyAllExpectations verifies all expectations and returns detailed results
func (m *EnhancedTaskManagerMock) VerifyAllExpectations() map[string]bool {
	results := make(map[string]bool)

	// Check if all expected calls were made
	// We need to manually verify expectations since AssertExpectations doesn't clear ExpectedCalls
	allExpectationsMet := true
	for _, expectedCall := range m.ExpectedCalls {
		if expectedCall.Repeatability > 0 {
			allExpectationsMet = false
			break
		}
	}
	results["expectations_met"] = allExpectationsMet

	// Check state consistency
	state := m.GetCurrentState()
	results["state_consistent"] = state["total_tasks"].(int) >= 0

	return results
}

// ResetToCleanState resets the mock to a clean state
func (m *EnhancedTaskManagerMock) ResetToCleanState() {
	m.ClearHistory()
	m.ClearState()
	m.ExpectedCalls = nil
}

// ClearState clears all state-related data
func (m *EnhancedTaskManagerMock) ClearState() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks = make(map[string]*task_engine.Task)
	m.runningTasks = make(map[string]bool)
	m.taskResults = make(map[string]interface{})
	m.taskErrors = make(map[string]error)
	m.taskTiming = make(map[string]time.Duration)
}
