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

// NewDockerPsAction creates an action to list Docker containers
func NewDockerPsAction(logger *slog.Logger, options ...DockerPsOption) *task_engine.Action[*DockerPsAction] {
	action := &DockerPsAction{
		BaseAction:       task_engine.BaseAction{Logger: logger},
		All:              false,
		Filter:           "",
		Format:           "",
		Last:             0,
		Latest:           false,
		NoTrunc:          false,
		Quiet:            false,
		Size:             false,
		CommandProcessor: command.NewDefaultCommandRunner(),
	}

	// Apply options
	for _, option := range options {
		option(action)
	}

	return &task_engine.Action[*DockerPsAction]{
		ID:      "docker-ps-action",
		Wrapped: action,
	}
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
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerPsAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerPsAction) Execute(execCtx context.Context) error {
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

	// Check if the next field looks like a port mapping (contains "/" or "->")
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
