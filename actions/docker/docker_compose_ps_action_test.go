package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

// DockerComposePsActionTestSuite tests the DockerComposePsAction
type DockerComposePsActionTestSuite struct {
	suite.Suite
}

// TestDockerComposePsActionTestSuite runs the DockerComposePsAction test suite
func TestDockerComposePsActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposePsActionTestSuite))
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsAction() {
	logger := slog.Default()
	services := []string{"web", "db"}

	action := NewDockerComposePsAction(logger, services)

	suite.NotNil(action)
	suite.Equal("docker-compose-ps-action", action.ID)
	suite.Equal(services, action.Wrapped.Services)
	suite.False(action.Wrapped.All)
	suite.Empty(action.Wrapped.Filter)
	suite.Empty(action.Wrapped.Format)
	suite.False(action.Wrapped.Quiet)
	suite.Empty(action.Wrapped.WorkingDir)
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionWithOptions() {
	logger := slog.Default()
	services := []string{"web"}

	action := NewDockerComposePsAction(logger, services,
		WithComposePsAll(),
		WithComposePsFilter("status=running"),
		WithComposePsFormat("table {{.Name}}\t{{.Status}}"),
		WithComposePsQuiet(),
		WithComposePsWorkingDir("/path/to/compose"),
	)

	suite.NotNil(action)
	suite.Equal(services, action.Wrapped.Services)
	suite.True(action.Wrapped.All)
	suite.Equal("status=running", action.Wrapped.Filter)
	suite.Equal("table {{.Name}}\t{{.Status}}", action.Wrapped.Format)
	suite.True(action.Wrapped.Quiet)
	suite.Equal("/path/to/compose", action.Wrapped.WorkingDir)
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_Success() {
	logger := slog.Default()
	services := []string{"web", "db"}
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
myapp_db_1          postgres:13         "docker-entrypoint.s"    db                  2 hours ago         Up 2 hours         5432/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web", "db").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 2)

	// Check first service
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	suite.Equal("nginx:latest", action.Wrapped.ServicesList[0].Image)
	suite.Equal("web", action.Wrapped.ServicesList[0].ServiceName)
	suite.Equal("Up 2 hours", action.Wrapped.ServicesList[0].Status)
	suite.Equal("0.0.0.0:8080->80/tcp", action.Wrapped.ServicesList[0].Ports)

	// Check second service
	suite.Equal("myapp_db_1", action.Wrapped.ServicesList[1].Name)
	suite.Equal("postgres:13", action.Wrapped.ServicesList[1].Image)
	suite.Equal("db", action.Wrapped.ServicesList[1].ServiceName)
	suite.Equal("Up 2 hours", action.Wrapped.ServicesList[1].Status)
	suite.Equal("5432/tcp", action.Wrapped.ServicesList[1].Ports)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_NoServices() {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, []string{})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 1)
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_WithAll() {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--all", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services, WithComposePsAll())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 1)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_WithFilter() {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--filter", "status=running", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services, WithComposePsFilter("status=running"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 1)
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_WithFormat() {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--format", "table {{.Name}}\t{{.Status}}", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services, WithComposePsFormat("table {{.Name}}\t{{.Status}}"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 1)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_WithQuiet() {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `myapp_web_1
myapp_db_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--quiet", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services, WithComposePsQuiet())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 2)
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	suite.Equal("myapp_db_1", action.Wrapped.ServicesList[1].Name)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_CommandError() {
	logger := slog.Default()
	services := []string{"web"}
	expectedError := errors.New("docker compose ps command failed")

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web").Return("", expectedError)

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), "docker compose ps command failed", "Error should contain the command failure message")
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.ServicesList)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_ContextCancellation() {
	logger := slog.Default()
	services := []string{"web"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web").Return("", context.Canceled)

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(ctx)

	suite.Error(err)
	suite.Contains(err.Error(), "context canceled", "Error should contain the context cancellation message")
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.ServicesList)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_parseServices() {
	logger := slog.Default()
	output := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
myapp_db_1          postgres:13         "docker-entrypoint.s"    db                  2 hours ago         Up 2 hours         5432/tcp`

	action := NewDockerComposePsAction(logger, []string{})
	action.Wrapped.Output = output
	action.Wrapped.parseServices(output)

	suite.Len(action.Wrapped.ServicesList, 2)

	// Check first service
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	suite.Equal("nginx:latest", action.Wrapped.ServicesList[0].Image)
	suite.Equal("web", action.Wrapped.ServicesList[0].ServiceName)
	suite.Equal("Up 2 hours", action.Wrapped.ServicesList[0].Status)
	suite.Equal("0.0.0.0:8080->80/tcp", action.Wrapped.ServicesList[0].Ports)

	// Check second service
	suite.Equal("myapp_db_1", action.Wrapped.ServicesList[1].Name)
	suite.Equal("postgres:13", action.Wrapped.ServicesList[1].Image)
	suite.Equal("db", action.Wrapped.ServicesList[1].ServiceName)
	suite.Equal("Up 2 hours", action.Wrapped.ServicesList[1].Status)
	suite.Equal("5432/tcp", action.Wrapped.ServicesList[1].Ports)
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_parseServiceLine() {
	logger := slog.Default()
	action := NewDockerComposePsAction(logger, []string{})

	// Test normal line
	line := "myapp_web_1         nginx:latest        \"nginx -g 'daemon off\"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp"
	service := action.Wrapped.parseServiceLine(line)

	suite.Equal("myapp_web_1", service.Name)
	suite.Equal("nginx:latest", service.Image)
	suite.Equal("web", service.ServiceName)
	suite.Equal("Up 2 hours", service.Status)
	suite.Equal("0.0.0.0:8080->80/tcp", service.Ports)

	// Test line with different status
	line = "myapp_db_1          postgres:13         \"docker-entrypoint.s\"    db                  2 hours ago         Exited (0) 2 hours ago         5432/tcp"
	service = action.Wrapped.parseServiceLine(line)

	suite.Equal("myapp_db_1", service.Name)
	suite.Equal("postgres:13", service.Image)
	suite.Equal("db", service.ServiceName)
	suite.Equal("Exited (0) 2 hours ago", service.Status)
	suite.Equal("5432/tcp", service.Ports)

	// Test line with extra whitespace
	line = "  myapp_web_1    nginx:latest    \"nginx -g 'daemon off\"    web   2 hours ago    Up 2 hours   0.0.0.0:8080->80/tcp  "
	service = action.Wrapped.parseServiceLine(line)

	suite.Equal("myapp_web_1", service.Name)
	suite.Equal("nginx:latest", service.Image)
	suite.Equal("web", service.ServiceName)
	suite.Equal("Up 2 hours", service.Status)
	suite.Equal("0.0.0.0:8080->80/tcp", service.Ports)

	// Test line with tab separators
	line = "myapp_web_1\tnginx:latest\t\"nginx -g 'daemon off\"\tweb\t2 hours ago\tUp 2 hours\t0.0.0.0:8080->80/tcp"
	service = action.Wrapped.parseServiceLine(line)

	suite.Equal("myapp_web_1", service.Name)
	suite.Equal("nginx:latest", service.Image)
	suite.Equal("web", service.ServiceName)
	suite.Equal("Up 2 hours", service.Status)
	suite.Equal("0.0.0.0:8080->80/tcp", service.Ports)

	// Test line with mixed separators
	line = "myapp_web_1\t nginx:latest \t\"nginx -g 'daemon off\"\t web \t2 hours ago\t Up 2 hours \t0.0.0.0:8080->80/tcp"
	service = action.Wrapped.parseServiceLine(line)

	suite.Equal("myapp_web_1", service.Name)
	suite.Equal("nginx:latest", service.Image)
	suite.Equal("web", service.ServiceName)
	suite.Equal("Up 2 hours", service.Status)
	suite.Equal("0.0.0.0:8080->80/tcp", service.Ports)
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_EmptyOutput() {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Empty(action.Wrapped.ServicesList)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_Execute_OutputWithTrailingWhitespace() {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 1)
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	mockRunner.AssertExpectations(suite.T())
}
