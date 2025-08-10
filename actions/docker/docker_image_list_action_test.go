package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

// DockerImageListActionTestSuite tests the DockerImageListAction
type DockerImageListActionTestSuite struct {
	suite.Suite
}

// TestDockerImageListActionTestSuite runs the DockerImageListAction test suite
func TestDockerImageListActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerImageListActionTestSuite))
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListAction() {
	logger := slog.Default()

	action := NewDockerImageListAction(logger)

	suite.NotNil(action)
	suite.Equal("docker-image-list-action", action.ID)
	suite.False(action.Wrapped.All)
	suite.False(action.Wrapped.Digests)
	suite.Empty(action.Wrapped.Filter)
	suite.Empty(action.Wrapped.Format)
	suite.False(action.Wrapped.NoTrunc)
	suite.False(action.Wrapped.Quiet)
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionWithOptions() {
	logger := slog.Default()

	action := NewDockerImageListAction(logger,
		WithAll(),
		WithDigests(),
		WithFilter("dangling=true"),
		WithFormat("table {{.Repository}}\t{{.Tag}}"),
		WithNoTrunc(),
		WithQuietOutput(),
	)

	suite.NotNil(action)
	suite.True(action.Wrapped.All)
	suite.True(action.Wrapped.Digests)
	suite.Equal("dangling=true", action.Wrapped.Filter)
	suite.Equal("table {{.Repository}}\t{{.Tag}}", action.Wrapped.Format)
	suite.True(action.Wrapped.NoTrunc)
	suite.True(action.Wrapped.Quiet)
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_Success() {
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

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Images, 3)

	// Check first image
	suite.Equal("nginx", action.Wrapped.Images[0].Repository)
	suite.Equal("latest", action.Wrapped.Images[0].Tag)
	suite.Equal("sha256:abc123def456", action.Wrapped.Images[0].ImageID)
	suite.Equal("2 weeks ago", action.Wrapped.Images[0].Created)
	suite.Equal("133MB", action.Wrapped.Images[0].Size)

	// Check second image
	suite.Equal("redis", action.Wrapped.Images[1].Repository)
	suite.Equal("alpine", action.Wrapped.Images[1].Tag)
	suite.Equal("sha256:def456ghi789", action.Wrapped.Images[1].ImageID)
	suite.Equal("3 weeks ago", action.Wrapped.Images[1].Created)
	suite.Equal("32.3MB", action.Wrapped.Images[1].Size)

	// Check third image
	suite.Equal("postgres", action.Wrapped.Images[2].Repository)
	suite.Equal("13.4", action.Wrapped.Images[2].Tag)
	suite.Equal("sha256:ghi789jkl012", action.Wrapped.Images[2].ImageID)
	suite.Equal("1 month ago", action.Wrapped.Images[2].Created)
	suite.Equal("314MB", action.Wrapped.Images[2].Size)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_WithAll() {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB
<none>              <none>              sha256:def456ghi789 3 weeks ago         0B`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--all").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithAll())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Images, 2)
	suite.Equal("<none>", action.Wrapped.Images[1].Repository)
	suite.Equal("<none>", action.Wrapped.Images[1].Tag)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_WithFilter() {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--filter", "dangling=true").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithFilter("dangling=true"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Images, 1)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_WithFormat() {
	logger := slog.Default()
	expectedOutput := "nginx:latest\nredis:alpine"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--format", "{{.Repository}}:{{.Tag}}").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithFormat("{{.Repository}}:{{.Tag}}"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_WithNoTrunc() {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456789abcdef123456789abcdef123456789abcdef123456789abcdef 2 weeks ago         133MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--no-trunc").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithNoTrunc())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Images, 1)
	suite.Equal("sha256:abc123def456789abcdef123456789abcdef123456789abcdef123456789abcdef", action.Wrapped.Images[0].ImageID)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_WithQuiet() {
	logger := slog.Default()
	expectedOutput := "sha256:abc123def456\nsha256:def456ghi789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--quiet").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger, WithQuietOutput())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_CommandError() {
	logger := slog.Default()
	expectedError := "docker image ls failed"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return("", errors.New(expectedError))

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), expectedError)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.Images)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_ContextCancellation() {
	logger := slog.Default()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return("", context.Canceled)

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.True(errors.Is(err, context.Canceled))
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.Images)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_parseImages() {
	logger := slog.Default()
	output := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB
redis               alpine              sha256:def456ghi789 3 weeks ago         32.3MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return(output, nil)

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(output, action.Wrapped.Output)
	suite.Len(action.Wrapped.Images, 2)
	suite.Equal("nginx", action.Wrapped.Images[0].Repository)
	suite.Equal("latest", action.Wrapped.Images[0].Tag)
	suite.Equal("sha256:abc123def456", action.Wrapped.Images[0].ImageID)
	suite.Equal("2 weeks ago", action.Wrapped.Images[0].Created)
	suite.Equal("133MB", action.Wrapped.Images[0].Size)
	suite.Equal("redis", action.Wrapped.Images[1].Repository)
	suite.Equal("alpine", action.Wrapped.Images[1].Tag)
	suite.Equal("sha256:def456ghi789", action.Wrapped.Images[1].ImageID)
	suite.Equal("3 weeks ago", action.Wrapped.Images[1].Created)
	suite.Equal("32.3MB", action.Wrapped.Images[1].Size)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_parseImageLine() {
	action := &DockerImageListAction{}

	// Test parsing a standard image line
	line := "nginx               latest              sha256:abc123def456 2 weeks ago         133MB"
	image := action.parseImageLine(line)

	suite.Equal("nginx", image.Repository)
	suite.Equal("latest", image.Tag)
	suite.Equal("sha256:abc123def456", image.ImageID)
	suite.Equal("2 weeks ago", image.Created)
	suite.Equal("133MB", image.Size)

	// Test parsing image with <none> values
	line = "<none>              <none>              sha256:def456ghi789 3 weeks ago         0B"
	image = action.parseImageLine(line)

	suite.Equal("<none>", image.Repository)
	suite.Equal("<none>", image.Tag)
	suite.Equal("sha256:def456ghi789", image.ImageID)
	suite.Equal("3 weeks ago", image.Created)
	suite.Equal("0B", image.Size)

	// Test parsing image with different time formats
	line = "postgres            13.4                sha256:ghi789jkl012 1 month ago         314MB"
	image = action.parseImageLine(line)

	suite.Equal("postgres", image.Repository)
	suite.Equal("13.4", image.Tag)
	suite.Equal("sha256:ghi789jkl012", image.ImageID)
	suite.Equal("1 month ago", image.Created)
	suite.Equal("314MB", image.Size)

	// Test parsing image with registry
	line = "docker.io/library/ubuntu 20.04              sha256:jkl012mno345 2 months ago        72.8MB"
	image = action.parseImageLine(line)

	suite.Equal("docker.io/library/ubuntu", image.Repository)
	suite.Equal("20.04", image.Tag)
	suite.Equal("sha256:jkl012mno345", image.ImageID)
	suite.Equal("2 months ago", image.Created)
	suite.Equal("72.8MB", image.Size)

	// Test parsing image with special characters
	line = "my-registry.com/my-project/my-app  v1.2.3            sha256:pqr678stu901 3 months ago        45.2MB"
	image = action.parseImageLine(line)

	suite.Equal("my-registry.com/my-project/my-app", image.Repository)
	suite.Equal("v1.2.3", image.Tag)
	suite.Equal("sha256:pqr678stu901", image.ImageID)
	suite.Equal("3 months ago", image.Created)
	suite.Equal("45.2MB", image.Size)
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_EmptyOutput() {
	logger := slog.Default()
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return(expectedOutput, nil)

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Empty(action.Wrapped.Images)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_Execute_OutputWithTrailingWhitespace() {
	logger := slog.Default()
	output := "REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE\nnginx               latest              sha256:abc123def456 2 weeks ago         133MB\n  \n"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return(output, nil)

	action := NewDockerImageListAction(logger)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(output, action.Wrapped.Output)
	suite.Len(action.Wrapped.Images, 1)
	suite.Equal("nginx", action.Wrapped.Images[0].Repository)
	mockRunner.AssertExpectations(suite.T())
}
