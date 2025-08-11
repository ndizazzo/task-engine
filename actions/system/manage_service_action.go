package system

import (
	"context"
	"fmt"
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewManageServiceAction creates a new ManageServiceAction with the given logger
func NewManageServiceAction(logger *slog.Logger) *ManageServiceAction {
	return &ManageServiceAction{
		BaseAction:       task_engine.NewBaseAction(logger),
		CommandProcessor: command.NewDefaultCommandRunner(),
	}
}

type ManageServiceAction struct {
	task_engine.BaseAction

	// Parameters
	ServiceNameParam task_engine.ActionParameter
	ActionTypeParam  task_engine.ActionParameter

	// Runtime resolved values
	ServiceName      string
	ActionType       string
	CommandProcessor command.CommandRunner
}

// WithParameters sets the parameters for service name and action type and returns a wrapped Action
func (a *ManageServiceAction) WithParameters(serviceNameParam, actionTypeParam task_engine.ActionParameter) (*task_engine.Action[*ManageServiceAction], error) {
	a.ServiceNameParam = serviceNameParam
	a.ActionTypeParam = actionTypeParam

	id := "manage-service-action"
	return &task_engine.Action[*ManageServiceAction]{
		ID:      id,
		Name:    "Manage Service",
		Wrapped: a,
	}, nil
}

func (a *ManageServiceAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve service name parameter if it exists
	if a.ServiceNameParam != nil {
		serviceNameValue, err := a.ServiceNameParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve service name parameter: %w", err)
		}
		if serviceNameStr, ok := serviceNameValue.(string); ok {
			a.ServiceName = serviceNameStr
		} else {
			return fmt.Errorf("service name parameter is not a string, got %T", serviceNameValue)
		}
	}

	// Resolve action type parameter if it exists
	if a.ActionTypeParam != nil {
		actionTypeValue, err := a.ActionTypeParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve action type parameter: %w", err)
		}
		if actionTypeStr, ok := actionTypeValue.(string); ok {
			a.ActionType = actionTypeStr
		} else {
			return fmt.Errorf("action type parameter is not a string, got %T", actionTypeValue)
		}
	}

	switch a.ActionType {
	case "start", "stop", "restart":
		// Dont allow anything except these commands to be passed to systemctl
	default:
		err := fmt.Errorf("invalid action type: %s; must be 'start', 'stop', or 'restart'", a.ActionType)
		return err
	}

	_, err := a.CommandProcessor.RunCommand("systemctl", a.ActionType, a.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to %s service %s: %w", a.ActionType, a.ServiceName, err)
	}

	return nil
}

// GetOutput returns the service operation performed
func (a *ManageServiceAction) GetOutput() interface{} {
	return map[string]interface{}{
		"service": a.ServiceName,
		"action":  a.ActionType,
		"success": true,
	}
}
