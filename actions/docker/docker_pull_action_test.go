package docker

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// DockerPullActionTestSuite tests the DockerPullAction functionality
type DockerPullActionTestSuite struct {
	suite.Suite
}

// TestDockerPullActionTestSuite runs the DockerPullAction test suite
func TestDockerPullActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerPullActionTestSuite))
}

func (suite *DockerPullActionTestSuite) TestNewDockerPullAction() {
	logger := slog.Default()
	images := map[string]ImageSpec{
		"nginx": {
			Image:        "nginx",
			Tag:          "latest",
			Architecture: "amd64",
		},
	}

	action := NewDockerPullAction(logger, images)
	assert.NotNil(suite.T(), action)
	assert.Equal(suite.T(), "docker-pull-action", action.ID)
	assert.NotNil(suite.T(), action.Wrapped)
}

func (suite *DockerPullActionTestSuite) TestNewDockerPullActionWithOptions() {
	logger := slog.Default()
	images := map[string]ImageSpec{
		"alpine": {
			Image:        "alpine",
			Tag:          "3.18",
			Architecture: "arm64",
		},
	}

	action := NewDockerPullAction(logger, images, WithPullQuietOutput(), WithPullPlatform("linux/amd64"))
	assert.NotNil(suite.T(), action)
	assert.True(suite.T(), action.Wrapped.Quiet)
	assert.Equal(suite.T(), "linux/amd64", action.Wrapped.Platform)
}

func (suite *DockerPullActionTestSuite) TestNewDockerPullMultiArchAction() {
	logger := slog.Default()
	multiArchImages := map[string]MultiArchImageSpec{
		"nginx": {
			Image:         "nginx",
			Tag:           "latest",
			Architectures: []string{"amd64", "arm64"},
		},
	}

	action := NewDockerPullMultiArchAction(logger, multiArchImages)
	assert.NotNil(suite.T(), action)
	assert.Equal(suite.T(), "docker-pull-multiarch-action", action.ID)
	assert.NotNil(suite.T(), action.Wrapped)
	assert.Len(suite.T(), action.Wrapped.MultiArchImages, 1)
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_Success() {
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

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), action.Wrapped.PulledImages, 1)
	assert.Equal(suite.T(), "nginx", action.Wrapped.PulledImages[0])
	assert.Len(suite.T(), action.Wrapped.FailedImages, 0)
	assert.Contains(suite.T(), action.Wrapped.Output, "Pulled 1 images, failed 0 images")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_SuccessMultipleImages() {
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

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), action.Wrapped.PulledImages, 3)
	assert.Contains(suite.T(), action.Wrapped.PulledImages, "nginx")
	assert.Contains(suite.T(), action.Wrapped.PulledImages, "alpine")
	assert.Contains(suite.T(), action.Wrapped.PulledImages, "redis")
	assert.Len(suite.T(), action.Wrapped.FailedImages, 0)
	assert.Contains(suite.T(), action.Wrapped.Output, "Pulled 3 images, failed 0 images")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_MultiArchSuccess() {
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

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), action.Wrapped.PulledImages, 1)
	assert.Equal(suite.T(), "nginx", action.Wrapped.PulledImages[0])
	assert.Len(suite.T(), action.Wrapped.FailedImages, 0)
	assert.Contains(suite.T(), action.Wrapped.Output, "Pulled 1 images, failed 0 images")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_MultiArchPartialFailure() {
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

	assert.NoError(suite.T(), err) // Should succeed because at least one architecture was pulled
	assert.Len(suite.T(), action.Wrapped.PulledImages, 1)
	assert.Equal(suite.T(), "nginx", action.Wrapped.PulledImages[0])
	assert.Len(suite.T(), action.Wrapped.FailedImages, 0)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_MultiArchCompleteFailure() {
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

	assert.Error(suite.T(), err)
	assert.Len(suite.T(), action.Wrapped.PulledImages, 0)
	assert.Len(suite.T(), action.Wrapped.FailedImages, 1)
	assert.Equal(suite.T(), "nginx", action.Wrapped.FailedImages[0])

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_MixedImages() {
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

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), action.Wrapped.PulledImages, 2)
	assert.Len(suite.T(), action.Wrapped.FailedImages, 0)
	assert.Contains(suite.T(), action.Wrapped.Output, "Pulled 2 images, failed 0 images")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_Failure() {
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

	assert.Error(suite.T(), err)
	assert.Len(suite.T(), action.Wrapped.PulledImages, 0)
	assert.Len(suite.T(), action.Wrapped.FailedImages, 1)
	assert.Equal(suite.T(), "nonexistent", action.Wrapped.FailedImages[0])
	assert.Contains(suite.T(), action.Wrapped.Output, "Pulled 0 images, failed 1 images")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_PartialFailure() {
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

	assert.Error(suite.T(), err)
	assert.Len(suite.T(), action.Wrapped.PulledImages, 1)
	assert.Len(suite.T(), action.Wrapped.FailedImages, 1)
	assert.Equal(suite.T(), "nginx", action.Wrapped.PulledImages[0])
	assert.Equal(suite.T(), "nonexistent", action.Wrapped.FailedImages[0])
	assert.Contains(suite.T(), action.Wrapped.Output, "Pulled 1 images, failed 1 images")

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_EmptyImages() {
	logger := slog.Default()
	images := map[string]ImageSpec{}

	action := NewDockerPullAction(logger, images)
	err := action.Wrapped.Execute(context.Background())

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no images specified for pulling")
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_EmptyMultiArchImages() {
	logger := slog.Default()
	multiArchImages := map[string]MultiArchImageSpec{}

	action := NewDockerPullMultiArchAction(logger, multiArchImages)
	err := action.Wrapped.Execute(context.Background())

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no images specified for pulling")
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_WithQuietOption() {
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

	assert.NoError(suite.T(), err)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_WithPlatformOption() {
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

	assert.NoError(suite.T(), err)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_WithArchitectureOverride() {
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

	assert.NoError(suite.T(), err)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_BuildImageReference() {
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
	assert.Equal(suite.T(), "nginx:latest", ref1)

	ref2 := action.Wrapped.buildImageReference(images["alpine"])
	assert.Equal(suite.T(), "alpine", ref2)
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_GetPulledImages() {
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
	assert.Equal(suite.T(), []string{"nginx", "alpine"}, result)
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_GetFailedImages() {
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
	assert.Equal(suite.T(), []string{"nonexistent"}, result)
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_GetOutput() {
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
	assert.Equal(suite.T(), "Test output", result)
}

func (suite *DockerPullActionTestSuite) TestDockerPullAction_Execute_ContextCancellation() {
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

	assert.Error(suite.T(), err)
	assert.Len(suite.T(), action.Wrapped.FailedImages, 1)

	mockRunner.AssertExpectations(suite.T())
}
