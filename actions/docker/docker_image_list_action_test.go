package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerImageListAction(t *testing.T) {
	logger := slog.Default()

	action := NewDockerImageListAction(logger)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-image-list-action", action.ID)
	assert.False(t, action.Wrapped.All)
	assert.False(t, action.Wrapped.Digests)
	assert.Empty(t, action.Wrapped.Filter)
	assert.Empty(t, action.Wrapped.Format)
	assert.False(t, action.Wrapped.NoTrunc)
	assert.False(t, action.Wrapped.Quiet)
}

func TestNewDockerImageListActionWithOptions(t *testing.T) {
	logger := slog.Default()

	action := NewDockerImageListAction(logger,
		WithAll(),
		WithDigests(),
		WithFilter("dangling=true"),
		WithFormat("table {{.Repository}}\t{{.Tag}}"),
		WithNoTrunc(),
		WithQuietOutput(),
	)

	assert.NotNil(t, action)
	assert.True(t, action.Wrapped.All)
	assert.True(t, action.Wrapped.Digests)
	assert.Equal(t, "dangling=true", action.Wrapped.Filter)
	assert.Equal(t, "table {{.Repository}}\t{{.Tag}}", action.Wrapped.Format)
	assert.True(t, action.Wrapped.NoTrunc)
	assert.True(t, action.Wrapped.Quiet)
}

func TestDockerImageListAction_Execute_Success(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB
redis               alpine              sha256:def456ghi789 3 weeks ago         32.3MB
postgres            13.4                sha256:ghi789jkl012 1 month ago         314MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Images, 3)

	// Check first image
	assert.Equal(t, "nginx", action.Wrapped.Images[0].Repository)
	assert.Equal(t, "latest", action.Wrapped.Images[0].Tag)
	assert.Equal(t, "sha256:abc123def456", action.Wrapped.Images[0].ImageID)
	assert.Equal(t, "2 weeks ago", action.Wrapped.Images[0].Created)
	assert.Equal(t, "133MB", action.Wrapped.Images[0].Size)

	// Check second image
	assert.Equal(t, "redis", action.Wrapped.Images[1].Repository)
	assert.Equal(t, "alpine", action.Wrapped.Images[1].Tag)
	assert.Equal(t, "sha256:def456ghi789", action.Wrapped.Images[1].ImageID)
	assert.Equal(t, "3 weeks ago", action.Wrapped.Images[1].Created)
	assert.Equal(t, "32.3MB", action.Wrapped.Images[1].Size)

	// Check third image
	assert.Equal(t, "postgres", action.Wrapped.Images[2].Repository)
	assert.Equal(t, "13.4", action.Wrapped.Images[2].Tag)
	assert.Equal(t, "sha256:ghi789jkl012", action.Wrapped.Images[2].ImageID)
	assert.Equal(t, "1 month ago", action.Wrapped.Images[2].Created)
	assert.Equal(t, "314MB", action.Wrapped.Images[2].Size)

	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_Execute_WithAll(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB
<none>              <none>              sha256:def456ghi789 3 weeks ago         0B`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--all").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithAll())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Images, 2)

	// Check dangling image
	assert.Empty(t, action.Wrapped.Images[1].Repository)
	assert.Empty(t, action.Wrapped.Images[1].Tag)
	assert.Equal(t, "sha256:def456ghi789", action.Wrapped.Images[1].ImageID)
	assert.Equal(t, "3 weeks ago", action.Wrapped.Images[1].Created)
	assert.Equal(t, "0B", action.Wrapped.Images[1].Size)

	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_Execute_WithFilter(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--filter", "dangling=false").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithFilter("dangling=false"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Images, 1)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_Execute_WithFormat(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `nginx:latest
redis:alpine`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--format", "{{.Repository}}:{{.Tag}}").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithFormat("{{.Repository}}:{{.Tag}}"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	// With custom format, we don't parse the output into structured data
	assert.Empty(t, action.Wrapped.Images)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_Execute_WithNoTrunc(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID                                                                    CREATED             SIZE
nginx               latest              sha256:abc123def456789012345678901234567890123456789012345678901234567890 2 weeks ago         133MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--no-trunc").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithNoTrunc())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Images, 1)
	assert.Equal(t, "sha256:abc123def456789012345678901234567890123456789012345678901234567890", action.Wrapped.Images[0].ImageID)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_Execute_WithQuiet(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `sha256:abc123def456
sha256:def456ghi789`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--quiet").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithQuietOutput())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	// With quiet mode, we don't parse the output into structured data
	assert.Empty(t, action.Wrapped.Images)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_Execute_CommandError(t *testing.T) {
	logger := slog.Default()
	expectedError := "permission denied"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return("", errors.New(expectedError))

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
	assert.Empty(t, action.Wrapped.Images)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_Execute_ContextCancellation(t *testing.T) {
	logger := slog.Default()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return("", context.Canceled)

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_parseImages(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedImages []DockerImage
	}{
		{
			name:           "empty output",
			output:         "",
			expectedImages: []DockerImage(nil),
		},
		{
			name:           "only header",
			output:         "REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE",
			expectedImages: []DockerImage(nil),
		},
		{
			name: "single image",
			output: `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB`,
			expectedImages: []DockerImage{
				{
					Repository: "nginx",
					Tag:        "latest",
					ImageID:    "sha256:abc123def456",
					Created:    "2 weeks ago",
					Size:       "133MB",
				},
			},
		},
		{
			name: "multiple images",
			output: `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB
redis               alpine              sha256:def456ghi789 3 weeks ago         32.3MB`,
			expectedImages: []DockerImage{
				{
					Repository: "nginx",
					Tag:        "latest",
					ImageID:    "sha256:abc123def456",
					Created:    "2 weeks ago",
					Size:       "133MB",
				},
				{
					Repository: "redis",
					Tag:        "alpine",
					ImageID:    "sha256:def456ghi789",
					Created:    "3 weeks ago",
					Size:       "32.3MB",
				},
			},
		},
		{
			name: "dangling images",
			output: `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
<none>              <none>              sha256:def456ghi789 3 weeks ago         0B`,
			expectedImages: []DockerImage{
				{
					Repository: "",
					Tag:        "",
					ImageID:    "sha256:def456ghi789",
					Created:    "3 weeks ago",
					Size:       "0B",
				},
			},
		},
		{
			name: "registry images",
			output: `REPOSITORY                    TAG                 IMAGE ID            CREATED             SIZE
docker.io/library/ubuntu   20.04               sha256:ghi789jkl012 1 month ago         72.8MB`,
			expectedImages: []DockerImage{
				{
					Repository: "docker.io/library/ubuntu",
					Tag:        "20.04",
					ImageID:    "sha256:ghi789jkl012",
					Created:    "1 month ago",
					Size:       "72.8MB",
				},
			},
		},
		{
			name: "image without tag (defaults to latest)",
			output: `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB`,
			expectedImages: []DockerImage{
				{
					Repository: "nginx",
					Tag:        "latest",
					ImageID:    "sha256:abc123def456",
					Created:    "2 weeks ago",
					Size:       "133MB",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			action := NewDockerImageListAction(logger)

			action.Wrapped.parseImages(tt.output)

			assert.Equal(t, tt.expectedImages, action.Wrapped.Images)
		})
	}
}

func TestDockerImageListAction_parseImageLine(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		expectedImage *DockerImage
	}{
		{
			name: "valid image line",
			line: "nginx               latest              sha256:abc123def456 2 weeks ago         133MB",
			expectedImage: &DockerImage{
				Repository: "nginx",
				Tag:        "latest",
				ImageID:    "sha256:abc123def456",
				Created:    "2 weeks ago",
				Size:       "133MB",
			},
		},
		{
			name: "dangling image",
			line: "<none>              <none>              sha256:def456ghi789 3 weeks ago         0B",
			expectedImage: &DockerImage{
				Repository: "",
				Tag:        "",
				ImageID:    "sha256:def456ghi789",
				Created:    "3 weeks ago",
				Size:       "0B",
			},
		},
		{
			name: "registry image",
			line: "docker.io/library/ubuntu   20.04               sha256:ghi789jkl012 1 month ago         72.8MB",
			expectedImage: &DockerImage{
				Repository: "docker.io/library/ubuntu",
				Tag:        "20.04",
				ImageID:    "sha256:ghi789jkl012",
				Created:    "1 month ago",
				Size:       "72.8MB",
			},
		},
		{
			name: "image without tag",
			line: "nginx               latest              sha256:abc123def456 2 weeks ago         133MB",
			expectedImage: &DockerImage{
				Repository: "nginx",
				Tag:        "latest",
				ImageID:    "sha256:abc123def456",
				Created:    "2 weeks ago",
				Size:       "133MB",
			},
		},
		{
			name:          "insufficient parts",
			line:          "nginx latest",
			expectedImage: nil,
		},
		{
			name:          "empty line",
			line:          "",
			expectedImage: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			action := NewDockerImageListAction(logger)

			result := action.Wrapped.parseImageLine(tt.line)

			assert.Equal(t, tt.expectedImage, result)
		})
	}
}

func TestDockerImageListAction_Execute_EmptyOutput(t *testing.T) {
	logger := slog.Default()
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Empty(t, action.Wrapped.Images)
	mockRunner.AssertExpectations(t)
}

func TestDockerImageListAction_Execute_OutputWithTrailingWhitespace(t *testing.T) {
	logger := slog.Default()
	rawOutput := "REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE\nnginx               latest              sha256:abc123def456 2 weeks ago         133MB\n  \n  "

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return(rawOutput, nil)

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, rawOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.Images, 1)
	assert.Equal(t, "nginx", action.Wrapped.Images[0].Repository)
	mockRunner.AssertExpectations(t)
}
