package system

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// ServiceStatus represents the status of a systemd service
type ServiceStatus struct {
	Name        string `json:"name"`
	Loaded      string `json:"loaded"`
	Active      string `json:"active"`
	Sub         string `json:"sub"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Vendor      string `json:"vendor"`
	Exists      bool   `json:"exists"`
}

// NewGetServiceStatusAction creates an action to get the status of specific services
func NewGetServiceStatusAction(logger *slog.Logger, serviceNames ...string) *task_engine.Action[*GetServiceStatusAction] {
	id := fmt.Sprintf("get-service-status-%s-action", strings.Join(serviceNames, "-"))
	return &task_engine.Action[*GetServiceStatusAction]{
		ID: id,
		Wrapped: &GetServiceStatusAction{
			BaseAction:       task_engine.BaseAction{Logger: logger},
			ServiceNames:     serviceNames,
			CommandProcessor: command.NewDefaultCommandRunner(),
		},
	}
}

// NewGetAllServicesStatusAction creates an action to get the status of all services
func NewGetAllServicesStatusAction(logger *slog.Logger) *task_engine.Action[*GetServiceStatusAction] {
	return &task_engine.Action[*GetServiceStatusAction]{
		ID: "get-all-services-status-action",
		Wrapped: &GetServiceStatusAction{
			BaseAction:       task_engine.BaseAction{Logger: logger},
			ServiceNames:     []string{}, // Empty means get all
			CommandProcessor: command.NewDefaultCommandRunner(),
		},
	}
}

// GetServiceStatusAction retrieves the status of systemd services
type GetServiceStatusAction struct {
	task_engine.BaseAction
	ServiceNames     []string
	CommandProcessor command.CommandRunner
	ServiceStatuses  []ServiceStatus
}

// SetCommandProcessor allows injecting a mock or alternative CommandProcessor for testing
func (a *GetServiceStatusAction) SetCommandProcessor(processor command.CommandRunner) {
	a.CommandProcessor = processor
}

func (a *GetServiceStatusAction) Execute(execCtx context.Context) error {
	a.Logger.Info("Getting service status", "serviceNames", a.ServiceNames)

	var serviceStatuses []ServiceStatus

	if len(a.ServiceNames) == 0 {
		// Get all services - this would be very slow and not practical
		// For now, return an error suggesting to use specific service names
		return fmt.Errorf("getting all services status is not supported; please specify service names")
	}

	// Get status for each service individually to handle mixed output properly
	for _, serviceName := range a.ServiceNames {
		status, err := a.getServiceStatus(execCtx, serviceName)
		if err != nil {
			a.Logger.Warn("Failed to get service status", "service", serviceName, "error", err)
			// Continue with other services even if one fails
			status = ServiceStatus{
				Name:   serviceName,
				Exists: false,
			}
		}
		serviceStatuses = append(serviceStatuses, status)
	}

	a.ServiceStatuses = serviceStatuses
	a.Logger.Info("Successfully retrieved service statuses", "count", len(serviceStatuses))
	return nil
}

// getServiceStatus gets the status of a single service using systemctl show
func (a *GetServiceStatusAction) getServiceStatus(execCtx context.Context, serviceName string) (ServiceStatus, error) {
	// Use systemctl show with specific properties for reliable parsing
	properties := []string{
		"LoadState",     // loaded, not-found, error, masked, bad-setting
		"ActiveState",   // active, inactive, activating, deactivating, failed
		"SubState",      // running, exited, dead, etc.
		"Description",   // human-readable description
		"FragmentPath",  // path to the unit file
		"Vendor",        // vendor information
		"UnitFileState", // enabled, disabled, masked, etc.
	}

	// Build the command with all properties
	args := []string{"show", "--property=" + strings.Join(properties, ","), serviceName}
	output, err := a.CommandProcessor.RunCommandWithContext(execCtx, "systemctl", args...)

	// Check if the service doesn't exist
	if err != nil || strings.Contains(output, "could not be found") || strings.Contains(output, "Unit not found") {
		return ServiceStatus{
			Name:   serviceName,
			Exists: false,
		}, nil
	}

	return a.parseServiceShowOutput(serviceName, output)
}

// parseServiceShowOutput parses the systemctl show output
func (a *GetServiceStatusAction) parseServiceShowOutput(serviceName, output string) (ServiceStatus, error) {
	status := ServiceStatus{
		Name:   serviceName,
		Exists: true,
	}

	lines := strings.Split(output, "\n")
	properties := make(map[string]string)

	// Parse each line to extract properties
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Each line is in format "Property=Value"
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				properties[parts[0]] = parts[1]
			}
		}
	}

	// Map properties to our ServiceStatus struct
	if loadState, exists := properties["LoadState"]; exists {
		status.Loaded = loadState
	}

	if activeState, exists := properties["ActiveState"]; exists {
		status.Active = activeState
	}

	if subState, exists := properties["SubState"]; exists {
		status.Sub = subState
		// If we have both ActiveState and SubState, combine them
		if status.Active != "" && subState != "" {
			status.Active = status.Active + " (" + subState + ")"
		}
	}

	if description, exists := properties["Description"]; exists {
		status.Description = description
	}

	if path, exists := properties["FragmentPath"]; exists {
		status.Path = path
	}

	if vendor, exists := properties["Vendor"]; exists {
		status.Vendor = vendor
	}

	// Determine if service exists based on LoadState
	if loadState, exists := properties["LoadState"]; exists {
		status.Exists = loadState != "not-found"
	}

	return status, nil
}
