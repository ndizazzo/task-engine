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

// DockerLoadActionTestSuite tests the DockerLoadAction
type DockerLoadActionTestSuite struct {
	suite.Suite
}

// TestDockerLoadActionTestSuite runs the DockerLoadAction test suite
func TestDockerLoadActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerLoadActionTestSuite))
}

func (suite *DockerLoadActionTestSuite) TestNewDockerLoadAction() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"

	action := NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: tarFilePath})

	suite.NotNil(action)
	suite.Equal("docker-load--path-to-image.tar-action", action.ID)
	// TarFilePath is resolved at runtime, not at construction
	suite.Equal("", action.Wrapped.TarFilePath)
	suite.Equal("", action.Wrapped.Platform)
	suite.False(action.Wrapped.Quiet)
}

func (suite *DockerLoadActionTestSuite) TestNewDockerLoadActionWithOptions() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"

	action := NewDockerLoadAction(logger).WithOptions(WithPlatform("linux/amd64"), WithQuiet()).WithParameters(task_engine.StaticParameter{Value: tarFilePath})

	suite.NotNil(action)
	suite.NotNil(action.Wrapped.TarFilePathParam)
	suite.True(action.Wrapped.Platform == "linux/amd64")
	suite.True(action.Wrapped.Quiet)
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_Success() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	expectedOutput := "Loaded image: nginx:latest\nLoaded image: redis:alpine"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: tarFilePath})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "redis:alpine"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_WithPlatform() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	platform := "linux/amd64"
	expectedOutput := "Loaded image: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath, "--platform", platform).Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger).WithOptions(WithPlatform(platform)).WithParameters(task_engine.StaticParameter{Value: tarFilePath})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_WithQuiet() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	expectedOutput := "Loaded image: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath, "-q").Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger).WithOptions(WithQuiet()).WithParameters(task_engine.StaticParameter{Value: tarFilePath})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_WithPlatformAndQuiet() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	platform := "linux/amd64"
	expectedOutput := "Loaded image: nginx:latest"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath, "--platform", platform, "-q").Return(expectedOutput, nil)

	action := NewDockerLoadAction(logger).WithOptions(WithPlatform(platform), WithQuiet()).WithParameters(task_engine.StaticParameter{Value: tarFilePath})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_CommandError() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	expectedError := "docker load failed"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return("", errors.New(expectedError))

	action := NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: tarFilePath})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), expectedError)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_ContextCancellation() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return("", context.Canceled)

	action := NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: tarFilePath})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.True(errors.Is(err, context.Canceled))
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_parseLoadedImages() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	output := `Loaded image: nginx:latest
Loaded image: redis:alpine
Loaded image: postgres:13`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return(output, nil)

	action := NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: tarFilePath})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(output, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "redis:alpine", "postgres:13"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_EmptyTarFilePath() {
	logger := slog.Default()
	tarFilePath := ""

	action := NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: tarFilePath})

	suite.NotNil(action)
	suite.Equal("docker-load--action", action.ID)
	suite.Equal("", action.Wrapped.TarFilePath)
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_SpecialCharactersInPath() {
	logger := slog.Default()
	tarFilePath := "/path/with spaces/and-special-chars@#$%.tar"

	action := NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: tarFilePath})

	suite.NotNil(action)
	suite.Equal("docker-load--path-with-spaces-and-special-chars.tar-action", action.ID)
	// TarFilePath is resolved at runtime, not at construction
	suite.Equal("", action.Wrapped.TarFilePath)
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_Execute_OutputWithTrailingWhitespace() {
	logger := slog.Default()
	tarFilePath := "/path/to/image.tar"
	output := "Loaded image: nginx:latest\nLoaded image: redis:alpine\n  \n"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "load", "-i", tarFilePath).Return(output, nil)

	action := NewDockerLoadAction(logger).WithParameters(task_engine.StaticParameter{Value: tarFilePath})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(output, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "redis:alpine"}, action.Wrapped.LoadedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerLoadActionTestSuite) TestDockerLoadAction_GetOutput() {
	action := &DockerLoadAction{
		TarFilePath:  "/path/to/image.tar",
		Output:       "Loaded image: nginx:latest\nLoaded image: redis:alpine",
		LoadedImages: []string{"nginx:latest", "redis:alpine"},
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(2, m["count"])
	suite.Equal("/path/to/image.tar", m["tarFile"])
	suite.Equal("Loaded image: nginx:latest\nLoaded image: redis:alpine", m["output"])
	suite.Equal(true, m["success"])
	suite.Len(m["loadedImages"], 2)
}
