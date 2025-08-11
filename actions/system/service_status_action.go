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

// NewServiceStatusAction creates a new ServiceStatusAction with the given logger
func NewServiceStatusAction(logger *slog.Logger) *ServiceStatusAction {
	return &ServiceStatusAction{
		BaseAction:       task_engine.NewBaseAction(logger),
		CommandProcessor: command.NewDefaultCommandRunner(),
	}
}

// NewGetAllServicesStatusAction creates an action to get the status of all services (backward compatibility)
// This function exists for backward compatibility and returns an action that will fail because no service names are provided
func NewGetAllServicesStatusAction(logger *slog.Logger) *task_engine.Action[*ServiceStatusAction] {
	return &task_engine.Action[*ServiceStatusAction]{
		ID:   "get-all-services-status-action",
		Name: "Get All Services Status",
		Wrapped: &ServiceStatusAction{
			BaseAction:       task_engine.NewBaseAction(logger),
			ServiceNames:     []string{}, // Empty means get all - will cause error
			CommandProcessor: command.NewDefaultCommandRunner(),
		},
	}
}

// ServiceStatusAction retrieves the status of systemd services
type ServiceStatusAction struct {
	task_engine.BaseAction

	// Parameters
	ServiceNamesParam task_engine.ActionParameter

	// Runtime resolved values
	ServiceNames    []string
	ServiceStatuses []ServiceStatus

	CommandProcessor command.CommandRunner
}

// WithParameters sets the service names parameter and returns a wrapped Action
func (a *ServiceStatusAction) WithParameters(serviceNamesParam task_engine.ActionParameter) (*task_engine.Action[*ServiceStatusAction], error) {
	a.ServiceNamesParam = serviceNamesParam

	id := "service-status-action"
	return &task_engine.Action[*ServiceStatusAction]{
		ID:      id,
		Name:    "Service Status",
		Wrapped: a,
	}, nil
}

// SetCommandProcessor allows injecting a mock or alternative CommandProcessor for testing
func (a *ServiceStatusAction) SetCommandProcessor(processor command.CommandRunner) {
	a.CommandProcessor = processor
}

func (a *ServiceStatusAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve service names parameter if it exists
	if a.ServiceNamesParam != nil {
		serviceNamesValue, err := a.ServiceNamesParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve service names parameter: %w", err)
		}
		if serviceNamesSlice, ok := serviceNamesValue.([]string); ok {
			a.ServiceNames = serviceNamesSlice
		} else {
			return fmt.Errorf("service names parameter is not a []string, got %T", serviceNamesValue)
		}
	}

	if len(a.ServiceNames) == 0 {
		return fmt.Errorf("no service names provided and no parameter to resolve")
	}

	a.Logger.Info("Getting service status", "serviceNames", a.ServiceNames)

	var serviceStatuses []ServiceStatus

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

// GetOutput returns the retrieved service statuses
func (a *ServiceStatusAction) GetOutput() interface{} {
	return map[string]interface{}{
		"services": a.ServiceStatuses,
		"count":    len(a.ServiceStatuses),
		"success":  true,
	}
}

// getServiceStatus gets the status of a single service using systemctl show
func (a *ServiceStatusAction) getServiceStatus(execCtx context.Context, serviceName string) (ServiceStatus, error) {
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
	if err != nil || strings.Contains(output, "could not be found") || strings.Contains(output, "Unit not found") {
		return ServiceStatus{
			Name:   serviceName,
			Exists: false,
		}, nil
	}

	return a.parseServiceShowOutput(serviceName, output)
}

// parseServiceShowOutput parses the systemctl show output
func (a *ServiceStatusAction) parseServiceShowOutput(serviceName, output string) (ServiceStatus, error) {
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
