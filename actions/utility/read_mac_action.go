package utility

import (
	"context"
	"log/slog"
	"os"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
)

type ReadMACAddressAction struct {
	task_engine.BaseAction

	Interface string
	MAC       string
}

func NewReadMacAction(netInterface string, logger *slog.Logger) *task_engine.Action[*ReadMACAddressAction] {
	return &task_engine.Action[*ReadMACAddressAction]{
		ID: "fetch-mac-action",
		Wrapped: &ReadMACAddressAction{
			BaseAction: task_engine.BaseAction{Logger: logger},
			Interface:  netInterface,
			MAC:        "",
		},
	}
}

func (a *ReadMACAddressAction) Execute(ctx context.Context) error {
	data, err := os.ReadFile("/sys/class/net/" + a.Interface + "/address")
	if err != nil {
		return err
	}

	a.MAC = strings.TrimSpace(string(data))
	return nil
}
