package utility_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/utility"
	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// PrerequisiteCheckActionTestSuite tests the PrerequisiteCheckAction functionality
type PrerequisiteCheckActionTestSuite struct {
	suite.Suite
}

// TestPrerequisiteCheckActionTestSuite runs the PrerequisiteCheckAction test suite
func TestPrerequisiteCheckActionTestSuite(t *testing.T) {
	suite.Run(t, new(PrerequisiteCheckActionTestSuite))
}

func (suite *PrerequisiteCheckActionTestSuite) TestPrerequisiteCheckAction_Execute() {
	tests := []struct {
		name           string
		description    string
		checkFunc      utility.PrerequisiteCheckFunc
		expectError    error
		expectContains string
	}{
		{
			name:        "Prerequisite Met",
			description: "Check if world is round",
			checkFunc: func(ctx context.Context, logger *slog.Logger) (abortTask bool, err error) {
				return false, nil // Prerequisite met
			},
			expectError: nil,
		},
		{
			name:        "Prerequisite Not Met",
			description: "Check if sky is green",
			checkFunc: func(ctx context.Context, logger *slog.Logger) (abortTask bool, err error) {
				return true, nil // Prerequisite NOT met, abort
			},
			expectError: task_engine.ErrPrerequisiteNotMet,
		},
		{
			name:        "Check Function Error",
			description: "Check with internal error",
			checkFunc: func(ctx context.Context, logger *slog.Logger) (abortTask bool, err error) {
				return false, errors.New("internal sensor failure")
			},
			expectError:    nil, // Not ErrPrerequisiteNotMet, but a wrapped error
			expectContains: "internal sensor failure",
		},
		{
			name:        "Context Canceled During Check",
			description: "Check with context cancellation",
			checkFunc: func(ctx context.Context, logger *slog.Logger) (abortTask bool, err error) {
				select {
				case <-ctx.Done():
					return false, ctx.Err()
				case <-time.After(1 * time.Millisecond):
					return false, nil
				}
			},
			expectError:    nil,
			expectContains: context.Canceled.Error(),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			logger := mocks.NewDiscardLogger()
			// Constructor now returns an error, handle it for valid test cases
			action, err := utility.NewPrerequisiteCheckAction(logger, tc.description, tc.checkFunc)
			assert.NoError(suite.T(), err, "NewPrerequisiteCheckAction should not return an error for valid test cases here")
			assert.NotNil(suite.T(), action)

			var ctx context.Context
			var cancel context.CancelFunc

			if tc.name == "Context Canceled During Check" {
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			defer cancel()

			execErr := action.Wrapped.Execute(ctx)

			switch {
			case tc.expectError != nil:
				assert.ErrorIs(suite.T(), execErr, tc.expectError, fmt.Sprintf("Expected error %v, got %v", tc.expectError, execErr))
			case tc.expectContains != "":
				assert.ErrorContains(suite.T(), execErr, tc.expectContains, fmt.Sprintf("Error message '%v' does not contain '%s'", execErr, tc.expectContains))
			default:
				assert.NoError(suite.T(), execErr, fmt.Sprintf("Expected no error, got %v", execErr))
			}
		})
	}
}

func (suite *PrerequisiteCheckActionTestSuite) TestNewPrerequisiteCheckAction_NilCheck() {
	logger := mocks.NewDiscardLogger()
	action, err := utility.NewPrerequisiteCheckAction(logger, "test", nil)
	assert.Error(suite.T(), err, "NewPrerequisiteCheckAction should return an error when check function is nil")
	assert.Nil(suite.T(), action, "NewPrerequisiteCheckAction should return nil action when check function is nil")
}
