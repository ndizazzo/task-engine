package common_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/stretchr/testify/suite"
)

type mockAction struct{}

func (m *mockAction) BeforeExecute(ctx context.Context) error {
	return nil
}

func (m *mockAction) Execute(ctx context.Context) error {
	return nil
}

func (m *mockAction) AfterExecute(ctx context.Context) error {
	return nil
}

func (m *mockAction) GetOutput() interface{} {
	return nil
}

type BaseConstructorTestSuite struct {
	suite.Suite
	constructor *common.BaseConstructor[task_engine.ActionInterface]
	logger      *slog.Logger
}

func (suite *BaseConstructorTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	suite.constructor = common.NewBaseConstructor[task_engine.ActionInterface](suite.logger)
}

func (suite *BaseConstructorTestSuite) TestNewBaseConstructor() {
	constructor := common.NewBaseConstructor[task_engine.ActionInterface](suite.logger)
	suite.NotNil(constructor)
}

func (suite *BaseConstructorTestSuite) TestGetLogger() {
	logger := suite.constructor.GetLogger()
	suite.NotNil(logger)
	suite.Equal(suite.logger, logger)
}

func (suite *BaseConstructorTestSuite) TestWrapAction_WithIDProvided() {
	action := &mockAction{}
	wrappedAction := suite.constructor.WrapAction(action, "test-action", "custom-id")

	suite.NotNil(wrappedAction)
	suite.Equal("custom-id", wrappedAction.ID)
	suite.Equal("test-action", wrappedAction.Name)
	suite.Equal(action, wrappedAction.Wrapped)
}

func (suite *BaseConstructorTestSuite) TestWrapAction_WithoutIDProvided() {
	action := &mockAction{}
	wrappedAction := suite.constructor.WrapAction(action, "test-action")

	suite.NotNil(wrappedAction)
	suite.NotEmpty(wrappedAction.ID)
	suite.Equal("test-action", wrappedAction.Name)
	suite.Equal(action, wrappedAction.Wrapped)
	suite.Contains(wrappedAction.ID, "test-action")
	suite.Contains(wrappedAction.ID, "action")
}

func (suite *BaseConstructorTestSuite) TestWrapAction_WithEmptyID() {
	action := &mockAction{}
	wrappedAction := suite.constructor.WrapAction(action, "my-test", "")

	suite.NotNil(wrappedAction)
	suite.NotEmpty(wrappedAction.ID)
	suite.Contains(wrappedAction.ID, "my-test")
}

func (suite *BaseConstructorTestSuite) TestWrapAction_WithSpecialCharactersInName() {
	action := &mockAction{}
	wrappedAction := suite.constructor.WrapAction(action, "test@action#name", "id123")

	suite.NotNil(wrappedAction)
	suite.Equal("id123", wrappedAction.ID)
	suite.Equal("test@action#name", wrappedAction.Name)
}

func (suite *BaseConstructorTestSuite) TestWrapAction_WithMultipleIDVarargs() {
	action := &mockAction{}
	wrappedAction := suite.constructor.WrapAction(action, "test-action", "id1", "id2")

	suite.NotNil(wrappedAction)
	suite.Equal("id1", wrappedAction.ID)
}

func TestBaseConstructorTestSuite(t *testing.T) {
	suite.Run(t, new(BaseConstructorTestSuite))
}
