package utility

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
	"gopkg.in/yaml.v3"
)

type WriteCloudInitConfigAction struct {
	task_engine.BaseAction

	MACAddress string
	ConfigPath string
	Interface  string // TODO: is eth0 the interface on Lemony? IDK, need to check this
}

func NewWriteCloudInitConfigAction(
	mac,
	configPath,
	netInterface string,
	logger *slog.Logger,
) *task_engine.Action[*WriteCloudInitConfigAction] {
	return &task_engine.Action[*WriteCloudInitConfigAction]{
		ID: "write-cloudinit-action",
		Wrapped: &WriteCloudInitConfigAction{
			BaseAction: task_engine.BaseAction{Logger: logger},
			MACAddress: mac,
			ConfigPath: configPath,
			Interface:  netInterface,
		},
	}
}

func (a *WriteCloudInitConfigAction) Execute(ctx context.Context) error {
	a.Logger.Error("Starting WriteCloudInitConfigAction", "ConfigPath", a.ConfigPath, "MACAddress", a.MACAddress, "Interface", a.Interface)

	if a.MACAddress == "" {
		return fmt.Errorf("MACAddress cannot be empty")
	}
	if a.ConfigPath == "" {
		return fmt.Errorf("ConfigPath cannot be empty")
	}

	data := map[string]interface{}{
		"network": map[string]interface{}{
			"version": 2,
			"ethernets": map[string]interface{}{
				a.Interface: map[string]interface{}{
					"match": map[string]interface{}{
						"macaddress": a.MACAddress,
					},
					"dhcp4":    true,
					"set-name": a.Interface,
				},
			},
		},
	}

	yamlContent, err := yaml.Marshal(data)
	if err != nil {
		a.Logger.Error("Failed to marshal YAML content", "error", err)
		return fmt.Errorf("failed to marshal YAML content: %w", err)
	}

	file, err := os.OpenFile(a.ConfigPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		a.Logger.Error("Failed to open config file", "ConfigPath", a.ConfigPath, "error", err)
		return fmt.Errorf("failed to open config file %s: %w", a.ConfigPath, err)
	}
	defer file.Close()

	_, err = file.Write(yamlContent)
	if err != nil {
		a.Logger.Error("Failed to write to config file", "ConfigPath", a.ConfigPath, "error", err)
		return fmt.Errorf("failed to write to config file %s: %w", a.ConfigPath, err)
	}

	a.Logger.Error("Successfully wrote to cloud-init config file", "ConfigPath", a.ConfigPath, "MACAddress", a.MACAddress)
	return nil
}
