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

	Operation        ShutdownCommandOperation
	Delay            time.Duration
	CommandProcessor command.CommandRunner
}

type ShutdownCommandOperation string

const (
	ShutdownOperation_Shutdown ShutdownCommandOperation = "shutdown"
	ShutdownOperation_Restart  ShutdownCommandOperation = "restart"
	ShutdownOperation_Sleep    ShutdownCommandOperation = "sleep"
)

func NewShutdownAction(delay time.Duration, operation ShutdownCommandOperation, logger *slog.Logger) *task_engine.Action[*ShutdownAction] {
	return &task_engine.Action[*ShutdownAction]{
		ID: "shutdown-host-action",
		Wrapped: &ShutdownAction{
			BaseAction:       task_engine.BaseAction{Logger: logger},
			Delay:            delay,
			Operation:        operation,
			CommandProcessor: command.NewDefaultCommandRunner(),
		},
	}
}

func (a *ShutdownAction) Execute(ctx context.Context) error {
	additionalFlags := shutdownArgs(a.Operation, a.Delay)
	_, err := a.CommandProcessor.RunCommand("shutdown", additionalFlags...)
	return err
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
