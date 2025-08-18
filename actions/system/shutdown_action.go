package system

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

type ShutdownAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
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

// NewShutdownAction creates a new ShutdownAction with the given logger
func NewShutdownAction(logger *slog.Logger) *ShutdownAction {
	return &ShutdownAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		CommandProcessor:  command.NewDefaultCommandRunner(),
	}
}

// WithParameters sets the parameters for shutdown and returns a wrapped Action
func (a *ShutdownAction) WithParameters(
	operationParam task_engine.ActionParameter,
	delayParam task_engine.ActionParameter,
) (*task_engine.Action[*ShutdownAction], error) {
	a.OperationParam = operationParam
	a.DelayParam = delayParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*ShutdownAction](a.Logger)
	return constructor.WrapAction(a, "Shutdown", "shutdown-action"), nil
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *ShutdownAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *ShutdownAction) Execute(ctx context.Context) error {
	// Resolve operation parameter using the ParameterResolver
	var operation ShutdownCommandOperation
	if a.OperationParam != nil {
		operationValue, err := a.ResolveStringParameter(ctx, a.OperationParam, "operation")
		if err != nil {
			return err
		}
		operation = ShutdownCommandOperation(operationValue)
	}

	// Resolve delay parameter using the ParameterResolver
	var delay time.Duration
	if a.DelayParam != nil {
		delayValue, err := a.ResolveDurationParameter(ctx, a.DelayParam, "delay")
		if err != nil {
			return err
		}
		delay = delayValue
	}

	additionalFlags := shutdownArgs(operation, delay)
	_, err := a.CommandProcessor.RunCommand("shutdown", additionalFlags...)
	return err
}

// GetOutput returns the requested shutdown operation and delay
func (a *ShutdownAction) GetOutput() interface{} {
	return a.BuildSimpleOutput(true, "")
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
