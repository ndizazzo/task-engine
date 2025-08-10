package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

// DockerImageRmActionTestSuite tests the DockerImageRmAction
type DockerImageRmActionTestSuite struct {
	suite.Suite
}

// TestDockerImageRmActionTestSuite runs the DockerImageRmAction test suite
func TestDockerImageRmActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerImageRmActionTestSuite))
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmByNameAction() {
	logger := slog.Default()
	imageName := "nginx:latest"

	action := NewDockerImageRmByNameAction(logger, imageName)

	suite.NotNil(action)
	suite.Equal("docker-image-rm-nginx:latest-action", action.ID)
	suite.Equal(imageName, action.Wrapped.ImageName)
	suite.Equal("", action.Wrapped.ImageID)
	suite.False(action.Wrapped.RemoveByID)
	suite.False(action.Wrapped.Force)
	suite.False(action.Wrapped.NoPrune)
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmByIDAction() {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"

	action := NewDockerImageRmByIDAction(logger, imageID)

	suite.NotNil(action)
	suite.Equal("docker-image-rm-id-sha256:abc123def456789-action", action.ID)
	suite.Equal("", action.Wrapped.ImageName)
	suite.Equal(imageID, action.Wrapped.ImageID)
	suite.True(action.Wrapped.RemoveByID)
	suite.False(action.Wrapped.Force)
	suite.False(action.Wrapped.NoPrune)
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmByNameActionWithOptions() {
	logger := slog.Default()
	imageName := "nginx:latest"

	action := NewDockerImageRmByNameAction(logger, imageName,
		WithForce(),
		WithNoPrune(),
	)

	suite.NotNil(action)
	suite.Equal(imageName, action.Wrapped.ImageName)
	suite.True(action.Wrapped.Force)
	suite.True(action.Wrapped.NoPrune)
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmByIDActionWithOptions() {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"

	action := NewDockerImageRmByIDAction(logger, imageID,
		WithForce(),
		WithNoPrune(),
	)

	suite.NotNil(action)
	suite.Equal(imageID, action.Wrapped.ImageID)
	suite.True(action.Wrapped.RemoveByID)
	suite.True(action.Wrapped.Force)
	suite.True(action.Wrapped.NoPrune)
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_ByName_Success() {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_ByID_Success() {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"
	expectedOutput := "Deleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageID).Return(expectedOutput, nil)

	action := NewDockerImageRmByIDAction(logger, imageID)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_WithForce() {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--force", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName, WithForce())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_WithNoPrune() {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--no-prune", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName, WithNoPrune())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_WithForceAndNoPrune() {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--force", "--no-prune", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName, WithForce(), WithNoPrune())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_CommandError() {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedError := errors.New("docker image rm command failed")

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return("", expectedError)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Equal(expectedError, err)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_ContextCancellation() {
	logger := slog.Default()
	imageName := "nginx:latest"
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return("", context.Canceled)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(ctx)

	suite.Error(err)
	suite.Equal(context.Canceled, err)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_parseRemovedImages() {
	logger := slog.Default()
	output := `Untagged: nginx:latest
Untagged: nginx:1.21
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012`

	action := NewDockerImageRmByNameAction(logger, "nginx")
	action.Wrapped.Output = output
	action.Wrapped.parseRemovedImages(output)

	suite.Len(action.Wrapped.RemovedImages, 4)
	suite.Equal("nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[2])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[3])
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_EmptyImageName() {
	logger := slog.Default()
	imageName := ""
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_EmptyImageID() {
	logger := slog.Default()
	imageID := ""
	expectedOutput := "Deleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageID).Return(expectedOutput, nil)

	action := NewDockerImageRmByIDAction(logger, imageID)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_SpecialCharactersInName() {
	logger := slog.Default()
	imageName := "my-app/nginx:latest"
	expectedOutput := "Untagged: my-app/nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"my-app/nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_OutputWithTrailingWhitespace() {
	logger := slog.Default()
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789\n  \n  "

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_VariousTagForms() {
	logger := slog.Default()
	imageName := "nginx"
	expectedOutput := `Untagged: nginx:latest
Untagged: nginx:1.21
Untagged: nginx:1.21-alpine
Untagged: nginx:alpine
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 6)
	suite.Equal("nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("nginx:1.21-alpine", action.Wrapped.RemovedImages[2])
	suite.Equal("nginx:alpine", action.Wrapped.RemovedImages[3])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[4])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[5])
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_RegistryImagesWithTags() {
	logger := slog.Default()
	imageName := "registry.example.com/myapp/nginx:latest"
	expectedOutput := `Untagged: registry.example.com/myapp/nginx:latest
Untagged: registry.example.com/myapp/nginx:1.21
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 4)
	suite.Equal("registry.example.com/myapp/nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("registry.example.com/myapp/nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[2])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[3])
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_EdgeCaseTags() {
	logger := slog.Default()
	imageName := "nginx"
	expectedOutput := `Untagged: nginx:latest
Untagged: nginx:1.21
Untagged: nginx:1.21-alpine
Untagged: nginx:alpine
Untagged: nginx:1.21-alpine-slim
Untagged: nginx:1.21-slim
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012
Deleted: sha256:ghi789jkl012345`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 9)
	suite.Equal("nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("nginx:1.21-alpine", action.Wrapped.RemovedImages[2])
	suite.Equal("nginx:alpine", action.Wrapped.RemovedImages[3])
	suite.Equal("nginx:1.21-alpine-slim", action.Wrapped.RemovedImages[4])
	suite.Equal("nginx:1.21-slim", action.Wrapped.RemovedImages[5])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[6])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[7])
	suite.Equal("sha256:ghi789jkl012345", action.Wrapped.RemovedImages[8])
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_MultipleVersions() {
	logger := slog.Default()
	imageName := "nginx"
	expectedOutput := `Untagged: nginx:latest
Untagged: nginx:1.21
Untagged: nginx:1.20
Untagged: nginx:1.19
Untagged: nginx:1.18
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012
Deleted: sha256:ghi789jkl012345
Deleted: sha256:jkl012mno345678`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 9)
	suite.Equal("nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("nginx:1.20", action.Wrapped.RemovedImages[2])
	suite.Equal("nginx:1.19", action.Wrapped.RemovedImages[3])
	suite.Equal("nginx:1.18", action.Wrapped.RemovedImages[4])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[5])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[6])
	suite.Equal("sha256:ghi789jkl012345", action.Wrapped.RemovedImages[7])
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_VersionDoesNotExist() {
	logger := slog.Default()
	imageName := "nginx:1.22"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return("", errors.New("No such image: nginx:1.22"))

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "No such image: nginx:1.22")
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_DanglingImages() {
	logger := slog.Default()
	imageName := "nginx"
	expectedOutput := `Untagged: nginx:latest
Untagged: nginx:1.21
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012
Deleted: sha256:ghi789jkl012345`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 5)
	suite.Equal("nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[2])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[3])
	suite.Equal("sha256:ghi789jkl012345", action.Wrapped.RemovedImages[4])
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_DanglingImagesByID() {
	logger := slog.Default()
	imageID := "sha256:abc123def456789"
	expectedOutput := "Deleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageID).Return(expectedOutput, nil)

	action := NewDockerImageRmByIDAction(logger, imageID)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 1)
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[0])
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_ForceRemoveNonExistent() {
	logger := slog.Default()
	imageName := "nonexistent:latest"
	expectedOutput := "Untagged: nonexistent:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--force", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName, WithForce())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nonexistent:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_MixedOutputScenarios() {
	logger := slog.Default()
	imageName := "nginx"
	expectedOutput := `Untagged: nginx:latest
Untagged: nginx:1.21
Untagged: nginx:1.20
Untagged: nginx:1.19
Untagged: nginx:1.18
Untagged: nginx:1.17
Untagged: nginx:1.16
Untagged: nginx:1.15
Untagged: nginx:1.14
Untagged: nginx:1.13
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012
Deleted: sha256:ghi789jkl012345
Deleted: sha256:jkl012mno345678
Deleted: sha256:mno345pqr678901
Deleted: sha256:pqr678stu901234
Deleted: sha256:stu901vwx234567
Deleted: sha256:vwx234yza567890
Deleted: sha256:yza567bcd890123
Deleted: sha256:bcd890efg123456`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	action := NewDockerImageRmByNameAction(logger, imageName)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 20)
	suite.Equal("nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("nginx:1.20", action.Wrapped.RemovedImages[2])
	suite.Equal("nginx:1.19", action.Wrapped.RemovedImages[3])
	suite.Equal("nginx:1.18", action.Wrapped.RemovedImages[4])
	suite.Equal("nginx:1.17", action.Wrapped.RemovedImages[5])
	suite.Equal("nginx:1.16", action.Wrapped.RemovedImages[6])
	suite.Equal("nginx:1.15", action.Wrapped.RemovedImages[7])
	suite.Equal("nginx:1.14", action.Wrapped.RemovedImages[8])
	suite.Equal("nginx:1.13", action.Wrapped.RemovedImages[9])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[10])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[11])
	suite.Equal("sha256:ghi789jkl012345", action.Wrapped.RemovedImages[12])
	suite.Equal("sha256:jkl012mno345678", action.Wrapped.RemovedImages[13])
	suite.Equal("sha256:mno345pqr678901", action.Wrapped.RemovedImages[14])
	suite.Equal("sha256:pqr678stu901234", action.Wrapped.RemovedImages[15])
	suite.Equal("sha256:stu901vwx234567", action.Wrapped.RemovedImages[16])
	suite.Equal("sha256:vwx234yza567890", action.Wrapped.RemovedImages[17])
	suite.Equal("sha256:yza567bcd890123", action.Wrapped.RemovedImages[18])
	mockRunner.AssertExpectations(suite.T())
}
