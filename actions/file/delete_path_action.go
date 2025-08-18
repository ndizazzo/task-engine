package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

type DeletePathAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	Path            string
	Recursive       bool
	DryRun          bool
	IncludeHidden   bool
	ExcludePatterns []string

	// Parameter-aware fields
	PathParam task_engine.ActionParameter
}

type DeleteEntry struct {
	Path  string
	Type  string // "file", "directory", "symlink", "special"
	Size  int64
	Mode  os.FileMode
	Error error
}

// NewDeletePathAction creates a new DeletePathAction with the given logger
func NewDeletePathAction(logger *slog.Logger) *DeletePathAction {
	return &DeletePathAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for path, recursive flag, dry run flag, include hidden flag, and exclude patterns
func (a *DeletePathAction) WithParameters(
	pathParam task_engine.ActionParameter,
	recursive bool,
	dryRun bool,
	includeHidden bool,
	excludePatterns []string,
) (*task_engine.Action[*DeletePathAction], error) {
	a.PathParam = pathParam
	a.Recursive = recursive
	a.DryRun = dryRun
	a.IncludeHidden = includeHidden
	a.ExcludePatterns = excludePatterns

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*DeletePathAction](a.Logger)
	return constructor.WrapAction(a, "Delete Path", "delete-path-action"), nil
}

func (a *DeletePathAction) Execute(execCtx context.Context) error {
	// Resolve path parameter using the ParameterResolver
	if a.PathParam != nil {
		pathValue, err := a.ResolveStringParameter(execCtx, a.PathParam, "path")
		if err != nil {
			return err
		}
		a.Path = pathValue
	}

	if a.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Sanitize path to prevent path traversal attacks
	sanitizedPath, err := SanitizePath(a.Path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	info, err := os.Stat(sanitizedPath)
	if os.IsNotExist(err) {
		a.Logger.Warn("Path does not exist, skipping deletion", "path", sanitizedPath)
		return nil
	}
	if err != nil {
		a.Logger.Error("Failed to stat path", "path", sanitizedPath, "error", err)
		return fmt.Errorf("failed to stat path %s: %w", sanitizedPath, err)
	}

	// If it's a directory and recursive flag is set, use recursive delete logic
	if info.IsDir() && a.Recursive {
		return a.executeRecursiveDelete()
	}

	// If it's a directory but recursive flag is not set, return error
	if info.IsDir() && !a.Recursive {
		a.Logger.Error("Cannot delete directory without recursive flag", "path", sanitizedPath)
		return fmt.Errorf("cannot delete directory %s without recursive flag", sanitizedPath)
	}

	// Otherwise, use the original file-based delete logic
	return a.executeFileDelete(sanitizedPath)
}

func (a *DeletePathAction) executeRecursiveDelete() error {
	a.Logger.Info("Executing recursive delete", "path", a.Path, "dryRun", a.DryRun)

	// Build list of entries to delete for dry-run and logging
	entries, err := a.buildDeleteList()
	if err != nil {
		return fmt.Errorf("failed to build delete list: %w", err)
	}

	// Log what would be deleted
	a.logDeletePlan(entries)

	// If dry run, just return success
	if a.DryRun {
		a.Logger.Info("Dry run completed - no files were actually deleted",
			"path", a.Path,
			"totalEntries", len(entries))
		return nil
	}

	// Use os.RemoveAll for actual deletion - it's more reliable and handles edge cases better
	err = os.RemoveAll(a.Path)
	if err != nil {
		a.Logger.Error("Failed to delete directory recursively", "path", a.Path, "error", err)
		return fmt.Errorf("failed to delete directory %s recursively: %w", a.Path, err)
	}

	a.Logger.Info("Successfully deleted directory recursively", "path", a.Path)
	return nil
}

func (a *DeletePathAction) buildDeleteList() ([]DeleteEntry, error) {
	var entries []DeleteEntry

	// Walk the directory tree and collect all paths to delete
	err := filepath.Walk(a.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log the error but continue with other files
			a.Logger.Warn("Error accessing path during walk", "path", path, "error", err)
			entries = append(entries, DeleteEntry{
				Path:  path,
				Type:  "error",
				Error: err,
			})
			return nil
		}

		// Skip the root directory itself (we'll delete it at the end)
		if path == a.Path {
			return nil
		}

		// Determine entry type
		entryType := "file"
		switch {
		case info.IsDir():
			entryType = "directory"
		case info.Mode()&os.ModeSymlink != 0:
			entryType = "symlink"
		case info.Mode()&os.ModeType != 0:
			entryType = "special"
		}

		entries = append(entries, DeleteEntry{
			Path: path,
			Type: entryType,
			Size: info.Size(),
			Mode: info.Mode(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Add the root directory at the end
	rootInfo, err := os.Stat(a.Path)
	if err == nil {
		entries = append(entries, DeleteEntry{
			Path: a.Path,
			Type: "directory",
			Size: rootInfo.Size(),
			Mode: rootInfo.Mode(),
		})
	}

	return entries, nil
}

func (a *DeletePathAction) logDeletePlan(entries []DeleteEntry) {
	a.Logger.Info("Delete plan",
		"path", a.Path,
		"totalEntries", len(entries),
		"dryRun", a.DryRun)

	// Count by type
	counts := make(map[string]int)
	for _, entry := range entries {
		if entry.Error != nil {
			counts["error"]++
			continue
		}
		counts[entry.Type]++
	}

	for entryType, count := range counts {
		a.Logger.Info("Entry count", "type", entryType, "count", count)
	}

	// Log first few entries for visibility
	maxLogEntries := 10
	if len(entries) > maxLogEntries {
		a.Logger.Info("Showing first entries", "maxShown", maxLogEntries)
	}

	for i, entry := range entries {
		if i >= maxLogEntries {
			break
		}
		if entry.Error != nil {
			a.Logger.Debug("Would skip (error)", "path", entry.Path, "error", entry.Error)
		} else {
			a.Logger.Debug("Would delete", "path", entry.Path, "type", entry.Type, "size", entry.Size)
		}
	}

	if len(entries) > maxLogEntries {
		a.Logger.Info("... and more entries", "remaining", len(entries)-maxLogEntries)
	}
}

func (a *DeletePathAction) executeFileDelete(sanitizedPath string) error {
	a.Logger.Info("Deleting file", "path", sanitizedPath, "dryRun", a.DryRun)

	// Get file info for logging
	info, err := os.Stat(sanitizedPath)
	if err == nil {
		a.Logger.Info("Would delete file", "path", sanitizedPath, "size", info.Size(), "mode", info.Mode())
	}

	if a.DryRun {
		a.Logger.Info("Dry run completed - file would be deleted", "path", sanitizedPath)
		return nil
	}

	err = os.Remove(sanitizedPath)
	if err != nil {
		if os.IsNotExist(err) {
			a.Logger.Warn("File does not exist, skipping deletion", "path", sanitizedPath)
			return nil
		}
		a.Logger.Error("Failed to delete file", "path", sanitizedPath, "error", err)
		return fmt.Errorf("failed to delete file %s: %w", sanitizedPath, err)
	}

	a.Logger.Info("Successfully deleted file", "path", sanitizedPath)
	return nil
}

// GetOutput returns metadata about the delete operation
func (a *DeletePathAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"path":      a.Path,
		"recursive": a.Recursive,
		"dryRun":    a.DryRun,
	})
}
