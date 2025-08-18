package file

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
)

type CreateSymlinkAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
	Target     string
	LinkPath   string
	Overwrite  bool
	CreateDirs bool

	// Parameter-aware fields
	TargetParam   task_engine.ActionParameter
	LinkPathParam task_engine.ActionParameter
}

// NewCreateSymlinkAction creates a new CreateSymlinkAction with the given logger
func NewCreateSymlinkAction(logger *slog.Logger) *CreateSymlinkAction {
	return &CreateSymlinkAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
	}
}

// WithParameters sets the parameters for target, link path, overwrite flag, and create directories flag
func (a *CreateSymlinkAction) WithParameters(
	targetParam task_engine.ActionParameter,
	linkPathParam task_engine.ActionParameter,
	overwrite bool,
	createDirs bool,
) (*task_engine.Action[*CreateSymlinkAction], error) {
	a.TargetParam = targetParam
	a.LinkPathParam = linkPathParam
	a.Overwrite = overwrite
	a.CreateDirs = createDirs

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*CreateSymlinkAction](a.Logger)
	return constructor.WrapAction(a, "Create Symlink", "create-symlink-action"), nil
}

func (a *CreateSymlinkAction) Execute(execCtx context.Context) error {
	// Resolve parameters using the ParameterResolver
	if a.TargetParam != nil {
		targetValue, err := a.ResolveStringParameter(execCtx, a.TargetParam, "target")
		if err != nil {
			return err
		}
		a.Target = targetValue
	}

	if a.LinkPathParam != nil {
		linkPathValue, err := a.ResolveStringParameter(execCtx, a.LinkPathParam, "link path")
		if err != nil {
			return err
		}
		a.LinkPath = linkPathValue
	}

	// Sanitize paths to prevent path traversal attacks
	sanitizedTarget, err := SanitizePath(a.Target)
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}
	sanitizedLinkPath, err := SanitizePath(a.LinkPath)
	if err != nil {
		return fmt.Errorf("invalid link path: %w", err)
	}

	a.Logger.Info("Creating symlink", "target", sanitizedTarget, "link", sanitizedLinkPath, "overwrite", a.Overwrite, "createDirs", a.CreateDirs)
	if _, err := os.Lstat(sanitizedLinkPath); err == nil {
		if !a.Overwrite {
			errMsg := fmt.Sprintf("symlink %s already exists and overwrite is set to false", sanitizedLinkPath)
			a.Logger.Error(errMsg)
			return errors.New(errMsg)
		}
		// Remove existing symlink if overwrite is enabled
		if err := os.Remove(sanitizedLinkPath); err != nil {
			a.Logger.Error("Failed to remove existing symlink", "path", sanitizedLinkPath, "error", err)
			return fmt.Errorf("failed to remove existing symlink %s: %w", sanitizedLinkPath, err)
		}
		a.Logger.Debug("Removed existing symlink", "path", sanitizedLinkPath)
	} else if !os.IsNotExist(err) {
		a.Logger.Error("Failed to check if symlink exists", "path", sanitizedLinkPath, "error", err)
		return fmt.Errorf("failed to stat symlink %s: %w", sanitizedLinkPath, err)
	}

	// Create parent directories if requested
	if a.CreateDirs {
		linkDir := filepath.Dir(sanitizedLinkPath)
		if err := os.MkdirAll(linkDir, 0o750); err != nil {
			a.Logger.Error("Failed to create parent directory for symlink", "path", linkDir, "error", err)
			return fmt.Errorf("failed to create directory %s for symlink: %w", linkDir, err)
		}
		a.Logger.Debug("Created parent directory", "path", linkDir)
	}

	// Create the symlink
	if err := os.Symlink(sanitizedTarget, sanitizedLinkPath); err != nil {
		a.Logger.Error("Failed to create symlink", "target", sanitizedTarget, "link", sanitizedLinkPath, "error", err)
		return fmt.Errorf("failed to create symlink %s -> %s: %w", sanitizedLinkPath, sanitizedTarget, err)
	}
	if err := a.verifySymlink(sanitizedLinkPath, sanitizedTarget); err != nil {
		a.Logger.Error("Failed to verify symlink", "link", sanitizedLinkPath, "error", err)
		return fmt.Errorf("failed to verify symlink %s: %w", sanitizedLinkPath, err)
	}

	a.Logger.Info("Successfully created symlink", "target", sanitizedTarget, "link", sanitizedLinkPath)
	return nil
}

// GetOutput returns metadata about the created symlink
func (a *CreateSymlinkAction) GetOutput() interface{} {
	return a.BuildStandardOutput(nil, true, map[string]interface{}{
		"target":    a.Target,
		"linkPath":  a.LinkPath,
		"overwrite": a.Overwrite,
		"created":   true,
	})
}

func (a *CreateSymlinkAction) verifySymlink(linkPath, expectedTarget string) error {
	if err := a.checkSymlinkExists(linkPath); err != nil {
		return err
	}

	// Read the symlink target
	actualTarget, err := a.readSymlinkTarget(linkPath)
	if err != nil {
		return err
	}

	// Compare targets (handle both absolute and relative paths)
	if err := a.compareSymlinkTargets(linkPath, expectedTarget, actualTarget); err != nil {
		return err
	}

	return nil
}

func (a *CreateSymlinkAction) checkSymlinkExists(linkPath string) error {
	info, err := os.Lstat(linkPath)
	if err != nil {
		return fmt.Errorf("failed to stat symlink: %w", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("created file is not a symlink")
	}

	return nil
}

func (a *CreateSymlinkAction) readSymlinkTarget(linkPath string) (string, error) {
	actualTarget, err := os.Readlink(linkPath)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink target: %w", err)
	}
	return actualTarget, nil
}

func (a *CreateSymlinkAction) compareSymlinkTargets(linkPath, expectedTarget, actualTarget string) error {
	if actualTarget == expectedTarget {
		return nil
	}

	// Try to resolve relative paths for comparison
	linkDir := filepath.Dir(linkPath)
	resolvedExpected := expectedTarget
	if !filepath.IsAbs(expectedTarget) {
		resolvedExpected = filepath.Join(linkDir, expectedTarget)
	}
	resolvedActual := actualTarget
	if !filepath.IsAbs(actualTarget) {
		resolvedActual = filepath.Join(linkDir, actualTarget)
	}

	if resolvedActual != resolvedExpected {
		return fmt.Errorf("symlink target mismatch: expected %s, got %s", expectedTarget, actualTarget)
	}

	return nil
}
