package task_engine_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ActionTestSuite struct {
	suite.Suite
}

func (suite *ActionTestSuite) TestAction_RunIDIsUnique() {
	action := PassingTestAction

	err := action.Execute(testContext())
	runID1 := action.RunID
	suite.NoError(err, "Execute should not return an error")
	suite.NotEmpty(runID1, "RunID should be set after execution")

	err = action.Execute(testContext())
	runID2 := action.RunID
	suite.NoError(err, "Execute should not return an error")
	suite.NotEmpty(runID2, "RunID should be set after execution")
	suite.NotEqual(runID1, runID2, "RunID should be unique for each execution")
}

func (suite *ActionTestSuite) TestAction_SucceedsWithoutError() {
	action := PassingTestAction
	err := action.Execute(testContext())
	suite.NoError(err, "Execute should not return an error")
}

func (suite *ActionTestSuite) TestAction_ExecutesFunc() {
	action := PassingTestAction
	err := action.Execute(testContext())
	require.NoError(suite.T(), err)
	suite.True(action.Wrapped.Called, "Execute should have been called")
}

func (suite *ActionTestSuite) TestAction_ComputesDuration() {
	action := PassingTestAction
	err := action.Execute(testContext())
	suite.NoError(err, "Execute should not return an error")
	suite.GreaterOrEqual(action.Duration, time.Duration(0), "Duration should be non-negative")
}

func (suite *ActionTestSuite) TestAction_ReturnsErrorOnFailure() {
	action := FailingTestAction
	err := action.Execute(testContext())
	suite.Error(err, "Execute should return an error when Execute fails")
}

func (suite *ActionTestSuite) TestAction_BeforeExecuteFailure() {
	action := BeforeExecuteFailingTestAction
	err := action.Execute(testContext())
	suite.Error(err, "Execute should return an error when BeforeExecute fails")
	suite.Contains(err.Error(), "simulated BeforeExecute failure", "Error should contain BeforeExecute failure message")
}

func (suite *ActionTestSuite) TestAction_AfterExecuteFailure() {
	action := AfterExecuteFailingTestAction
	err := action.Execute(testContext())
	suite.Error(err, "Execute should return an error when AfterExecute fails")
	suite.Contains(err.Error(), "simulated AfterExecute failure", "Error should contain AfterExecute failure message")
}

func (suite *ActionTestSuite) TestAction_GetLogger() {
	action := PassingTestAction
	logger := action.GetLogger()
	suite.Nil(logger, "GetLogger should return nil when no logger is set")

	action.Logger = noOpLogger
	logger = action.GetLogger()
	suite.NotNil(logger, "GetLogger should return the logger when set")
	suite.Equal(noOpLogger, logger, "GetLogger should return the same logger that was set")
}

func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}
