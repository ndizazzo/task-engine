package task_engine

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestAction struct for testing basic action functionality
type TestAction struct {
	BaseAction
	Called     bool
	ShouldFail bool
}

func (a *TestAction) Execute(ctx context.Context) error {
	a.Called = true
	if a.ShouldFail {
		return fmt.Errorf("simulated failure")
	}
	return nil
}

// DelayAction for testing timing functionality
type DelayAction struct {
	BaseAction
	Delay time.Duration
}

func (a *DelayAction) Execute(ctx context.Context) error {
	time.Sleep(a.Delay)
	return nil
}

// BeforeExecuteFailingAction for testing BeforeExecute failures
type BeforeExecuteFailingAction struct {
	BaseAction
	ShouldFailBefore bool
}

func (a *BeforeExecuteFailingAction) BeforeExecute(ctx context.Context) error {
	if a.ShouldFailBefore {
		return fmt.Errorf("simulated BeforeExecute failure")
	}
	return nil
}

func (a *BeforeExecuteFailingAction) Execute(ctx context.Context) error {
	return nil
}

// AfterExecuteFailingAction for testing AfterExecute failures
type AfterExecuteFailingAction struct {
	BaseAction
	ShouldFailAfter bool
}

func (a *AfterExecuteFailingAction) BeforeExecute(ctx context.Context) error {
	return nil
}

func (a *AfterExecuteFailingAction) Execute(ctx context.Context) error {
	return nil
}

func (a *AfterExecuteFailingAction) AfterExecute(ctx context.Context) error {
	if a.ShouldFailAfter {
		return fmt.Errorf("simulated AfterExecute failure")
	}
	return nil
}

// ActionTestSuite contains all the action tests
type ActionTestSuite struct {
	suite.Suite
}

// TestActionTestSuite runs the ActionTestSuite
func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}

// TestAction_RunIDIsUnique tests that each execution gets a unique RunID
func (suite *ActionTestSuite) TestAction_RunIDIsUnique() {
	action := &Action[*TestAction]{
		ID: "test-action",
		Wrapped: &TestAction{
			BaseAction: BaseAction{},
		},
	}

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

// TestAction_SucceedsWithoutError tests successful execution
func (suite *ActionTestSuite) TestAction_SucceedsWithoutError() {
	action := &Action[*TestAction]{
		ID: "test-action",
		Wrapped: &TestAction{
			BaseAction: BaseAction{},
		},
	}
	err := action.Execute(testContext())
	suite.NoError(err, "Execute should not return an error")
}

// TestAction_ExecutesFunc tests that the wrapped action is called
func (suite *ActionTestSuite) TestAction_ExecutesFunc() {
	action := &Action[*TestAction]{
		ID: "test-action",
		Wrapped: &TestAction{
			BaseAction: BaseAction{},
		},
	}
	err := action.Execute(testContext())
	require.NoError(suite.T(), err)
	suite.True(action.Wrapped.Called, "Execute should have been called")
}

// TestAction_ComputesDuration tests duration calculation
func (suite *ActionTestSuite) TestAction_ComputesDuration() {
	action := &Action[*TestAction]{
		ID: "test-action",
		Wrapped: &TestAction{
			BaseAction: BaseAction{},
		},
	}
	err := action.Execute(testContext())
	suite.NoError(err, "Execute should not return an error")
	suite.GreaterOrEqual(action.Duration, time.Duration(0), "Duration should be non-negative")
}

// TestAction_ReturnsErrorOnFailure tests error handling
func (suite *ActionTestSuite) TestAction_ReturnsErrorOnFailure() {
	action := &Action[*TestAction]{
		ID: "test-action",
		Wrapped: &TestAction{
			BaseAction: BaseAction{},
			ShouldFail: true,
		},
	}
	err := action.Execute(testContext())
	suite.Error(err, "Execute should return an error when Execute fails")
}

// TestAction_BeforeExecuteFailure tests BeforeExecute error handling
func (suite *ActionTestSuite) TestAction_BeforeExecuteFailure() {
	action := &Action[*BeforeExecuteFailingAction]{
		ID: "test-action",
		Wrapped: &BeforeExecuteFailingAction{
			BaseAction:       BaseAction{},
			ShouldFailBefore: true,
		},
	}
	err := action.Execute(testContext())
	suite.Error(err, "Execute should return an error when BeforeExecute fails")
	suite.Contains(err.Error(), "simulated BeforeExecute failure", "Error should contain BeforeExecute failure message")
}

// TestAction_AfterExecuteFailure tests AfterExecute error handling
func (suite *ActionTestSuite) TestAction_AfterExecuteFailure() {
	action := &Action[*AfterExecuteFailingAction]{
		ID: "test-action",
		Wrapped: &AfterExecuteFailingAction{
			BaseAction:      BaseAction{},
			ShouldFailAfter: true,
		},
	}
	err := action.Execute(testContext())
	suite.Error(err, "Execute should return an error when AfterExecute fails")
	suite.Contains(err.Error(), "simulated AfterExecute failure", "Error should contain AfterExecute failure message")
}

// TestAction_GetLogger tests logger access
func (suite *ActionTestSuite) TestAction_GetLogger() {
	action := &Action[*TestAction]{
		ID: "test-action",
		Wrapped: &TestAction{
			BaseAction: BaseAction{},
		},
	}
	logger := action.GetLogger()
	suite.Nil(logger, "GetLogger should return nil when no logger is set")

	action.Logger = NewDiscardLogger()
	logger = action.GetLogger()
	suite.NotNil(logger, "GetLogger should return the logger when set")
	suite.Equal(NewDiscardLogger(), logger, "GetLogger should return the same logger that was set")
}

// TestResolveString tests string parameter resolution
func TestResolveString(t *testing.T) {
	ctx := context.Background()
	gc := NewGlobalContext()

	tests := []struct {
		name     string
		param    ActionParameter
		expected string
		hasError bool
	}{
		{
			name:     "nil parameter",
			param:    nil,
			expected: "",
			hasError: false,
		},
		{
			name:     "string parameter",
			param:    StaticParameter{Value: "test string"},
			expected: "test string",
			hasError: false,
		},
		{
			name:     "byte slice parameter",
			param:    StaticParameter{Value: []byte("test bytes")},
			expected: "test bytes",
			hasError: false,
		},
		{
			name:     "int parameter",
			param:    StaticParameter{Value: 42},
			expected: "42",
			hasError: false,
		},
		{
			name:     "bool parameter",
			param:    StaticParameter{Value: true},
			expected: "true",
			hasError: false,
		},
		{
			name:     "float parameter",
			param:    StaticParameter{Value: 3.14},
			expected: "3.14",
			hasError: false,
		},
		{
			name:     "stringer parameter",
			param:    StaticParameter{Value: testStringer{}},
			expected: "test stringer",
			hasError: false,
		},
		{
			name:     "unsupported type",
			param:    StaticParameter{Value: make(chan int)},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveString(ctx, tt.param, gc)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestResolveBool tests boolean parameter resolution
func TestResolveBool(t *testing.T) {
	ctx := context.Background()
	gc := NewGlobalContext()

	tests := []struct {
		name     string
		param    ActionParameter
		expected bool
		hasError bool
	}{
		{
			name:     "nil parameter",
			param:    nil,
			expected: false,
			hasError: false,
		},
		{
			name:     "true bool",
			param:    StaticParameter{Value: true},
			expected: true,
			hasError: false,
		},
		{
			name:     "false bool",
			param:    StaticParameter{Value: false},
			expected: false,
			hasError: false,
		},
		{
			name:     "true string",
			param:    StaticParameter{Value: "true"},
			expected: true,
			hasError: false,
		},
		{
			name:     "false string",
			param:    StaticParameter{Value: "false"},
			expected: false,
			hasError: false,
		},
		{
			name:     "yes string",
			param:    StaticParameter{Value: "yes"},
			expected: true,
			hasError: false,
		},
		{
			name:     "no string",
			param:    StaticParameter{Value: "no"},
			expected: false,
			hasError: false,
		},
		{
			name:     "1 string",
			param:    StaticParameter{Value: "1"},
			expected: true,
			hasError: false,
		},
		{
			name:     "0 string",
			param:    StaticParameter{Value: "0"},
			expected: false,
			hasError: false,
		},
		{
			name:     "positive int",
			param:    StaticParameter{Value: 42},
			expected: true,
			hasError: false,
		},
		{
			name:     "zero int",
			param:    StaticParameter{Value: 0},
			expected: false,
			hasError: false,
		},
		{
			name:     "positive int64",
			param:    StaticParameter{Value: int64(42)},
			expected: true,
			hasError: false,
		},
		{
			name:     "positive uint",
			param:    StaticParameter{Value: uint(42)},
			expected: true,
			hasError: false,
		},
		{
			name:     "invalid string",
			param:    StaticParameter{Value: "invalid"},
			expected: false,
			hasError: true,
		},
		{
			name:     "unsupported type",
			param:    StaticParameter{Value: 3.14},
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveBool(ctx, tt.param, gc)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestResolveStringSlice tests string slice parameter resolution
func TestResolveStringSlice(t *testing.T) {
	ctx := context.Background()
	gc := NewGlobalContext()

	tests := []struct {
		name     string
		param    ActionParameter
		expected []string
		hasError bool
	}{
		{
			name:     "nil parameter",
			param:    nil,
			expected: nil,
			hasError: false,
		},
		{
			name:     "string slice parameter",
			param:    StaticParameter{Value: []string{"a", "b", "c"}},
			expected: []string{"a", "b", "c"},
			hasError: false,
		},
		{
			name:     "empty string",
			param:    StaticParameter{Value: ""},
			expected: []string{},
			hasError: false,
		},
		{
			name:     "comma separated string",
			param:    StaticParameter{Value: "a,b,c"},
			expected: []string{"a", "b", "c"},
			hasError: false,
		},
		{
			name:     "space separated string",
			param:    StaticParameter{Value: "a b c"},
			expected: []string{"a", "b", "c"},
			hasError: false,
		},
		{
			name:     "mixed separators",
			param:    StaticParameter{Value: "a, b , c"},
			expected: []string{"a", "b", "c"},
			hasError: false,
		},
		{
			name:     "unsupported type",
			param:    StaticParameter{Value: 42},
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveStringSlice(ctx, tt.param, gc)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestSanitizeIDPart tests ID sanitization
func TestSanitizeIDPart(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "simple lowercase",
			input:    "test",
			expected: "test",
		},
		{
			name:     "with spaces",
			input:    "test name",
			expected: "test-name",
		},
		{
			name:     "with slashes",
			input:    "test/name",
			expected: "test-name",
		},
		{
			name:     "with special characters",
			input:    "test@name#123",
			expected: "testname123",
		},
		{
			name:     "mixed case",
			input:    "TestName",
			expected: "testname",
		},
		{
			name:     "with numbers and underscores",
			input:    "test_123_name",
			expected: "test_123_name",
		},
		{
			name:     "with colons and dots",
			input:    "test:name.txt",
			expected: "test:name.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeIDPart(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildActionID tests action ID construction
func TestBuildActionID(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		parts    []string
		expected string
	}{
		{
			name:     "empty prefix and parts",
			prefix:   "",
			parts:    []string{},
			expected: "action-action",
		},
		{
			name:     "with prefix only",
			prefix:   "test",
			parts:    []string{},
			expected: "test-action",
		},
		{
			name:     "with prefix and parts",
			prefix:   "docker",
			parts:    []string{"image", "pull"},
			expected: "docker-image-pull-action",
		},
		{
			name:     "with empty parts",
			prefix:   "test",
			parts:    []string{"", "valid", ""},
			expected: "test-valid-action",
		},
		{
			name:     "with special characters",
			prefix:   "test@name",
			parts:    []string{"part 1", "part/2"},
			expected: "testname-part-1-part-2-action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildActionID(tt.prefix, tt.parts...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTypedOutputKey_Validate tests TypedOutputKey validation
func TestTypedOutputKey_Validate(t *testing.T) {
	type TestStruct struct {
		ValidField string
		ValidInt   int
	}

	type EmptyStruct struct{}

	tests := []struct {
		name     string
		key      TypedOutputKey[TestStruct]
		hasError bool
	}{
		{
			name:     "valid field",
			key:      TypedOutputKey[TestStruct]{ActionID: "test", Key: "ValidField"},
			hasError: false,
		},
		{
			name:     "valid int field",
			key:      TypedOutputKey[TestStruct]{ActionID: "test", Key: "ValidInt"},
			hasError: false,
		},
		{
			name:     "invalid field",
			key:      TypedOutputKey[TestStruct]{ActionID: "test", Key: "InvalidField"},
			hasError: true,
		},
		{
			name:     "empty field",
			key:      TypedOutputKey[TestStruct]{ActionID: "test", Key: ""},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.key.Validate()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Test with non-struct type
	t.Run("non-struct type", func(t *testing.T) {
		key := TypedOutputKey[string]{ActionID: "test", Key: "any"}
		err := key.Validate()
		assert.NoError(t, err) // Should not error for non-struct types
	})

	// Test with empty struct
	t.Run("empty struct", func(t *testing.T) {
		key := TypedOutputKey[EmptyStruct]{ActionID: "test", Key: "any"}
		err := key.Validate()
		assert.Error(t, err) // Should error for empty struct
	})
}

// TestNewAction tests action creation
func TestNewAction(t *testing.T) {
	logger := NewDiscardLogger()
	mockAction := &MockAction{}

	tests := []struct {
		name     string
		wrapped  *MockAction
		nameStr  string
		logger   *slog.Logger
		id       []string
		expected *Action[*MockAction]
	}{
		{
			name:     "with custom ID",
			wrapped:  mockAction,
			nameStr:  "Test Action",
			logger:   logger,
			id:       []string{"custom-id"},
			expected: &Action[*MockAction]{ID: "custom-id", Name: "Test Action", Wrapped: mockAction, Logger: logger},
		},
		{
			name:     "without ID, generates from name",
			wrapped:  mockAction,
			nameStr:  "Test Action",
			logger:   logger,
			id:       []string{},
			expected: &Action[*MockAction]{ID: "test-action", Name: "Test Action", Wrapped: mockAction, Logger: logger},
		},
		{
			name:     "with empty name, generates ID",
			wrapped:  mockAction,
			nameStr:  "",
			logger:   logger,
			id:       []string{},
			expected: &Action[*MockAction]{ID: "", Name: "", Wrapped: mockAction, Logger: logger},
		},
		{
			name:     "with empty ID string",
			wrapped:  mockAction,
			nameStr:  "Test Action",
			logger:   logger,
			id:       []string{""},
			expected: &Action[*MockAction]{ID: "test-action", Name: "Test Action", Wrapped: mockAction, Logger: logger},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewAction(tt.wrapped, tt.nameStr, tt.logger, tt.id...)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Wrapped, result.Wrapped)
			assert.Equal(t, tt.expected.Logger, result.Logger)
		})
	}
}

// TestGenerateIDFromName tests ID generation from names
func TestGenerateIDFromName(t *testing.T) {
	logger := NewDiscardLogger()
	mockAction := &MockAction{}

	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "simple name",
			expected: "simple-name",
		},
		{
			name:     "with spaces",
			expected: "with-spaces",
		},
		{
			name:     "with underscores",
			expected: "with-underscores",
		},
		{
			name:     "with mixed case",
			expected: "with-mixed-case",
		},
		{
			name:     "with special characters",
			expected: "with-special-characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewAction(mockAction, tt.name, logger)
			assert.Equal(t, tt.expected, result.ID)
		})
	}
}

// Helper types for testing
type testStringer struct{}

func (t testStringer) String() string {
	return "test stringer"
}

type MockAction struct{}

func (m *MockAction) Execute(ctx context.Context) error {
	return nil
}

func (m *MockAction) GetOutput() interface{} {
	return nil
}

func (m *MockAction) BeforeExecute(ctx context.Context) error {
	return nil
}

func (m *MockAction) AfterExecute(ctx context.Context) error {
	return nil
}

// Helper function to create test context
func testContext() context.Context {
	return context.Background()
}

// NewDiscardLogger creates a logger that discards all output
func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
