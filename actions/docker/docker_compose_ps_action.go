package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// ComposeService represents a Docker Compose service with its metadata
type ComposeService struct {
	Name        string
	Image       string
	ServiceName string
	Status      string
	Ports       string
}

// DockerComposePsActionWrapper provides a consistent interface for DockerComposePsAction
type DockerComposePsActionWrapper struct {
	ID      string
	Wrapped *DockerComposePsAction
}

// DockerComposePsActionConstructor provides the modern constructor pattern
type DockerComposePsActionConstructor struct {
	logger *slog.Logger
}

// NewDockerComposePsAction creates a new DockerComposePsAction constructor
func NewDockerComposePsAction(logger *slog.Logger) *DockerComposePsActionConstructor {
	return &DockerComposePsActionConstructor{
		logger: logger,
	}
}

// WithParameters creates a DockerComposePsAction with the given parameters
func (c *DockerComposePsActionConstructor) WithParameters(
	servicesParam task_engine.ActionParameter,
	allParam task_engine.ActionParameter,
	filterParam task_engine.ActionParameter,
	formatParam task_engine.ActionParameter,
	quietParam task_engine.ActionParameter,
	workingDirParam task_engine.ActionParameter,
) (*DockerComposePsActionWrapper, error) {
	action := &DockerComposePsAction{
		BaseAction:       task_engine.BaseAction{Logger: c.logger},
		Services:         []string{},
		All:              false,
		Filter:           "",
		Format:           "",
		Quiet:            false,
		WorkingDir:       "",
		CommandProcessor: command.NewDefaultCommandRunner(),
		ServicesParam:    servicesParam,
		AllParam:         allParam,
		FilterParam:      filterParam,
		FormatParam:      formatParam,
		QuietParam:       quietParam,
		WorkingDirParam:  workingDirParam,
	}

	return &DockerComposePsActionWrapper{
		ID:      "docker-compose-ps-action",
		Wrapped: action,
	}, nil
}

// DockerComposePsOption is a function type for configuring DockerComposePsAction
type DockerComposePsOption func(*DockerComposePsAction)

// WithComposePsAll shows all containers (default shows only running)
func WithComposePsAll() DockerComposePsOption {
	return func(a *DockerComposePsAction) {
		a.All = true
	}
}

// WithComposePsFilter filters output based on conditions provided
func WithComposePsFilter(filter string) DockerComposePsOption {
	return func(a *DockerComposePsAction) {
		a.Filter = filter
	}
}

// WithComposePsFormat uses a custom template for output
func WithComposePsFormat(format string) DockerComposePsOption {
	return func(a *DockerComposePsAction) {
		a.Format = format
	}
}

// WithComposePsQuiet only show container IDs
func WithComposePsQuiet() DockerComposePsOption {
	return func(a *DockerComposePsAction) {
		a.Quiet = true
	}
}

// WithComposePsWorkingDir sets the working directory for the compose command
func WithComposePsWorkingDir(workingDir string) DockerComposePsOption {
	return func(a *DockerComposePsAction) {
		a.WorkingDir = workingDir
	}
}

// DockerComposePsAction lists Docker Compose services
type DockerComposePsAction struct {
	task_engine.BaseAction
	Services         []string
	All              bool
	Filter           string
	Format           string
	Quiet            bool
	WorkingDir       string
	CommandProcessor command.CommandRunner
	Output           string
	ServicesList     []ComposeService // Stores the parsed services

	// Parameter-aware fields
	ServicesParam   task_engine.ActionParameter
	AllParam        task_engine.ActionParameter
	FilterParam     task_engine.ActionParameter
	FormatParam     task_engine.ActionParameter
	QuietParam      task_engine.ActionParameter
	WorkingDirParam task_engine.ActionParameter
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerComposePsAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerComposePsAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve services parameter if it exists
	if a.ServicesParam != nil {
		servicesValue, err := a.ServicesParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve services parameter: %w", err)
		}
		if servicesSlice, ok := servicesValue.([]string); ok {
			a.Services = servicesSlice
		} else if servicesStr, ok := servicesValue.(string); ok {
			// If it's a single string, split by comma or space
			if strings.Contains(servicesStr, ",") {
				a.Services = strings.Split(servicesStr, ",")
			} else {
				a.Services = strings.Fields(servicesStr)
			}
		} else {
			return fmt.Errorf("services parameter is not a string slice or string, got %T", servicesValue)
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

	// Resolve filter parameter if it exists
	if a.FilterParam != nil {
		filterValue, err := a.FilterParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve filter parameter: %w", err)
		}
		if filterStr, ok := filterValue.(string); ok {
			a.Filter = filterStr
		} else {
			return fmt.Errorf("filter parameter is not a string, got %T", filterValue)
		}
	}

	// Resolve format parameter if it exists
	if a.FormatParam != nil {
		formatValue, err := a.FormatParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve format parameter: %w", err)
		}
		if formatStr, ok := formatValue.(string); ok {
			a.Format = formatStr
		} else {
			return fmt.Errorf("format parameter is not a string, got %T", formatValue)
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

	// Resolve working directory parameter if it exists
	if a.WorkingDirParam != nil {
		workingDirValue, err := a.WorkingDirParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve working directory parameter: %w", err)
		}
		if workingDirStr, ok := workingDirValue.(string); ok {
			a.WorkingDir = workingDirStr
		} else {
			return fmt.Errorf("working directory parameter is not a string, got %T", workingDirValue)
		}
	}

	args := []string{"compose", "ps"}

	if a.All {
		args = append(args, "--all")
	}
	if a.Filter != "" {
		args = append(args, "--filter", a.Filter)
	}
	if a.Format != "" {
		args = append(args, "--format", a.Format)
	}
	if a.Quiet {
		args = append(args, "--quiet")
	}

	// Add service names if specified
	if len(a.Services) > 0 {
		args = append(args, a.Services...)
	}

	a.Logger.Info("Executing docker compose ps",
		"services", a.Services,
		"all", a.All,
		"filter", a.Filter,
		"format", a.Format,
		"quiet", a.Quiet,
		"workingDir", a.WorkingDir,
	)

	output, err := a.CommandProcessor.RunCommand("docker", args...)
	if err != nil {
		a.Logger.Error("Failed to list Docker Compose services", "error", err.Error(), "output", output)
		return fmt.Errorf("failed to list Docker Compose services: %w", err)
	}

	a.Output = output
	a.parseServices(output)

	a.Logger.Info("Docker compose ps finished successfully",
		"serviceCount", len(a.ServicesList),
		"output", output,
	)

	return nil
}

// GetOutput returns parsed services information and raw output metadata
func (a *DockerComposePsAction) GetOutput() interface{} {
	return map[string]interface{}{
		"services": a.ServicesList,
		"count":    len(a.ServicesList),
		"output":   a.Output,
		"success":  true,
	}
}

// parseServices parses the docker compose ps output and populates the ServicesList slice
func (a *DockerComposePsAction) parseServices(output string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		a.ServicesList = []ComposeService{}
		return
	}

	// Skip header line
	if len(lines) > 0 && strings.Contains(lines[0], "NAME") {
		lines = lines[1:]
	}

	var services []ComposeService
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse the line
		service := a.parseServiceLine(line)
		if service != nil {
			services = append(services, *service)
		}
	}

	a.ServicesList = services
}

// parseServiceLine parses a single line from docker compose ps output
func (a *DockerComposePsAction) parseServiceLine(line string) *ComposeService {
	// Format: NAME IMAGE COMMAND SERVICE CREATED STATUS PORTS
	// Example: myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
	// Example: myapp_db_1          postgres:13         "docker-entrypoint.s"    db                  2 hours ago         Up 2 hours         5432/tcp
	// Quiet format: just the service name
	// Example: myapp_web_1

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	// If quiet mode, we only have the service name
	if a.Quiet {
		return &ComposeService{
			Name:        parts[0],
			Image:       "",
			ServiceName: "",
			Status:      "",
			Ports:       "",
		}
	}

	// Regular mode requires at least 6 fields
	if len(parts) < 6 {
		return nil
	}

	name := parts[0]
	image := parts[1]

	// Find the service name by looking for the pattern after the command
	// The command is typically quoted, so we need to find where it ends
	serviceName := ""
	status := ""
	ports := ""

	// Find the end of the command (look for the closing quote)
	commandEndIndex := -1
	for i, part := range parts {
		if strings.HasSuffix(part, "\"") && commandEndIndex == -1 {
			commandEndIndex = i
			break
		}
	}

	if commandEndIndex == -1 {
		// If no quoted command found, assume simple format
		if len(parts) >= 4 {
			serviceName = parts[3]
		}
		if len(parts) >= 6 {
			status = parts[5]
		}
		if len(parts) > 6 {
			ports = strings.Join(parts[6:], " ")
		}
	} else {
		// Command ends at commandEndIndex, so service name should be at commandEndIndex + 1
		if commandEndIndex+1 < len(parts) {
			serviceName = parts[commandEndIndex+1]
		}

		// Find status and ports after the service name
		// The format after service name is: CREATED STATUS PORTS
		// CREATED is typically "2 hours ago" or "1 hour ago"
		// STATUS is typically "Up 2 hours" or "Exited (0) 1 hour ago"
		// PORTS is the rest

		// Find the status by looking for patterns like "Up", "Exited", "Created", etc.
		statusStartIndex := -1
		for i := commandEndIndex + 2; i < len(parts); i++ {
			if parts[i] == "Up" || parts[i] == "Exited" || parts[i] == "Created" || parts[i] == "Restarting" {
				statusStartIndex = i
				break
			}
		}

		if statusStartIndex != -1 {
			// Find where status ends (look for port patterns or end of line)
			statusEndIndex := len(parts)
			for i := statusStartIndex + 1; i < len(parts); i++ {
				if strings.Contains(parts[i], ":") || strings.Contains(parts[i], "->") || strings.Contains(parts[i], "/tcp") || strings.Contains(parts[i], "/udp") {
					statusEndIndex = i
					break
				}
			}

			status = strings.Join(parts[statusStartIndex:statusEndIndex], " ")

			// Everything after status is ports
			if statusEndIndex < len(parts) {
				ports = strings.Join(parts[statusEndIndex:], " ")
			}
		}
	}

	return &ComposeService{
		Name:        name,
		Image:       image,
		ServiceName: serviceName,
		Status:      status,
		Ports:       ports,
	}
}
