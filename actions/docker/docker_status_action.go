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

// GetContainerStateActionBuilder provides a fluent interface for building GetContainerStateAction
type GetContainerStateActionBuilder struct {
	logger             *slog.Logger
	containerNameParam task_engine.ActionParameter
}

// NewGetContainerStateAction creates a fluent builder for GetContainerStateAction
func NewGetContainerStateAction(logger *slog.Logger) *GetContainerStateActionBuilder {
	return &GetContainerStateActionBuilder{
		logger: logger,
	}
}

// WithParameters sets the parameters for container name
func (b *GetContainerStateActionBuilder) WithParameters(containerNameParam task_engine.ActionParameter) (*task_engine.Action[*GetContainerStateAction], error) {
	b.containerNameParam = containerNameParam

	id := "get-container-state-action"
	return &task_engine.Action[*GetContainerStateAction]{
		ID:   id,
		Name: "Get Container State",
		Wrapped: &GetContainerStateAction{
			BaseAction:         task_engine.NewBaseAction(b.logger),
			ContainerName:      "",
			CommandProcessor:   command.NewDefaultCommandRunner(),
			ContainerNameParam: b.containerNameParam,
		},
	}, nil
}

// GetContainerStateAction retrieves the state of Docker containers
type GetContainerStateAction struct {
	task_engine.BaseAction
	ContainerIDs     []string
	CommandProcessor command.CommandRunner
	ContainerStates  []ContainerState

	// Parameter-aware fields
	ContainerName      string
	ContainerNameParam task_engine.ActionParameter
}

// SetCommandProcessor allows injecting a mock or alternative CommandProcessor for testing
func (a *GetContainerStateAction) SetCommandProcessor(processor command.CommandRunner) {
	a.CommandProcessor = processor
}

func (a *GetContainerStateAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve container name parameter if it exists
	if a.ContainerNameParam != nil {
		containerNameValue, err := a.ContainerNameParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve container name parameter: %w", err)
		}
		switch v := containerNameValue.(type) {
		case string:
			a.ContainerName = v
			if strings.TrimSpace(v) != "" {
				a.ContainerIDs = []string{v}
			}
		case []string:
			filtered := make([]string, 0, len(v))
			for _, name := range v {
				if strings.TrimSpace(name) != "" {
					filtered = append(filtered, name)
				}
			}
			if len(filtered) > 0 {
				a.ContainerIDs = filtered
			}
		default:
			return fmt.Errorf("container name parameter is not a string, got %T", containerNameValue)
		}
	}

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

// GetOutput returns the retrieved container states
func (a *GetContainerStateAction) GetOutput() interface{} {
	return map[string]interface{}{
		"containers": a.ContainerStates,
		"count":      len(a.ContainerStates),
		"success":    true,
	}
}
