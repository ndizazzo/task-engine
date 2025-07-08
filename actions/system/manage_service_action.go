package system

import (
	"context"
	"fmt"
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

func NewManageServiceAction(serviceName, actionType string, logger *slog.Logger) *task_engine.Action[*ManageServiceAction] {
	return &task_engine.Action[*ManageServiceAction]{
		ID: fmt.Sprintf("%s-%s-action", actionType, serviceName),
		Wrapped: &ManageServiceAction{
			BaseAction:       task_engine.BaseAction{Logger: logger},
			ServiceName:      serviceName,
			ActionType:       actionType,
			CommandProcessor: command.NewDefaultCommandRunner(),
		},
	}
}

type ManageServiceAction struct {
	task_engine.BaseAction

	ServiceName      string
	ActionType       string
	CommandProcessor command.CommandRunner
}

func (a *ManageServiceAction) Execute(execCtx context.Context) error {
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
