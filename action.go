package task_engine

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// GlobalContextKey is the key used to store the global context in the context
const GlobalContextKey contextKey = "globalContext"

// ActionInterface defines the contract for actions
type ActionInterface interface {
	BeforeExecute(ctx context.Context) error
	Execute(ctx context.Context) error
	AfterExecute(ctx context.Context) error
	// GetOutput returns the action's execution results for parameter passing
	// between actions and tasks. Return a map[string]interface{} for structured output.
	GetOutput() interface{}
}

// ActionWithResults interface for actions that can optionally provide rich results
type ActionWithResults interface {
	ActionInterface
	ResultProvider
}

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

	switch p.EntityType {
	case "action":
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

	case "task":
		output, exists := globalContext.TaskOutputs[p.EntityID]
		if !exists {
			return nil, fmt.Errorf("EntityOutputParameter: task '%s' not found in context", p.EntityID)
		}
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

// --- Consistent action ID helpers ---

var idSanitizer = regexp.MustCompile(`[^a-z0-9_:\-.]+`)

// SanitizeIDPart makes a value safe for inclusion in an action ID.
// Lowercases, trims, replaces spaces and slashes with '-', and strips
// characters outside [a-z0-9_:\-.].
func SanitizeIDPart(s string) string {
	n := strings.TrimSpace(strings.ToLower(s))
	if n == "" {
		return ""
	}
	n = strings.ReplaceAll(n, " ", "-")
	n = strings.ReplaceAll(n, "/", "-")
	n = idSanitizer.ReplaceAllString(n, "")
	return n
}

// BuildActionID constructs a consistent action ID with a prefix and
// optional sanitized parts: prefix-part1-part2-action. Empty parts are skipped.
func BuildActionID(prefix string, parts ...string) string {
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		sp := SanitizeIDPart(p)
		if sp != "" {
			cleaned = append(cleaned, sp)
		}
	}
	base := SanitizeIDPart(prefix)
	if base == "" {
		base = "action"
	}
	if len(cleaned) == 0 {
		return base + "-action"
	}
	return base + "-" + strings.Join(cleaned, "-") + "-action"
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

// EntityOutput creates a parameter reference to any entity type (action or task)
func EntityOutput(entityType, entityID string) EntityOutputParameter {
	return EntityOutputParameter{EntityType: entityType, EntityID: entityID}
}

// EntityOutputField creates a parameter reference to a specific field in any entity output
func EntityOutputField(entityType, entityID, field string) EntityOutputParameter {
	return EntityOutputParameter{EntityType: entityType, EntityID: entityID, OutputKey: field}
}

// --- Phase 5 Ergonomics ---

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

// BaseAction is used as a composite struct for newly defined actions, to provide a default no-op implementation of the before/after
// hooks. It also has a logger passed from the action that wraps it.
// BaseAction provides common functionality for actions including logging
// and default implementations of common methods. Embed this in your
// custom actions to get standard behavior.
type BaseAction struct {
	Logger *slog.Logger // Logger for action execution
}

// NewBaseAction creates a new BaseAction with a logger. If logger is nil, it uses a discard logger.
func NewBaseAction(logger *slog.Logger) BaseAction {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return BaseAction{Logger: logger}
}

func (ba *BaseAction) BeforeExecute(ctx context.Context) error {
	return nil
}

func (a *BaseAction) AfterExecute(ctx context.Context) error {
	return nil
}

// GetOutput provides a default no-op implementation for actions that don't produce outputs
func (ba *BaseAction) GetOutput() interface{} {
	return nil
}

// ---

// GlobalContext maintains state across the entire system for parameter resolution.
// This enables cross-task and cross-action parameter passing by storing outputs
// from all executed entities.
type GlobalContext struct {
	ActionOutputs map[string]interface{}    // Outputs from individual actions
	ActionResults map[string]ResultProvider // Actions implementing ResultProvider
	TaskOutputs   map[string]interface{}    // Outputs from completed tasks
	TaskResults   map[string]ResultProvider // Tasks implementing ResultProvider
	mu            sync.RWMutex              // Protects concurrent access
}

// NewGlobalContext creates a new GlobalContext instance
func NewGlobalContext() *GlobalContext {
	return &GlobalContext{
		ActionOutputs: make(map[string]interface{}),
		ActionResults: make(map[string]ResultProvider),
		TaskOutputs:   make(map[string]interface{}),
		TaskResults:   make(map[string]ResultProvider),
	}
}

// StoreActionOutput stores the output from an action
func (gc *GlobalContext) StoreActionOutput(actionID string, output interface{}) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.ActionOutputs[actionID] = output
}

// StoreActionResult stores the result provider from an action
func (gc *GlobalContext) StoreActionResult(actionID string, resultProvider ResultProvider) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.ActionResults[actionID] = resultProvider
}

// StoreTaskOutput stores the output from a task
func (gc *GlobalContext) StoreTaskOutput(taskID string, output interface{}) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.TaskOutputs[taskID] = output
}

// StoreTaskResult stores the result provider from a task
func (gc *GlobalContext) StoreTaskResult(taskID string, resultProvider ResultProvider) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.TaskResults[taskID] = resultProvider
}

// ActionWrapper interface for actions that can be executed by tasks.
// This interface provides the contract that tasks use to interact with actions,
// including execution, metadata access, and output retrieval.
type ActionWrapper interface {
	Execute(ctx context.Context) error
	GetDuration() time.Duration
	GetLogger() *slog.Logger
	GetID() string
	SetID(string)
	GetName() string
	GetOutput() interface{} // Returns action execution results for parameter passing
}

// Action[T] wraps an ActionInterface implementation with execution tracking,
// lifecycle management, and parameter passing support. This is the main
// type used to create and execute actions in the task engine.
type Action[T ActionInterface] struct {
	ID        string        // Unique identifier for the action
	Name      string        // Human-readable name for the action
	RunID     string        // Unique identifier for this execution run
	Wrapped   T             // The actual action implementation
	StartTime time.Time     // When execution started
	EndTime   time.Time     // When execution completed
	Duration  time.Duration // Total execution time
	Logger    *slog.Logger  // Logger for the action
	mu        sync.RWMutex  // Protects concurrent access to time fields
}

func (a *Action[T]) Execute(ctx context.Context) error {
	return a.InternalExecute(ctx)
}

func (a *Action[T]) GetDuration() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Duration
}

func (a *Action[T]) GetLogger() *slog.Logger {
	return a.Logger
}

func (a *Action[T]) GetID() string {
	return a.ID
}

func (a *Action[T]) SetID(id string) {
	a.ID = id
}

func (a *Action[T]) GetName() string {
	if strings.TrimSpace(a.Name) != "" {
		return a.Name
	}
	return a.ID
}

// GetOutput delegates to the wrapped action's GetOutput method
func (a *Action[T]) GetOutput() interface{} {
	if actionWithOutput, ok := any(a.Wrapped).(interface{ GetOutput() interface{} }); ok {
		return actionWithOutput.GetOutput()
	}
	return nil
}

func (a *Action[T]) InternalExecute(ctx context.Context) error {
	// Auto-generate ID if missing using the action name
	if strings.TrimSpace(a.ID) == "" && strings.TrimSpace(a.Name) != "" {
		a.ID = generateIDFromName(a.Name)
	}
	a.mu.Lock()
	a.RunID = uuid.New().String()
	runID := a.RunID // Store locally to avoid race conditions in logging
	a.mu.Unlock()

	a.log("Starting action", "actionID", a.ID, "runID", runID)

	a.mu.Lock()
	if a.StartTime.IsZero() {
		a.StartTime = time.Now()
	}
	a.mu.Unlock()

	// Ensure context has GlobalContext for parameter resolution
	execCtx := ctx
	if _, ok := ctx.Value(GlobalContextKey).(*GlobalContext); !ok {
		// Context doesn't have GlobalContext, add an empty one for standalone execution
		execCtx = context.WithValue(ctx, GlobalContextKey, NewGlobalContext())
	}

	if err := a.Wrapped.BeforeExecute(execCtx); err != nil {
		a.log("BeforeExecute failed", "actionID", a.ID, "runID", runID, "error", err)
		return err
	}

	if err := a.Wrapped.Execute(execCtx); err != nil {
		a.log("Execute failed", "actionID", a.ID, "runID", runID, "error", err)
		return err
	}

	a.mu.Lock()
	if a.EndTime.IsZero() {
		a.EndTime = time.Now()
	}
	duration := a.EndTime.Sub(a.StartTime)
	a.Duration = duration
	a.mu.Unlock()

	if err := a.Wrapped.AfterExecute(execCtx); err != nil {
		a.log("AfterExecute failed", "actionID", a.ID, "runID", runID, "error", err)
		return err
	}

	a.log("Action completed", "actionID", a.ID, "runID", runID, "duration", duration)
	return nil
}

func (a *Action[T]) log(message string, keyvals ...interface{}) {
	if a.Logger != nil {
		a.Logger.Info(message, keyvals...)
	}
}

// NewAction creates a new Action instance with the given wrapped action, name, and logger.
// Optionally provide a custom ID; if omitted, one will be generated from the name.
func NewAction[T ActionInterface](wrapped T, name string, logger *slog.Logger, id ...string) *Action[T] {
	actionID := ""
	if len(id) > 0 && strings.TrimSpace(id[0]) != "" {
		actionID = id[0]
	} else if strings.TrimSpace(name) != "" {
		actionID = generateIDFromName(name)
	}
	return &Action[T]{
		ID:      actionID,
		Name:    name,
		Wrapped: wrapped,
		Logger:  logger,
	}
}

func generateIDFromName(name string) string {
	n := strings.TrimSpace(name)
	n = strings.ToLower(n)
	n = strings.ReplaceAll(n, " ", "-")
	n = strings.ReplaceAll(n, "_", "-")
	return n
}
