package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// Container represents a Docker container with its metadata
type Container struct {
	ContainerID string
	Image       string
	Command     string
	Created     string
	Status      string
	Ports       string
	Names       string
}

// DockerPsOption is a function type for configuring DockerPsAction
type DockerPsOption func(*DockerPsAction)

// WithPsAll shows all containers (default shows only running)
func WithPsAll() DockerPsOption {
	return func(a *DockerPsAction) {
		a.All = true
	}
}

// WithPsFilter filters output based on conditions provided
func WithPsFilter(filter string) DockerPsOption {
	return func(a *DockerPsAction) {
		a.Filter = filter
	}
}

// WithPsFormat uses a custom template for output
func WithPsFormat(format string) DockerPsOption {
	return func(a *DockerPsAction) {
		a.Format = format
	}
}

// WithPsLast shows n last created containers (includes all states)
func WithPsLast(n int) DockerPsOption {
	return func(a *DockerPsAction) {
		a.Last = n
	}
}

// WithPsLatest shows the latest created container (includes all states)
func WithPsLatest() DockerPsOption {
	return func(a *DockerPsAction) {
		a.Latest = true
	}
}

// WithPsNoTrunc don't truncate output
func WithPsNoTrunc() DockerPsOption {
	return func(a *DockerPsAction) {
		a.NoTrunc = true
	}
}

// WithPsQuiet only display container IDs
func WithPsQuiet() DockerPsOption {
	return func(a *DockerPsAction) {
		a.Quiet = true
	}
}

// WithPsSize display total file sizes
func WithPsSize() DockerPsOption {
	return func(a *DockerPsAction) {
		a.Size = true
	}
}

// DockerPsAction lists Docker containers
type DockerPsAction struct {
	task_engine.BaseAction
	All              bool
	Filter           string
	Format           string
	Last             int
	Latest           bool
	NoTrunc          bool
	Quiet            bool
	Size             bool
	CommandProcessor command.CommandRunner
	Output           string
	Containers       []Container // Stores the parsed containers

	// Parameter-aware fields
	FilterParam  task_engine.ActionParameter
	AllParam     task_engine.ActionParameter
	QuietParam   task_engine.ActionParameter
	NoTruncParam task_engine.ActionParameter
	SizeParam    task_engine.ActionParameter
	LatestParam  task_engine.ActionParameter
	LastParam    task_engine.ActionParameter
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerPsAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerPsAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve filter parameter if it exists
	if a.FilterParam != nil {
		filterValue, err := a.FilterParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve filter parameter: %w", err)
		}
		if filterStr, ok := filterValue.(string); ok {
			// Only override if a non-empty filter is provided via parameter
			if strings.TrimSpace(filterStr) != "" {
				a.Filter = filterStr
			}
		} else {
			return fmt.Errorf("filter parameter is not a string, got %T", filterValue)
		}
	}

	// Resolve all parameter if it exists
	if a.AllParam != nil {
		allValue, err := a.AllParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve all parameter: %w", err)
		}
		if allBool, ok := allValue.(bool); ok {
			a.All = allBool
		} else {
			return fmt.Errorf("all parameter is not a bool, got %T", allValue)
		}
	}

	// Resolve quiet parameter if it exists
	if a.QuietParam != nil {
		quietValue, err := a.QuietParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve quiet parameter: %w", err)
		}
		if quietBool, ok := quietValue.(bool); ok {
			a.Quiet = quietBool
		} else {
			return fmt.Errorf("quiet parameter is not a bool, got %T", quietValue)
		}
	}

	// Resolve noTrunc parameter if it exists
	if a.NoTruncParam != nil {
		noTruncValue, err := a.NoTruncParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve noTrunc parameter: %w", err)
		}
		if noTruncBool, ok := noTruncValue.(bool); ok {
			a.NoTrunc = noTruncBool
		} else {
			return fmt.Errorf("noTrunc parameter is not a bool, got %T", noTruncValue)
		}
	}

	// Resolve size parameter if it exists
	if a.SizeParam != nil {
		sizeValue, err := a.SizeParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve size parameter: %w", err)
		}
		if sizeBool, ok := sizeValue.(bool); ok {
			a.Size = sizeBool
		} else {
			return fmt.Errorf("size parameter is not a bool, got %T", sizeValue)
		}
	}

	// Resolve latest parameter if it exists
	if a.LatestParam != nil {
		latestValue, err := a.LatestParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve latest parameter: %w", err)
		}
		if latestBool, ok := latestValue.(bool); ok {
			a.Latest = latestBool
		} else {
			return fmt.Errorf("latest parameter is not a bool, got %T", latestValue)
		}
	}

	// Resolve last parameter if it exists
	if a.LastParam != nil {
		lastValue, err := a.LastParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve last parameter: %w", err)
		}
		if lastInt, ok := lastValue.(int); ok {
			a.Last = lastInt
		} else {
			return fmt.Errorf("last parameter is not an int, got %T", lastValue)
		}
	}

	args := []string{"ps"}

	if a.All {
		args = append(args, "--all")
	}
	if a.Filter != "" {
		args = append(args, "--filter", a.Filter)
	}
	if a.Format != "" {
		args = append(args, "--format", a.Format)
	}
	if a.Last > 0 {
		args = append(args, "--last", fmt.Sprintf("%d", a.Last))
	}
	if a.Latest {
		args = append(args, "--latest")
	}
	if a.NoTrunc {
		args = append(args, "--no-trunc")
	}
	if a.Quiet {
		args = append(args, "--quiet")
	}
	if a.Size {
		args = append(args, "--size")
	}

	a.Logger.Info("Executing docker ps",
		"all", a.All,
		"filter", a.Filter,
		"format", a.Format,
		"last", a.Last,
		"latest", a.Latest,
		"noTrunc", a.NoTrunc,
		"quiet", a.Quiet,
		"size", a.Size,
	)

	output, err := a.CommandProcessor.RunCommand("docker", args...)
	if err != nil {
		a.Logger.Error("Failed to list Docker containers", "error", err.Error(), "output", output)
		return fmt.Errorf("failed to list Docker containers: %w", err)
	}

	a.Output = output
	a.parseContainers(output)

	a.Logger.Info("Docker ps finished successfully",
		"containerCount", len(a.Containers),
		"output", output,
	)

	return nil
}

// GetOutput returns parsed container information and raw output metadata
func (a *DockerPsAction) GetOutput() interface{} {
	return map[string]interface{}{
		"containers": a.Containers,
		"count":      len(a.Containers),
		"output":     a.Output,
		"success":    true,
	}
}

// parseContainers parses the docker ps output and populates the Containers slice
func (a *DockerPsAction) parseContainers(output string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		a.Containers = []Container{}
		return
	}

	// Skip header line
	if len(lines) > 0 && strings.Contains(lines[0], "CONTAINER ID") {
		lines = lines[1:]
	}

	var containers []Container
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse the line
		container := a.parseContainerLine(line)
		if container != nil {
			containers = append(containers, *container)
		}
	}

	a.Containers = containers
}

// parseContainerLine parses a single line from docker ps output
func (a *DockerPsAction) parseContainerLine(line string) *Container {
	// Format: CONTAINER ID IMAGE COMMAND CREATED STATUS PORTS NAMES
	// Example: abc123def456 nginx:latest "nginx -g 'daemon off" 2 hours ago Up 2 hours 0.0.0.0:8080->80/tcp myapp_web_1
	// Example: def456ghi789 postgres:13 "docker-entrypoint.s" 2 hours ago Exited (0) 1 hour ago 6379/tcp myapp_db_1

	parts := strings.Fields(line)
	if len(parts) < 7 {
		return nil
	}

	containerID := parts[0]
	image := parts[1]

	// Find the command (enclosed in quotes)
	commandStart := -1
	commandEnd := -1
	for i, part := range parts {
		if strings.HasPrefix(part, "\"") && commandStart == -1 {
			commandStart = i
		}
		if commandStart != -1 && strings.HasSuffix(part, "\"") {
			commandEnd = i
			break
		}
	}

	if commandStart == -1 || commandEnd == -1 {
		return nil
	}

	command := strings.Join(parts[commandStart:commandEnd+1], " ")
	// Remove the quotes
	command = strings.Trim(command, "\"")

	// The remaining parts after command should be: CREATED STATUS PORTS NAMES
	remainingParts := parts[commandEnd+1:]
	if len(remainingParts) < 4 {
		return nil
	}

	// Parse the remaining fields more carefully
	// The pattern is: CREATED STATUS PORTS NAMES
	// Where CREATED can be "2 hours ago", STATUS can be "Up 2 hours" or "Exited (0) 1 hour ago", etc.

	// Find where CREATED ends (it contains "ago")
	createdEnd := -1
	for i, part := range remainingParts {
		if strings.Contains(part, "ago") {
			createdEnd = i
			break
		}
	}

	if createdEnd == -1 {
		return nil
	}

	created := strings.Join(remainingParts[:createdEnd+1], " ")

	// Find where STATUS ends
	statusStart := createdEnd + 1
	if statusStart >= len(remainingParts) {
		return nil
	}

	// STATUS parsing - handle different patterns
	statusEnd := statusStart
	status := ""

	if statusStart < len(remainingParts) {
		statusPart := remainingParts[statusStart]
		switch {
		case strings.HasPrefix(statusPart, "Up"):
			// "Up X hours" pattern - look for the next field that contains "/" (ports) or doesn't contain time units
			for i := statusStart + 1; i < len(remainingParts); i++ {
				part := remainingParts[i]
				// If this part contains "/" it's likely a port mapping
				if strings.Contains(part, "/") {
					statusEnd = i - 1
					break
				}
				// If this part doesn't contain time units and isn't a number, it might be the start of names
				if !strings.Contains(part, "hour") && !strings.Contains(part, "minute") && !strings.Contains(part, "second") &&
					!strings.Contains(part, "ago") && !isNumeric(part) {
					statusEnd = i - 1
					break
				}
			}
			// If we didn't find a clear boundary, assume status is 2 words (Up + duration)
			if statusEnd == statusStart && statusStart+1 < len(remainingParts) {
				statusEnd = statusStart + 1
			}
		case strings.HasPrefix(statusPart, "Exited"):
			// "Exited (X) Y ago" pattern - look for the "ago" to find the end
			for i := statusStart; i < len(remainingParts); i++ {
				if strings.Contains(remainingParts[i], "ago") {
					statusEnd = i
					break
				}
			}
		case strings.HasPrefix(statusPart, "Restarting"):
			// "Restarting (X) Y ago" pattern - look for the "ago" to find the end
			for i := statusStart; i < len(remainingParts); i++ {
				if strings.Contains(remainingParts[i], "ago") {
					statusEnd = i
					break
				}
			}
		case strings.HasPrefix(statusPart, "Created"), strings.HasPrefix(statusPart, "Paused"),
			strings.HasPrefix(statusPart, "Dead"), strings.HasPrefix(statusPart, "Removing"):
			// Single word statuses
			statusEnd = statusStart
		default:
			// Default case for unknown status patterns
			statusEnd = statusStart
		}
	}

	status = strings.Join(remainingParts[statusStart:statusEnd+1], " ")

	// The next field is PORTS
	portsStart := statusEnd + 1
	if portsStart >= len(remainingParts) {
		return &Container{
			ContainerID: containerID,
			Image:       image,
			Command:     command,
			Created:     created,
			Status:      status,
			Ports:       "",
			Names:       "",
		}
	}
	ports := ""
	namesStart := portsStart
	if portsStart < len(remainingParts) {
		potentialPorts := remainingParts[portsStart]
		if strings.Contains(potentialPorts, "/") || strings.Contains(potentialPorts, "->") {
			// This is a port mapping - collect all consecutive port-related fields
			portParts := []string{potentialPorts}
			namesStart = portsStart + 1

			// Look for more port mappings (they might be comma-separated or in separate fields)
			for i := portsStart + 1; i < len(remainingParts); i++ {
				part := remainingParts[i]
				// If this part contains port indicators, it's part of the ports
				if strings.Contains(part, "/") || strings.Contains(part, "->") ||
					strings.Contains(part, ",") || strings.HasPrefix(part, "0.0.0.0:") {
					portParts = append(portParts, part)
					namesStart = i + 1
				} else {
					// This is likely the start of names
					break
				}
			}
			ports = strings.Join(portParts, " ")
		} else {
			// This is likely the start of names (no ports)
			ports = ""
			namesStart = portsStart
		}
	}

	// Everything after PORTS is NAMES
	names := ""
	if namesStart < len(remainingParts) {
		names = strings.Join(remainingParts[namesStart:], " ")
	}

	return &Container{
		ContainerID: containerID,
		Image:       image,
		Command:     command,
		Created:     created,
		Status:      status,
		Ports:       ports,
		Names:       names,
	}
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// DockerPsActionConstructor provides the new constructor pattern
type DockerPsActionConstructor struct {
	logger *slog.Logger
}

// NewDockerPsAction creates a new DockerPsAction constructor
func NewDockerPsAction(logger *slog.Logger) *DockerPsActionConstructor {
	return &DockerPsActionConstructor{
		logger: logger,
	}
}

// WithParameters creates a DockerPsAction with the specified parameters
func (c *DockerPsActionConstructor) WithParameters(
	filterParam task_engine.ActionParameter,
	allParam task_engine.ActionParameter,
	quietParam task_engine.ActionParameter,
	noTruncParam task_engine.ActionParameter,
	sizeParam task_engine.ActionParameter,
	latestParam task_engine.ActionParameter,
	lastParam task_engine.ActionParameter,
) (*task_engine.Action[*DockerPsAction], error) {
	action := &DockerPsAction{
		BaseAction:       task_engine.NewBaseAction(c.logger),
		All:              false,
		Filter:           "",
		Format:           "",
		Last:             0,
		Latest:           false,
		NoTrunc:          false,
		Quiet:            false,
		Size:             false,
		CommandProcessor: command.NewDefaultCommandRunner(),
		FilterParam:      filterParam,
		AllParam:         allParam,
		QuietParam:       quietParam,
		NoTruncParam:     noTruncParam,
		SizeParam:        sizeParam,
		LatestParam:      latestParam,
		LastParam:        lastParam,
	}

	id := "docker-ps-action"
	return &task_engine.Action[*DockerPsAction]{
		ID:      id,
		Name:    "Docker PS",
		Wrapped: action,
	}, nil
}
