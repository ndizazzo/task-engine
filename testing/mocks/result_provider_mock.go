package mocks

import (
	"sync"

	"github.com/stretchr/testify/mock"
)

// ResultProviderMock provides mocking capabilities for tasks that produce results
type ResultProviderMock struct {
	mock.Mock
	mu sync.RWMutex

	// Result data
	result interface{}
	err    error

	// Call tracking
	getResultCalls int
	getErrorCalls  int
}

// NewResultProviderMock creates a new result provider mock
func NewResultProviderMock() *ResultProviderMock {
	return &ResultProviderMock{}
}

// GetResult returns the stored result
func (m *ResultProviderMock) GetResult() interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getResultCalls++
	args := m.Called()

	if args.Get(0) != nil {
		return args.Get(0)
	}
	return m.result
}

// GetError returns the stored error
func (m *ResultProviderMock) GetError() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getErrorCalls++
	args := m.Called()

	if args.Error(0) != nil {
		return args.Error(0)
	}
	return m.err
}

// SetResult sets the result to return
func (m *ResultProviderMock) SetResult(result interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.result = result
}

// SetError sets the error to return
func (m *ResultProviderMock) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// GetResultCallCount returns the number of times GetResult was called
func (m *ResultProviderMock) GetResultCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getResultCalls
}

// GetErrorCallCount returns the number of times GetError was called
func (m *ResultProviderMock) GetErrorCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getErrorCalls
}

// ResetState resets all internal state
func (m *ResultProviderMock) ResetState() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.result = nil
	m.err = nil
	m.getResultCalls = 0
	m.getErrorCalls = 0
}
