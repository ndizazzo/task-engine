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

type FetchNetInterfacesAction struct {
	task_engine.BaseAction

	NetDevicePath string
	Interfaces    []string
}

func NewFetchNetInterfacesAction(devicePath string, logger *slog.Logger) *task_engine.Action[*FetchNetInterfacesAction] {
	return &task_engine.Action[*FetchNetInterfacesAction]{
		ID: "fetch-interfaces-action",
		Wrapped: &FetchNetInterfacesAction{
			BaseAction:    task_engine.BaseAction{Logger: logger},
			NetDevicePath: devicePath,
			Interfaces:    []string{},
		},
	}
}

// gathers and sorts the network interfaces from the specified device path
func (a *FetchNetInterfacesAction) Execute(ctx context.Context) error {
	if a.NetDevicePath == "" {
		return fmt.Errorf("NetDevicePath cannot be empty")
	}

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
