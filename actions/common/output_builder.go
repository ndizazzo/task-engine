package common

import (
	"log/slog"
	"reflect"
	"strings"
)

// OutputBuilder provides common output building functionality
// that can be embedded into actions to eliminate duplicate code
type OutputBuilder struct {
	logger *slog.Logger
}

// NewOutputBuilder creates a new output builder with the given logger
func NewOutputBuilder(logger *slog.Logger) *OutputBuilder {
	return &OutputBuilder{logger: logger}
}

// GetLogger returns the logger from the output builder
func (ob *OutputBuilder) GetLogger() *slog.Logger {
	return ob.logger
}

// BuildStandardOutput creates a standard output map with common fields
// This eliminates the repetitive pattern of building map[string]interface{} outputs
func (ob *OutputBuilder) BuildStandardOutput(
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
func (ob *OutputBuilder) BuildOutputFromStruct(
	action interface{},
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
func (ob *OutputBuilder) BuildOutputWithCount(
	items interface{},
	success bool,
	additionalFields map[string]interface{},
) map[string]interface{} {
	result := ob.BuildStandardOutput(items, success, additionalFields)

	// Add count if items is a slice
	if items != nil {
		v := reflect.ValueOf(items)
		if v.Kind() == reflect.Slice {
			result["count"] = v.Len()
		}
	}

	return result
}

// BuildSimpleOutput creates a simple output with just success and optional message
func (ob *OutputBuilder) BuildSimpleOutput(
	success bool,
	message string,
) map[string]interface{} {
	result := map[string]interface{}{
		"success": success,
	}

	if message != "" {
		result["message"] = message
	}

	return result
}

// BuildErrorOutput creates an output for error cases
func (ob *OutputBuilder) BuildErrorOutput(
	error interface{},
	additionalFields map[string]interface{},
) map[string]interface{} {
	result := map[string]interface{}{
		"success": false,
		"error":   error,
	}

	// Add any additional fields
	for key, value := range additionalFields {
		result[key] = value
	}

	return result
}
