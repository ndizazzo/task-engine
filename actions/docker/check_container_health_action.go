package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

type CheckContainerHealthAction struct {
	task_engine.BaseAction
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

// NewCheckContainerHealthAction creates the action instance (single builder pattern)
func NewCheckContainerHealthAction(logger *slog.Logger) *CheckContainerHealthAction {
	return &CheckContainerHealthAction{
		BaseAction:    task_engine.BaseAction{Logger: logger},
		commandRunner: command.NewDefaultCommandRunner(),
	}
}

// WithParameters validates and attaches parameters, returning the wrapped action
func (a *CheckContainerHealthAction) WithParameters(
	workingDirParam task_engine.ActionParameter,
	serviceNameParam task_engine.ActionParameter,
	checkCommandParam task_engine.ActionParameter,
	maxRetriesParam task_engine.ActionParameter,
	retryDelayParam task_engine.ActionParameter,
) (*task_engine.Action[*CheckContainerHealthAction], error) {
	if workingDirParam == nil || serviceNameParam == nil || checkCommandParam == nil || maxRetriesParam == nil || retryDelayParam == nil {
		return nil, fmt.Errorf("parameters cannot be nil")
	}
	a.WorkingDirParam = workingDirParam
	a.ServiceNameParam = serviceNameParam
	a.CheckCommandParam = checkCommandParam
	a.MaxRetriesParam = maxRetriesParam
	a.RetryDelayParam = retryDelayParam

	return &task_engine.Action[*CheckContainerHealthAction]{
		ID:      "check-container-health-action",
		Name:    "Check Container Health",
		Wrapped: a,
	}, nil
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *CheckContainerHealthAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *CheckContainerHealthAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve working directory parameter
	workingDirValue, err := a.WorkingDirParam.Resolve(execCtx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve working directory parameter: %w", err)
	}
	if workingDirStr, ok := workingDirValue.(string); ok {
		a.ResolvedWorkingDir = workingDirStr
	} else {
		return fmt.Errorf("working directory parameter is not a string, got %T", workingDirValue)
	}

	// Resolve service name parameter
	serviceNameValue, err := a.ServiceNameParam.Resolve(execCtx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve service name parameter: %w", err)
	}
	if serviceNameStr, ok := serviceNameValue.(string); ok {
		a.ResolvedServiceName = serviceNameStr
	} else {
		return fmt.Errorf("service name parameter is not a string, got %T", serviceNameValue)
	}

	// Resolve check command parameter
	checkCommandValue, err := a.CheckCommandParam.Resolve(execCtx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve check command parameter: %w", err)
	}
	if checkCommandSlice, ok := checkCommandValue.([]string); ok {
		a.ResolvedCheckCommand = checkCommandSlice
	} else if checkCommandStr, ok := checkCommandValue.(string); ok {
		a.ResolvedCheckCommand = strings.Fields(checkCommandStr)
	} else {
		return fmt.Errorf("check command parameter is not a string slice or string, got %T", checkCommandValue)
	}

	// Resolve max retries parameter
	maxRetriesValue, err := a.MaxRetriesParam.Resolve(execCtx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve max retries parameter: %w", err)
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
	retryDelayValue, err := a.RetryDelayParam.Resolve(execCtx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve retry delay parameter: %w", err)
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
	return map[string]interface{}{
		"service":    a.ResolvedServiceName,
		"command":    a.ResolvedCheckCommand,
		"maxRetries": a.ResolvedMaxRetries,
		"retryDelay": a.ResolvedRetryDelay.String(),
		"workingDir": a.ResolvedWorkingDir,
		"success":    true,
	}
}
