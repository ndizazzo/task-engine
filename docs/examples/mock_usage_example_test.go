package examples

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
)

// ExampleTaskProcessor demonstrates how downstream projects can use the enhanced mocks
type ExampleTaskProcessor struct {
	taskManager task_engine.TaskManagerInterface
}

// NewExampleTaskProcessor creates a new task processor
func NewExampleTaskProcessor(taskManager task_engine.TaskManagerInterface) *ExampleTaskProcessor {
	return &ExampleTaskProcessor{
		taskManager: taskManager,
	}
}

// ProcessTask demonstrates a simple task processing workflow
func (p *ExampleTaskProcessor) ProcessTask(taskID string) error {
	if p.taskManager.IsTaskRunning(taskID) {
		return nil // Task already running
	}

	// Create a simple task
	task := &task_engine.Task{
		ID:   taskID,
		Name: "Example Task",
	}

	// Add the task
	if err := p.taskManager.AddTask(task); err != nil {
		return err
	}

	// Run the task
	return p.taskManager.RunTask(taskID)
}

// MockUsageExampleTestSuite tests the mock usage examples
type MockUsageExampleTestSuite struct {
	suite.Suite
}

// TestMockUsageExampleTestSuite runs the MockUsageExample test suite
func TestMockUsageExampleTestSuite(t *testing.T) {
	suite.Run(t, new(MockUsageExampleTestSuite))
}

// TestExampleTaskProcessor_ProcessTask tests the ExampleTaskProcessor with mocks
func (suite *MockUsageExampleTestSuite) TestExampleTaskProcessor_ProcessTask() {
	// Create mocks
	taskManagerMock := mocks.NewEnhancedTaskManagerMock()

	// Set up mock expectations
	taskManagerMock.On("IsTaskRunning", "test-task").Return(false)
	taskManagerMock.On("AddTask", mock.AnythingOfType("*task_engine.Task")).Return(nil)
	taskManagerMock.On("RunTask", "test-task").Return(nil)

	// Create processor and process task
	processor := NewExampleTaskProcessor(taskManagerMock)
	err := processor.ProcessTask("test-task")

	// Assertions
	suite.NoError(err)
	taskManagerMock.AssertExpectations(suite.T())
}

// TestExampleTaskProcessor_AlreadyRunning tests the case when task is already running
func (suite *MockUsageExampleTestSuite) TestExampleTaskProcessor_AlreadyRunning() {
	// Create mocks
	taskManagerMock := mocks.NewEnhancedTaskManagerMock()

	// Set up mock expectations for already running task
	taskManagerMock.On("IsTaskRunning", "running-task").Return(true)

	// Create processor and process task
	processor := NewExampleTaskProcessor(taskManagerMock)
	err := processor.ProcessTask("running-task")

	// Assertions
	suite.NoError(err)
	taskManagerMock.AssertExpectations(suite.T())
}

// TestExampleTaskProcessor_AddTaskFailure tests the case when adding task fails
func (suite *MockUsageExampleTestSuite) TestExampleTaskProcessor_AddTaskFailure() {
	// Create mocks
	taskManagerMock := mocks.NewEnhancedTaskManagerMock()

	// Set up mock expectations for failure
	taskManagerMock.On("IsTaskRunning", "failing-task").Return(false)
	taskManagerMock.On("AddTask", mock.AnythingOfType("*task_engine.Task")).Return(assert.AnError)

	// Create processor and process task
	processor := NewExampleTaskProcessor(taskManagerMock)
	err := processor.ProcessTask("failing-task")

	// Assertions
	suite.Error(err)
	taskManagerMock.AssertExpectations(suite.T())
}
