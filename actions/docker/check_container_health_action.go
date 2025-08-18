package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

type CheckContainerHealthAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	// Parameter-only inputs
	WorkingDirParam   task_engine.ActionParameter
	ServiceNameParam  task_engine.ActionParameter
	CheckCommandParam task_engine.ActionParameter
	MaxRetriesParam   task_engine.ActionParameter
	RetryDelayParam   task_engine.ActionParameter

	// Execution dependency
	commandRunner command.CommandRunner

	// Resolved/output fields
	ResolvedWorkingDir   string
	ResolvedServiceName  string
	ResolvedCheckCommand []string
	ResolvedMaxRetries   int
	ResolvedRetryDelay   time.Duration
}

// NewCheckContainerHealthAction creates a new CheckContainerHealthAction with the given logger
func NewCheckContainerHealthAction(logger *slog.Logger) *CheckContainerHealthAction {
	return &CheckContainerHealthAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		commandRunner:     command.NewDefaultCommandRunner(),
	}
}

// WithParameters sets the parameters for container health check and returns a wrapped Action
func (a *CheckContainerHealthAction) WithParameters(
	workingDirParam task_engine.ActionParameter,
	serviceNameParam task_engine.ActionParameter,
	checkCommandParam task_engine.ActionParameter,
	maxRetriesParam task_engine.ActionParameter,
	retryDelayParam task_engine.ActionParameter,
) (*task_engine.Action[*CheckContainerHealthAction], error) {
	a.WorkingDirParam = workingDirParam
	a.ServiceNameParam = serviceNameParam
	a.CheckCommandParam = checkCommandParam
	a.MaxRetriesParam = maxRetriesParam
	a.RetryDelayParam = retryDelayParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*CheckContainerHealthAction](a.Logger)
	return constructor.WrapAction(a, "Check Container Health", "check-container-health-action"), nil
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *CheckContainerHealthAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *CheckContainerHealthAction) Execute(execCtx context.Context) error {
	// Resolve working directory parameter
	workingDirValue, err := a.ResolveStringParameter(execCtx, a.WorkingDirParam, "working directory")
	if err != nil {
		return err
	}
	a.ResolvedWorkingDir = workingDirValue

	// Resolve service name parameter
	serviceNameValue, err := a.ResolveStringParameter(execCtx, a.ServiceNameParam, "service name")
	if err != nil {
		return err
	}
	a.ResolvedServiceName = serviceNameValue

	// Resolve check command parameter
	checkCommandValue, err := a.ResolveParameter(execCtx, a.CheckCommandParam, "check command")
	if err != nil {
		return err
	}
	if checkCommandSlice, ok := checkCommandValue.([]string); ok {
		a.ResolvedCheckCommand = checkCommandSlice
	} else if checkCommandStr, ok := checkCommandValue.(string); ok {
		a.ResolvedCheckCommand = strings.Fields(checkCommandStr)
	} else {
		return fmt.Errorf("check command parameter is not a string slice or string, got %T", checkCommandValue)
	}

	// Resolve max retries parameter
	maxRetriesValue, err := a.ResolveParameter(execCtx, a.MaxRetriesParam, "max retries")
	if err != nil {
		return err
	}
	switch v := maxRetriesValue.(type) {
	case int:
		a.ResolvedMaxRetries = v
	case int64:
		a.ResolvedMaxRetries = int(v)
	case string:
		// Accept numeric strings
		// Trim spaces and try to parse
		s := strings.TrimSpace(v)
		// Simple manual parse to avoid extra imports
		n := 0
		for i := 0; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return fmt.Errorf("max retries parameter is not an int, got %T", maxRetriesValue)
			}
			n = n*10 + int(s[i]-'0')
		}
		a.ResolvedMaxRetries = n
	default:
		return fmt.Errorf("max retries parameter is not an int, got %T", maxRetriesValue)
	}

	// Resolve retry delay parameter
	retryDelayValue, err := a.ResolveParameter(execCtx, a.RetryDelayParam, "retry delay")
	if err != nil {
		return err
	}
	switch v := retryDelayValue.(type) {
	case time.Duration:
		a.ResolvedRetryDelay = v
	case string:
		d, perr := time.ParseDuration(strings.TrimSpace(v))
		if perr != nil {
			return fmt.Errorf("retry delay parameter is not a duration or duration string, got %T", retryDelayValue)
		}
		a.ResolvedRetryDelay = d
	default:
		return fmt.Errorf("retry delay parameter is not a duration or duration string, got %T", retryDelayValue)
	}

	cmdArgs := append([]string{"compose", "exec", a.ResolvedServiceName}, a.ResolvedCheckCommand...)

	for i := 0; i < a.ResolvedMaxRetries; i++ {
		a.Logger.Info("Checking container health", "service", a.ResolvedServiceName, "attempt", i+1, "workingDir", a.ResolvedWorkingDir)

		var output string
		var err error
		if a.ResolvedWorkingDir != "" {
			output, err = a.commandRunner.RunCommandInDirWithContext(execCtx, a.ResolvedWorkingDir, "docker", cmdArgs...)
		} else {
			output, err = a.commandRunner.RunCommandWithContext(execCtx, "docker", cmdArgs...)
		}

		if err == nil {
			a.Logger.Info("Container health check passed", "service", a.ResolvedServiceName, "output", output)
			return nil
		}

		a.Logger.Warn("Container health check failed", "service", a.ResolvedServiceName, "error", err, "output", output, "attempt", i+1)
		select {
		case <-execCtx.Done():
			a.Logger.Info("Context cancelled, stopping health check retries", "service", a.ResolvedServiceName)
			return execCtx.Err()
		case <-time.After(a.ResolvedRetryDelay):
			// Continue to next retry
		}
	}

	return fmt.Errorf("container %s failed health check after %d retries", a.ResolvedServiceName, a.ResolvedMaxRetries)
}

// GetOutput returns details about the health check configuration
func (a *CheckContainerHealthAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"service":    a.ResolvedServiceName,
		"command":    a.ResolvedCheckCommand,
		"maxRetries": a.ResolvedMaxRetries,
		"retryDelay": a.ResolvedRetryDelay.String(),
		"workingDir": a.ResolvedWorkingDir,
	})
}
