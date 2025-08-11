package utility

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
)

// ReadMACAddressAction represents an action that reads the MAC address of a network interface
type ReadMACAddressAction struct {
	task_engine.BaseAction
	// Parameter fields
	InterfaceNameParam task_engine.ActionParameter
	// Execution result fields (not parameters)
	Interface string // Resolved interface name
	MAC       string // Read MAC address
}

// NewReadMACAddressAction creates a new ReadMACAddressAction with the provided logger
func NewReadMACAddressAction(logger *slog.Logger) *ReadMACAddressAction {
	return &ReadMACAddressAction{
		BaseAction: task_engine.BaseAction{Logger: logger},
	}
}

// WithParameters sets the interface name parameter and returns the wrapped action
func (a *ReadMACAddressAction) WithParameters(interfaceNameParam task_engine.ActionParameter) (*task_engine.Action[*ReadMACAddressAction], error) {
	if interfaceNameParam == nil {
		return nil, fmt.Errorf("interface name parameter cannot be nil")
	}

	a.InterfaceNameParam = interfaceNameParam

	return &task_engine.Action[*ReadMACAddressAction]{
		ID:      "read-mac-action",
		Name:    "Read MAC Address",
		Wrapped: a,
	}, nil
}

func (a *ReadMACAddressAction) Execute(ctx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := ctx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve interface name parameter
	interfaceNameValue, err := a.InterfaceNameParam.Resolve(ctx, globalContext)
	if err != nil {
		return fmt.Errorf("failed to resolve interface name parameter: %w", err)
	}

	interfaceName, ok := interfaceNameValue.(string)
	if !ok {
		return fmt.Errorf("interface name parameter is not a string, got %T", interfaceNameValue)
	}

	if interfaceName == "" {
		return fmt.Errorf("interface name cannot be empty")
	}

	// Store resolved interface name for GetOutput
	a.Interface = interfaceName

	data, err := os.ReadFile("/sys/class/net/" + interfaceName + "/address")
	if err != nil {
		return fmt.Errorf("failed to read MAC address for interface %s: %w", interfaceName, err)
	}

	mac := strings.TrimSpace(string(data))
	if mac == "" {
		return fmt.Errorf("empty MAC address for interface %s", interfaceName)
	}

	// Store the result for GetOutput
	a.MAC = mac
	return nil
}

func (a *ReadMACAddressAction) GetOutput() interface{} {
	return map[string]interface{}{
		"interface": a.Interface,
		"mac":       a.MAC,
		"success":   a.MAC != "",
	}
}
