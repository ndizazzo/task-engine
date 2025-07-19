package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	task_engine "github.com/ndizazzo/task-engine"
)

type DeletePathAction struct {
	task_engine.BaseAction
	Path      string
	Recursive bool
	DryRun    bool
}

type DeleteEntry struct {
	Path  string
	Type  string // "file", "directory", "symlink", "special"
	Size  int64
	Mode  os.FileMode
	Error error
}

func NewDeletePathAction(path string, recursive bool, dryRun bool, logger *slog.Logger) *task_engine.Action[*DeletePathAction] {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	if path == "" {
		logger.Error("Invalid parameter: path cannot be empty")
		return nil
	}

	return &task_engine.Action[*DeletePathAction]{
		ID: "delete-path-action",
		Wrapped: &DeletePathAction{
			BaseAction: task_engine.BaseAction{Logger: logger},
			Path:       path,
			Recursive:  recursive,
			DryRun:     dryRun,
		},
	}
}

func (a *DeletePathAction) Execute(execCtx context.Context) error {
	// Check if path exists
	info, err := os.Stat(a.Path)
	if os.IsNotExist(err) {
		a.Logger.Warn("Path does not exist, skipping deletion", "path", a.Path)
		return nil
	}
	if err != nil {
		a.Logger.Error("Failed to stat path", "path", a.Path, "error", err)
		return fmt.Errorf("failed to stat path %s: %w", a.Path, err)
	}

	// If it's a directory and recursive flag is set, use recursive delete logic
	if info.IsDir() && a.Recursive {
		return a.executeRecursiveDelete(execCtx)
	}

	// If it's a directory but recursive flag is not set, return error
	if info.IsDir() && !a.Recursive {
		a.Logger.Error("Cannot delete directory without recursive flag", "path", a.Path)
		return fmt.Errorf("cannot delete directory %s without recursive flag", a.Path)
	}

	// Otherwise, use the original file-based delete logic
	return a.executeFileDelete(execCtx)
}

func (a *DeletePathAction) executeRecursiveDelete(execCtx context.Context) error {
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

func (a *DeletePathAction) executeFileDelete(execCtx context.Context) error {
	a.Logger.Info("Deleting file", "path", a.Path, "dryRun", a.DryRun)

	// Get file info for logging
	info, err := os.Stat(a.Path)
	if err == nil {
		a.Logger.Info("Would delete file", "path", a.Path, "size", info.Size(), "mode", info.Mode())
	}

	if a.DryRun {
		a.Logger.Info("Dry run completed - file would be deleted", "path", a.Path)
		return nil
	}

	err = os.Remove(a.Path)
	if err != nil {
		if os.IsNotExist(err) {
			a.Logger.Warn("File does not exist, skipping deletion", "path", a.Path)
			return nil
		}
		a.Logger.Error("Failed to delete file", "path", a.Path, "error", err)
		return fmt.Errorf("failed to delete file %s: %w", a.Path, err)
	}

	a.Logger.Info("Successfully deleted file", "path", a.Path)
	return nil
}
