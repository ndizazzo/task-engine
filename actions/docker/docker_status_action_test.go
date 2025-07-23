package docker_test

import (
	"context"
	"testing"

	"github.com/ndizazzo/task-engine/actions/docker"
	command_mock "github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DockerStatusActionTestSuite struct {
	suite.Suite
	mockProcessor *command_mock.MockCommandRunner
}

func (suite *DockerStatusActionTestSuite) SetupTest() {
	suite.mockProcessor = new(command_mock.MockCommandRunner)
}

func (suite *DockerStatusActionTestSuite) TestGetSpecificContainerState() {
	containerName := "test-container"
	expectedOutput := `{"ID":"abc123","Names":"test-container","Image":"nginx:latest","Status":"Up 2 hours"}`
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, containerName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result
	suite.Len(action.Wrapped.ContainerStates, 1)
	container := action.Wrapped.ContainerStates[0]
	suite.Equal("abc123", container.ID)
	suite.Equal([]string{"test-container"}, container.Names)
	suite.Equal("nginx:latest", container.Image)
	suite.Equal("Up 2 hours", container.Status)
}

func (suite *DockerStatusActionTestSuite) TestGetMultipleContainerStates() {
	containerNames := []string{"container1", "container2"}
	expectedOutput := `{"ID":"abc123","Names":"container1","Image":"nginx:latest","Status":"Up 2 hours"}
{"ID":"def456","Names":"container2","Image":"redis:alpine","Status":"Exited (0) 1 hour ago"}`
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, containerNames...)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=container1", "--filter", "name=container2").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed results
	suite.Len(action.Wrapped.ContainerStates, 2)

	container1 := action.Wrapped.ContainerStates[0]
	suite.Equal("abc123", container1.ID)
	suite.Equal([]string{"container1"}, container1.Names)
	suite.Equal("nginx:latest", container1.Image)
	suite.Equal("Up 2 hours", container1.Status)

	container2 := action.Wrapped.ContainerStates[1]
	suite.Equal("def456", container2.ID)
	suite.Equal([]string{"container2"}, container2.Names)
	suite.Equal("redis:alpine", container2.Image)
	suite.Equal("Exited (0) 1 hour ago", container2.Status)
}

func (suite *DockerStatusActionTestSuite) TestGetAllContainersState() {
	expectedOutput := `{"ID":"abc123","Names":"container1","Image":"nginx:latest","Status":"Up 2 hours"}
{"ID":"def456","Names":"container2","Image":"redis:alpine","Status":"Exited (0) 1 hour ago"}
{"ID":"ghi789","Names":"container3","Image":"postgres:13","Status":"Paused"}`
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed results
	suite.Len(action.Wrapped.ContainerStates, 3)

	container1 := action.Wrapped.ContainerStates[0]
	suite.Equal("abc123", container1.ID)
	suite.Equal([]string{"container1"}, container1.Names)
	suite.Equal("nginx:latest", container1.Image)
	suite.Equal("Up 2 hours", container1.Status)

	container2 := action.Wrapped.ContainerStates[1]
	suite.Equal("def456", container2.ID)
	suite.Equal([]string{"container2"}, container2.Names)
	suite.Equal("redis:alpine", container2.Image)
	suite.Equal("Exited (0) 1 hour ago", container2.Status)

	container3 := action.Wrapped.ContainerStates[2]
	suite.Equal("ghi789", container3.ID)
	suite.Equal([]string{"container3"}, container3.Names)
	suite.Equal("postgres:13", container3.Image)
	suite.Equal("Paused", container3.Status)
}

func (suite *DockerStatusActionTestSuite) TestContainerWithMultipleNames() {
	containerName := "test-container"
	expectedOutput := `{"ID":"abc123","Names":"test-container,my-container,alias1","Image":"nginx:latest","Status":"Up 2 hours"}`
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, containerName)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result with multiple names
	suite.Len(action.Wrapped.ContainerStates, 1)
	container := action.Wrapped.ContainerStates[0]
	suite.Equal("abc123", container.ID)
	suite.Equal([]string{"test-container", "my-container", "alias1"}, container.Names)
	suite.Equal("nginx:latest", container.Image)
	suite.Equal("Up 2 hours", container.Status)
}

func (suite *DockerStatusActionTestSuite) TestEmptyOutput() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return("", nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Len(action.Wrapped.ContainerStates, 0)
}

func (suite *DockerStatusActionTestSuite) TestWhitespaceOnlyOutput() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return("   \n  \t  ", nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Len(action.Wrapped.ContainerStates, 0)
}

func (suite *DockerStatusActionTestSuite) TestMalformedJSONLine() {
	expectedOutput := `{"ID":"abc123","Names":"container1","Image":"nginx:latest","Status":"Up 2 hours"}
{"ID":"def456","Names":"container2","Image":"redis:alpine","Status":"Exited (0) 1 hour ago"}
{"ID":"ghi789","Names":"container3","Image":"postgres:13","Status":"Paused"}
{"ID":"jkl012","Names":"container4","Image":"invalid-json","Status":"Up 1 hour`
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Should parse the valid containers and skip the malformed one
	suite.Len(action.Wrapped.ContainerStates, 3)

	container1 := action.Wrapped.ContainerStates[0]
	suite.Equal("abc123", container1.ID)
	suite.Equal("container1", container1.Names[0])

	container2 := action.Wrapped.ContainerStates[1]
	suite.Equal("def456", container2.ID)
	suite.Equal("container2", container2.Names[0])

	container3 := action.Wrapped.ContainerStates[2]
	suite.Equal("ghi789", container3.ID)
	suite.Equal("container3", container3.Names[0])
}

func (suite *DockerStatusActionTestSuite) TestCommandFailure() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, "test-container")
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return("", assert.AnError)

	err := action.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get container state")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerStatusActionTestSuite) TestContainerWithEmptyNames() {
	expectedOutput := `{"ID":"abc123","Names":"","Image":"nginx:latest","Status":"Up 2 hours"}`
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result with empty names
	suite.Len(action.Wrapped.ContainerStates, 1)
	container := action.Wrapped.ContainerStates[0]
	suite.Equal("abc123", container.ID)
	suite.Equal([]string{}, container.Names) // Should be empty slice, not nil
	suite.Equal("nginx:latest", container.Image)
	suite.Equal("Up 2 hours", container.Status)
}

func (suite *DockerStatusActionTestSuite) TestContainerWithWhitespaceInNames() {
	expectedOutput := `{"ID":"abc123","Names":"  container1  ,  container2  ,  container3  ","Image":"nginx:latest","Status":"Up 2 hours"}`
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result with whitespace trimmed
	suite.Len(action.Wrapped.ContainerStates, 1)
	container := action.Wrapped.ContainerStates[0]
	suite.Equal("abc123", container.ID)
	suite.Equal([]string{"container1", "container2", "container3"}, container.Names)
	suite.Equal("nginx:latest", container.Image)
	suite.Equal("Up 2 hours", container.Status)
}

func (suite *DockerStatusActionTestSuite) TestContainerWithCommasInNames() {
	expectedOutput := `{"ID":"abc123","Names":"container1,container2,container3","Image":"nginx:latest","Status":"Up 2 hours"}`
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify the parsed result with comma-separated names
	suite.Len(action.Wrapped.ContainerStates, 1)
	container := action.Wrapped.ContainerStates[0]
	suite.Equal("abc123", container.ID)
	suite.Equal([]string{"container1", "container2", "container3"}, container.Names)
	suite.Equal("nginx:latest", container.Image)
	suite.Equal("Up 2 hours", container.Status)
}

func (suite *DockerStatusActionTestSuite) TestAllContainerStates() {
	expectedOutput := `{"ID":"abc123","Names":"running-container","Image":"nginx:latest","Status":"Up 2 hours"}
{"ID":"def456","Names":"exited-container","Image":"redis:alpine","Status":"Exited (0) 1 hour ago"}
{"ID":"ghi789","Names":"paused-container","Image":"postgres:13","Status":"Paused"}
{"ID":"jkl012","Names":"created-container","Image":"alpine:latest","Status":"Created"}
{"ID":"mno345","Names":"restarting-container","Image":"nginx:latest","Status":"Restarting (1) 2 minutes ago"}
{"ID":"pqr678","Names":"dead-container","Image":"alpine:latest","Status":"Dead"}
{"ID":"stu901","Names":"removing-container","Image":"nginx:latest","Status":"Removing"}`

	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify all container states are parsed correctly
	suite.Len(action.Wrapped.ContainerStates, 7)

	// Check running container
	suite.Equal("running-container", action.Wrapped.ContainerStates[0].Names[0])
	suite.Equal("Up 2 hours", action.Wrapped.ContainerStates[0].Status)

	// Check exited container
	suite.Equal("exited-container", action.Wrapped.ContainerStates[1].Names[0])
	suite.Equal("Exited (0) 1 hour ago", action.Wrapped.ContainerStates[1].Status)

	// Check paused container
	suite.Equal("paused-container", action.Wrapped.ContainerStates[2].Names[0])
	suite.Equal("Paused", action.Wrapped.ContainerStates[2].Status)

	// Check created container
	suite.Equal("created-container", action.Wrapped.ContainerStates[3].Names[0])
	suite.Equal("Created", action.Wrapped.ContainerStates[3].Status)

	// Check restarting container
	suite.Equal("restarting-container", action.Wrapped.ContainerStates[4].Names[0])
	suite.Equal("Restarting (1) 2 minutes ago", action.Wrapped.ContainerStates[4].Status)

	// Check dead container
	suite.Equal("dead-container", action.Wrapped.ContainerStates[5].Names[0])
	suite.Equal("Dead", action.Wrapped.ContainerStates[5].Status)

	// Check removing container
	suite.Equal("removing-container", action.Wrapped.ContainerStates[6].Names[0])
	suite.Equal("Removing", action.Wrapped.ContainerStates[6].Status)
}

func (suite *DockerStatusActionTestSuite) TestRealisticDockerOutput() {
	// Test with realistic Docker output that includes various edge cases
	expectedOutput := `{"ID":"a1b2c3d4e5f6","Names":"/web-server","Image":"nginx:1.25-alpine","Status":"Up 3 days"}
{"ID":"f6e5d4c3b2a1","Names":"/db-server,/postgres","Image":"postgres:15-alpine","Status":"Up 1 week"}
{"ID":"123456789abc","Names":"/redis-cache","Image":"redis:7-alpine","Status":"Exited (139) 2 hours ago"}
{"ID":"abcdef123456","Names":"/app-server","Image":"node:18-alpine","Status":"Restarting (2) 5 minutes ago"}
{"ID":"deadbeef1234","Names":"/test-container","Image":"alpine:latest","Status":"Created"}
{"ID":"feedcafe5678","Names":"/temp-container","Image":"busybox:latest","Status":"Dead"}`

	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetAllContainersStateAction(logger)
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json").Return(expectedOutput, nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())

	// Verify realistic output is parsed correctly
	suite.Len(action.Wrapped.ContainerStates, 6)

	// Check web server
	suite.Equal("/web-server", action.Wrapped.ContainerStates[0].Names[0])
	suite.Equal("nginx:1.25-alpine", action.Wrapped.ContainerStates[0].Image)
	suite.Equal("Up 3 days", action.Wrapped.ContainerStates[0].Status)

	// Check database with multiple names
	suite.Equal("/db-server", action.Wrapped.ContainerStates[1].Names[0])
	suite.Equal("/postgres", action.Wrapped.ContainerStates[1].Names[1])
	suite.Equal("postgres:15-alpine", action.Wrapped.ContainerStates[1].Image)
	suite.Equal("Up 1 week", action.Wrapped.ContainerStates[1].Status)

	// Check crashed container
	suite.Equal("/redis-cache", action.Wrapped.ContainerStates[2].Names[0])
	suite.Equal("Exited (139) 2 hours ago", action.Wrapped.ContainerStates[2].Status)

	// Check restarting container
	suite.Equal("/app-server", action.Wrapped.ContainerStates[3].Names[0])
	suite.Equal("Restarting (2) 5 minutes ago", action.Wrapped.ContainerStates[3].Status)

	// Check created container
	suite.Equal("/test-container", action.Wrapped.ContainerStates[4].Names[0])
	suite.Equal("Created", action.Wrapped.ContainerStates[4].Status)

	// Check dead container
	suite.Equal("/temp-container", action.Wrapped.ContainerStates[5].Names[0])
	suite.Equal("Dead", action.Wrapped.ContainerStates[5].Status)
}

func (suite *DockerStatusActionTestSuite) TestNonExistentContainer() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, "non-existent-container")
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	// Docker returns empty output for non-existent containers
	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=non-existent-container").Return("", nil)

	err := action.Execute(context.Background())

	suite.NoError(err)
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Len(action.Wrapped.ContainerStates, 0)
}

func (suite *DockerStatusActionTestSuite) TestContextCancellation() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, "test-container")
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	suite.mockProcessor.On("RunCommandWithContext", ctx, "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return("", context.Canceled)

	err := action.Execute(ctx)

	suite.Error(err)
	suite.Contains(err.Error(), "failed to get container state")
	suite.mockProcessor.AssertExpectations(suite.T())
}

func (suite *DockerStatusActionTestSuite) TestParseContainerOutputWithMissingRequiredFields() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, "test-container")
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	// Output with missing ID field
	outputWithMissingID := `{"Names":"/test-container","Image":"nginx:latest","Status":"Up 2 hours"}
{"ID":"abc123","Names":"/valid-container","Image":"redis:latest","Status":"Up 1 hour"}`

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return(outputWithMissingID, nil)

	err := action.Execute(context.Background())

	suite.NoError(err) // Should still succeed as one valid container was parsed
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Len(action.Wrapped.ContainerStates, 1)
	suite.Equal("abc123", action.Wrapped.ContainerStates[0].ID)
}

func (suite *DockerStatusActionTestSuite) TestParseContainerOutputWithMissingStatusField() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, "test-container")
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	// Output with missing Status field
	outputWithMissingStatus := `{"ID":"abc123","Names":"/test-container","Image":"nginx:latest"}
{"ID":"def456","Names":"/valid-container","Image":"redis:latest","Status":"Up 1 hour"}`

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return(outputWithMissingStatus, nil)

	err := action.Execute(context.Background())

	suite.NoError(err) // Should still succeed as one valid container was parsed
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Len(action.Wrapped.ContainerStates, 1)
	suite.Equal("def456", action.Wrapped.ContainerStates[0].ID)
}

func (suite *DockerStatusActionTestSuite) TestParseContainerOutputWithAllInvalidLines() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, "test-container")
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	// Output with all invalid lines (missing required fields)
	outputWithAllInvalid := `{"Names":"/test-container","Image":"nginx:latest"}
{"Image":"redis:latest","Status":"Up 1 hour"}
{"ID":"abc123","Names":"/another-container"}`

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return(outputWithAllInvalid, nil)

	err := action.Execute(context.Background())

	suite.Error(err) // Should fail as no valid containers were parsed
	suite.Contains(err.Error(), "failed to parse any valid containers")
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Len(action.Wrapped.ContainerStates, 0)
}

func (suite *DockerStatusActionTestSuite) TestParseContainerOutputWithTooManyErrors() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, "test-container")
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	// Output with more than 50% invalid lines (5 invalid, 3 valid = 62.5% error rate)
	outputWithTooManyErrors := `{"ID":"abc123","Names":"/valid1","Image":"nginx","Status":"Up"}
{"Names":"/invalid1","Image":"nginx"}
{"ID":"def456","Names":"/valid2","Image":"redis","Status":"Up"}
{"Image":"invalid2","Status":"Up"}
{"ID":"ghi789","Names":"/valid3","Image":"postgres","Status":"Up"}
{"Names":"/invalid3","Image":"postgres"}
{"Image":"invalid4","Status":"Up"}
{"Image":"invalid5","Status":"Up"}`

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return(outputWithTooManyErrors, nil)

	err := action.Execute(context.Background())

	suite.Error(err) // Should fail as more than 50% of lines had errors
	suite.Contains(err.Error(), "too many parsing errors")
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Len(action.Wrapped.ContainerStates, 0)
}

func (suite *DockerStatusActionTestSuite) TestParseContainerOutputWithMixedValidAndInvalid() {
	logger := command_mock.NewDiscardLogger()
	action := docker.NewGetContainerStateAction(logger, "test-container")
	action.Wrapped.SetCommandProcessor(suite.mockProcessor)

	// Output with some valid and some invalid lines (less than 50% errors)
	outputWithMixed := `{"ID":"abc123","Names":"/valid1","Image":"nginx","Status":"Up"}
{"Names":"/invalid1","Image":"nginx"}
{"ID":"def456","Names":"/valid2","Image":"redis","Status":"Up"}
{"ID":"ghi789","Names":"/valid3","Image":"postgres","Status":"Up"}
{"Image":"invalid2","Status":"Up"}`

	suite.mockProcessor.On("RunCommandWithContext", context.Background(), "docker", "ps", "-a", "--format", "json", "--filter", "name=test-container").Return(outputWithMixed, nil)

	err := action.Execute(context.Background())

	suite.NoError(err) // Should succeed as less than 50% of lines had errors
	suite.mockProcessor.AssertExpectations(suite.T())
	suite.Len(action.Wrapped.ContainerStates, 3) // Only valid containers should be parsed
	suite.Equal("abc123", action.Wrapped.ContainerStates[0].ID)
	suite.Equal("def456", action.Wrapped.ContainerStates[1].ID)
	suite.Equal("ghi789", action.Wrapped.ContainerStates[2].ID)
}

func TestDockerStatusActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerStatusActionTestSuite))
}
