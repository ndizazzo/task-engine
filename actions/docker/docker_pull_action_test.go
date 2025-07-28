package docker

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerPullAction(t *testing.T) {
	logger := slog.Default()
	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images)
	assert.NotNil(t, action)
	assert.Equal(t, "docker-pull-action", action.ID)
	assert.NotNil(t, action.Wrapped)
}

func TestNewDockerPullActionWithOptions(t *testing.T) {
	logger := slog.Default()
	images := map[string]ImageSpec{
		"alpine": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "arm64",
		},
	}

	action := NewDockerPullAction(logger, images, WithPullQuietOutput(), WithPullPlatform("linux/amd64"))
	assert.NotNil(t, action)
	assert.True(t, action.Wrapped.Quiet)
	assert.Equal(t, "linux/amd64", action.Wrapped.Platform)
}

func TestNewDockerPullMultiArchAction(t *testing.T) {
	logger := slog.Default()
	multiArchImages := map[string]MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	action := NewDockerPullMultiArchAction(logger, multiArchImages)
	assert.NotNil(t, action)
	assert.Equal(t, "docker-pull-multiarch-action", action.ID)
	assert.NotNil(t, action.Wrapped)
	assert.Len(t, action.Wrapped.MultiArchImages, 1)
}

func TestDockerPullAction_Execute_Success(t *testing.T) {
	logger := slog.Default()
	expectedOutput := "nginx:latest: Pulling from library/nginx\nDigest: sha256:...\nStatus: Downloaded newer image for nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nginx:latest").Return(expectedOutput, nil)

	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Len(t, action.Wrapped.PulledImages, 1)
	assert.Equal(t, "nginx", action.Wrapped.PulledImages[0])
	assert.Len(t, action.Wrapped.FailedImages, 0)
	assert.Contains(t, action.Wrapped.Output, "Pulled 1 images, failed 0 images")

	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_SuccessMultipleImages(t *testing.T) {
	logger := slog.Default()
	expectedOutput := "Image pulled successfully"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nginx:latest").Return(expectedOutput, nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "arm64", "alpine:3.18").Return(expectedOutput, nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "redis:7-alpine").Return(expectedOutput, nil)

	images := map[string]ImageSpec{
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
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Len(t, action.Wrapped.PulledImages, 3)
	assert.Len(t, action.Wrapped.FailedImages, 0)
	assert.Contains(t, action.Wrapped.Output, "Pulled 3 images, failed 0 images")

	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_MultiArchSuccess(t *testing.T) {
	logger := slog.Default()
	expectedOutput := "Image pulled successfully"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nginx:latest").Return(expectedOutput, nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "arm64", "nginx:latest").Return(expectedOutput, nil)

	multiArchImages := map[string]MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	action := NewDockerPullMultiArchAction(logger, multiArchImages)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Len(t, action.Wrapped.PulledImages, 1)
	assert.Equal(t, "nginx", action.Wrapped.PulledImages[0])
	assert.Len(t, action.Wrapped.FailedImages, 0)
	assert.Contains(t, action.Wrapped.Output, "Pulled 1 images, failed 0 images")

	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_MultiArchPartialFailure(t *testing.T) {
	logger := slog.Default()
	expectedOutput := "Image pulled successfully"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nginx:latest").Return(expectedOutput, nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "arm64", "nginx:latest").Return("", assert.AnError)

	multiArchImages := map[string]MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	action := NewDockerPullMultiArchAction(logger, multiArchImages)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err) // Should succeed because at least one architecture was pulled
	assert.Len(t, action.Wrapped.PulledImages, 1)
	assert.Equal(t, "nginx", action.Wrapped.PulledImages[0])
	assert.Len(t, action.Wrapped.FailedImages, 0)

	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_MultiArchCompleteFailure(t *testing.T) {
	logger := slog.Default()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nginx:latest").Return("", assert.AnError)
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "arm64", "nginx:latest").Return("", assert.AnError)

	multiArchImages := map[string]MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	action := NewDockerPullMultiArchAction(logger, multiArchImages)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Len(t, action.Wrapped.PulledImages, 0)
	assert.Len(t, action.Wrapped.FailedImages, 1)
	assert.Equal(t, "nginx", action.Wrapped.FailedImages[0])

	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_MixedImages(t *testing.T) {
	logger := slog.Default()
	expectedOutput := "Image pulled successfully"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "alpine:3.18").Return(expectedOutput, nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nginx:latest").Return(expectedOutput, nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "arm64", "nginx:latest").Return(expectedOutput, nil)

	images := map[string]ImageSpec{
		"alpine": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "",
		},
	}

	multiArchImages := map[string]MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.MultiArchImages = multiArchImages
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Len(t, action.Wrapped.PulledImages, 2)
	assert.Len(t, action.Wrapped.FailedImages, 0)
	assert.Contains(t, action.Wrapped.Output, "Pulled 2 images, failed 0 images")

	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_Failure(t *testing.T) {
	logger := slog.Default()
	expectedError := "Error response from daemon: manifest for nonexistent:latest not found"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nonexistent:latest").Return(expectedError, assert.AnError)

	images := map[string]ImageSpec{
		"nonexistent": {
			Image:        "nonexistent",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Len(t, action.Wrapped.PulledImages, 0)
	assert.Len(t, action.Wrapped.FailedImages, 1)
	assert.Equal(t, "nonexistent", action.Wrapped.FailedImages[0])
	assert.Contains(t, action.Wrapped.Output, "Pulled 0 images, failed 1 images")

	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_PartialFailure(t *testing.T) {
	logger := slog.Default()
	successOutput := "nginx:latest: Pulling from library/nginx\nStatus: Downloaded newer image for nginx:latest"
	errorOutput := "Error response from daemon: manifest for nonexistent:latest not found"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nginx:latest").Return(successOutput, nil)
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "amd64", "nonexistent:latest").Return(errorOutput, assert.AnError)

	images := map[string]ImageSpec{
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
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Len(t, action.Wrapped.PulledImages, 1)
	assert.Len(t, action.Wrapped.FailedImages, 1)
	assert.Equal(t, "nginx", action.Wrapped.PulledImages[0])
	assert.Equal(t, "nonexistent", action.Wrapped.FailedImages[0])
	assert.Contains(t, action.Wrapped.Output, "Pulled 1 images, failed 1 images")

	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_EmptyImages(t *testing.T) {
	logger := slog.Default()
	images := map[string]ImageSpec{}

	action := NewDockerPullAction(logger, images)
	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no images specified for pulling")
}

func TestDockerPullAction_Execute_EmptyMultiArchImages(t *testing.T) {
	logger := slog.Default()
	multiArchImages := map[string]MultiArchImageSpec{}

	action := NewDockerPullMultiArchAction(logger, multiArchImages)
	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no images specified for pulling")
}

func TestDockerPullAction_Execute_WithQuietOption(t *testing.T) {
	logger := slog.Default()
	expectedOutput := "Image pulled successfully"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--quiet", "--platform", "amd64", "nginx:latest").Return(expectedOutput, nil)

	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images, WithPullQuietOutput())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_WithPlatformOption(t *testing.T) {
	logger := slog.Default()
	expectedOutput := "Image pulled successfully"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "linux/amd64", "nginx:latest").Return(expectedOutput, nil)

	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images, WithPullPlatform("linux/amd64"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_Execute_WithArchitectureOverride(t *testing.T) {
	logger := slog.Default()
	expectedOutput := "Image pulled successfully"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", context.Background(), "docker", "pull", "--platform", "arm64", "nginx:latest").Return(expectedOutput, nil)

	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "arm64",
		},
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	mockRunner.AssertExpectations(t)
}

func TestDockerPullAction_BuildImageReference(t *testing.T) {
	logger := slog.Default()
	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
		"alpine": {
			Image:        "alpine",
			Tag:          "",
			Architecture: "arm64",
		},
	}

	action := NewDockerPullAction(logger, images)

	ref1 := action.Wrapped.buildImageReference(images["nginx"])
	assert.Equal(t, "nginx:latest", ref1)

	ref2 := action.Wrapped.buildImageReference(images["alpine"])
	assert.Equal(t, "alpine", ref2)
}

func TestDockerPullAction_GetPulledImages(t *testing.T) {
	logger := slog.Default()
	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.PulledImages = []string{"nginx", "alpine"}

	result := action.Wrapped.GetPulledImages()
	assert.Equal(t, []string{"nginx", "alpine"}, result)
}

func TestDockerPullAction_GetFailedImages(t *testing.T) {
	logger := slog.Default()
	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.FailedImages = []string{"nonexistent"}

	result := action.Wrapped.GetFailedImages()
	assert.Equal(t, []string{"nonexistent"}, result)
}

func TestDockerPullAction_GetOutput(t *testing.T) {
	logger := slog.Default()
	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.Output = "Test output"

	result := action.Wrapped.GetOutput()
	assert.Equal(t, "Test output", result)
}

func TestDockerPullAction_Execute_ContextCancellation(t *testing.T) {
	logger := slog.Default()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommandWithContext", ctx, "docker", "pull", "--platform", "amd64", "nginx:latest").Return("", context.Canceled)

	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(ctx)

	assert.Error(t, err)
	assert.Len(t, action.Wrapped.FailedImages, 1)

	mockRunner.AssertExpectations(t)
}
