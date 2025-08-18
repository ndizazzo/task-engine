package common

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
)

// BaseOutputBuilder provides generic functionality for building action outputs
// and resolving parameters, eliminating duplicate code across actions
type BaseOutputBuilder[T any] struct {
	logger *slog.Logger
}

// NewBaseOutputBuilder creates a new base output builder with the given logger
func NewBaseOutputBuilder[T any](logger *slog.Logger) *BaseOutputBuilder[T] {
	return &BaseOutputBuilder[T]{logger: logger}
}

// GetLogger returns the logger from the base output builder
func (b *BaseOutputBuilder[T]) GetLogger() *slog.Logger {
	return b.logger
}

// ResolveParameter is a generic helper for resolving action parameters
// It handles the common pattern of extracting GlobalContext and calling Resolve
func (b *BaseOutputBuilder[T]) ResolveParameter(
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
func (b *BaseOutputBuilder[T]) ResolveStringParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (string, error) {
	value, err := b.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return "", err
	}

	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("%s parameter resolved to non-string value: %T", paramName, value)
}

// ResolveBoolParameter resolves a parameter and converts it to a boolean
func (b *BaseOutputBuilder[T]) ResolveBoolParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (bool, error) {
	value, err := b.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return false, err
	}

	if b, ok := value.(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("%s parameter resolved to non-boolean value: %T", paramName, value)
}

// ResolveIntParameter resolves a parameter and converts it to an integer
func (b *BaseOutputBuilder[T]) ResolveIntParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) (int, error) {
	value, err := b.ResolveParameter(ctx, param, paramName)
	if err != nil {
		return 0, err
	}

	if i, ok := value.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("%s parameter resolved to non-integer value: %T", paramName, value)
}

// ResolveStringSliceParameter resolves a parameter and converts it to a string slice
func (b *BaseOutputBuilder[T]) ResolveStringSliceParameter(
	ctx context.Context,
	param task_engine.ActionParameter,
	paramName string,
) ([]string, error) {
	value, err := b.ResolveParameter(ctx, param, paramName)
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

// BuildStandardOutput creates a standard output map with common fields
// This eliminates the repetitive pattern of building map[string]interface{} outputs
func (b *BaseOutputBuilder[T]) BuildStandardOutput(
	output interface{},
	success bool,
	additionalFields map[string]interface{},
) map[string]interface{} {
	result := map[string]interface{}{
		"output":  output,
		"success": success,
	}

	// Add any additional fields
	for key, value := range additionalFields {
		result[key] = value
	}

	return result
}

// BuildOutputFromStruct automatically generates an output map from a struct
// by reflecting over its fields and including non-zero values
func (b *BaseOutputBuilder[T]) BuildOutputFromStruct(
	action T,
	success bool,
	excludeFields []string,
) map[string]interface{} {
	result := map[string]interface{}{
		"success": success,
	}

	// Create a set of fields to exclude for faster lookup
	excludeSet := make(map[string]bool)
	for _, field := range excludeFields {
		excludeSet[field] = true
	}

	// Use reflection to get struct fields
	v := reflect.ValueOf(action)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// Fall back to standard output if not a struct
		return result
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		fieldName := fieldType.Name
		fieldValue := field.Interface()

		// Skip excluded fields
		if excludeSet[fieldName] {
			continue
		}

		// Skip zero values (nil, empty string, 0, false)
		if !field.IsZero() {
			// Convert field name to camelCase for consistency
			camelCaseName := strings.ToLower(fieldName[:1]) + fieldName[1:]
			result[camelCaseName] = fieldValue
		}
	}

	return result
}

// BuildOutputWithCount creates an output with a count field for slice results
func (b *BaseOutputBuilder[T]) BuildOutputWithCount(
	items interface{},
	success bool,
	additionalFields map[string]interface{},
) map[string]interface{} {
	result := b.BuildStandardOutput(items, success, additionalFields)

	// Add count if items is a slice
	if items != nil {
		v := reflect.ValueOf(items)
		if v.Kind() == reflect.Slice {
			result["count"] = v.Len()
		}
	}

	return result
}
