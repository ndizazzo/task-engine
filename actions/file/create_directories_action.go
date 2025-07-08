package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	task_engine "github.com/ndizazzo/task-engine"
)

// NewCreateDirectoriesAction creates an action that creates multiple directories
// relative to the given installation path.
func NewCreateDirectoriesAction(logger *slog.Logger, rootPath string, directories []string) *task_engine.Action[*CreateDirectoriesAction] {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	if rootPath == "" {
		logger.Error("Invalid parameter: rootPath cannot be empty")
		return nil
	}
	if len(directories) == 0 {
		logger.Error("Invalid parameter: directories list cannot be empty")
		return nil
	}

	return &task_engine.Action[*CreateDirectoriesAction]{
		ID: "create-directories-action",
		Wrapped: &CreateDirectoriesAction{
			BaseAction:  task_engine.BaseAction{Logger: logger},
			RootPath:    rootPath,
			Directories: directories,
		},
	}
}

// CreateDirectoriesAction creates multiple directories relative to an installation path
type CreateDirectoriesAction struct {
	task_engine.BaseAction
	RootPath         string
	Directories      []string
	CreatedDirsCount int
}

func (a *CreateDirectoriesAction) Execute(ctx context.Context) error {
	if a.RootPath == "" {
		return fmt.Errorf("root path cannot be empty")
	}

	if len(a.Directories) == 0 {
		a.Logger.Info("No directories to create")
		return nil
	}

	a.Logger.Info("Creating directories", "count", len(a.Directories), "root_path", a.RootPath)

	for _, dir := range a.Directories {
		if dir == "" {
			a.Logger.Warn("Skipping empty directory path")
			continue
		}

		// Create full path by joining installation path with relative directory
		fullPath := filepath.Join(a.RootPath, dir)

		a.Logger.Debug("Creating directory", "path", fullPath)

		// Create directory with proper permissions
		err := os.MkdirAll(fullPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}

		a.CreatedDirsCount++
		a.Logger.Debug("Successfully created directory", "path", fullPath)
	}

	a.Logger.Info("Successfully created directories", "created_count", a.CreatedDirsCount, "total_count", len(a.Directories))
	return nil
}
