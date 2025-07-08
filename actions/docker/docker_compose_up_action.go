package docker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

func NewDockerComposeUpAction(logger *slog.Logger, workingDir string, services ...string) *task_engine.Action[*DockerComposeUpAction] {
	var id string
	if len(services) == 0 {
		id = "docker-compose-up-all-action"
	} else {
		// Sort services for deterministic ID
		sortedServices := make([]string, len(services))
		copy(sortedServices, services)
		sort.Strings(sortedServices)

		// Create a canonical representation
		canonicalString := strings.Join(sortedServices, "\x00")
		hashBytes := sha256.Sum256([]byte(canonicalString))
		hexHash := hex.EncodeToString(hashBytes[:])

		id = fmt.Sprintf("docker-compose-up-%s-action", hexHash)
	}

	return &task_engine.Action[*DockerComposeUpAction]{
		ID: id,
		Wrapped: &DockerComposeUpAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			Services:      services,
			WorkingDir:    workingDir,
			commandRunner: command.NewDefaultCommandRunner(),
		},
	}
}

type DockerComposeUpAction struct {
	task_engine.BaseAction
	Services      []string
	WorkingDir    string
	commandRunner command.CommandRunner
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing.
func (a *DockerComposeUpAction) SetCommandRunner(runner command.CommandRunner) {
	a.commandRunner = runner
}

func (a *DockerComposeUpAction) Execute(execCtx context.Context) error {
	args := []string{"compose", "up", "-d"}
	args = append(args, a.Services...)

	a.Logger.Info("Executing docker compose up", "services", a.Services, "workingDir", a.WorkingDir)

	var output string
	var err error
	if a.WorkingDir != "" {
		output, err = a.commandRunner.RunCommandInDirWithContext(execCtx, a.WorkingDir, "docker", args...)
	} else {
		output, err = a.commandRunner.RunCommandWithContext(execCtx, "docker", args...)
	}

	if err != nil {
		a.Logger.Error("Failed to run docker compose up", "error", err, "output", output)
		return fmt.Errorf("failed to run docker compose up for services %v in dir %s: %w. Output: %s", a.Services, a.WorkingDir, err, output)
	}
	a.Logger.Info("Docker compose up finished successfully", "output", output)
	return nil
}
