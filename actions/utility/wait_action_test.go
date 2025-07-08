package utility_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/ndizazzo/task-engine/actions/utility"
	command_mock "github.com/ndizazzo/task-engine/mocks"
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
	duration := 10 * time.Millisecond
	action := utility.NewWaitAction(suite.logger, duration)

	start := time.Now()
	err := action.Wrapped.Execute(context.Background())
	elapsed := time.Since(start)

	suite.NoError(err)
	suite.GreaterOrEqual(elapsed, duration)
}

func (suite *WaitActionTestSuite) TestExecuteContextCancellation() {
	duration := 100 * time.Millisecond
	action := utility.NewWaitAction(suite.logger, duration)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := action.Wrapped.Execute(ctx)
	elapsed := time.Since(start)

	suite.Error(err)
	suite.ErrorIs(err, context.Canceled)
	suite.Less(elapsed, duration)
}

func (suite *WaitActionTestSuite) TestExecuteZeroDuration() {
	duration := 0 * time.Second
	action := utility.NewWaitAction(suite.logger, duration)

	start := time.Now()
	err := action.Wrapped.Execute(context.Background())
	elapsed := time.Since(start)

	suite.NoError(err)
	suite.Less(elapsed, 10*time.Millisecond)
}

func TestWaitActionTestSuite(t *testing.T) {
	suite.Run(t, new(WaitActionTestSuite))
}
