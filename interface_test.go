package task_engine

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

// InterfaceTestSuite tests the interface implementations
type InterfaceTestSuite struct {
	suite.Suite
}

// TestInterfaceTestSuite runs the Interface test suite
func TestInterfaceTestSuite(t *testing.T) {
	suite.Run(t, new(InterfaceTestSuite))
}

// TestTaskManagerImplementsInterface verifies that TaskManager implements TaskManagerInterface
func (suite *InterfaceTestSuite) TestTaskManagerImplementsInterface() {
	var _ TaskManagerInterface = (*TaskManager)(nil)
}

// TestTaskImplementsInterface verifies that Task implements TaskInterface
func (suite *InterfaceTestSuite) TestTaskImplementsInterface() {
	var _ TaskInterface = (*Task)(nil)
}

// TestNewTaskManagerCreatesValidInterface verifies that NewTaskManager returns a valid TaskManagerInterface
func (suite *InterfaceTestSuite) TestNewTaskManagerCreatesValidInterface() {
	// Use a discard logger to prevent test output
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	taskManager := NewTaskManager(discardLogger)

	// This should compile and run without errors
	var _ TaskManagerInterface = taskManager
}
