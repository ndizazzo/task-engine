package tasks

import (
	"log/slog"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/docker"
)

func ExampleDockerPullOperations() *task_engine.Task {
	logger := slog.Default()

	images := map[string]docker.ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
		"alpine": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "arm64",
		},
		"redis": {
			Image:        "redis",
			Tag:          "7-alpine",
			Architecture: "amd64",
		},
		"postgres": {
			Image:        "postgres",
			Tag:          "15",
			Architecture: "amd64",
		},
	}

	pullAction := docker.NewDockerPullAction(
		logger,
		images,
		docker.WithPullQuietOutput(),
		docker.WithPullPlatform("linux/amd64"),
	)

	task := &task_engine.Task{
		ID:   "docker-pull-example",
		Name: "Pull multiple Docker images with different specifications",
		Actions: []task_engine.ActionWrapper{
			pullAction,
		},
		Logger: logger,
	}

	return task
}

func ExampleDockerPullOperationsWithErrorHandling() *task_engine.Task {
	logger := slog.Default()

	images := map[string]docker.ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
		"nonexistent": {
			Image:        "nonexistent",
			Tag:          "latest",
			Architecture: "amd64",
		},
		"alpine": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "arm64",
		},
	}

	pullAction := docker.NewDockerPullAction(logger, images)

	task := &task_engine.Task{
		ID:   "docker-pull-with-error-handling",
		Name: "Pull Docker images with error handling for partial failures",
		Actions: []task_engine.ActionWrapper{
			pullAction,
		},
		Logger: logger,
	}

	return task
}

func ExampleDockerPullOperationsForMultiArch() *task_engine.Task {
	logger := slog.Default()

	images := map[string]docker.ImageSpec{
		"nginx-amd64": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
		"nginx-arm64": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "arm64",
		},
		"alpine-amd64": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "amd64",
		},
		"alpine-arm64": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "arm64",
		},
	}

	pullAction := docker.NewDockerPullAction(logger, images)

	task := &task_engine.Task{
		ID:   "docker-pull-multi-arch",
		Name: "Pull Docker images for multiple architectures",
		Actions: []task_engine.ActionWrapper{
			pullAction,
		},
		Logger: logger,
	}

	return task
}

func ExampleDockerPullOperationsWithCustomPlatform() *task_engine.Task {
	logger := slog.Default()

	images := map[string]docker.ImageSpec{
		"nginx-linux": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "linux/amd64",
		},
		"alpine-linux": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "linux/arm64",
		},
		"redis-linux": {
			Image:        "redis",
			Tag:          "7-alpine",
			Architecture: "linux/amd64",
		},
	}

	pullAction := docker.NewDockerPullAction(
		logger,
		images,
		docker.WithPullPlatform("linux/amd64"),
	)

	task := &task_engine.Task{
		ID:   "docker-pull-custom-platform",
		Name: "Pull Docker images with custom platform specifications",
		Actions: []task_engine.ActionWrapper{
			pullAction,
		},
		Logger: logger,
	}

	return task
}

func ExampleDockerPullOperationsMinimal() *task_engine.Task {
	logger := slog.Default()

	images := map[string]docker.ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	pullAction := docker.NewDockerPullAction(logger, images)

	task := &task_engine.Task{
		ID:   "docker-pull-minimal",
		Name: "Pull a single Docker image",
		Actions: []task_engine.ActionWrapper{
			pullAction,
		},
		Logger: logger,
	}

	return task
}

func ExampleDockerPullMultiArchOperations() *task_engine.Task {
	logger := slog.Default()

	multiArchImages := map[string]docker.MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64", "arm/v7"},
		},
		"alpine": {
			Image:         "alpine",
			Tag:           "3.18",
			Architectures: []string{"amd64", "arm64"},
		},
		"redis": {
			Image:         "redis",
			Tag:           "7-alpine",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	pullAction := docker.NewDockerPullMultiArchAction(logger, multiArchImages)

	task := &task_engine.Task{
		ID:   "docker-pull-multiarch",
		Name: "Pull Docker images for multiple architectures using MultiArchImageSpec",
		Actions: []task_engine.ActionWrapper{
			pullAction,
		},
		Logger: logger,
	}

	return task
}

func ExampleDockerPullMultiArchOperationsWithOptions() *task_engine.Task {
	logger := slog.Default()

	multiArchImages := map[string]docker.MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64"},
		},
		"postgres": {
			Image:         "postgres",
			Tag:           "15",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	pullAction := docker.NewDockerPullMultiArchAction(
		logger,
		multiArchImages,
		docker.WithPullQuietOutput(),
		docker.WithPullPlatform("linux/amd64"),
	)

	task := &task_engine.Task{
		ID:   "docker-pull-multiarch-with-options",
		Name: "Pull multi-architecture Docker images with quiet output and platform override",
		Actions: []task_engine.ActionWrapper{
			pullAction,
		},
		Logger: logger,
	}

	return task
}

func ExampleDockerPullMixedOperations() *task_engine.Task {
	logger := slog.Default()

	images := map[string]docker.ImageSpec{
		"alpine": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "",
		},
		"redis": {
			Image:        "redis",
			Tag:          "7-alpine",
			Architecture: "amd64",
		},
	}

	multiArchImages := map[string]docker.MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64"},
		},
		"postgres": {
			Image:         "postgres",
			Tag:           "15",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	pullAction := docker.NewDockerPullAction(logger, images)
	pullAction.Wrapped.MultiArchImages = multiArchImages

	task := &task_engine.Task{
		ID:   "docker-pull-mixed",
		Name: "Pull Docker images using both single and multi-architecture specifications",
		Actions: []task_engine.ActionWrapper{
			pullAction,
		},
		Logger: logger,
	}

	return task
}
