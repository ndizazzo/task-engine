package common

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
)

// ParameterResolver provides common parameter resolution functionality
// that can be embedded into actions to eliminate duplicate code
type ParameterResolver struct {
	logger *slog.Logger
}

// NewParameterResolver creates a new parameter resolver with the given logger
func NewParameterResolver(logger *slog.Logger) *ParameterResolver {
	return &ParameterResolver{logger: logger}
}

// GetLogger returns the logger from the parameter resolver
func (pr *ParameterResolver) GetLogger() *slog.Logger {
	return pr.logger
}

// ResolveParameter is a generic helper for resolving action parameters
// It handles the common pattern of extracting GlobalContext and calling Resolve
func (pr *ParameterResolver) ResolveParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (interface{}, error) {
	if param == nil {
		return nil, fmt.Errorf("%s parameter cannot be nil", paramName)
	}

	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve the parameter
	value, err := param.Resolve(ctx, globalContext)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s parameter: %w", paramName, err)
	}

	return value, nil
}

// ResolveStringParameter resolves a parameter and converts it to a string
func (pr *ParameterResolver) ResolveStringParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (string, error) {
	value, err := pr.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return "", err
	}

	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("%s parameter resolved to non-string value: %T", paramName, value)
}

// ResolveBoolParameter resolves a parameter and converts it to a boolean
func (pr *ParameterResolver) ResolveBoolParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (bool, error) {
	value, err := pr.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return false, err
	}

	if b, ok := value.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("%s parameter resolved to non-boolean value: %T", paramName, value)
}

// ResolveIntParameter resolves a parameter and converts it to an integer
func (pr *ParameterResolver) ResolveIntParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (int, error) {
	value, err := pr.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return 0, err
	}

	if i, ok := value.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("%s parameter resolved to non-integer value: %T", paramName, value)
}

// ResolveStringSliceParameter resolves a parameter and converts it to a string slice
func (pr *ParameterResolver) ResolveStringSliceParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) ([]string, error) {
	value, err := pr.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return nil, err
	}

	if slice, ok := value.([]string); ok {
		return slice, nil
	}

	// Handle single string case
	if str, ok := value.(string); ok {
		return []string{str}, nil
	}

	return nil, fmt.Errorf("%s parameter resolved to non-string-slice value: %T", paramName, value)
}

// ResolveDurationParameter resolves a parameter and converts it to a time.Duration
func (pr *ParameterResolver) ResolveDurationParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (time.Duration, error) {
	value, err := pr.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return 0, err
	}

	// Handle different duration formats
	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		// Parse duration string (e.g., "5s", "1m", "2h")
		parsedDuration, err := time.ParseDuration(v)
		if err != nil {
			return 0, fmt.Errorf("failed to parse duration string '%s': %w", v, err)
		}
		return parsedDuration, nil
	case int:
		// Treat as seconds
		return time.Duration(v) * time.Second, nil
	default:
		return 0, fmt.Errorf("%s parameter resolved to unsupported duration type: %T", paramName, value)
	}
}

// ResolveMapParameter resolves a parameter and converts it to a map
func (pr *ParameterResolver) ResolveMapParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (map[string]interface{}, error) {
	value, err := pr.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return nil, err
	}

	if m, ok := value.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, fmt.Errorf("%s parameter resolved to non-map value: %T", paramName, value)
}

// ResolveSliceParameter resolves a parameter and converts it to a slice
func (pr *ParameterResolver) ResolveSliceParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) ([]interface{}, error) {
	value, err := pr.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return nil, err
	}

	if slice, ok := value.([]interface{}); ok {
		return slice, nil
	}

	// Handle typed slices by converting to interface slice
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Slice {
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result, nil
	}

	return nil, fmt.Errorf("%s parameter resolved to non-slice value: %T", paramName, value)
}
