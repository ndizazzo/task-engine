package tasks

import (
	"context"
	"log/slog"
	"os"

	"github.com/ndizazzo/task-engine/actions/system"
)

// ExampleUpdatePackages demonstrates how to use the UpdatePackagesAction
func ExampleUpdatePackages() {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create an action to install multiple packages
	packageNames := []string{"curl", "wget", "git"}
	action := system.NewUpdatePackagesAction(packageNames, logger)

	// Execute the action
	ctx := context.Background()
	err := action.Execute(ctx)
	if err != nil {
		logger.Error("Failed to update packages", "error", err)
		return
	}

	logger.Info("Successfully updated packages", "packages", packageNames)
}

// ExampleUpdateSinglePackage demonstrates installing a single package
func ExampleUpdateSinglePackage() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Install a single package
	packageNames := []string{"htop"}
	action := system.NewUpdatePackagesAction(packageNames, logger)

	ctx := context.Background()
	err := action.Execute(ctx)
	if err != nil {
		logger.Error("Failed to install package", "package", packageNames[0], "error", err)
		return
	}

	logger.Info("Successfully installed package", "package", packageNames[0])
}
