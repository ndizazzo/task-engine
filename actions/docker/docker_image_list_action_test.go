package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
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

// Tests for new constructor pattern with parameters
func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_WithParameters() {
	logger := slog.Default()

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: false}, // all
		task_engine.StaticParameter{Value: false}, // digests
		task_engine.StaticParameter{Value: ""},    // filter
		task_engine.StaticParameter{Value: ""},    // format
		task_engine.StaticParameter{Value: false}, // noTrunc
		task_engine.StaticParameter{Value: false}, // quiet
	)

	suite.Require().NoError(err)
	suite.NotNil(action)
	suite.Equal("docker-image-list-action", action.ID)
	suite.NotNil(action.Wrapped)
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_Execute_WithParameters() {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB
redis               alpine              sha256:def456ghi789 3 weeks ago         32.3MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return(expectedOutput, nil)

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: false}, // all
		task_engine.StaticParameter{Value: false}, // digests
		task_engine.StaticParameter{Value: ""},    // filter
		task_engine.StaticParameter{Value: ""},    // format
		task_engine.StaticParameter{Value: false}, // noTrunc
		task_engine.StaticParameter{Value: false}, // quiet
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Images, 2)
	suite.Equal("nginx", action.Wrapped.Images[0].Repository)
	suite.Equal("latest", action.Wrapped.Images[0].Tag)
	suite.Equal("sha256:abc123def456", action.Wrapped.Images[0].ImageID)
	suite.Equal("2 weeks ago", action.Wrapped.Images[0].Created)
	suite.Equal("133MB", action.Wrapped.Images[0].Size)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_Execute_WithAllParameter() {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB
<none>              <none>              sha256:def456ghi789 3 weeks ago         0B`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--all").Return(expectedOutput, nil)

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: true},  // all = true
		task_engine.StaticParameter{Value: false}, // digests
		task_engine.StaticParameter{Value: ""},    // filter
		task_engine.StaticParameter{Value: ""},    // format
		task_engine.StaticParameter{Value: false}, // noTrunc
		task_engine.StaticParameter{Value: false}, // quiet
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Images, 2)
	suite.Equal("<none>", action.Wrapped.Images[1].Repository)
	suite.Equal("<none>", action.Wrapped.Images[1].Tag)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_Execute_WithFilterParameter() {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--filter", "dangling=true").Return(expectedOutput, nil)

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: false},           // all
		task_engine.StaticParameter{Value: false},           // digests
		task_engine.StaticParameter{Value: "dangling=true"}, // filter
		task_engine.StaticParameter{Value: ""},              // format
		task_engine.StaticParameter{Value: false},           // noTrunc
		task_engine.StaticParameter{Value: false},           // quiet
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal("dangling=true", action.Wrapped.Filter)
	suite.Len(action.Wrapped.Images, 1)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_Execute_WithQuietParameter() {
	logger := slog.Default()
	expectedOutput := "sha256:abc123def456\nsha256:def456ghi789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--quiet").Return(expectedOutput, nil)

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: false}, // all
		task_engine.StaticParameter{Value: false}, // digests
		task_engine.StaticParameter{Value: ""},    // filter
		task_engine.StaticParameter{Value: ""},    // format
		task_engine.StaticParameter{Value: false}, // noTrunc
		task_engine.StaticParameter{Value: true},  // quiet = true
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.True(action.Wrapped.Quiet)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_Execute_WithFormatParameter() {
	logger := slog.Default()
	expectedOutput := "nginx:latest\nredis:alpine"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--format", "{{.Repository}}:{{.Tag}}").Return(expectedOutput, nil)

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: false},                      // all
		task_engine.StaticParameter{Value: false},                      // digests
		task_engine.StaticParameter{Value: ""},                         // filter
		task_engine.StaticParameter{Value: "{{.Repository}}:{{.Tag}}"}, // format
		task_engine.StaticParameter{Value: false},                      // noTrunc
		task_engine.StaticParameter{Value: false},                      // quiet
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal("{{.Repository}}:{{.Tag}}", action.Wrapped.Format)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_Execute_WithDigestsAndNoTruncParameters() {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456789abcdef123456789abcdef123456789abcdef123456789abcdef 2 weeks ago         133MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--digests", "--no-trunc").Return(expectedOutput, nil)

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: false}, // all
		task_engine.StaticParameter{Value: true},  // digests = true
		task_engine.StaticParameter{Value: ""},    // filter
		task_engine.StaticParameter{Value: ""},    // format
		task_engine.StaticParameter{Value: true},  // noTrunc = true
		task_engine.StaticParameter{Value: false}, // quiet
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.True(action.Wrapped.Digests)
	suite.True(action.Wrapped.NoTrunc)
	suite.Len(action.Wrapped.Images, 1)
	suite.Equal("sha256:abc123def456789abcdef123456789abcdef123456789abcdef123456789abcdef", action.Wrapped.Images[0].ImageID)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_Execute_InvalidParameterType() {
	logger := slog.Default()

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: "invalid"}, // all should be bool, not string
		task_engine.StaticParameter{Value: false},     // digests
		task_engine.StaticParameter{Value: ""},        // filter
		task_engine.StaticParameter{Value: ""},        // format
		task_engine.StaticParameter{Value: false},     // noTrunc
		task_engine.StaticParameter{Value: false},     // quiet
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "all parameter resolved to non-boolean value")
}

func (suite *DockerImageListActionTestSuite) TestNewDockerImageListActionConstructor_Execute_CommandFailure() {
	logger := slog.Default()
	expectedError := "docker image ls failed"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls").Return("", errors.New(expectedError))

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: false}, // all
		task_engine.StaticParameter{Value: false}, // digests
		task_engine.StaticParameter{Value: ""},    // filter
		task_engine.StaticParameter{Value: ""},    // format
		task_engine.StaticParameter{Value: false}, // noTrunc
		task_engine.StaticParameter{Value: false}, // quiet
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), expectedError)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.Images)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_parseImageLine() {
	action := &DockerImageListAction{}
	line := "nginx               latest              sha256:abc123def456 2 weeks ago         133MB"
	image := action.parseImageLine(line)

	suite.Equal("nginx", image.Repository)
	suite.Equal("latest", image.Tag)
	suite.Equal("sha256:abc123def456", image.ImageID)
	suite.Equal("2 weeks ago", image.Created)
	suite.Equal("133MB", image.Size)
	line = "<none>              <none>              sha256:def456ghi789 3 weeks ago         0B"
	image = action.parseImageLine(line)

	suite.Equal("<none>", image.Repository)
	suite.Equal("<none>", image.Tag)
	suite.Equal("sha256:def456ghi789", image.ImageID)
	suite.Equal("3 weeks ago", image.Created)
	suite.Equal("0B", image.Size)
	line = "docker.io/library/ubuntu 20.04              sha256:jkl012mno345 2 months ago        72.8MB"
	image = action.parseImageLine(line)

	suite.Equal("docker.io/library/ubuntu", image.Repository)
	suite.Equal("20.04", image.Tag)
	suite.Equal("sha256:jkl012mno345", image.ImageID)
	suite.Equal("2 months ago", image.Created)
	suite.Equal("72.8MB", image.Size)
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_GetOutput() {
	action := &DockerImageListAction{
		Output: "raw output",
		Images: []DockerImage{{Repository: "nginx", Tag: "latest", ImageID: "sha256:abc", Size: "133MB", Created: "2 weeks ago"}},
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(1, m["count"])
	suite.Equal("raw output", m["output"])
	suite.Equal(true, m["success"])
	suite.Len(m["images"], 1)
}

func (suite *DockerImageListActionTestSuite) TestDockerImageListAction_SetOptions() {
	logger := slog.Default()
	expectedOutput := `REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
nginx               latest              sha256:abc123def456 2 weeks ago         133MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "ls", "--all", "--digests", "--filter", "dangling=true", "--format", "{{.Repository}}", "--no-trunc", "--quiet").Return(expectedOutput, nil)

	constructor := NewDockerImageListAction(logger)
	action, err := constructor.WithParameters(
		nil, // all
		nil, // digests
		nil, // filter
		nil, // format
		nil, // noTrunc
		nil, // quiet
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)
	action.Wrapped.SetOptions(
		WithAll(),
		WithDigests(),
		WithFilter("dangling=true"),
		WithFormat("{{.Repository}}"),
		WithNoTrunc(),
		WithQuietOutput(),
	)

	err = action.Wrapped.Execute(context.Background())
	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.True(action.Wrapped.All)
	suite.True(action.Wrapped.Digests)
	suite.Equal("dangling=true", action.Wrapped.Filter)
	suite.Equal("{{.Repository}}", action.Wrapped.Format)
	suite.True(action.Wrapped.NoTrunc)
	suite.True(action.Wrapped.Quiet)

	mockRunner.AssertExpectations(suite.T())
}
