package common_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/stretchr/testify/suite"
)

type MockActionParameter struct {
	value interface{}
	err   error
}

func (m *MockActionParameter) Resolve(ctx context.Context, globalContext *task_engine.GlobalContext) (interface{}, error) {
	return m.value, m.err
}

type ParameterResolverTestSuite struct {
	suite.Suite
	resolver *common.ParameterResolver
	logger   *slog.Logger
}

func (suite *ParameterResolverTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	suite.resolver = common.NewParameterResolver(suite.logger)
}

func (suite *ParameterResolverTestSuite) TestNewParameterResolver() {
	pr := common.NewParameterResolver(suite.logger)
	suite.NotNil(pr)
}

func (suite *ParameterResolverTestSuite) TestGetLogger() {
	logger := suite.resolver.GetLogger()
	suite.NotNil(logger)
	suite.Equal(suite.logger, logger)
}

func (suite *ParameterResolverTestSuite) TestResolveParameter_Success() {
	param := &MockActionParameter{value: "test_value"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveParameter(ctx, param, "testParam")

	suite.NoError(err)
	suite.Equal("test_value", result)
}

func (suite *ParameterResolverTestSuite) TestResolveParameter_NilParameter() {
	ctx := context.Background()

	result, err := suite.resolver.ResolveParameter(ctx, nil, "testParam")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "testParam parameter cannot be nil")
}

func (suite *ParameterResolverTestSuite) TestResolveParameter_ResolutionError() {
	param := &MockActionParameter{err: errors.New("resolution failed")}
	ctx := context.Background()

	result, err := suite.resolver.ResolveParameter(ctx, param, "testParam")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to resolve testParam parameter")
}

func (suite *ParameterResolverTestSuite) TestResolveParameter_WithGlobalContext() {
	globalCtx := &task_engine.GlobalContext{}
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalCtx)
	param := &MockActionParameter{value: "with_global_context"}

	result, err := suite.resolver.ResolveParameter(ctx, param, "testParam")

	suite.NoError(err)
	suite.Equal("with_global_context", result)
}

func (suite *ParameterResolverTestSuite) TestResolveStringParameter_Success() {
	param := &MockActionParameter{value: "string_value"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveStringParameter(ctx, param, "stringParam")

	suite.NoError(err)
	suite.Equal("string_value", result)
}

func (suite *ParameterResolverTestSuite) TestResolveStringParameter_NonStringValue() {
	param := &MockActionParameter{value: 123}
	ctx := context.Background()

	result, err := suite.resolver.ResolveStringParameter(ctx, param, "stringParam")

	suite.Error(err)
	suite.Empty(result)
	suite.Contains(err.Error(), "resolved to non-string value")
}

func (suite *ParameterResolverTestSuite) TestResolveStringParameter_EmptyString() {
	param := &MockActionParameter{value: ""}
	ctx := context.Background()

	result, err := suite.resolver.ResolveStringParameter(ctx, param, "stringParam")

	suite.NoError(err)
	suite.Equal("", result)
}

func (suite *ParameterResolverTestSuite) TestResolveBoolParameter_TrueValue() {
	param := &MockActionParameter{value: true}
	ctx := context.Background()

	result, err := suite.resolver.ResolveBoolParameter(ctx, param, "boolParam")

	suite.NoError(err)
	suite.True(result)
}

func (suite *ParameterResolverTestSuite) TestResolveBoolParameter_FalseValue() {
	param := &MockActionParameter{value: false}
	ctx := context.Background()

	result, err := suite.resolver.ResolveBoolParameter(ctx, param, "boolParam")

	suite.NoError(err)
	suite.False(result)
}

func (suite *ParameterResolverTestSuite) TestResolveBoolParameter_NonBoolValue() {
	param := &MockActionParameter{value: "true"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveBoolParameter(ctx, param, "boolParam")

	suite.Error(err)
	suite.False(result)
	suite.Contains(err.Error(), "resolved to non-boolean value")
}

func (suite *ParameterResolverTestSuite) TestResolveIntParameter_Success() {
	param := &MockActionParameter{value: 42}
	ctx := context.Background()

	result, err := suite.resolver.ResolveIntParameter(ctx, param, "intParam")

	suite.NoError(err)
	suite.Equal(42, result)
}

func (suite *ParameterResolverTestSuite) TestResolveIntParameter_ZeroValue() {
	param := &MockActionParameter{value: 0}
	ctx := context.Background()

	result, err := suite.resolver.ResolveIntParameter(ctx, param, "intParam")

	suite.NoError(err)
	suite.Equal(0, result)
}

func (suite *ParameterResolverTestSuite) TestResolveIntParameter_NegativeValue() {
	param := &MockActionParameter{value: -100}
	ctx := context.Background()

	result, err := suite.resolver.ResolveIntParameter(ctx, param, "intParam")

	suite.NoError(err)
	suite.Equal(-100, result)
}

func (suite *ParameterResolverTestSuite) TestResolveIntParameter_NonIntValue() {
	param := &MockActionParameter{value: "42"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveIntParameter(ctx, param, "intParam")

	suite.Error(err)
	suite.Equal(0, result)
	suite.Contains(err.Error(), "resolved to non-integer value")
}

func (suite *ParameterResolverTestSuite) TestResolveStringSliceParameter_StringSlice() {
	param := &MockActionParameter{value: []string{"a", "b", "c"}}
	ctx := context.Background()

	result, err := suite.resolver.ResolveStringSliceParameter(ctx, param, "sliceParam")

	suite.NoError(err)
	suite.Equal([]string{"a", "b", "c"}, result)
}

func (suite *ParameterResolverTestSuite) TestResolveStringSliceParameter_SingleString() {
	param := &MockActionParameter{value: "single_string"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveStringSliceParameter(ctx, param, "sliceParam")

	suite.NoError(err)
	suite.Equal([]string{"single_string"}, result)
}

func (suite *ParameterResolverTestSuite) TestResolveStringSliceParameter_EmptySlice() {
	param := &MockActionParameter{value: []string{}}
	ctx := context.Background()

	result, err := suite.resolver.ResolveStringSliceParameter(ctx, param, "sliceParam")

	suite.NoError(err)
	suite.Equal([]string{}, result)
	suite.Len(result, 0)
}

func (suite *ParameterResolverTestSuite) TestResolveStringSliceParameter_NonSliceValue() {
	param := &MockActionParameter{value: 123}
	ctx := context.Background()

	result, err := suite.resolver.ResolveStringSliceParameter(ctx, param, "sliceParam")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "resolved to non-string-slice value")
}

func (suite *ParameterResolverTestSuite) TestResolveDurationParameter_DurationValue() {
	duration := 5 * time.Second
	param := &MockActionParameter{value: duration}
	ctx := context.Background()

	result, err := suite.resolver.ResolveDurationParameter(ctx, param, "durationParam")

	suite.NoError(err)
	suite.Equal(5*time.Second, result)
}

func (suite *ParameterResolverTestSuite) TestResolveDurationParameter_StringValue() {
	param := &MockActionParameter{value: "10s"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveDurationParameter(ctx, param, "durationParam")

	suite.NoError(err)
	suite.Equal(10*time.Second, result)
}

func (suite *ParameterResolverTestSuite) TestResolveDurationParameter_StringValueMinutes() {
	param := &MockActionParameter{value: "5m"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveDurationParameter(ctx, param, "durationParam")

	suite.NoError(err)
	suite.Equal(5*time.Minute, result)
}

func (suite *ParameterResolverTestSuite) TestResolveDurationParameter_StringValueHours() {
	param := &MockActionParameter{value: "2h"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveDurationParameter(ctx, param, "durationParam")

	suite.NoError(err)
	suite.Equal(2*time.Hour, result)
}

func (suite *ParameterResolverTestSuite) TestResolveDurationParameter_IntValue() {
	param := &MockActionParameter{value: 30}
	ctx := context.Background()

	result, err := suite.resolver.ResolveDurationParameter(ctx, param, "durationParam")

	suite.NoError(err)
	suite.Equal(30*time.Second, result)
}

func (suite *ParameterResolverTestSuite) TestResolveDurationParameter_InvalidStringValue() {
	param := &MockActionParameter{value: "invalid"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveDurationParameter(ctx, param, "durationParam")

	suite.Error(err)
	suite.Equal(time.Duration(0), result)
	suite.Contains(err.Error(), "failed to parse duration string")
}

func (suite *ParameterResolverTestSuite) TestResolveDurationParameter_UnsupportedType() {
	param := &MockActionParameter{value: []int{1, 2, 3}}
	ctx := context.Background()

	result, err := suite.resolver.ResolveDurationParameter(ctx, param, "durationParam")

	suite.Error(err)
	suite.Equal(time.Duration(0), result)
	suite.Contains(err.Error(), "unsupported duration type")
}

func (suite *ParameterResolverTestSuite) TestResolveMapParameter_Success() {
	mapValue := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	param := &MockActionParameter{value: mapValue}
	ctx := context.Background()

	result, err := suite.resolver.ResolveMapParameter(ctx, param, "mapParam")

	suite.NoError(err)
	suite.Equal(mapValue, result)
}

func (suite *ParameterResolverTestSuite) TestResolveMapParameter_EmptyMap() {
	mapValue := map[string]interface{}{}
	param := &MockActionParameter{value: mapValue}
	ctx := context.Background()

	result, err := suite.resolver.ResolveMapParameter(ctx, param, "mapParam")

	suite.NoError(err)
	suite.Equal(mapValue, result)
	suite.Len(result, 0)
}

func (suite *ParameterResolverTestSuite) TestResolveMapParameter_NonMapValue() {
	param := &MockActionParameter{value: "not a map"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveMapParameter(ctx, param, "mapParam")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "resolved to non-map value")
}

func (suite *ParameterResolverTestSuite) TestResolveSliceParameter_InterfaceSlice() {
	sliceValue := []interface{}{"a", 1, true}
	param := &MockActionParameter{value: sliceValue}
	ctx := context.Background()

	result, err := suite.resolver.ResolveSliceParameter(ctx, param, "sliceParam")

	suite.NoError(err)
	suite.Equal(sliceValue, result)
}

func (suite *ParameterResolverTestSuite) TestResolveSliceParameter_StringSlice() {
	sliceValue := []string{"a", "b", "c"}
	param := &MockActionParameter{value: sliceValue}
	ctx := context.Background()

	result, err := suite.resolver.ResolveSliceParameter(ctx, param, "sliceParam")

	suite.NoError(err)
	suite.Len(result, 3)
	suite.Equal("a", result[0])
	suite.Equal("b", result[1])
	suite.Equal("c", result[2])
}

func (suite *ParameterResolverTestSuite) TestResolveSliceParameter_IntSlice() {
	sliceValue := []int{1, 2, 3, 4, 5}
	param := &MockActionParameter{value: sliceValue}
	ctx := context.Background()

	result, err := suite.resolver.ResolveSliceParameter(ctx, param, "sliceParam")

	suite.NoError(err)
	suite.Len(result, 5)
}

func (suite *ParameterResolverTestSuite) TestResolveSliceParameter_EmptySlice() {
	sliceValue := []interface{}{}
	param := &MockActionParameter{value: sliceValue}
	ctx := context.Background()

	result, err := suite.resolver.ResolveSliceParameter(ctx, param, "sliceParam")

	suite.NoError(err)
	suite.Len(result, 0)
}

func (suite *ParameterResolverTestSuite) TestResolveSliceParameter_NonSliceValue() {
	param := &MockActionParameter{value: "not a slice"}
	ctx := context.Background()

	result, err := suite.resolver.ResolveSliceParameter(ctx, param, "sliceParam")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "resolved to non-slice value")
}

func TestParameterResolverTestSuite(t *testing.T) {
	suite.Run(t, new(ParameterResolverTestSuite))
}
