package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerImageRmByNameAction(t *testing.T) {
	logger := slog.Default()
	imageName := "nginx:latest"

	action := NewDockerImageRmByNameAction(logger, imageName)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-image-rm-nginx:latest-action", action.ID)
	assert.Equal(t, imageName, action.Wrapped.ImageName)
	assert.Equal(t, "", action.Wrapped.ImageID)
	assert.False(t, action.Wrapped.RemoveByID)
	assert.False(t, action.Wrapped.Force)
	assert.False(t, action.Wrapped.NoPrune)
}

func TestNewDockerImageRmByIDAction(t *testing.T) {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"

	action := NewDockerImageRmByIDAction(logger, imageID)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-image-rm-id-sha256:abc123def456789-action", action.ID)
	assert.Equal(t, "", action.Wrapped.ImageName)
	assert.Equal(t, imageID, action.Wrapped.ImageID)
	assert.True(t, action.Wrapped.RemoveByID)
	assert.False(t, action.Wrapped.Force)
	assert.False(t, action.Wrapped.NoPrune)
}

func TestNewDockerImageRmByNameActionWithOptions(t *testing.T) {
	logger := slog.Default()
	imageName := "nginx:latest"

	action := NewDockerImageRmByNameAction(logger, imageName,
		WithForce(),
		WithNoPrune(),
	)

	assert.NotNil(t, action)
	assert.Equal(t, imageName, action.Wrapped.ImageName)
	assert.True(t, action.Wrapped.Force)
	assert.True(t, action.Wrapped.NoPrune)
}

func TestNewDockerImageRmByIDActionWithOptions(t *testing.T) {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"

	action := NewDockerImageRmByIDAction(logger, imageID,
		WithForce(),
		WithNoPrune(),
	)

	assert.NotNil(t, action)
	assert.Equal(t, imageID, action.Wrapped.ImageID)
	assert.True(t, action.Wrapped.RemoveByID)
	assert.True(t, action.Wrapped.Force)
	assert.True(t, action.Wrapped.NoPrune)
}

func TestDockerImageRmAction_Execute_ByName_Success(t *testing.T) {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_Execute_ByID_Success(t *testing.T) {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"
	expectedOutput := "Deleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageID).Return(expectedOutput, nil)

	action := NewDockerImageRmByIDAction(logger, imageID)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_Execute_WithForce(t *testing.T) {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "-f", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName, WithForce())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_Execute_WithNoPrune(t *testing.T) {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"
	expectedOutput := "Deleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--no-prune", imageID).Return(expectedOutput, nil)

	action := NewDockerImageRmByIDAction(logger, imageID, WithNoPrune())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_Execute_WithForceAndNoPrune(t *testing.T) {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "-f", "--no-prune", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName,
		WithForce(),
		WithNoPrune(),
	)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_Execute_CommandError(t *testing.T) {
	logger := slog.Default()
	imageName := "nonexistent:latest"
	expectedError := errors.New("image not found")
	expectedOutput := "Error: No such image: nonexistent:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, expectedError)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove Docker image nonexistent:latest")
	assert.Contains(t, err.Error(), expectedOutput)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_Execute_ContextCancellation(t *testing.T) {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageID).Return("", context.Canceled)

	action := NewDockerImageRmByIDAction(logger, imageID)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove Docker image sha256:abc123def456789")
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_parseRemovedImages(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedImages []string
	}{
		{
			name:           "untagged only",
			output:         "Untagged: nginx:latest",
			expectedImages: []string{"nginx:latest"},
		},
		{
			name:           "deleted only",
			output:         "Deleted: sha256:abc123def456789",
			expectedImages: []string{"sha256:abc123def456789"},
		},
		{
			name:           "untagged and deleted",
			output:         "Untagged: nginx:latest\nDeleted: sha256:abc123def456789",
			expectedImages: []string{"nginx:latest", "sha256:abc123def456789"},
		},
		{
			name:           "multiple untagged",
			output:         "Untagged: nginx:latest\nUntagged: nginx:alpine\nDeleted: sha256:abc123def456789",
			expectedImages: []string{"nginx:latest", "nginx:alpine", "sha256:abc123def456789"},
		},
		{
			name:           "empty output",
			output:         "",
			expectedImages: []string{},
		},
		{
			name:           "output with extra whitespace",
			output:         "  Untagged: nginx:latest  \n  Deleted: sha256:abc123def456789  ",
			expectedImages: []string{"nginx:latest", "sha256:abc123def456789"},
		},
		{
			name:           "output with unrelated lines",
			output:         "Some other output\nUntagged: nginx:latest\nMore output\nDeleted: sha256:abc123def456789",
			expectedImages: []string{"nginx:latest", "sha256:abc123def456789"},
		},
		{
			name:           "partial matches should be ignored",
			output:         "Untagged: nginx:latest\nNot untagged: something\nDeleted: sha256:abc123",
			expectedImages: []string{"nginx:latest", "sha256:abc123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &DockerImageRmAction{}
			action.parseRemovedImages(tt.output)
			assert.Equal(t, tt.expectedImages, action.RemovedImages)
		})
	}
}

func TestDockerImageRmAction_Execute_EmptyImageName(t *testing.T) {
	logger := slog.Default()
	imageName := ""

	action := NewDockerImageRmByNameAction(logger, imageName)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-image-rm--action", action.ID)
}

func TestDockerImageRmAction_Execute_EmptyImageID(t *testing.T) {
	logger := slog.Default()
	imageID := ""

	action := NewDockerImageRmByIDAction(logger, imageID)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-image-rm-id--action", action.ID)
}

func TestDockerImageRmAction_Execute_SpecialCharactersInName(t *testing.T) {
	logger := slog.Default()
	imageName := "my-registry.com/namespace/image:latest"
	expectedOutput := "Untagged: my-registry.com/namespace/image:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"my-registry.com/namespace/image:latest"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_Execute_OutputWithTrailingWhitespace(t *testing.T) {
	logger := slog.Default()
	imageName := "nginx:latest"
	rawOutput := "Untagged: nginx:latest\n  \n  "
	expectedOutput := "Untagged: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(rawOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Equal(t, []string{"nginx:latest"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageRmAction_Execute_VariousTagForms(t *testing.T) {
	tests := []struct {
		name           string
		imageName      string
		expectedOutput string
		expectedImages []string
	}{
		{
			name:           "latest tag",
			imageName:      "nginx:latest",
			expectedOutput: "Untagged: nginx:latest\nDeleted: sha256:abc123def456789",
			expectedImages: []string{"nginx:latest", "sha256:abc123def456789"},
		},
		{
			name:           "specific version tag",
			imageName:      "nginx:1.21.0",
			expectedOutput: "Untagged: nginx:1.21.0\nDeleted: sha256:def456ghi789012",
			expectedImages: []string{"nginx:1.21.0", "sha256:def456ghi789012"},
		},
		{
			name:           "alpine tag",
			imageName:      "redis:alpine",
			expectedOutput: "Untagged: redis:alpine\nDeleted: sha256:ghi789jkl012345",
			expectedImages: []string{"redis:alpine", "sha256:ghi789jkl012345"},
		},
		{
			name:           "semantic version tag",
			imageName:      "postgres:13.4",
			expectedOutput: "Untagged: postgres:13.4\nDeleted: sha256:jkl012mno345678",
			expectedImages: []string{"postgres:13.4", "sha256:jkl012mno345678"},
		},
		{
			name:           "beta tag",
			imageName:      "node:16-beta",
			expectedOutput: "Untagged: node:16-beta\nDeleted: sha256:mno345pqr678901",
			expectedImages: []string{"node:16-beta", "sha256:mno345pqr678901"},
		},
		{
			name:           "rc tag",
			imageName:      "python:3.9-rc",
			expectedOutput: "Untagged: python:3.9-rc\nDeleted: sha256:pqr678stu901234",
			expectedImages: []string{"python:3.9-rc", "sha256:pqr678stu901234"},
		},
		{
			name:           "date tag",
			imageName:      "ubuntu:2023-01-15",
			expectedOutput: "Untagged: ubuntu:2023-01-15\nDeleted: sha256:stu901vwx234567",
			expectedImages: []string{"ubuntu:2023-01-15", "sha256:stu901vwx234567"},
		},
		{
			name:           "hash tag",
			imageName:      "golang:1.19.4-bullseye",
			expectedOutput: "Untagged: golang:1.19.4-bullseye\nDeleted: sha256:vwx234yza567890",
			expectedImages: []string{"golang:1.19.4-bullseye", "sha256:vwx234yza567890"},
		},
		{
			name:           "no tag (defaults to latest)",
			imageName:      "nginx",
			expectedOutput: "Untagged: nginx:latest\nDeleted: sha256:abc123def456789",
			expectedImages: []string{"nginx:latest", "sha256:abc123def456789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", tt.imageName).Return(tt.expectedOutput, nil)

			action := NewDockerImageRmByNameAction(logger, tt.imageName)
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Equal(t, tt.expectedImages, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}

func TestDockerImageRmAction_Execute_RegistryImagesWithTags(t *testing.T) {
	tests := []struct {
		name           string
		imageName      string
		expectedOutput string
		expectedImages []string
	}{
		{
			name:           "docker hub with tag",
			imageName:      "library/nginx:latest",
			expectedOutput: "Untagged: library/nginx:latest\nDeleted: sha256:abc123def456789",
			expectedImages: []string{"library/nginx:latest", "sha256:abc123def456789"},
		},
		{
			name:           "private registry with version tag",
			imageName:      "my-registry.com/namespace/app:v1.2.3",
			expectedOutput: "Untagged: my-registry.com/namespace/app:v1.2.3\nDeleted: sha256:def456ghi789012",
			expectedImages: []string{"my-registry.com/namespace/app:v1.2.3", "sha256:def456ghi789012"},
		},
		{
			name:           "private registry with latest tag",
			imageName:      "registry.example.com/project/service:latest",
			expectedOutput: "Untagged: registry.example.com/project/service:latest\nDeleted: sha256:ghi789jkl012345",
			expectedImages: []string{"registry.example.com/project/service:latest", "sha256:ghi789jkl012345"},
		},
		{
			name:           "private registry with beta tag",
			imageName:      "internal.registry.com/team/product:beta-2023-12-01",
			expectedOutput: "Untagged: internal.registry.com/team/product:beta-2023-12-01\nDeleted: sha256:jkl012mno345678",
			expectedImages: []string{"internal.registry.com/team/product:beta-2023-12-01", "sha256:jkl012mno345678"},
		},
		{
			name:           "AWS ECR with tag",
			imageName:      "123456789012.dkr.ecr.us-west-2.amazonaws.com/my-app:prod",
			expectedOutput: "Untagged: 123456789012.dkr.ecr.us-west-2.amazonaws.com/my-app:prod\nDeleted: sha256:mno345pqr678901",
			expectedImages: []string{"123456789012.dkr.ecr.us-west-2.amazonaws.com/my-app:prod", "sha256:mno345pqr678901"},
		},
		{
			name:           "Google GCR with tag",
			imageName:      "gcr.io/my-project/my-service:v2.1.0",
			expectedOutput: "Untagged: gcr.io/my-project/my-service:v2.1.0\nDeleted: sha256:pqr678stu901234",
			expectedImages: []string{"gcr.io/my-project/my-service:v2.1.0", "sha256:pqr678stu901234"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", tt.imageName).Return(tt.expectedOutput, nil)

			action := NewDockerImageRmByNameAction(logger, tt.imageName)
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Equal(t, tt.expectedImages, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}

func TestDockerImageRmAction_Execute_EdgeCaseTags(t *testing.T) {
	tests := []struct {
		name           string
		imageName      string
		expectedOutput string
		expectedImages []string
	}{
		{
			name:           "tag with dots",
			imageName:      "image:1.2.3.4",
			expectedOutput: "Untagged: image:1.2.3.4\nDeleted: sha256:abc123def456789",
			expectedImages: []string{"image:1.2.3.4", "sha256:abc123def456789"},
		},
		{
			name:           "tag with underscores",
			imageName:      "service:api_v2_1",
			expectedOutput: "Untagged: service:api_v2_1\nDeleted: sha256:def456ghi789012",
			expectedImages: []string{"service:api_v2_1", "sha256:def456ghi789012"},
		},
		{
			name:           "tag with hyphens",
			imageName:      "app:release-2023-12-01",
			expectedOutput: "Untagged: app:release-2023-12-01\nDeleted: sha256:ghi789jkl012345",
			expectedImages: []string{"app:release-2023-12-01", "sha256:ghi789jkl012345"},
		},
		{
			name:           "tag with mixed characters",
			imageName:      "test:alpha-1.2.3_beta",
			expectedOutput: "Untagged: test:alpha-1.2.3_beta\nDeleted: sha256:jkl012mno345678",
			expectedImages: []string{"test:alpha-1.2.3_beta", "sha256:jkl012mno345678"},
		},
		{
			name:           "tag with numbers only",
			imageName:      "version:12345",
			expectedOutput: "Untagged: version:12345\nDeleted: sha256:mno345pqr678901",
			expectedImages: []string{"version:12345", "sha256:mno345pqr678901"},
		},
		{
			name:           "tag with special characters",
			imageName:      "build:test@sha256:abc123",
			expectedOutput: "Untagged: build:test@sha256:abc123\nDeleted: sha256:pqr678stu901234",
			expectedImages: []string{"build:test@sha256:abc123", "sha256:pqr678stu901234"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", tt.imageName).Return(tt.expectedOutput, nil)

			action := NewDockerImageRmByNameAction(logger, tt.imageName)
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Equal(t, tt.expectedImages, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}

func TestDockerImageRmAction_Execute_MultipleVersions(t *testing.T) {
	tests := []struct {
		name           string
		imageName      string
		expectedOutput string
		expectedImages []string
	}{
		{
			name:      "Multiple versions of same image",
			imageName: "nginx",
			expectedOutput: `Untagged: nginx:latest
Untagged: nginx:1.21
Untagged: nginx:1.20
Deleted: sha256:abc123def456789`,
			expectedImages: []string{"nginx:latest", "nginx:1.21", "nginx:1.20", "sha256:abc123def456789"},
		},
		{
			name:      "Multiple versions with different tags",
			imageName: "node",
			expectedOutput: `Untagged: node:latest
Untagged: node:18-alpine
Untagged: node:16-slim
Deleted: sha256:def456ghi789012`,
			expectedImages: []string{"node:latest", "node:18-alpine", "node:16-slim", "sha256:def456ghi789012"},
		},
		{
			name:      "Multiple versions with registry",
			imageName: "docker.io/library/ubuntu",
			expectedOutput: `Untagged: docker.io/library/ubuntu:latest
Untagged: docker.io/library/ubuntu:20.04
Untagged: docker.io/library/ubuntu:18.04
Deleted: sha256:ghi789jkl012345`,
			expectedImages: []string{"docker.io/library/ubuntu:latest", "docker.io/library/ubuntu:20.04", "docker.io/library/ubuntu:18.04", "sha256:ghi789jkl012345"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", tt.imageName).Return(tt.expectedOutput, nil)

			action := NewDockerImageRmByNameAction(logger, tt.imageName)
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Equal(t, tt.expectedImages, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}

func TestDockerImageRmAction_Execute_VersionDoesNotExist(t *testing.T) {
	tests := []struct {
		name           string
		imageName      string
		expectedError  string
		expectedOutput string
	}{
		{
			name:           "Non-existent version",
			imageName:      "nginx:999.999.999",
			expectedError:  "Error: No such image: nginx:999.999.999",
			expectedOutput: "Error: No such image: nginx:999.999.999",
		},
		{
			name:           "Non-existent image entirely",
			imageName:      "nonexistent-image:latest",
			expectedError:  "Error: No such image: nonexistent-image:latest",
			expectedOutput: "Error: No such image: nonexistent-image:latest",
		},
		{
			name:           "Non-existent version of existing image",
			imageName:      "ubuntu:999.999.999",
			expectedError:  "Error: No such image: ubuntu:999.999.999",
			expectedOutput: "Error: No such image: ubuntu:999.999.999",
		},
		{
			name:           "Invalid version format",
			imageName:      "nginx:invalid-version-format",
			expectedError:  "Error: No such image: nginx:invalid-version-format",
			expectedOutput: "Error: No such image: nginx:invalid-version-format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", tt.imageName).Return(tt.expectedOutput, errors.New(tt.expectedOutput))

			action := NewDockerImageRmByNameAction(logger, tt.imageName)
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Empty(t, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}

func TestDockerImageRmAction_Execute_DanglingImages(t *testing.T) {
	tests := []struct {
		name           string
		imageName      string
		expectedOutput string
		expectedImages []string
	}{
		{
			name:           "Dangling image with <none>:<none>",
			imageName:      "<none>:<none>",
			expectedOutput: `Deleted: sha256:abc123def456789`,
			expectedImages: []string{"sha256:abc123def456789"},
		},
		{
			name:      "Multiple dangling images",
			imageName: "<none>:<none>",
			expectedOutput: `Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012
Deleted: sha256:ghi789jkl012345`,
			expectedImages: []string{"sha256:abc123def456789", "sha256:def456ghi789012", "sha256:ghi789jkl012345"},
		},
		{
			name:           "Dangling image with name <none>",
			imageName:      "<none>",
			expectedOutput: `Deleted: sha256:abc123def456789`,
			expectedImages: []string{"sha256:abc123def456789"},
		},
		{
			name:           "Dangling image with ID only",
			imageName:      "sha256:abc123def456789",
			expectedOutput: `Deleted: sha256:abc123def456789`,
			expectedImages: []string{"sha256:abc123def456789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", tt.imageName).Return(tt.expectedOutput, nil)

			action := NewDockerImageRmByNameAction(logger, tt.imageName)
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Equal(t, tt.expectedImages, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}

func TestDockerImageRmAction_Execute_DanglingImagesByID(t *testing.T) {
	tests := []struct {
		name           string
		imageID        string
		expectedOutput string
		expectedImages []string
	}{
		{
			name:           "Dangling image by ID",
			imageID:        "sha256:abc123def456789",
			expectedOutput: `Deleted: sha256:abc123def456789`,
			expectedImages: []string{"sha256:abc123def456789"},
		},
		{
			name:           "Short image ID",
			imageID:        "abc123",
			expectedOutput: `Deleted: abc123`,
			expectedImages: []string{"abc123"},
		},
		{
			name:    "Multiple dangling images by ID",
			imageID: "sha256:abc123def456789",
			expectedOutput: `Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012`,
			expectedImages: []string{"sha256:abc123def456789", "sha256:def456ghi789012"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", tt.imageID).Return(tt.expectedOutput, nil)

			action := NewDockerImageRmByIDAction(logger, tt.imageID)
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Equal(t, tt.expectedImages, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}

func TestDockerImageRmAction_Execute_ForceRemoveNonExistent(t *testing.T) {
	tests := []struct {
		name           string
		imageName      string
		expectedOutput string
		expectedImages []string
	}{
		{
			name:           "Force remove non-existent version",
			imageName:      "nginx:999.999.999",
			expectedOutput: "Error: No such image: nginx:999.999.999",
			expectedImages: []string{},
		},
		{
			name:           "Force remove non-existent image",
			imageName:      "nonexistent-image:latest",
			expectedOutput: "Error: No such image: nonexistent-image:latest",
			expectedImages: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", "-f", tt.imageName).Return(tt.expectedOutput, errors.New(tt.expectedOutput))

			action := NewDockerImageRmByNameAction(logger, tt.imageName, WithForce())
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedOutput)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Empty(t, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}

func TestDockerImageRmAction_Execute_MixedOutputScenarios(t *testing.T) {
	tests := []struct {
		name           string
		imageName      string
		expectedOutput string
		expectedImages []string
	}{
		{
			name:      "Mixed untagged and deleted output",
			imageName: "nginx",
			expectedOutput: `Untagged: nginx:latest
Untagged: nginx:1.21
Deleted: sha256:abc123def456789
Untagged: nginx:1.20
Deleted: sha256:def456ghi789012`,
			expectedImages: []string{"nginx:latest", "nginx:1.21", "sha256:abc123def456789", "nginx:1.20", "sha256:def456ghi789012"},
		},
		{
			name:           "Only untagged (image still referenced)",
			imageName:      "nginx:latest",
			expectedOutput: `Untagged: nginx:latest`,
			expectedImages: []string{"nginx:latest"},
		},
		{
			name:           "Only deleted (no tags)",
			imageName:      "sha256:abc123def456789",
			expectedOutput: `Deleted: sha256:abc123def456789`,
			expectedImages: []string{"sha256:abc123def456789"},
		},
		{
			name:           "Empty output (nothing to remove)",
			imageName:      "nginx:latest",
			expectedOutput: "",
			expectedImages: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()

			mockRunner := &mocks.MockCommandRunner{}
			mockRunner.On("RunCommand", "docker", "image", "rm", tt.imageName).Return(tt.expectedOutput, nil)

			action := NewDockerImageRmByNameAction(logger, tt.imageName)
			action.Wrapped.SetCommandRunner(mockRunner)

			err := action.Wrapped.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, action.Wrapped.Output)
			assert.Equal(t, tt.expectedImages, action.Wrapped.RemovedImages)
			mockRunner.AssertExpectations(t)
		})
	}
}
