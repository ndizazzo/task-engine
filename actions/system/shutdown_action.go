package system

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

type ShutdownAction struct {
	task_engine.BaseAction
	CommandProcessor command.CommandRunner

	// Parameter-only fields
	OperationParam task_engine.ActionParameter
	DelayParam     task_engine.ActionParameter
}

type ShutdownCommandOperation string

const (
	ShutdownOperation_Shutdown ShutdownCommandOperation = "shutdown"
	ShutdownOperation_Restart  ShutdownCommandOperation = "restart"
	ShutdownOperation_Sleep    ShutdownCommandOperation = "sleep"
)

// NewShutdownAction creates a ShutdownAction instance
func NewShutdownAction(logger *slog.Logger) *ShutdownAction {
	return &ShutdownAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		CommandProcessor: command.NewDefaultCommandRunner(),
	}
}

// WithParameters sets the parameters and returns a wrapped Action
func (a *ShutdownAction) WithParameters(operationParam, delayParam task_engine.ActionParameter) (*task_engine.Action[*ShutdownAction], error) {
	if operationParam == nil || delayParam == nil {
		return nil, fmt.Errorf("operationParam and delayParam cannot be nil")
	}

	a.OperationParam = operationParam
	a.DelayParam = delayParam

	return &task_engine.Action[*ShutdownAction]{
		ID:      "shutdown-action",
		Name:    "Shutdown",
		Wrapped: a,
	}, nil
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *ShutdownAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *ShutdownAction) Execute(ctx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve operation parameter
	var operation ShutdownCommandOperation
	if a.OperationParam != nil {
		operationValue, err := a.OperationParam.Resolve(ctx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve operation parameter: %w", err)
		}
		if operationStr, ok := operationValue.(string); ok {
			operation = ShutdownCommandOperation(operationStr)
		} else {
			return fmt.Errorf("operation parameter is not a string, got %T", operationValue)
		}
	}

	// Resolve delay parameter
	var delay time.Duration
	if a.DelayParam != nil {
		delayValue, err := a.DelayParam.Resolve(ctx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve delay parameter: %w", err)
		}
		switch v := delayValue.(type) {
		case time.Duration:
			delay = v
		case int:
			delay = time.Duration(v) * time.Second
		case int64:
			delay = time.Duration(v) * time.Second
		case string:
			parsed, parseErr := time.ParseDuration(v)
			if parseErr != nil {
				return fmt.Errorf("delay parameter could not be parsed as duration: %w", parseErr)
			}
			delay = parsed
		default:
			return fmt.Errorf("delay parameter is not a valid type, got %T", delayValue)
		}
	}

	additionalFlags := shutdownArgs(operation, delay)
	_, err := a.CommandProcessor.RunCommand("shutdown", additionalFlags...)
	return err
}

// GetOutput returns the requested shutdown operation and delay
func (a *ShutdownAction) GetOutput() interface{} {
	return map[string]interface{}{
		"success": true,
	}
}

func shutdownArgs(operation ShutdownCommandOperation, duration time.Duration) []string {
	flags := []string{}

	switch operation {
	case ShutdownOperation_Shutdown:
		flags = append(flags, "-h")
	case ShutdownOperation_Restart:
		flags = append(flags, "-r")
	case ShutdownOperation_Sleep:
		flags = append(flags, "-s")
	}

	if duration <= 0 {
		return append(flags, "now")
	}

	seconds := int(duration.Seconds())
	return append(flags, fmt.Sprintf("+%d", seconds))
}
