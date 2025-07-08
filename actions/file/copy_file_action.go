package file

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	task_engine "github.com/ndizazzo/task-engine"
)

type CopyFileAction struct {
	task_engine.BaseAction

	Source      string
	Destination string
	CreateDir   bool
}

func NewCopyFileAction(source, destination string, createDir bool, logger *slog.Logger) *task_engine.Action[*CopyFileAction] {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	if source == "" {
		logger.Error("Invalid parameter: source cannot be empty")
		return nil
	}
	if destination == "" {
		logger.Error("Invalid parameter: destination cannot be empty")
		return nil
	}
	if source == destination {
		logger.Error("Invalid parameter: source and destination cannot be the same")
		return nil
	}

	return &task_engine.Action[*CopyFileAction]{
		ID: "copy-file-action",
		Wrapped: &CopyFileAction{
			BaseAction:  task_engine.BaseAction{Logger: logger},
			Source:      source,
			Destination: destination,
			CreateDir:   createDir,
		},
	}
}

func (a *CopyFileAction) Execute(execCtx context.Context) error {
	if a.CreateDir {
		destDir := filepath.Dir(a.Destination)
		if err := os.MkdirAll(destDir, 0750); err != nil {
			a.Logger.Debug("Failed to create destination directory", "error", err, "directory", destDir)
			return err
		}
	}

	srcFile, err := os.Open(a.Source)
	if err != nil {
		a.Logger.Debug("Failed to open source file", "error", err, "file", a.Source)
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(a.Destination)
	if err != nil {
		a.Logger.Debug("Failed to create destination file", "error", err, "file", a.Destination)
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		a.Logger.Debug("Failed to copy file", "error", err, "source", a.Source, "destination", a.Destination)
		return err
	}

	return nil
}
