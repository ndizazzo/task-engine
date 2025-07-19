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

// NewUpdatePackagesAction creates an action that updates packages using the appropriate package manager
func NewUpdatePackagesAction(packageNames []string, logger *slog.Logger) *task_engine.Action[*UpdatePackagesAction] {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	// Detect package manager based on OS
	packageManager := detectPackageManager()

	id := fmt.Sprintf("update-packages-%s-%s", packageManager, strings.Join(packageNames, "-"))
	return &task_engine.Action[*UpdatePackagesAction]{
		ID: id,
		Wrapped: &UpdatePackagesAction{
			BaseAction:     task_engine.BaseAction{Logger: logger},
			PackageNames:   packageNames,
			PackageManager: packageManager,
			CommandRunner:  command.NewDefaultCommandRunner(),
		},
	}
}

// UpdatePackagesAction updates packages using the appropriate package manager
type UpdatePackagesAction struct {
	task_engine.BaseAction
	PackageNames   []string
	PackageManager PackageManager
	CommandRunner  command.CommandRunner
}

func (a *UpdatePackagesAction) Execute(execCtx context.Context) error {
	a.Logger.Info("Attempting to update packages",
		"packages", a.PackageNames,
		"packageManager", a.PackageManager)

	// Validate package names
	if len(a.PackageNames) == 0 {
		errMsg := "no package names provided"
		a.Logger.Error(errMsg)
		return errors.New(errMsg)
	}

	// Check if package manager is supported
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

// detectPackageManager detects the appropriate package manager based on the operating system
func detectPackageManager() PackageManager {
	switch runtime.GOOS {
	case "linux":
		// Check if apt is available
		if isCommandAvailable("apt") {
			return AptPackageManager
		}
	case "darwin":
		// Check if brew is available
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
