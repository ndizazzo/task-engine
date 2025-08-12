package system

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// PackageManager represents the type of package manager to use
type PackageManager string

const (
	// AptPackageManager represents apt package manager for Debian-based systems
	AptPackageManager PackageManager = "apt"
	// BrewPackageManager represents Homebrew package manager for macOS
	BrewPackageManager PackageManager = "brew"
)

// UpdatePackagesActionConstructor provides the modern constructor pattern
type UpdatePackagesActionConstructor struct {
	logger *slog.Logger
}

// NewUpdatePackagesAction creates a new UpdatePackagesAction constructor
func NewUpdatePackagesAction(logger *slog.Logger) *UpdatePackagesActionConstructor {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return &UpdatePackagesActionConstructor{
		logger: logger,
	}
}

// WithParameters creates an UpdatePackagesAction with the given parameters
func (c *UpdatePackagesActionConstructor) WithParameters(
	packageNamesParam task_engine.ActionParameter,
	packageManagerParam task_engine.ActionParameter,
) (*task_engine.Action[*UpdatePackagesAction], error) {
	// Detect package manager based on OS if not explicitly provided
	defaultPackageManager := detectPackageManager()

	action := &UpdatePackagesAction{
		BaseAction:          task_engine.NewBaseAction(c.logger),
		PackageNames:        []string{},
		PackageManager:      defaultPackageManager, // May be overridden at runtime
		CommandRunner:       command.NewDefaultCommandRunner(),
		PackageNamesParam:   packageNamesParam,
		PackageManagerParam: packageManagerParam,
	}

	return &task_engine.Action[*UpdatePackagesAction]{
		ID:      "update-packages-action",
		Name:    "Update Packages",
		Wrapped: action,
	}, nil
}

// UpdatePackagesAction updates packages using the appropriate package manager
type UpdatePackagesAction struct {
	task_engine.BaseAction
	PackageNames   []string
	PackageManager PackageManager
	CommandRunner  command.CommandRunner

	// Parameter-aware fields
	PackageNamesParam   task_engine.ActionParameter
	PackageManagerParam task_engine.ActionParameter
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *UpdatePackagesAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandRunner = runner
}

func (a *UpdatePackagesAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve package names parameter if it exists
	if a.PackageNamesParam != nil {
		packageNamesValue, err := a.PackageNamesParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve package names parameter: %w", err)
		}
		if packageNamesSlice, ok := packageNamesValue.([]string); ok {
			a.PackageNames = packageNamesSlice
		} else if packageNamesStr, ok := packageNamesValue.(string); ok {
			// If it's a single string, split by comma or space
			if strings.Contains(packageNamesStr, ",") {
				a.PackageNames = strings.Split(packageNamesStr, ",")
			} else {
				a.PackageNames = strings.Fields(packageNamesStr)
			}
		} else {
			return fmt.Errorf("package names parameter is not a string slice or string, got %T", packageNamesValue)
		}
	}

	// Resolve package manager parameter if it exists
	if a.PackageManagerParam != nil {
		packageManagerValue, err := a.PackageManagerParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve package manager parameter: %w", err)
		}
		if packageManagerStr, ok := packageManagerValue.(string); ok {
			a.PackageManager = PackageManager(packageManagerStr)
		} else {
			return fmt.Errorf("package manager parameter is not a string, got %T", packageManagerValue)
		}
	}

	a.Logger.Info("Attempting to update packages",
		"packages", a.PackageNames,
		"packageManager", a.PackageManager)

	// Validate package names
	if len(a.PackageNames) == 0 {
		errMsg := "no package names provided"
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if a.PackageManager == "" {
		errMsg := "unsupported operating system for package management"
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	// Execute package installation based on package manager
	switch a.PackageManager {
	case AptPackageManager:
		return a.installWithApt(execCtx)
	case BrewPackageManager:
		return a.installWithBrew(execCtx)
	default:
		errMsg := fmt.Sprintf("unsupported package manager: %s", a.PackageManager)
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
}

// installWithApt installs packages using apt
func (a *UpdatePackagesAction) installWithApt(execCtx context.Context) error {
	// First update package list
	a.Logger.Info("Updating apt package list")
	_, err := a.CommandRunner.RunCommandWithContext(execCtx, "apt", "update")
	if err != nil {
		a.Logger.Error("Failed to update apt package list", "error", err)
		return fmt.Errorf("failed to update apt package list: %w", err)
	}

	// Build install command with -y flag to avoid prompts
	args := append([]string{"install", "-y"}, a.PackageNames...)
	a.Logger.Info("Installing packages with apt", "packages", a.PackageNames)

	output, err := a.CommandRunner.RunCommandWithContext(execCtx, "apt", args...)
	if err != nil {
		a.Logger.Error("Failed to install packages with apt", "packages", a.PackageNames, "error", err)
		return fmt.Errorf("failed to install packages with apt: %w", err)
	}

	a.Logger.Info("Successfully installed packages with apt",
		"packages", a.PackageNames,
		"output", output)
	return nil
}

// installWithBrew installs packages using Homebrew
func (a *UpdatePackagesAction) installWithBrew(execCtx context.Context) error {
	// Build install command
	args := append([]string{"install"}, a.PackageNames...)
	a.Logger.Info("Installing packages with brew", "packages", a.PackageNames)

	output, err := a.CommandRunner.RunCommandWithContext(execCtx, "brew", args...)
	if err != nil {
		a.Logger.Error("Failed to install packages with brew", "packages", a.PackageNames, "error", err)
		return fmt.Errorf("failed to install packages with brew: %w", err)
	}

	a.Logger.Info("Successfully installed packages with brew",
		"packages", a.PackageNames,
		"output", output)
	return nil
}

// GetOutput returns information about attempted package installation
func (a *UpdatePackagesAction) GetOutput() interface{} {
	return map[string]interface{}{
		"packages":       a.PackageNames,
		"packageManager": string(a.PackageManager),
		"success":        true,
	}
}

// detectPackageManager detects the appropriate package manager based on the operating system
func detectPackageManager() PackageManager {
	switch runtime.GOOS {
	case "linux":
		if isCommandAvailable("apt") {
			return AptPackageManager
		}
	case "darwin":
		if isCommandAvailable("brew") {
			return BrewPackageManager
		}
	}
	return ""
}

// isCommandAvailable checks if a command is available in the system PATH
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
