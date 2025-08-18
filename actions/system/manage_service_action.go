package system

import (
	"context"
	"fmt"
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

// NewManageServiceAction creates a new ManageServiceAction with the given logger
func NewManageServiceAction(logger *slog.Logger) *ManageServiceAction {
	return &ManageServiceAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		CommandProcessor:  command.NewDefaultCommandRunner(),
	}
}

type ManageServiceAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder

	// Parameters
	ServiceNameParam task_engine.ActionParameter
	OperationParam   task_engine.ActionParameter

	// Runtime resolved values
	ServiceName      string
	ActionType       string
	CommandProcessor command.CommandRunner
}

// WithParameters sets the parameters for service management and returns a wrapped Action
func (a *ManageServiceAction) WithParameters(
	serviceNameParam task_engine.ActionParameter,
	operationParam task_engine.ActionParameter,
) (*task_engine.Action[*ManageServiceAction], error) {
	a.ServiceNameParam = serviceNameParam
	a.OperationParam = operationParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*ManageServiceAction](a.Logger)
	return constructor.WrapAction(a, "Manage Service", "manage-service-action"), nil
}

func (a *ManageServiceAction) Execute(execCtx context.Context) error {
	// Resolve required parameters
	serviceName, err := a.ResolveStringParameter(execCtx, a.ServiceNameParam, "service name")
	if err != nil {
		return err
	}
	a.ServiceName = serviceName

	actionType, err := a.ResolveStringParameter(execCtx, a.OperationParam, "action type")
	if err != nil {
		return err
	}
	a.ActionType = actionType

	switch a.ActionType {
	case "start", "stop", "restart":
		// valid
	default:
		return fmt.Errorf("invalid action type: %s; must be 'start', 'stop', or 'restart'", a.ActionType)
	}

	_, err = a.CommandProcessor.RunCommand("systemctl", a.ActionType, a.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to %s service %s: %w", a.ActionType, a.ServiceName, err)
	}

	return nil
}

// GetOutput returns the service operation performed
func (a *ManageServiceAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"service": a.ServiceName,
		"action":  a.ActionType,
	})
}
