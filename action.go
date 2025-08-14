package task_engine

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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

// --- Typed convenience helpers (simplest way to fetch data) ---

// ActionResultAs returns a typed action result from an action implementing ResultProvider.
func ActionResultAs[T any](gc *GlobalContext, actionID string) (T, bool) {
	gc.mu.RLock()
	rp, ok := gc.ActionResults[actionID]
	gc.mu.RUnlock()
	var zero T
	if !ok || rp == nil {
		return zero, false
	}
	v, ok := rp.GetResult().(T)
	return v, ok
}

// TaskResultAs returns a typed task result from a task implementing ResultProvider.
func TaskResultAs[T any](gc *GlobalContext, taskID string) (T, bool) {
	gc.mu.RLock()
	rp, ok := gc.TaskResults[taskID]
	gc.mu.RUnlock()
	var zero T
	if !ok || rp == nil {
		return zero, false
	}
	v, ok := rp.GetResult().(T)
	return v, ok
}

// ActionOutputFieldAs returns a typed value from an action's output map.
func ActionOutputFieldAs[T any](gc *GlobalContext, actionID, key string) (T, error) {
	gc.mu.RLock()
	output, exists := gc.ActionOutputs[actionID]
	gc.mu.RUnlock()
	var zero T
	if !exists {
		return zero, fmt.Errorf("action '%s' not found in context", actionID)
	}
	if key == "" {
		if v, ok := output.(T); ok {
			return v, nil
		}
		return zero, fmt.Errorf("action '%s' output is not %T", actionID, zero)
	}
	m, ok := output.(map[string]interface{})
	if !ok {
		return zero, fmt.Errorf("action '%s' output is not a map, cannot extract key '%s'", actionID, key)
	}
	val, exists := m[key]
	if !exists {
		return zero, fmt.Errorf("output key '%s' not found in action '%s'", key, actionID)
	}
	typed, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("action '%s' output key '%s' is not %T", actionID, key, zero)
	}
	return typed, nil
}

// TaskOutputFieldAs returns a typed value from a task's output map.
func TaskOutputFieldAs[T any](gc *GlobalContext, taskID, key string) (T, error) {
	gc.mu.RLock()
	output, exists := gc.TaskOutputs[taskID]
	gc.mu.RUnlock()
	var zero T
	if !exists {
		return zero, fmt.Errorf("task '%s' not found in context", taskID)
	}
	if key == "" {
		if v, ok := output.(T); ok {
			return v, nil
		}
		return zero, fmt.Errorf("task '%s' output is not %T", taskID, zero)
	}
	m, ok := output.(map[string]interface{})
	if !ok {
		return zero, fmt.Errorf("task '%s' output is not a map, cannot extract key '%s'", taskID, key)
	}
	val, exists := m[key]
	if !exists {
		return zero, fmt.Errorf("output key '%s' not found in task '%s'", key, taskID)
	}
	typed, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("task '%s' output key '%s' is not %T", taskID, key, zero)
	}
	return typed, nil
}

// EntityValue returns a value from either outputs or results for the given entity.
// For actions, tries ActionOutputs then ActionResults. For tasks, tries TaskOutputs then TaskResults.
func EntityValue(gc *GlobalContext, entityType, id, key string) (interface{}, error) {
	switch entityType {
	case "action":
		if key == "" {
			gc.mu.RLock()
			out, exists := gc.ActionOutputs[id]
			gc.mu.RUnlock()
			if exists {
				return out, nil
			}
			gc.mu.RLock()
			rp, exists := gc.ActionResults[id]
			gc.mu.RUnlock()
			if exists && rp != nil {
				return rp.GetResult(), nil
			}
			return nil, fmt.Errorf("action '%s' not found in context", id)
		}
		return ActionOutputFieldAs[interface{}](gc, id, key)
	case "task":
		if key == "" {
			gc.mu.RLock()
			out, exists := gc.TaskOutputs[id]
			gc.mu.RUnlock()
			if exists {
				return out, nil
			}
			gc.mu.RLock()
			rp, exists := gc.TaskResults[id]
			gc.mu.RUnlock()
			if exists && rp != nil {
				return rp.GetResult(), nil
			}
			return nil, fmt.Errorf("task '%s' not found in context", id)
		}
		return TaskOutputFieldAs[interface{}](gc, id, key)
	default:
		return nil, fmt.Errorf("invalid entity type '%s'", entityType)
	}
}

// EntityValueAs returns a typed value from either outputs or results for the given entity.
func EntityValueAs[T any](gc *GlobalContext, entityType, id, key string) (T, error) {
	var zero T
	v, err := EntityValue(gc, entityType, id, key)
	if err != nil {
		return zero, err
	}
	out, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("entity '%s' value is not %T", id, zero)
	}
	return out, nil
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
