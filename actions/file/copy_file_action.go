package file

import (
	"context"
	"fmt"
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
	Recursive   bool
}

func NewCopyFileAction(source, destination string, createDir, recursive bool, logger *slog.Logger) (*task_engine.Action[*CopyFileAction], error) {
	if err := ValidateSourcePath(source); err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}
	if err := ValidateDestinationPath(destination); err != nil {
		return nil, fmt.Errorf("invalid destination path: %w", err)
	}
	if source == destination {
		return nil, fmt.Errorf("invalid parameter: source and destination cannot be the same")
	}

	return &task_engine.Action[*CopyFileAction]{
		ID: "copy-file-action",
		Wrapped: &CopyFileAction{
			BaseAction:  task_engine.NewBaseAction(logger),
			Source:      source,
			Destination: destination,
			CreateDir:   createDir,
			Recursive:   recursive,
		},
	}, nil
}

func (a *CopyFileAction) Execute(execCtx context.Context) error {
	// Check if source exists
	if _, err := os.Stat(a.Source); os.IsNotExist(err) {
		a.Logger.Error("Source path does not exist", "source", a.Source)
		return err
	}

	// If recursive flag is set, use recursive copy logic
	if a.Recursive {
		return a.executeRecursiveCopy()
	}

	// Otherwise, use the original file-based copy logic
	return a.executeFileCopy()
}

func (a *CopyFileAction) executeRecursiveCopy() error {
	sourceInfo, err := os.Stat(a.Source)
	if err != nil {
		a.Logger.Error("Failed to stat source", "source", a.Source, "error", err)
		return err
	}

	// If source is a file, just copy it normally
	if !sourceInfo.IsDir() {
		return a.executeFileCopy()
	}

	// For directories, create destination directory and copy contents recursively
	if a.CreateDir {
		if err := os.MkdirAll(a.Destination, 0750); err != nil {
			a.Logger.Debug("Failed to create destination directory", "error", err, "directory", a.Destination)
			return err
		}
	}

	a.Logger.Info("Executing recursive copy", "source", a.Source, "destination", a.Destination)

	return filepath.Walk(a.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log the error but continue with other files
			a.Logger.Warn("Error accessing path during walk", "path", path, "error", err)
			return nil
		}

		// Calculate the relative path from source
		relPath, err := filepath.Rel(a.Source, path)
		if err != nil {
			a.Logger.Error("Failed to calculate relative path", "path", path, "source", a.Source, "error", err)
			return err
		}

		// Skip the source directory itself
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(a.Destination, relPath)

		// Handle different file types
		switch {
		case info.IsDir():
			// Create directory in destination
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				a.Logger.Error("Failed to create directory", "path", destPath, "error", err)
				return err
			}
			a.Logger.Debug("Created directory", "path", destPath)

		case info.Mode()&os.ModeSymlink != 0:
			// Handle symlinks - copy the symlink itself, not the target
			if err := a.copySymlink(path, destPath); err != nil {
				a.Logger.Warn("Failed to copy symlink", "source", path, "destination", destPath, "error", err)
				// Continue with other files instead of failing the entire operation
				return nil
			}
			a.Logger.Debug("Copied symlink", "source", path, "destination", destPath)

		case info.Mode()&os.ModeType == 0:
			// Regular file
			if err := a.copyFile(path, destPath, info.Mode()); err != nil {
				a.Logger.Error("Failed to copy file", "source", path, "destination", destPath, "error", err)
				return err
			}
			a.Logger.Debug("Copied file", "source", path, "destination", destPath)

		default:
			// Skip other special files (sockets, devices, etc.)
			a.Logger.Debug("Skipping special file", "path", path, "mode", info.Mode())
		}

		return nil
	})
}

func (a *CopyFileAction) copyFile(src, dst string, mode os.FileMode) error {
	// Sanitize paths to prevent path traversal attacks
	sanitizedSrc, err := SanitizePath(src)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	sanitizedDst, err := SanitizePath(dst)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(sanitizedDst)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return err
	}

	// Open source file
	// nosec G304 - Path is sanitized by SanitizePath function
	srcFile, err := os.Open(sanitizedSrc)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	// nosec G304 - Path is sanitized by SanitizePath function
	dstFile, err := os.Create(sanitizedDst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Set file mode
	return dstFile.Chmod(mode)
}

func (a *CopyFileAction) copySymlink(src, dst string) error {
	// Sanitize paths to prevent path traversal attacks
	sanitizedSrc, err := SanitizePath(src)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	sanitizedDst, err := SanitizePath(dst)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// Read the symlink target
	target, err := os.Readlink(sanitizedSrc)
	if err != nil {
		return err
	}

	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(sanitizedDst)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return err
	}

	// Create the symlink in the destination
	return os.Symlink(target, sanitizedDst)
}

func (a *CopyFileAction) executeFileCopy() error {
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
