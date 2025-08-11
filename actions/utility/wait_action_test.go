package utility_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/utility"
	command_mock "github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

type WaitActionTestSuite struct {
	suite.Suite
	logger *slog.Logger
}

func (suite *WaitActionTestSuite) SetupTest() {
	suite.logger = command_mock.NewDiscardLogger()
}

func (suite *WaitActionTestSuite) TestExecuteSuccess() {
	durationStr := "10ms"
	duration, _ := time.ParseDuration(durationStr)
	action, err := utility.NewWaitAction(suite.logger).WithParameters(task_engine.StaticParameter{Value: durationStr})
	suite.Require().NoError(err)

	start := time.Now()
	execErr := action.Wrapped.Execute(context.Background())
	elapsed := time.Since(start)

	suite.NoError(execErr)
	suite.GreaterOrEqual(elapsed, duration)
}

func (suite *WaitActionTestSuite) TestExecuteContextCancellation() {
	durationStr := "100ms"
	duration, _ := time.ParseDuration(durationStr)
	action, err := utility.NewWaitAction(suite.logger).WithParameters(task_engine.StaticParameter{Value: durationStr})
	suite.Require().NoError(err)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	execErr := action.Wrapped.Execute(ctx)
	elapsed := time.Since(start)

	suite.Error(execErr)
	suite.ErrorIs(execErr, context.Canceled)
	suite.Less(elapsed, duration)
}

func (suite *WaitActionTestSuite) TestExecuteZeroDuration() {
	duration := "0s"
	action, err := utility.NewWaitAction(suite.logger).WithParameters(task_engine.StaticParameter{Value: duration})
	suite.Require().NoError(err)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "invalid duration: must be positive")
}

func (suite *WaitActionTestSuite) TestWaitAction_GetOutput() {
	action := &utility.WaitAction{}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(true, m["success"])
}

func TestWaitActionTestSuite(t *testing.T) {
	suite.Run(t, new(WaitActionTestSuite))
}
