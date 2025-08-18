package docker

import (
	"context"
	"log/slog"
	"strings"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/command"
)

// NewDockerImageRmAction creates a new DockerImageRmAction with the given logger
func NewDockerImageRmAction(logger *slog.Logger) *DockerImageRmAction {
	return &DockerImageRmAction{
		BaseAction:        task_engine.NewBaseAction(logger),
		ParameterResolver: *common.NewParameterResolver(logger),
		OutputBuilder:     *common.NewOutputBuilder(logger),
		CommandProcessor:  command.NewDefaultCommandRunner(),
	}
}

// DockerImageRmAction removes Docker images by name/tag or ID
type DockerImageRmAction struct {
	task_engine.BaseAction
	common.ParameterResolver
	common.OutputBuilder
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

// WithParameters sets the parameters for image removal and returns a wrapped Action
func (a *DockerImageRmAction) WithParameters(
	imageNameParam task_engine.ActionParameter,
	imageIDParam task_engine.ActionParameter,
	removeByIDParam task_engine.ActionParameter,
	forceParam task_engine.ActionParameter,
	noPruneParam task_engine.ActionParameter,
) (*task_engine.Action[*DockerImageRmAction], error) {
	a.ImageNameParam = imageNameParam
	a.ImageIDParam = imageIDParam
	a.RemoveByIDParam = removeByIDParam
	a.ForceParam = forceParam
	a.NoPruneParam = noPruneParam

	// Create a temporary constructor to use the base functionality
	constructor := common.NewBaseConstructor[*DockerImageRmAction](a.Logger)
	return constructor.WrapAction(a, "Docker Image RM", "docker-image-rm-action"), nil
}

// SetCommandRunner allows injecting a mock or alternative CommandRunner for testing
func (a *DockerImageRmAction) SetCommandRunner(runner command.CommandRunner) {
	a.CommandProcessor = runner
}

func (a *DockerImageRmAction) Execute(execCtx context.Context) error {
	// Resolve image name parameter
	var imageName string
	if a.ImageNameParam != nil {
		imageNameValue, err := a.ResolveStringParameter(execCtx, a.ImageNameParam, "image name")
		if err != nil {
			return err
		}
		imageName = imageNameValue
	}

	// Resolve image ID parameter
	var imageID string
	if a.ImageIDParam != nil {
		imageIDValue, err := a.ResolveStringParameter(execCtx, a.ImageIDParam, "image ID")
		if err != nil {
			return err
		}
		imageID = imageIDValue
	}

	// Resolve removeByID parameter
	var removeByID bool
	if a.RemoveByIDParam != nil {
		removeByIDValue, err := a.ResolveBoolParameter(execCtx, a.RemoveByIDParam, "removeByID")
		if err != nil {
			return err
		}
		removeByID = removeByIDValue
	}

	// Resolve force parameter
	var force bool
	if a.ForceParam != nil {
		forceValue, err := a.ResolveBoolParameter(execCtx, a.ForceParam, "force")
		if err != nil {
			return err
		}
		force = forceValue
	}

	// Resolve noPrune parameter
	var noPrune bool
	if a.NoPruneParam != nil {
		noPruneValue, err := a.ResolveBoolParameter(execCtx, a.NoPruneParam, "noPrune")
		if err != nil {
			return err
		}
		noPrune = noPruneValue
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
	return a.BuildOutputWithCount(a.RemovedImages, len(a.RemovedImages) > 0, map[string]interface{}{
		"removed": a.RemovedImages,
		"output":  a.Output,
	})
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
