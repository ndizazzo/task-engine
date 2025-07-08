package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
)

// Takes a list of fully qualified file paths to remove
type RemoveFilesAction struct {
	task_engine.BaseAction

	Paths []string
}

func NewRemoveFilesAction(filePaths []string, logger *slog.Logger) *task_engine.Action[*RemoveFilesAction] {
	return &task_engine.Action[*RemoveFilesAction]{
		ID: "fetch-interfaces-action",
		Wrapped: &RemoveFilesAction{
			BaseAction: task_engine.BaseAction{Logger: logger},
			Paths:      filePaths,
		},
	}
}

func (a *RemoveFilesAction) Execute(ctx context.Context) error {
	if len(a.Paths) == 0 {
		return fmt.Errorf("no file paths provided to RemoveFilesAction")
	}

	for _, filePath := range a.Paths {
		err := os.Remove(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				a.Logger.Error("File not found (skipping)", "filePath", filePath)
				continue
			}

			a.Logger.Error("Failed to remove file", "filePath", filePath, "error", err)
			return fmt.Errorf("failure removing %s: %w", filePath, err)
		}
	}

	return nil
}
