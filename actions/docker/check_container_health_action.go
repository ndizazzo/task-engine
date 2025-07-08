package docker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewCheckContainerHealthAction creates an action that repeatedly runs a command inside a container
// via docker compose exec until it succeeds or retries are exhausted.
func NewCheckContainerHealthAction(workingDir string, serviceName string, checkCommand []string, maxRetries int, retryDelay time.Duration, logger *slog.Logger) *task_engine.Action[*CheckContainerHealthAction] {
	id := fmt.Sprintf("check-health-%s-%s", serviceName, checkCommand[0])
	return &task_engine.Action[*CheckContainerHealthAction]{
		ID: id,
		Wrapped: &CheckContainerHealthAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			WorkingDir:    workingDir,
			ServiceName:   serviceName,
			CheckCommand:  checkCommand,
			MaxRetries:    maxRetries,
			RetryDelay:    retryDelay,
			commandRunner: command.NewDefaultCommandRunner(),
		},
	}
}

type CheckContainerHealthAction struct {
	task_engine.BaseAction
	WorkingDir    string
	ServiceName   string
	CheckCommand  []string
	MaxRetries    int
	RetryDelay    time.Duration
	commandRunner command.CommandRunner
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *CheckContainerHealthAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *CheckContainerHealthAction) Execute(execCtx context.Context) error {
	cmdArgs := append([]string{"compose", "exec", a.ServiceName}, a.CheckCommand...)

	for i := 0; i < a.MaxRetries; i++ {
		a.Logger.Info("Checking container health", "service", a.ServiceName, "attempt", i+1, "workingDir", a.WorkingDir)

		var output string
		var err error
		if a.WorkingDir != "" {
			output, err = a.commandRunner.RunCommandInDirWithContext(execCtx, a.WorkingDir, "docker", cmdArgs...)
		} else {
			output, err = a.commandRunner.RunCommandWithContext(execCtx, "docker", cmdArgs...)
		}

		if err == nil {
			a.Logger.Info("Container health check passed", "service", a.ServiceName, "output", output)
			return nil
		}

		a.Logger.Warn("Container health check failed", "service", a.ServiceName, "error", err, "output", output, "attempt", i+1)
		select {
		case <-execCtx.Done():
			a.Logger.Info("Context cancelled, stopping health check retries", "service", a.ServiceName)
			return execCtx.Err()
		case <-time.After(a.RetryDelay):
			// Continue to next retry
		}
	}

	return fmt.Errorf("container %s failed health check after %d retries", a.ServiceName, a.MaxRetries)
}
