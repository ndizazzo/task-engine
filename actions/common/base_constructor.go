package common

import (
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
)

// BaseConstructor provides common constructor functionality for all actions
type BaseConstructor[T task_engine.ActionInterface] struct {
	logger *slog.Logger
}

// NewBaseConstructor creates a new base constructor with the given logger
func NewBaseConstructor[T task_engine.ActionInterface](logger *slog.Logger) *BaseConstructor[T] {
	return &BaseConstructor[T]{logger: logger}
}

// GetLogger returns the logger from the base constructor
func (c *BaseConstructor[T]) GetLogger() *slog.Logger {
	return c.logger
}

// WrapAction wraps an action with common fields and handles ID generation
func (c *BaseConstructor[T]) WrapAction(
	action T,
	name string,
	id ...string,
) *task_engine.Action[T] {
	actionID := ""
	if len(id) > 0 && id[0] != "" {
		actionID = id[0]
	} else {
		actionID = generateActionID(name)
	}

	return &task_engine.Action[T]{
		ID:      actionID,
		Name:    name,
		Wrapped: action,
	}
}

// generateActionID creates a consistent action ID from name
func generateActionID(name string) string {
	return task_engine.SanitizeIDPart(name) + "-action"
}
