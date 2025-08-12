package docker

import (
	"context"
	"errors"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
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
	imageName := "nginx:latest"

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(
		task_engine.StaticParameter{Value: imageName},
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false}, // removeByID
		task_engine.StaticParameter{Value: false}, // force
		task_engine.StaticParameter{Value: false}, // noPrune
	)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("docker-image-rm-action", action.ID)
	suite.NotNil(action.Wrapped.ImageNameParam)
	suite.NotNil(action.Wrapped.ImageIDParam)
	suite.NotNil(action.Wrapped.RemoveByIDParam)
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmByIDAction() {
	imageID := "sha256:abc123def456789"

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: imageID},
		task_engine.StaticParameter{Value: true},  // removeByID
		task_engine.StaticParameter{Value: false}, // force
		task_engine.StaticParameter{Value: false}, // noPrune
	)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("docker-image-rm-action", action.ID)
	suite.NotNil(action.Wrapped.ImageNameParam)
	suite.NotNil(action.Wrapped.ImageIDParam)
	suite.NotNil(action.Wrapped.RemoveByIDParam)
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmByNameActionWithOptions() {
	imageName := "nginx:latest"
	logger := mocks.NewDiscardLogger()

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(logger).WithParameters(
		task_engine.StaticParameter{Value: imageName},
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false}, // removeByID
		task_engine.StaticParameter{Value: true},  // force
		task_engine.StaticParameter{Value: true},  // noPrune
	)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.NotNil(action.Wrapped.ImageNameParam)
	suite.NotNil(action.Wrapped.ForceParam)
	suite.NotNil(action.Wrapped.NoPruneParam)
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmByIDActionWithOptions() {
	imageID := "sha256:abc123def456789"
	logger := mocks.NewDiscardLogger()

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: imageID},
		task_engine.StaticParameter{Value: true}, // removeByID
		task_engine.StaticParameter{Value: true}, // force
		task_engine.StaticParameter{Value: true}, // noPrune
	)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.NotNil(action.Wrapped.ImageIDParam)
	suite.NotNil(action.Wrapped.RemoveByIDParam)
	suite.NotNil(action.Wrapped.ForceParam)
	suite.NotNil(action.Wrapped.NoPruneParam)
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_ByName_Success() {
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_ByID_Success() {
	imageID := "sha256:abc123def456789"
	expectedOutput := "Deleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageID).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: imageID}, task_engine.StaticParameter{Value: true}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_WithForce() {
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--force", imageName).Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(nil).WithParameters(
		task_engine.StaticParameter{Value: imageName},
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false}, // removeByID
		task_engine.StaticParameter{Value: true},  // force
		task_engine.StaticParameter{Value: false}, // noPrune
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_WithNoPrune() {
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--no-prune", imageName).Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(nil).WithParameters(
		task_engine.StaticParameter{Value: imageName},
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false}, // removeByID
		task_engine.StaticParameter{Value: false}, // force
		task_engine.StaticParameter{Value: true},  // noPrune
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_WithForceAndNoPrune() {
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--force", "--no-prune", imageName).Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(nil).WithParameters(
		task_engine.StaticParameter{Value: imageName},
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false}, // removeByID
		task_engine.StaticParameter{Value: true},  // force
		task_engine.StaticParameter{Value: true},  // noPrune
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_CommandError() {
	imageName := "nginx:latest"
	expectedError := errors.New("docker image rm command failed")

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return("", expectedError)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Equal(expectedError, execErr)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_ContextCancellation() {
	imageName := "nginx:latest"
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return("", context.Canceled)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(ctx)

	suite.Error(execErr)
	suite.Equal(context.Canceled, execErr)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_parseRemovedImages() {
	output := `Untagged: nginx:latest
Untagged: nginx:1.21
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012`

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: "nginx:latest"}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.Output = output
	action.Wrapped.parseRemovedImages(output)

	suite.Len(action.Wrapped.RemovedImages, 4)
	suite.Equal("nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[2])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[3])
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_SpecialCharactersInName() {
	imageName := "my-app/nginx:latest"
	expectedOutput := "Untagged: my-app/nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"my-app/nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_OutputWithTrailingWhitespace() {
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789\n  \n  "

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_VariousTagForms() {
	imageName := "nginx"
	expectedOutput := `Untagged: nginx:latest
Untagged: nginx:1.21
Untagged: nginx:1.21-alpine
Untagged: nginx:alpine
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
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
	imageName := "registry.example.com/myapp/nginx:latest"
	expectedOutput := `Untagged: registry.example.com/myapp/nginx:latest
Untagged: registry.example.com/myapp/nginx:1.21
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 4)
	suite.Equal("registry.example.com/myapp/nginx:latest", action.Wrapped.RemovedImages[0])
	suite.Equal("registry.example.com/myapp/nginx:1.21", action.Wrapped.RemovedImages[1])
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[2])
	suite.Equal("sha256:def456ghi789012", action.Wrapped.RemovedImages[3])
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_EdgeCaseTags() {
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

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
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

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
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
	imageName := "nginx:1.22"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return("", errors.New("No such image: nginx:1.22"))

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.Contains(execErr.Error(), "No such image: nginx:1.22")
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_DanglingImages() {
	imageName := "nginx"
	expectedOutput := `Untagged: nginx:latest
Untagged: nginx:1.21
Deleted: sha256:abc123def456789
Deleted: sha256:def456ghi789012
Deleted: sha256:ghi789jkl012345`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageName).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
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
	imageID := "sha256:abc123def456789"
	expectedOutput := "Deleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", imageID).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: imageID}, task_engine.StaticParameter{Value: true}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.RemovedImages, 1)
	suite.Equal("sha256:abc123def456789", action.Wrapped.RemovedImages[0])
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_ForceRemoveNonExistent() {
	imageName := "nginx:latest"
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "--force", imageName).Return(expectedOutput, nil)

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: true}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_Execute_MixedOutputScenarios() {
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

	var action *task_engine.Action[*DockerImageRmAction]
	var err error
	action, err = NewDockerImageRmAction(nil).WithParameters(task_engine.StaticParameter{Value: imageName}, task_engine.StaticParameter{Value: ""}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
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

// ===== PARAMETER-AWARE CONSTRUCTOR TESTS =====

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmActionWithParams() {
	imageNameParam := task_engine.StaticParameter{Value: "nginx:latest"}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123def456789"}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("docker-image-rm-action", action.ID) // Modern constructor uses consistent ID
	suite.NotNil(action.Wrapped.ImageNameParam)
	suite.NotNil(action.Wrapped.ImageIDParam)
	suite.Equal(imageNameParam, action.Wrapped.ImageNameParam)
	suite.Equal(imageIDParam, action.Wrapped.ImageIDParam)
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmActionWithParams_RemoveByID() {
	imageNameParam := task_engine.StaticParameter{Value: "nginx:latest"}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123def456789"}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: true}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("docker-image-rm-action", action.ID) // Modern constructor uses consistent ID
}

func (suite *DockerImageRmActionTestSuite) TestNewDockerImageRmActionWithParams_WithForceAndNoPrune() {
	imageNameParam := task_engine.StaticParameter{Value: "nginx:latest"}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123def456789"}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: true}, task_engine.StaticParameter{Value: true})
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("docker-image-rm-action", action.ID) // Modern constructor uses consistent ID
}

// ===== PARAMETER RESOLUTION TESTS =====

func (suite *DockerImageRmActionTestSuite) TestExecute_WithStaticParameters() {
	imageNameParam := task_engine.StaticParameter{Value: "nginx:latest"}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123def456789"}

	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "nginx:latest").Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"nginx:latest", "sha256:abc123def456789"}, action.Wrapped.RemovedImages)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestExecute_WithStaticParameters_RemoveByID() {
	imageNameParam := task_engine.StaticParameter{Value: "nginx:latest"}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123def456789"}

	expectedOutput := "Deleted: sha256:abc123def456789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "sha256:abc123def456789").Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: true}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"sha256:abc123def456789"}, action.Wrapped.RemovedImages)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestExecute_WithActionOutputParameter() {
	// Create a mock global context with action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("list-images", map[string]interface{}{
		"imageName": "redis:alpine",
		"imageID":   "sha256:def456ghi789",
	})

	imageNameParam := task_engine.ActionOutputParameter{
		ActionID:  "list-images",
		OutputKey: "imageName",
	}
	imageIDParam := task_engine.ActionOutputParameter{
		ActionID:  "list-images",
		OutputKey: "imageID",
	}

	expectedOutput := "Untagged: redis:alpine\nDeleted: sha256:def456ghi789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "redis:alpine").Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"redis:alpine", "sha256:def456ghi789"}, action.Wrapped.RemovedImages)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestExecute_WithTaskOutputParameter() {
	// Create a mock global context with task output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreTaskOutput("build-task", map[string]interface{}{
		"builtImage": "myapp:v1.0.0",
		"imageHash":  "sha256:abc123def456",
	})

	imageNameParam := task_engine.TaskOutputParameter{
		TaskID:    "build-task",
		OutputKey: "builtImage",
	}
	imageIDParam := task_engine.TaskOutputParameter{
		TaskID:    "build-task",
		OutputKey: "imageHash",
	}

	expectedOutput := "Untagged: myapp:v1.0.0\nDeleted: sha256:abc123def456"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "myapp:v1.0.0").Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"myapp:v1.0.0", "sha256:abc123def456"}, action.Wrapped.RemovedImages)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestExecute_WithEntityOutputParameter() {
	// Create a mock global context with entity output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("docker-build", map[string]interface{}{
		"imageName": "prod-app:latest",
		"imageID":   "sha256:prod123hash456",
	})

	imageNameParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "imageName",
	}
	imageIDParam := task_engine.EntityOutputParameter{
		EntityType: "action",
		EntityID:   "docker-build",
		OutputKey:  "imageID",
	}

	expectedOutput := "Untagged: prod-app:latest\nDeleted: sha256:prod123hash456"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "prod-app:latest").Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"prod-app:latest", "sha256:prod123hash456"}, action.Wrapped.RemovedImages)

	mockRunner.AssertExpectations(suite.T())
}

// ===== PARAMETER ERROR HANDLING TESTS =====

func (suite *DockerImageRmActionTestSuite) TestExecute_WithInvalidActionOutputParameter() {
	// Create a mock global context without the referenced action
	globalContext := task_engine.NewGlobalContext()

	imageNameParam := task_engine.ActionOutputParameter{
		ActionID:  "non-existent-action",
		OutputKey: "imageName",
	}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123"}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.Error(execErr)
	suite.ErrorContains(execErr, "action 'non-existent-action' not found in context")
}

func (suite *DockerImageRmActionTestSuite) TestExecute_WithInvalidOutputKey() {
	// Create a mock global context with action output but missing key
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("list-images", map[string]interface{}{
		"otherField": "value",
	})

	imageNameParam := task_engine.ActionOutputParameter{
		ActionID:  "list-images",
		OutputKey: "imageName", // This key doesn't exist in the output
	}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123"}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.Error(execErr)
	suite.ErrorContains(execErr, "output key 'imageName' not found in action 'list-images'")
}

func (suite *DockerImageRmActionTestSuite) TestExecute_WithEmptyActionID() {
	imageNameParam := task_engine.ActionOutputParameter{
		ActionID:  "", // Empty ActionID
		OutputKey: "imageName",
	}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123"}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.ErrorContains(execErr, "ActionID cannot be empty")
}

func (suite *DockerImageRmActionTestSuite) TestExecute_WithNonMapOutput() {
	// Create a mock global context with non-map action output
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("list-images", "not-a-map")

	imageNameParam := task_engine.ActionOutputParameter{
		ActionID:  "list-images",
		OutputKey: "imageName",
	}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123"}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.Error(execErr)
	suite.ErrorContains(execErr, "action 'list-images' output is not a map, cannot extract key 'imageName'")
}

// ===== PARAMETER TYPE VALIDATION TESTS =====

func (suite *DockerImageRmActionTestSuite) TestExecute_WithNonStringImageNameParameter() {
	imageNameParam := task_engine.StaticParameter{Value: 123}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123"}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.ErrorContains(execErr, "image name parameter is not a string, got int")
}

func (suite *DockerImageRmActionTestSuite) TestExecute_WithNonStringImageIDParameter() {
	imageNameParam := task_engine.StaticParameter{Value: "nginx:latest"}
	imageIDParam := task_engine.StaticParameter{Value: 456}

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	execErr := action.Wrapped.Execute(context.Background())

	suite.Error(execErr)
	suite.ErrorContains(execErr, "image ID parameter is not a string, got int")
}

// ===== COMPLEX PARAMETER SCENARIOS =====

func (suite *DockerImageRmActionTestSuite) TestExecute_WithComplexImageNameResolution() {
	globalContext := task_engine.NewGlobalContext()
	globalContext.StoreActionOutput("build-action", map[string]interface{}{
		"imageName": "myapp:v1.0.0",
	})
	globalContext.StoreTaskOutput("deploy-task", map[string]interface{}{
		"imageID": "sha256:deploy123hash456",
	})

	imageNameParam := task_engine.ActionOutputParameter{
		ActionID:  "build-action",
		OutputKey: "imageName",
	}
	imageIDParam := task_engine.TaskOutputParameter{
		TaskID:    "deploy-task",
		OutputKey: "imageID",
	}

	expectedOutput := "Untagged: myapp:v1.0.0\nDeleted: sha256:deploy123hash456"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "myapp:v1.0.0").Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	// Create context with global context
	ctx := context.WithValue(context.Background(), task_engine.GlobalContextKey, globalContext)

	execErr := action.Wrapped.Execute(ctx)

	suite.NoError(execErr)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal([]string{"myapp:v1.0.0", "sha256:deploy123hash456"}, action.Wrapped.RemovedImages)

	mockRunner.AssertExpectations(suite.T())
}

// ===== BACKWARD COMPATIBILITY TESTS =====

func (suite *DockerImageRmActionTestSuite) TestBackwardCompatibility_ExecuteWithoutGlobalContext() {
	imageNameParam := task_engine.StaticParameter{Value: "nginx:latest"}
	imageIDParam := task_engine.StaticParameter{Value: "sha256:abc123"}
	expectedOutput := "Untagged: nginx:latest\nDeleted: sha256:abc123"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "image", "rm", "nginx:latest").Return(expectedOutput, nil)

	action, err := NewDockerImageRmAction(mocks.NewDiscardLogger()).WithParameters(imageNameParam, imageIDParam, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false}, task_engine.StaticParameter{Value: false})
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	execErr := action.Wrapped.Execute(context.Background())

	suite.NoError(execErr)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerImageRmActionTestSuite) TestDockerImageRmAction_GetOutput() {
	action := &DockerImageRmAction{
		Output:        "Untagged: nginx:latest\nDeleted: sha256:abc123def456789",
		RemovedImages: []string{"nginx:latest", "sha256:abc123def456789"},
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(2, m["count"])
	suite.Equal("Untagged: nginx:latest\nDeleted: sha256:abc123def456789", m["output"])
	suite.Equal(true, m["success"])
	suite.Len(m["removed"], 2)
}
