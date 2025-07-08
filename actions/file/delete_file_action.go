package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	task_engine "github.com/ndizazzo/task-engine"
)

// NewDeleteFileAction creates an action to delete a file
func NewDeleteFileAction(logger *slog.Logger, filePath string) *task_engine.Action[*DeleteFileAction] {
	id := fmt.Sprintf("delete-file-%s-action", filePath)
	return &task_engine.Action[*DeleteFileAction]{
		ID: id,
		Wrapped: &DeleteFileAction{
			BaseAction: task_engine.BaseAction{Logger: logger},
			FilePath:   filePath,
		},
	}
}

type DeleteFileAction struct {
	task_engine.BaseAction
	FilePath string
}

func (a *DeleteFileAction) Execute(execCtx context.Context) error {
	a.Logger.Info("Deleting file", "path", a.FilePath)

	err := os.Remove(a.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			a.Logger.Warn("File does not exist, skipping deletion", "path", a.FilePath)
			return nil
		}
		a.Logger.Error("Failed to delete file", "path", a.FilePath, "error", err)
		return fmt.Errorf("failed to delete file %s: %w", a.FilePath, err)
	}

	a.Logger.Info("Successfully deleted file", "path", a.FilePath)
	return nil
}
