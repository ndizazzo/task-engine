package task_engine

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// ActionParameter interface for all parameter types that can be resolved at runtime
// to provide values for action execution. Parameters support references to outputs
// from other actions, tasks, or static values.
type ActionParameter interface {
	// Resolve returns the actual value for this parameter by looking up
	// references in the global context or returning static values.
	Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error)
}

// StaticParameter represents a fixed value that doesn't need resolution.
// Use this for values known at task creation time.
type StaticParameter struct {
	Value interface{} // The static value to use
}

func (p StaticParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error) {
	return p.Value, nil
}

// ActionOutputParameter references output from a specific action.
// Use this to pass data between actions within the same task.
type ActionOutputParameter struct {
	ActionID  string // Required: ID of the action to reference
	OutputKey string // Optional: specific output field to extract (omit for entire output)
}

func (p ActionOutputParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error) {
	if p.ActionID == "" {
		return nil, fmt.Errorf("ActionOutputParameter: ActionID cannot be empty")
	}

	output, exists := globalContext.ActionOutputs[p.ActionID]
	if !exists {
		return nil, fmt.Errorf("ActionOutputParameter: action '%s' not found in context", p.ActionID)
	}

	if p.OutputKey != "" {
		// Validate OutputKey exists in output
		if outputMap, ok := output.(map[string]interface{}); ok {
			if value, exists := outputMap[p.OutputKey]; exists {
				return value, nil
			}
			return nil, fmt.Errorf("ActionOutputParameter: output key '%s' not found in action '%s'", p.OutputKey, p.ActionID)
		}
		return nil, fmt.Errorf("ActionOutputParameter: action '%s' output is not a map, cannot extract key '%s'", p.ActionID, p.OutputKey)
	}

	return output, nil
}

// ActionResultParameter references results from actions implementing ResultProvider
type ActionResultParameter struct {
	ActionID  string // Required: ID of the action to reference
	ResultKey string // Optional: specific result field to extract
}

func (p ActionResultParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error) {
	if p.ActionID == "" {
		return nil, fmt.Errorf("ActionResultParameter: ActionID cannot be empty")
	}

	resultProvider, exists := globalContext.ActionResults[p.ActionID]
	if !exists {
		return nil, fmt.Errorf("ActionResultParameter: action '%s' not found in context", p.ActionID)
	}

	result := resultProvider.GetResult()
	if p.ResultKey != "" {
		// Extract specific field from result
		if resultMap, ok := result.(map[string]interface{}); ok {
			if value, exists := resultMap[p.ResultKey]; exists {
				return value, nil
			}
			return nil, fmt.Errorf("ActionResultParameter: result key '%s' not found in action '%s'", p.ResultKey, p.ActionID)
		}
		return nil, fmt.Errorf("ActionResultParameter: action '%s' result is not a map, cannot extract key '%s'", p.ActionID, p.ResultKey)
	}

	return result, nil
}

// TaskResultParameter references results from tasks implementing ResultProvider
type TaskResultParameter struct {
	TaskID    string // Required: ID of the task to reference
	ResultKey string // Optional: specific result field to extract
}

func (p TaskResultParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error) {
	if p.TaskID == "" {
		return nil, fmt.Errorf("TaskResultParameter: TaskID cannot be empty")
	}

	resultProvider, exists := globalContext.TaskResults[p.TaskID]
	if !exists {
		return nil, fmt.Errorf("TaskResultParameter: task '%s' not found in context", p.TaskID)
	}

	result := resultProvider.GetResult()
	if p.ResultKey != "" {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if value, exists := resultMap[p.ResultKey]; exists {
				return value, nil
			}
			return nil, fmt.Errorf("TaskResultParameter: result key '%s' not found in task '%s'", p.ResultKey, p.TaskID)
		}
		return nil, fmt.Errorf("TaskResultParameter: task '%s' result is not a map, cannot extract key '%s'", p.TaskID, p.ResultKey)
	}

	return result, nil
}

// TaskOutputParameter references output from a specific task
type TaskOutputParameter struct {
	TaskID    string // Required: ID of the task to reference
	OutputKey string // Optional: specific output field to extract
}

func (p TaskOutputParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error) {
	if p.TaskID == "" {
		return nil, fmt.Errorf("TaskOutputParameter: TaskID cannot be empty")
	}

	output, exists := globalContext.TaskOutputs[p.TaskID]
	if !exists {
		return nil, fmt.Errorf("TaskOutputParameter: task '%s' not found in context", p.TaskID)
	}

	if p.OutputKey != "" {
		// Extract specific field from output
		if outputMap, ok := output.(map[string]interface{}); ok {
			if value, exists := outputMap[p.OutputKey]; exists {
				return value, nil
			}
			return nil, fmt.Errorf("TaskOutputParameter: output key '%s' not found in task '%s'", p.OutputKey, p.TaskID)
		}
		return nil, fmt.Errorf("TaskOutputParameter: task '%s' output is not a map, cannot extract key '%s'", p.TaskID, p.OutputKey)
	}

	return output, nil
}

// EntityOutputParameter references output from any entity (action or task)
type EntityOutputParameter struct {
	EntityType string // Required: "action" or "task"
	EntityID   string // Required: ID of the entity to reference
	OutputKey  string // Optional: specific output field to extract
}

func (p EntityOutputParameter) Resolve(ctx context.Context, globalContext *GlobalContext) (interface{}, error) {
	if p.EntityType == "" || p.EntityID == "" {
		return nil, fmt.Errorf("EntityOutputParameter: EntityType and EntityID cannot be empty")
	}

	const (
		entityTypeAction = "action"
		entityTypeTask   = "task"
	)

	switch p.EntityType {
	case entityTypeAction:
		// Try ActionOutputs first
		if output, exists := globalContext.ActionOutputs[p.EntityID]; exists {
			if p.OutputKey != "" {
				if outputMap, ok := output.(map[string]interface{}); ok {
					if value, exists := outputMap[p.OutputKey]; exists {
						return value, nil
					}
					return nil, fmt.Errorf("EntityOutputParameter: output key '%s' not found in action '%s'", p.OutputKey, p.EntityID)
				}
				return nil, fmt.Errorf("EntityOutputParameter: action '%s' output is not a map, cannot extract key '%s'", p.EntityID, p.OutputKey)
			}
			return output, nil
		}
		// Try ActionResults if ActionOutputs doesn't have it
		if resultProvider, exists := globalContext.ActionResults[p.EntityID]; exists {
			result := resultProvider.GetResult()
			if p.OutputKey != "" {
				if resultMap, ok := result.(map[string]interface{}); ok {
					if value, exists := resultMap[p.OutputKey]; exists {
						return value, nil
					}
					return nil, fmt.Errorf("EntityOutputParameter: result key '%s' not found in action '%s'", p.OutputKey, p.EntityID)
				}
				return nil, fmt.Errorf("EntityOutputParameter: action '%s' result is not a map, cannot extract key '%s'", p.EntityID, p.OutputKey)
			}
			return result, nil
		}
		return nil, fmt.Errorf("EntityOutputParameter: action '%s' not found in context", p.EntityID)

	case entityTypeTask:
		// Try TaskOutputs first
		if output, exists := globalContext.TaskOutputs[p.EntityID]; exists {
			if p.OutputKey != "" {
				if outputMap, ok := output.(map[string]interface{}); ok {
					if value, exists := outputMap[p.OutputKey]; exists {
						return value, nil
					}
					return nil, fmt.Errorf("EntityOutputParameter: output key '%s' not found in task '%s'", p.OutputKey, p.EntityID)
				}
				return nil, fmt.Errorf("EntityOutputParameter: task '%s' output is not a map, cannot extract key '%s'", p.EntityID, p.OutputKey)
			}
			return output, nil
		}
		// Try TaskResults if TaskOutputs doesn't have it
		if resultProvider, exists := globalContext.TaskResults[p.EntityID]; exists {
			result := resultProvider.GetResult()
			if p.OutputKey != "" {
				if resultMap, ok := result.(map[string]interface{}); ok {
					if value, exists := resultMap[p.OutputKey]; exists {
						return value, nil
					}
					return nil, fmt.Errorf("EntityOutputParameter: result key '%s' not found in task '%s'", p.OutputKey, p.EntityID)
				}
				return nil, fmt.Errorf("EntityOutputParameter: task '%s' result is not a map, cannot extract key '%s'", p.EntityID, p.OutputKey)
			}
			return result, nil
		}
		return nil, fmt.Errorf("EntityOutputParameter: task '%s' not found in context", p.EntityID)

	default:
		return nil, fmt.Errorf("EntityOutputParameter: invalid entity type '%s', must be 'action' or 'task'", p.EntityType)
	}
}

// --- Typed parameter resolution helpers ---

// ResolveString resolves an ActionParameter to a string with helpful
// conversions and clear error messages. When the parameter is nil,
// it returns an empty string without error.
func ResolveString(ctx context.Context, p ActionParameter, globalContext *GlobalContext) (string, error) {
	if p == nil {
		return "", nil
	}
	v, err := p.Resolve(ctx, globalContext)
	if err != nil {
		return "", err
	}
	switch t := v.(type) {
	case string:
		return t, nil
	case []byte:
		return string(t), nil
	case fmt.Stringer:
		return t.String(), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return fmt.Sprint(v), nil
	default:
		return "", fmt.Errorf("parameter is not a string, got %T", v)
	}
}

// ResolveBool resolves an ActionParameter to a bool with common coercions.
// If parameter is nil, returns false.
func ResolveBool(ctx context.Context, p ActionParameter, globalContext *GlobalContext) (bool, error) {
	if p == nil {
		return false, nil
	}
	v, err := p.Resolve(ctx, globalContext)
	if err != nil {
		return false, err
	}
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		s := strings.TrimSpace(strings.ToLower(t))
		if s == "true" || s == "1" || s == "yes" || s == "y" { // common truthy strings
			return true, nil
		}
		if s == "false" || s == "0" || s == "no" || s == "n" {
			return false, nil
		}
		return false, fmt.Errorf("cannot convert string '%s' to bool", t)
	case int:
		return t != 0, nil
	case int64:
		return t != 0, nil
	case uint:
		return t != 0, nil
	default:
		return false, fmt.Errorf("parameter is not a bool, got %T", v)
	}
}

// ResolveStringSlice resolves an ActionParameter into a []string.
// Accepts []string directly, or splits a string by comma or spaces.
func ResolveStringSlice(ctx context.Context, p ActionParameter, globalContext *GlobalContext) ([]string, error) {
	if p == nil {
		return nil, nil
	}
	v, err := p.Resolve(ctx, globalContext)
	if err != nil {
		return nil, err
	}
	switch t := v.(type) {
	case []string:
		return t, nil
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return []string{}, nil
		}
		if strings.Contains(s, ",") {
			parts := strings.Split(s, ",")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			return out, nil
		}
		return strings.Fields(s), nil
	default:
		return nil, fmt.Errorf("parameter is not a string slice or string, got %T", v)
	}
}

// ResolveAs provides a generic typed resolver using existing parameter resolution.
func ResolveAs[T any](ctx context.Context, p ActionParameter, globalContext *GlobalContext) (T, error) {
	var zero T
	if p == nil {
		return zero, nil
	}
	v, err := p.Resolve(ctx, globalContext)
	if err != nil {
		return zero, err
	}
	out, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("expected %T, got %T", zero, v)
	}
	return out, nil
}

// Helper functions for common parameter patterns
// ActionOutput creates a parameter reference to an entire action output
func ActionOutput(actionID string) ActionOutputParameter {
	return ActionOutputParameter{ActionID: actionID}
}

// ActionOutputField creates a parameter reference to a specific field in an action output
func ActionOutputField(actionID, field string) ActionOutputParameter {
	return ActionOutputParameter{ActionID: actionID, OutputKey: field}
}

// ActionResult creates a parameter reference to an action result (for ResultProvider actions)
func ActionResult(actionID string) ActionResultParameter {
	return ActionResultParameter{ActionID: actionID}
}

// ActionResultField creates a parameter reference to a specific field in an action result
func ActionResultField(actionID, field string) ActionResultParameter {
	return ActionResultParameter{ActionID: actionID, ResultKey: field}
}

// TaskOutput creates a parameter reference to an entire task output
func TaskOutput(taskID string) TaskOutputParameter {
	return TaskOutputParameter{TaskID: taskID}
}

// TaskOutputField creates a parameter reference to a specific field in a task output
func TaskOutputField(taskID, field string) TaskOutputParameter {
	return TaskOutputParameter{TaskID: taskID, OutputKey: field}
}

// TaskResult creates a parameter reference to an entire task result (for ResultProvider tasks)
func TaskResult(taskID string) TaskResultParameter {
	return TaskResultParameter{TaskID: taskID}
}

// TaskResultField creates a parameter reference to a specific field in a task result
func TaskResultField(taskID, field string) TaskResultParameter {
	return TaskResultParameter{TaskID: taskID, ResultKey: field}
}

// EntityOutput creates a parameter reference to any entity type (action or task)
func EntityOutput(entityType, entityID string) EntityOutputParameter {
	return EntityOutputParameter{EntityType: entityType, EntityID: entityID}
}

// EntityOutputField creates a parameter reference to a specific field in any entity output
func EntityOutputField(entityType, entityID, field string) EntityOutputParameter {
	return EntityOutputParameter{EntityType: entityType, EntityID: entityID, OutputKey: field}
}

// TypedOutputKey provides a way to associate an output field name with an expected
// struct type T. Validate can be used to check that the field exists on T at runtime.
// Note: This is a runtime validation helper; compile-time validation would require codegen.
// TypedOutputKey provides compile-time validation of output keys for type-safe
// parameter references. Use this when you want to ensure output keys exist
// in your output types at compile time.
type TypedOutputKey[T any] struct {
	ActionID string // ID of the action to reference
	Key      string // Field name to extract from the output
}

// Validate checks whether Key is a valid exported field on T when T is a struct.
// If T is not a struct, Validate returns nil (no validation performed).
func (k TypedOutputKey[T]) Validate() error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return nil
	}
	if _, exists := t.FieldByName(k.Key); !exists {
		return fmt.Errorf("field '%s' does not exist on output type %s", k.Key, t.Name())
	}
	return nil
}
