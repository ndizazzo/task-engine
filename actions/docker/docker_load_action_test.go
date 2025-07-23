package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerLoadAction(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"

	action := NewDockerLoadAction(logger, tarFilePath)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-load--path-to-image.tar-action", action.ID)
	assert.Equal(t, tarFilePath, action.Wrapped.TarFilePath)
	assert.Equal(t, "", action.Wrapped.Platform)
	assert.False(t, action.Wrapped.Quiet)
}

func TestNewDockerLoadActionWithOptions(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"

	action := NewDockerLoadAction(logger, tarFilePath,
		WithPlatform("linux/amd64"),
		WithQuiet(),
	)

	assert.NotNil(t, action)
	assert.Equal(t, "linux/amd64", action.Wrapped.Platform)
	assert.True(t, action.Wrapped.Quiet)
}

func TestDockerLoadAction_Execute_Success(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	expectedOutput := "Loaded image: nginx:latest\nLoaded image: redis:alpine"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger, tarFilePath)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest", "redis:alpine"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerLoadAction_Execute_WithPlatform(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	platform := "linux/amd64"
	expectedOutput := "Loaded image: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath, "--platform", platform).Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger, tarFilePath, WithPlatform(platform))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerLoadAction_Execute_WithQuiet(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	expectedOutput := "Loaded image: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath, "-q").Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger, tarFilePath, WithQuiet())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerLoadAction_Execute_WithPlatformAndQuiet(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	platform := "linux/arm64"
	expectedOutput := "Loaded image: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath, "--platform", platform, "-q").Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger, tarFilePath,
		WithPlatform(platform),
		WithQuiet(),
	)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerLoadAction_Execute_CommandError(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	expectedError := errors.New("file not found")
	expectedOutput := "Error: open /path/to/image.tar: no such file or directory"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return(expectedOutput, expectedError)

	action := NewDockerLoadAction(logger, tarFilePath)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load Docker image from /path/to/image.tar")
	assert.Contains(t, err.Error(), expectedOutput)
	mockRunner.AssertExpectations(t)
}

func TestDockerLoadAction_Execute_ContextCancellation(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return("", context.Canceled)

	action := NewDockerLoadAction(logger, tarFilePath)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load Docker image from /path/to/image.tar")
	mockRunner.AssertExpectations(t)
}

func TestDockerLoadAction_parseLoadedImages(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedImages []string
	}{
		{
			name:           "single image",
			output:         "Loaded image: nginx:latest",
			expectedImages: []string{"nginx:latest"},
		},
		{
			name:           "multiple images",
			output:         "Loaded image: nginx:latest\nLoaded image: redis:alpine\nLoaded image: postgres:13",
			expectedImages: []string{"nginx:latest", "redis:alpine", "postgres:13"},
		},
		{
			name:           "image with ID",
			output:         "Loaded image ID: sha256:abc123def456789",
			expectedImages: []string{"sha256:abc123def456789"},
		},
		{
			name:           "mixed image names and IDs",
			output:         "Loaded image: nginx:latest\nLoaded image ID: sha256:abc123def456789\nLoaded image: redis:alpine",
			expectedImages: []string{"nginx:latest", "sha256:abc123def456789", "redis:alpine"},
		},
		{
			name:           "empty output",
			output:         "",
			expectedImages: []string{},
		},
		{
			name:           "output with extra whitespace",
			output:         "  Loaded image: nginx:latest  \n  Loaded image: redis:alpine  ",
			expectedImages: []string{"nginx:latest", "redis:alpine"},
		},
		{
			name:           "output with unrelated lines",
			output:         "Some other output\nLoaded image: nginx:latest\nMore output\nLoaded image: redis:alpine",
			expectedImages: []string{"nginx:latest", "redis:alpine"},
		},
		{
			name:           "partial matches should be ignored",
			output:         "Loaded image: nginx:latest\nNot loaded image: something\nLoaded image ID: sha256:abc123",
			expectedImages: []string{"nginx:latest", "sha256:abc123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &DockerLoadAction{}
			action.parseLoadedImages(tt.output)
			assert.Equal(t, tt.expectedImages, action.LoadedImages)
		})
	}
}

func TestDockerLoadAction_Execute_EmptyTarFilePath(t *testing.T) {
	logger := slog.Default()
	tarFilePath := ""

	action := NewDockerLoadAction(logger, tarFilePath)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-load--action", action.ID)
}

func TestDockerLoadAction_Execute_SpecialCharactersInPath(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/with/special/chars/image.tar"
	expectedOutput := "Loaded image: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger, tarFilePath)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	mockRunner.AssertExpectations(t)
}

func TestDockerLoadAction_Execute_OutputWithTrailingWhitespace(t *testing.T) {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	rawOutput := "Loaded image: nginx:latest\n  \n  "
	expectedOutput := "Loaded image: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return(rawOutput, nil)

	action := NewDockerLoadAction(logger, tarFilePath)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(t)
}
