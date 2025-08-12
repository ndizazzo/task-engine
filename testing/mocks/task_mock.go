package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

// EnhancedTaskMock provides comprehensive mocking capabilities for individual tasks
type EnhancedTaskMock struct {
	mock.Mock
	mu sync.RWMutex

	// Task properties
	id             string
	name           string
	completedTasks int
	totalTime      time.Duration
	shouldFail     bool
	customError    error
	customResult   interface{}

	// State tracking
	isRunning        bool
	hasRun           bool
	runCount         int
	contextCancelled bool

	// Call tracking
	runCalls []context.Context
}

// NewEnhancedTaskMock creates a new enhanced task mock
func NewEnhancedTaskMock(id, name string) *EnhancedTaskMock {
	return &EnhancedTaskMock{
		id:             id,
		name:           name,
		completedTasks: 0,
		totalTime:      0,
		shouldFail:     false,
	}
}

// GetID returns the task ID
func (m *EnhancedTaskMock) GetID() string {
	return m.id
}

// GetName returns the task name
func (m *EnhancedTaskMock) GetName() string {
	return m.name
}

// Run mocks the task execution
func (m *EnhancedTaskMock) Run(ctx context.Context) error {
	args := m.Called(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.runCalls = append(m.runCalls, ctx)
	m.runCount++
	m.hasRun = true
	select {
	case <-ctx.Done():
		m.contextCancelled = true
		return ctx.Err()
	default:
		// Context not cancelled, continue
	}

	// Simulate task execution
	if m.shouldFail {
		if m.customError != nil {
			return m.customError
		}
		return args.Error(0)
	}

	// Simulate successful completion
	m.completedTasks++
	m.totalTime = time.Duration(m.runCount) * time.Millisecond

	return args.Error(0)
}

// GetCompletedTasks returns the completed tasks count
func (m *EnhancedTaskMock) GetCompletedTasks() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.completedTasks
}

// GetTotalTime returns the total execution time
func (m *EnhancedTaskMock) GetTotalTime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalTime
}

// SetShouldFail configures the task to fail on execution
func (m *EnhancedTaskMock) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

// SetCustomError sets a custom error to return when the task fails
func (m *EnhancedTaskMock) SetCustomError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customError = err
}

// SetCustomResult sets a custom result for the task
func (m *EnhancedTaskMock) SetCustomResult(result interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customResult = result
}

// GetCustomResult returns the custom result
func (m *EnhancedTaskMock) GetCustomResult() interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.customResult
}

// GetCustomError returns the custom error
func (m *EnhancedTaskMock) GetCustomError() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.customError
}

// SetCompletedTasks sets the completed tasks count
func (m *EnhancedTaskMock) SetCompletedTasks(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completedTasks = count
}

// SetTotalTime sets the total execution time
func (m *EnhancedTaskMock) SetTotalTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalTime = duration
}

// IsRunning returns whether the task is currently running
func (m *EnhancedTaskMock) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

// SetRunning sets the running state
func (m *EnhancedTaskMock) SetRunning(running bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = running
}

// HasRun returns whether the task has been executed
func (m *EnhancedTaskMock) HasRun() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hasRun
}

// GetRunCount returns the number of times Run was called
func (m *EnhancedTaskMock) GetRunCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.runCount
}

// WasContextCancelled returns whether the context was cancelled during execution
func (m *EnhancedTaskMock) WasContextCancelled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.contextCancelled
}

// GetRunCalls returns all contexts used in Run calls
func (m *EnhancedTaskMock) GetRunCalls() []context.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]context.Context{}, m.runCalls...)
}

// ResetState resets all internal state
func (m *EnhancedTaskMock) ResetState() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.completedTasks = 0
	m.totalTime = 0
	m.shouldFail = false
	m.customError = nil
	m.customResult = nil
	m.isRunning = false
	m.hasRun = false
	m.runCount = 0
	m.contextCancelled = false
	m.runCalls = nil
}
