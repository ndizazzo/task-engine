package utility

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
)

// FetchNetInterfacesAction represents an action that fetches network interfaces
type FetchNetInterfacesAction struct {
	task_engine.BaseAction
	// Parameter fields
	NetDevicePathParam task_engine.ActionParameter
	InterfacesParam    task_engine.ActionParameter // Optional: if provided, use these interfaces instead of scanning
	// Result fields
	NetDevicePath string   // Resolved device path
	Interfaces    []string // Discovered or provided interfaces
}

// NewFetchNetInterfacesAction creates a new FetchNetInterfacesAction with the provided logger
func NewFetchNetInterfacesAction(logger *slog.Logger) *FetchNetInterfacesAction {
	return &FetchNetInterfacesAction{
		BaseAction: task_engine.BaseAction{Logger: logger},
	}
}

// WithParameters sets the parameters for device path and optional interfaces
func (a *FetchNetInterfacesAction) WithParameters(netDevicePathParam task_engine.ActionParameter, interfacesParam task_engine.ActionParameter) (*task_engine.Action[*FetchNetInterfacesAction], error) {
	if netDevicePathParam == nil {
		return nil, fmt.Errorf("net device path parameter cannot be nil")
	}
	// interfacesParam can be nil - it's optional

	a.NetDevicePathParam = netDevicePathParam
	a.InterfacesParam = interfacesParam

	return &task_engine.Action[*FetchNetInterfacesAction]{
		ID:      "fetch-interfaces-action",
		Name:    "Fetch Network Interfaces",
		Wrapped: a,
	}, nil
}

// gathers and sorts the network interfaces from the specified device path
func (a *FetchNetInterfacesAction) Execute(ctx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve interfaces parameter if it exists - if provided, use these instead of scanning
	if a.InterfacesParam != nil {
		interfacesValue, err := a.InterfacesParam.Resolve(ctx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve interfaces parameter: %w", err)
		}
		if interfacesSlice, ok := interfacesValue.([]string); ok {
			a.Interfaces = interfacesSlice
			// If we have resolved interfaces, we don't need to scan the device path
			a.Logger.Info("Using resolved interfaces parameter", "count", len(a.Interfaces))
			return nil
		} else {
			return fmt.Errorf("interfaces parameter is not a []string, got %T", interfacesValue)
		}
	}

	// Resolve the net device path parameter
	netDevicePathValue, err := a.NetDevicePathParam.Resolve(ctx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve net device path parameter: %w", err)
	}

	netDevicePath, ok := netDevicePathValue.(string)
	if !ok {
		return fmt.Errorf("net device path parameter is not a string, got %T", netDevicePathValue)
	}

	if netDevicePath == "" {
		return fmt.Errorf("net device path cannot be empty")
	}

	// Store resolved path for GetOutput
	a.NetDevicePath = netDevicePath

	entries, err := os.ReadDir(a.NetDevicePath)
	if err != nil {
		a.Logger.Error("Failed to read NetDevicePath", "NetDevicePath", a.NetDevicePath, "error", err)
		return fmt.Errorf("failed to read NetDevicePath %s: %w", a.NetDevicePath, err)
	}

	var physical []string
	var usbEthernet []string
	var wireless []string
	var other []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		interfaceName := entry.Name()

		if _, err := os.Stat(fmt.Sprintf("%s/%s/wireless", a.NetDevicePath, interfaceName)); err == nil {
			wireless = append(wireless, interfaceName)
			continue
		}

		if strings.HasPrefix(interfaceName, "enx") {
			usbEthernet = append(usbEthernet, interfaceName)
			continue
		}

		if strings.HasPrefix(interfaceName, "en") || strings.HasPrefix(interfaceName, "eth") {
			physical = append(physical, interfaceName)
			continue
		}

		other = append(other, interfaceName)
	}

	sort.Strings(physical)
	sort.Strings(usbEthernet)
	sort.Strings(wireless)
	sort.Strings(other)

	// We want to prioritize physical and ethernet devices in our list - append them first
	a.Interfaces = append(a.Interfaces, physical...)
	a.Interfaces = append(a.Interfaces, usbEthernet...)
	a.Interfaces = append(a.Interfaces, wireless...)
	a.Interfaces = append(a.Interfaces, other...)

	return nil
}

// GetOutput returns the discovered interfaces
func (a *FetchNetInterfacesAction) GetOutput() interface{} {
	return map[string]interface{}{
		"interfaces": a.Interfaces,
		"count":      len(a.Interfaces),
		"devicePath": a.NetDevicePath,
		"success":    true,
	}
}
