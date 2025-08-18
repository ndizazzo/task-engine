package utility

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

// ReadMACAddressAction represents an action that reads the MAC address of a network interface
type ReadMACAddressAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	// Parameter fields
	InterfaceNameParam task_engine.ActionParameter
	// Execution result fields (not parameters)
	Interface string // Resolved interface name
	MAC       string // Read MAC address
	// Direct logger field for backward compatibility
	Logger *slog.Logger
}

// NewReadMACAddressAction creates a new ReadMACAddressAction with the given logger
func NewReadMACAddressAction(logger *slog.Logger) *ReadMACAddressAction {
	return &ReadMACAddressAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		Logger:            logger,
	}
}

// WithParameters sets the parameters for MAC address reading and returns a wrapped Action
func (a *ReadMACAddressAction) WithParameters(
	interfaceParam task_engine.ActionParameter,
) (*task_engine.Action[*ReadMACAddressAction], error) {
	if interfaceParam == nil {
		return nil, fmt.Errorf("interface name parameter cannot be nil")
	}

	a.InterfaceNameParam = interfaceParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*ReadMACAddressAction](a.Logger)
	return constructor.WrapAction(a, "Read MAC Address", "read-mac-action"), nil
}

func (a *ReadMACAddressAction) Execute(ctx context.Context) error {
	// Use the new parameter resolver to handle interface name parameter
	interfaceName, err := a.ResolveStringParameter(ctx, a.InterfaceNameParam, "interface name")
	if err != nil {
		return err
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
	// Use the new output builder to create the output
	return a.BuildStandardOutput(nil, a.MAC != "", map[string]interface{}{
		"interface": a.Interface,
		"mac":       a.MAC,
	})
}
