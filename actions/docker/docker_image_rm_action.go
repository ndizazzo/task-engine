package docker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerImageRmAction creates a DockerImageRmAction instance
func NewDockerImageRmAction(logger *slog.Logger) *DockerImageRmAction {
	return &DockerImageRmAction{
		BaseAction:       task_engine.NewBaseAction(logger),
		CommandProcessor: command.NewDefaultCommandRunner(),
		RemovedImages:    []string{},
	}
}

// DockerImageRmAction removes Docker images by name/tag or ID
type DockerImageRmAction struct {
	task_engine.BaseAction
	CommandProcessor command.CommandRunner
	RemovedImages    []string // Stores the IDs of removed images for GetOutput
	Output           string   // Stores the command output for GetOutput

	// Parameter-only fields
	ImageNameParam  task_engine.ActionParameter
	ImageIDParam    task_engine.ActionParameter
	RemoveByIDParam task_engine.ActionParameter
	ForceParam      task_engine.ActionParameter
	NoPruneParam    task_engine.ActionParameter
}

// WithParameters sets the parameters and returns a wrapped Action
func (a *DockerImageRmAction) WithParameters(imageNameParam, imageIDParam, removeByIDParam, forceParam, noPruneParam task_engine.ActionParameter) (*task_engine.Action[*DockerImageRmAction], error) {
	if imageNameParam == nil || imageIDParam == nil || removeByIDParam == nil {
		return nil, fmt.Errorf("imageNameParam, imageIDParam, and removeByIDParam cannot be nil")
	}

	a.ImageNameParam = imageNameParam
	a.ImageIDParam = imageIDParam
	a.RemoveByIDParam = removeByIDParam
	a.ForceParam = forceParam
	a.NoPruneParam = noPruneParam

	return &task_engine.Action[*DockerImageRmAction]{
		ID:      "docker-image-rm-action",
		Name:    "Docker Image Remove",
		Wrapped: a,
	}, nil
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerImageRmAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerImageRmAction) Execute(execCtx context.Context) error {
	// Extract GlobalContext from context
	var globalContext *task_engine.GlobalContext
	if gc, ok := execCtx.Value(task_engine.GlobalContextKey).(*task_engine.GlobalContext); ok {
		globalContext = gc
	}

	// Resolve image name parameter
	var imageName string
	if a.ImageNameParam != nil {
		imageNameValue, err := a.ImageNameParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve image name parameter: %w", err)
		}
		if v, ok := imageNameValue.(string); ok {
			imageName = v
		} else {
			return fmt.Errorf("image name parameter is not a string, got %T", imageNameValue)
		}
	}

	// Resolve image ID parameter
	var imageID string
	if a.ImageIDParam != nil {
		imageIDValue, err := a.ImageIDParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve image ID parameter: %w", err)
		}
		if v, ok := imageIDValue.(string); ok {
			imageID = v
		} else {
			return fmt.Errorf("image ID parameter is not a string, got %T", imageIDValue)
		}
	}

	// Resolve removeByID parameter
	var removeByID bool
	if a.RemoveByIDParam != nil {
		removeByIDValue, err := a.RemoveByIDParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve removeByID parameter: %w", err)
		}
		if v, ok := removeByIDValue.(bool); ok {
			removeByID = v
		} else {
			return fmt.Errorf("removeByID parameter is not a bool, got %T", removeByIDValue)
		}
	}

	// Resolve force parameter
	var force bool
	if a.ForceParam != nil {
		forceValue, err := a.ForceParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve force parameter: %w", err)
		}
		if v, ok := forceValue.(bool); ok {
			force = v
		} else {
			return fmt.Errorf("force parameter is not a bool, got %T", forceValue)
		}
	}

	// Resolve noPrune parameter
	var noPrune bool
	if a.NoPruneParam != nil {
		noPruneValue, err := a.NoPruneParam.Resolve(execCtx, globalContext)
		if err != nil {
			return fmt.Errorf("failed to resolve noPrune parameter: %w", err)
		}
		if v, ok := noPruneValue.(bool); ok {
			noPrune = v
		} else {
			return fmt.Errorf("noPrune parameter is not a bool, got %T", noPruneValue)
		}
	}

	args := []string{"image", "rm"}

	// Add force flag if specified
	if force {
		args = append(args, "--force")
	}

	// Add no-prune flag if specified
	if noPrune {
		args = append(args, "--no-prune")
	}

	// Add the image identifier (name/tag or ID)
	var identifier string
	if removeByID {
		args = append(args, imageID)
		identifier = imageID
	} else {
		args = append(args, imageName)
		identifier = imageName
	}

	a.Logger.Info("Executing docker image rm", "identifier", identifier, "force", force, "noPrune", noPrune)
	output, err := a.CommandProcessor.RunCommand("docker", args...)
	a.Output = output

	if err != nil {
		a.Logger.Error("Failed to remove Docker image", "error", err, "output", output)
		return err
	}

	// Parse removed image IDs from output
	a.parseRemovedImages(output)

	a.Logger.Info("Docker image rm finished successfully", "removedImages", a.RemovedImages, "output", a.Output)
	return nil
}

// GetOutput returns information about removed images and raw output
func (a *DockerImageRmAction) GetOutput() interface{} {
	return map[string]interface{}{
		"removed": a.RemovedImages,
		"count":   len(a.RemovedImages),
		"output":  a.Output,
		"success": len(a.RemovedImages) > 0,
	}
}

// parseRemovedImages extracts image IDs from the docker image rm output
// Example output: "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"
func (a *DockerImageRmAction) parseRemovedImages(output string) {
	lines := strings.Split(output, "\n")
	images := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Deleted: ") {
			// Extract image ID from "Deleted: sha256:abc123def456789"
			imageID := strings.TrimPrefix(line, "Deleted: ")
			if imageID != "" {
				images = append(images, imageID)
			}
		} else if strings.HasPrefix(line, "Untagged: ") {
			// Extract image name from "Untagged: nginx:latest"
			imageName := strings.TrimPrefix(line, "Untagged: ")
			if imageName != "" {
				images = append(images, imageName)
			}
		}
	}

	a.RemovedImages = images
}
