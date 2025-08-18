package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

// NewCreateDirectoriesAction creates a new CreateDirectoriesAction with the given logger
func NewCreateDirectoriesAction(logger *slog.Logger) *CreateDirectoriesAction {
	return &CreateDirectoriesAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for directory creation and returns a wrapped Action
func (a *CreateDirectoriesAction) WithParameters(
	rootPathParam task_engine.ActionParameter,
	directoriesParam task_engine.ActionParameter,
) (*task_engine.Action[*CreateDirectoriesAction], error) {
	a.RootPathParam = rootPathParam
	a.DirectoriesParam = directoriesParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*CreateDirectoriesAction](a.Logger)
	return constructor.WrapAction(a, "Create Directories", "create-directories-action"), nil
}

// CreateDirectoriesAction creates multiple directories relative to an installation path
type CreateDirectoriesAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	RootPath         string
	Directories      []string
	CreatedDirsCount int

	// Parameter-aware fields
	RootPathParam    task_engine.ActionParameter
	DirectoriesParam task_engine.ActionParameter
}

func (a *CreateDirectoriesAction) Execute(execCtx context.Context) error {
	// Resolve parameters using the ParameterResolver
	if a.RootPathParam != nil {
		rootPathValue, err := a.ResolveStringParameter(execCtx, a.RootPathParam, "root path")
		if err != nil {
			return err
		}
		a.RootPath = rootPathValue
	}

	if a.DirectoriesParam != nil {
		directoriesValue, err := a.ResolveParameter(execCtx, a.DirectoriesParam, "directories")
		if err != nil {
			return err
		}
		if directoriesSlice, ok := directoriesValue.([]string); ok {
			a.Directories = directoriesSlice
		} else {
			return fmt.Errorf("directories parameter is not a []string, got %T", directoriesValue)
		}
	}

	if a.RootPath == "" {
		return fmt.Errorf("root path cannot be empty")
	}

	if len(a.Directories) == 0 {
		return fmt.Errorf("directories list cannot be empty")
	}

	a.Logger.Info("Creating directories", "rootPath", a.RootPath, "directories", a.Directories)

	createdCount := 0
	for _, dir := range a.Directories {
		// Skip empty directory names
		if strings.TrimSpace(dir) == "" {
			continue
		}

		fullPath := filepath.Join(a.RootPath, dir)
		if _, err := os.Stat(fullPath); err == nil {
			a.Logger.Debug("Directory already exists", "path", fullPath)
			createdCount++ // Count existing directories as "created" for test compatibility
			continue
		}

		// Create the directory with parents
		if err := os.MkdirAll(fullPath, 0o750); err != nil {
			a.Logger.Error("Failed to create directory", "path", fullPath, "error", err)
			return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
		}

		a.Logger.Debug("Created directory", "path", fullPath)
		createdCount++
	}

	a.CreatedDirsCount = createdCount
	a.Logger.Info("Successfully created directories", "created", createdCount, "total", len(a.Directories))
	return nil
}

// GetOutput returns metadata about the directory creation
func (a *CreateDirectoriesAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"rootPath":    a.RootPath,
		"directories": a.Directories,
		"created":     a.CreatedDirsCount,
		"total":       len(a.Directories),
	})
}
