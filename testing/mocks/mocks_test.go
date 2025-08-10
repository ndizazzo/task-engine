package mocks

import (
	"context"
	"errors"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MocksTestSuite tests the enhanced mocks functionality
type MocksTestSuite struct {
	suite.Suite
}

// TestMocksTestSuite runs the Mocks test suite
func TestMocksTestSuite(t *testing.T) {
	suite.Run(t, new(MocksTestSuite))
}

// TestEnhancedTaskManagerMock tests the enhanced task manager mock
func (suite *MocksTestSuite) TestEnhancedTaskManagerMock() {
	suite.Run("NewEnhancedTaskManagerMock", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		assert.NotNil(suite.T(), taskManagerMock)
		assert.Empty(suite.T(), taskManagerMock.GetAddedTasks())
		assert.Empty(suite.T(), taskManagerMock.GetRunTaskCalls())
		assert.Empty(suite.T(), taskManagerMock.GetStopTaskCalls())
	})

	suite.Run("AddTask tracking", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		taskManagerMock.Mock.On("AddTask", mock.Anything).Return(nil)

		// Create a real task for testing
		task := &task_engine.Task{ID: "test-task", Name: "Test Task"}
		err := taskManagerMock.AddTask(task)

		assert.NoError(suite.T(), err)
		assert.Len(suite.T(), taskManagerMock.GetAddedTasks(), 1)
		assert.Equal(suite.T(), task, taskManagerMock.GetAddedTasks()[0])
		taskManagerMock.AssertExpectations(suite.T())
	})

	suite.Run("RunTask tracking", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		taskManagerMock.Mock.On("RunTask", "test-task").Return(nil)
		taskManagerMock.Mock.On("IsTaskRunning", "test-task").Return(true)

		err := taskManagerMock.RunTask("test-task")

		assert.NoError(suite.T(), err)
		assert.Len(suite.T(), taskManagerMock.GetRunTaskCalls(), 1)
		assert.Equal(suite.T(), "test-task", taskManagerMock.GetRunTaskCalls()[0])
		assert.True(suite.T(), taskManagerMock.IsTaskRunning("test-task"))
		taskManagerMock.AssertExpectations(suite.T())
	})

	suite.Run("StopTask tracking", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		taskManagerMock.Mock.On("StopTask", "test-task").Return(nil)
		taskManagerMock.Mock.On("IsTaskRunning", "test-task").Return(false)
		taskManagerMock.runningTasks["test-task"] = true

		err := taskManagerMock.StopTask("test-task")

		assert.NoError(suite.T(), err)
		assert.Len(suite.T(), taskManagerMock.GetRunTaskCalls(), 0)
		assert.Len(suite.T(), taskManagerMock.GetStopTaskCalls(), 1)
		assert.Equal(suite.T(), "test-task", taskManagerMock.GetStopTaskCalls()[0])
		assert.False(suite.T(), taskManagerMock.IsTaskRunning("test-task"))
		taskManagerMock.AssertExpectations(suite.T())
	})

	suite.Run("StopAllTasks", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		taskManagerMock.Mock.On("StopAllTasks").Return()
		taskManagerMock.Mock.On("GetRunningTasks").Return([]string{})
		taskManagerMock.runningTasks["task1"] = true
		taskManagerMock.runningTasks["task2"] = true

		taskManagerMock.StopAllTasks()

		assert.Empty(suite.T(), taskManagerMock.GetRunningTasks())
		taskManagerMock.AssertExpectations(suite.T())
	})

	suite.Run("GetRunningTasks", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		taskManagerMock.Mock.On("GetRunningTasks").Return([]string{"task1", "task3"})
		taskManagerMock.runningTasks["task1"] = true
		taskManagerMock.runningTasks["task2"] = false
		taskManagerMock.runningTasks["task3"] = true

		running := taskManagerMock.GetRunningTasks()

		assert.Len(suite.T(), running, 2)
		assert.Contains(suite.T(), running, "task1")
		assert.Contains(suite.T(), running, "task3")
		assert.NotContains(suite.T(), running, "task2")
		taskManagerMock.AssertExpectations(suite.T())
	})

	suite.Run("State management", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		taskManagerMock.SetTaskResult("task1", "result1")
		taskManagerMock.SetTaskError("task2", errors.New("error2"))

		result := taskManagerMock.GetTaskResult("task1")
		err := taskManagerMock.GetTaskError("task2")

		assert.Equal(suite.T(), "result1", result)
		assert.Equal(suite.T(), "error2", err.Error())
	})

	suite.Run("ClearHistory", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		taskManagerMock.Mock.On("AddTask", mock.Anything).Return(nil)
		taskManagerMock.Mock.On("RunTask", "test-task").Return(nil)

		task := &task_engine.Task{ID: "test-task", Name: "Test Task"}
		taskManagerMock.AddTask(task)
		taskManagerMock.RunTask("test-task")

		assert.Len(suite.T(), taskManagerMock.GetAddedTasks(), 1)
		assert.Len(suite.T(), taskManagerMock.GetRunTaskCalls(), 1)

		taskManagerMock.ClearHistory()

		assert.Empty(suite.T(), taskManagerMock.GetAddedTasks())
		assert.Empty(suite.T(), taskManagerMock.GetRunTaskCalls())
	})

	suite.Run("ResetState", func() {
		taskManagerMock := NewEnhancedTaskManagerMock()
		taskManagerMock.runningTasks["task1"] = true
		taskManagerMock.runningTasks["task2"] = true

		assert.Len(suite.T(), taskManagerMock.runningTasks, 2)

		taskManagerMock.ResetState()

		assert.Empty(suite.T(), taskManagerMock.runningTasks)
		assert.Empty(suite.T(), taskManagerMock.taskResults)
		assert.Empty(suite.T(), taskManagerMock.taskErrors)
	})
}

// TestEnhancedTaskMock tests the enhanced task mock
func (suite *MocksTestSuite) TestEnhancedTaskMock() {
	suite.Run("NewEnhancedTaskMock", func() {
		taskMock := NewEnhancedTaskMock("test-task", "Test Task")
		assert.NotNil(suite.T(), taskMock)
		assert.Equal(suite.T(), "test-task", taskMock.GetID())
		assert.Equal(suite.T(), "Test Task", taskMock.GetName())
		assert.Equal(suite.T(), 0, taskMock.GetRunCount())
	})

	suite.Run("Run tracking", func() {
		taskMock := NewEnhancedTaskMock("test-task", "Test Task")
		taskMock.Mock.On("Run", mock.Anything).Return(nil)

		ctx := context.Background()
		err := taskMock.Run(ctx)

		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, taskMock.GetRunCount())
		taskMock.AssertExpectations(suite.T())
	})

	suite.Run("Result and error setting", func() {
		taskMock := NewEnhancedTaskMock("test-task", "Test Task")
		taskMock.SetCustomResult("test result")
		taskMock.SetCustomError(errors.New("test error"))

		result := taskMock.GetCustomResult()
		err := taskMock.GetCustomError()

		assert.Equal(suite.T(), "test result", result)
		assert.Equal(suite.T(), "test error", err.Error())
	})

	suite.Run("ResetState", func() {
		taskMock := NewEnhancedTaskMock("test-task", "Test Task")
		taskMock.Mock.On("Run", mock.Anything).Return(nil)

		ctx := context.Background()
		taskMock.Run(ctx)

		assert.Equal(suite.T(), 1, taskMock.GetRunCount())

		taskMock.ResetState()

		assert.Equal(suite.T(), 0, taskMock.GetRunCount())
		assert.Nil(suite.T(), taskMock.GetCustomResult())
		assert.Nil(suite.T(), taskMock.GetCustomError())
	})
}

// TestResultProviderMock tests the result provider mock
func (suite *MocksTestSuite) TestResultProviderMock() {
	suite.Run("NewResultProviderMock", func() {
		resultProviderMock := NewResultProviderMock()
		assert.NotNil(suite.T(), resultProviderMock)
		assert.Equal(suite.T(), 0, resultProviderMock.GetResultCallCount())
		assert.Equal(suite.T(), 0, resultProviderMock.GetErrorCallCount())
	})

	suite.Run("GetResult tracking", func() {
		resultProviderMock := NewResultProviderMock()
		resultProviderMock.Mock.On("GetResult").Return("test result")

		result := resultProviderMock.GetResult()

		assert.Equal(suite.T(), "test result", result)
		assert.Equal(suite.T(), 1, resultProviderMock.GetResultCallCount())
		resultProviderMock.AssertExpectations(suite.T())
	})

	suite.Run("GetError tracking", func() {
		resultProviderMock := NewResultProviderMock()
		resultProviderMock.Mock.On("GetError").Return(errors.New("test error"))

		err := resultProviderMock.GetError()

		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), "test error", err.Error())
		assert.Equal(suite.T(), 1, resultProviderMock.GetErrorCallCount())
		resultProviderMock.AssertExpectations(suite.T())
	})

	suite.Run("Result and error setting", func() {
		resultProviderMock := NewResultProviderMock()
		resultProviderMock.SetResult("test result")
		resultProviderMock.SetError(errors.New("test error"))

		// Set up mock expectations for the methods that use m.Called()
		resultProviderMock.Mock.On("GetResult").Return("test result")
		resultProviderMock.Mock.On("GetError").Return(errors.New("test error"))

		result := resultProviderMock.GetResult()
		err := resultProviderMock.GetError()

		assert.Equal(suite.T(), "test result", result)
		assert.Equal(suite.T(), "test error", err.Error())
		resultProviderMock.AssertExpectations(suite.T())
	})

	suite.Run("ResetState", func() {
		resultProviderMock := NewResultProviderMock()
		resultProviderMock.Mock.On("GetResult").Return("test result")
		resultProviderMock.Mock.On("GetError").Return(errors.New("test error"))

		resultProviderMock.GetResult()
		resultProviderMock.GetError()

		assert.Equal(suite.T(), 1, resultProviderMock.GetResultCallCount())
		assert.Equal(suite.T(), 1, resultProviderMock.GetErrorCallCount())

		resultProviderMock.ResetState()

		assert.Equal(suite.T(), 0, resultProviderMock.GetResultCallCount())
		assert.Equal(suite.T(), 0, resultProviderMock.GetErrorCallCount())
	})
}

// mockTask is a simple task for testing
type mockTask struct {
	id   string
	name string
}

func (t *mockTask) GetID() string   { return t.id }
func (t *mockTask) GetName() string { return t.name }
