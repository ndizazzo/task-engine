package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// ContainerState represents the state of a Docker container
type ContainerState struct {
	ID     string   `json:"id"`
	Names  []string `json:"names"`
	Image  string   `json:"image"`
	Status string   `json:"status"`
}

// NewGetContainerStateAction creates an action to get the state of specific containers by ID or name
func NewGetContainerStateAction(logger *slog.Logger, containerIdentifiers ...string) *task_engine.Action[*GetContainerStateAction] {
	id := fmt.Sprintf("get-container-state-%s-action", strings.Join(containerIdentifiers, "-"))
	return &task_engine.Action[*GetContainerStateAction]{
		ID: id,
		Wrapped: &GetContainerStateAction{
			BaseAction:       task_engine.BaseAction{Logger: logger},
			ContainerIDs:     containerIdentifiers,
			CommandProcessor: command.NewDefaultCommandRunner(),
		},
	}
}

// NewGetAllContainersStateAction creates an action to get the state of all containers
func NewGetAllContainersStateAction(logger *slog.Logger) *task_engine.Action[*GetContainerStateAction] {
	return &task_engine.Action[*GetContainerStateAction]{
		ID: "get-all-containers-state-action",
		Wrapped: &GetContainerStateAction{
			BaseAction:       task_engine.BaseAction{Logger: logger},
			ContainerIDs:     []string{}, // Empty means get all
			CommandProcessor: command.NewDefaultCommandRunner(),
		},
	}
}

// GetContainerStateAction retrieves the state of Docker containers
type GetContainerStateAction struct {
	task_engine.BaseAction
	ContainerIDs     []string
	CommandProcessor command.CommandRunner
	ContainerStates  []ContainerState
}

// SetCommandProcessor allows injecting a mock or alternative CommandProcessor for testing
func (a *GetContainerStateAction) SetCommandProcessor(processor command.CommandRunner) {
	a.CommandProcessor = processor
}

func (a *GetContainerStateAction) Execute(execCtx context.Context) error {
	a.Logger.Info("Getting container state", "containerIDs", a.ContainerIDs)

	var output string
	var err error

	if len(a.ContainerIDs) == 0 {
		// Get all containers
		output, err = a.CommandProcessor.RunCommandWithContext(execCtx, "docker", "ps", "-a", "--format", "json")
	} else {
		// Get specific containers
		args := []string{"ps", "-a", "--format", "json"}
		for _, id := range a.ContainerIDs {
			args = append(args, "--filter", fmt.Sprintf("name=%s", id))
		}
		output, err = a.CommandProcessor.RunCommandWithContext(execCtx, "docker", args...)
	}

	if err != nil {
		a.Logger.Error("Failed to get container state", "error", err, "output", output)
		return fmt.Errorf("failed to get container state: %w. Output: %s", err, output)
	}

	// Parse the output
	containerStates, err := a.parseContainerOutput(output)
	if err != nil {
		a.Logger.Error("Failed to parse container output", "error", err, "output", output)
		return fmt.Errorf("failed to parse container output: %w", err)
	}

	a.ContainerStates = containerStates
	a.Logger.Info("Successfully retrieved container states", "count", len(containerStates))
	return nil
}

// parseContainerOutput parses the JSON output from docker ps command
func (a *GetContainerStateAction) parseContainerOutput(output string) ([]ContainerState, error) {
	if strings.TrimSpace(output) == "" {
		return []ContainerState{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	containerStates := make([]ContainerState, 0, len(lines))

	// Track parsing errors
	var parsingErrors []string
	validContainers := 0
	nonEmptyLines := 0

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		nonEmptyLines++

		var containerInfo struct {
			ID     string `json:"ID"`
			Names  string `json:"Names"`
			Image  string `json:"Image"`
			Status string `json:"Status"`
		}

		if err := json.Unmarshal([]byte(line), &containerInfo); err != nil {
			errorMsg := fmt.Sprintf("line %d: invalid JSON format: %v", i+1, err)
			parsingErrors = append(parsingErrors, errorMsg)
			a.Logger.Warn("Failed to parse container line", "line", line, "error", err)
			continue
		}

		// Validate required fields
		if containerInfo.ID == "" {
			errorMsg := fmt.Sprintf("line %d: missing required field 'ID'", i+1)
			parsingErrors = append(parsingErrors, errorMsg)
			a.Logger.Warn("Container line missing ID", "line", line)
			continue
		}

		if containerInfo.Status == "" {
			errorMsg := fmt.Sprintf("line %d: missing required field 'Status'", i+1)
			parsingErrors = append(parsingErrors, errorMsg)
			a.Logger.Warn("Container line missing Status", "line", line)
			continue
		}

		// Split names by comma and clean them up
		names := strings.Split(containerInfo.Names, ",")
		cleanNames := make([]string, 0, len(names))
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name != "" {
				cleanNames = append(cleanNames, name)
			}
		}

		containerState := ContainerState{
			ID:     containerInfo.ID,
			Names:  cleanNames,
			Image:  containerInfo.Image,
			Status: containerInfo.Status,
		}

		containerStates = append(containerStates, containerState)
		validContainers++
	}

	// Return error if no valid containers were parsed and there were parsing errors
	if validContainers == 0 && len(parsingErrors) > 0 {
		return nil, fmt.Errorf("failed to parse any valid containers: %s", strings.Join(parsingErrors, "; "))
	}

	// Return error if more than 50% of non-empty lines failed to parse
	if len(parsingErrors) > 0 && len(parsingErrors) > nonEmptyLines/2 {
		return nil, fmt.Errorf("too many parsing errors (%d/%d non-empty lines): %s", len(parsingErrors), nonEmptyLines, strings.Join(parsingErrors[:5], "; "))
	}

	return containerStates, nil
}
