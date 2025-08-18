package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	task_engine "github.com/ndizazzo/task-engine"
	"github.com/ndizazzo/task-engine/actions/common"
	"github.com/ndizazzo/task-engine/testing/mocks"
	"github.com/stretchr/testify/suite"
)

// DockerPsActionTestSuite tests the DockerPsAction
type DockerPsActionTestSuite struct {
	suite.Suite
}

// TestDockerPsActionTestSuite runs the DockerPsAction test suite
func TestDockerPsActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerPsActionTestSuite))
}

func (suite *DockerPsActionTestSuite) TestNewDockerPsAction() {
	logger := slog.Default()

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)

	suite.NotNil(action)
	suite.Equal("docker-ps-action", action.ID)
	suite.False(action.Wrapped.All)
	suite.Empty(action.Wrapped.Filter)
	suite.Empty(action.Wrapped.Format)
	suite.Equal(0, action.Wrapped.Last)
	suite.False(action.Wrapped.Latest)
	suite.False(action.Wrapped.NoTrunc)
	suite.False(action.Wrapped.Quiet)
	suite.False(action.Wrapped.Size)
}

func (suite *DockerPsActionTestSuite) TestNewDockerPsActionWithParameters() {
	logger := slog.Default()

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "status=running"},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: 5},
	)
	suite.Require().NoError(err)

	action.Wrapped.Format = "table {{.Names}}\t{{.Status}}"

	suite.NotNil(action)
	suite.False(action.Wrapped.All)
	suite.Empty(action.Wrapped.Filter)
	suite.Equal("table {{.Names}}\t{{.Status}}", action.Wrapped.Format)
	suite.Equal(0, action.Wrapped.Last)
	suite.False(action.Wrapped.Latest)
	suite.False(action.Wrapped.NoTrunc)
	suite.False(action.Wrapped.Quiet)
	suite.False(action.Wrapped.Size)
	suite.NotNil(action.Wrapped.FilterParam)
	suite.NotNil(action.Wrapped.AllParam)
	suite.NotNil(action.Wrapped.QuietParam)
	suite.NotNil(action.Wrapped.NoTruncParam)
	suite.NotNil(action.Wrapped.SizeParam)
	suite.NotNil(action.Wrapped.LatestParam)
	suite.NotNil(action.Wrapped.LastParam)
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_Success() {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1
def456ghi789   redis     "docker-entrypoint.s"    1 hour ago      Up 1 hour      6379/tcp   myapp_redis_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 2)
	suite.Equal("abc123def456", action.Wrapped.Containers[0].ContainerID)
	suite.Equal("nginx", action.Wrapped.Containers[0].Image)
	suite.Equal("nginx -g 'daemon off", action.Wrapped.Containers[0].Command)
	suite.Equal("2 hours ago", action.Wrapped.Containers[0].Created)
	suite.Equal("Up 2 hours", action.Wrapped.Containers[0].Status)
	suite.Equal("0.0.0.0:8080->80/tcp", action.Wrapped.Containers[0].Ports)
	suite.Equal("myapp_web_1", action.Wrapped.Containers[0].Names)
	suite.Equal("def456ghi789", action.Wrapped.Containers[1].ContainerID)
	suite.Equal("redis", action.Wrapped.Containers[1].Image)
	suite.Equal("docker-entrypoint.s", action.Wrapped.Containers[1].Command)
	suite.Equal("1 hour ago", action.Wrapped.Containers[1].Created)
	suite.Equal("Up 1 hour", action.Wrapped.Containers[1].Status)
	suite.Equal("6379/tcp", action.Wrapped.Containers[1].Ports)
	suite.Equal("myapp_redis_1", action.Wrapped.Containers[1].Names)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WithAll() {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS                     PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours                0.0.0.0:8080->80/tcp   myapp_web_1
def456ghi789   redis     "docker-entrypoint.s"    1 hour ago      Exited (0) 1 hour ago     6379/tcp   myapp_redis_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--all").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 2)
	suite.Equal("Exited (0) 1 hour ago", action.Wrapped.Containers[1].Status)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WithFilter() {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--filter", "status=running").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: "status=running"},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 1)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WithFormat() {
	logger := slog.Default()
	expectedOutput := "myapp_web_1\tUp 2 hours\nmyapp_redis_1\tUp 1 hour"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--format", "{{.Names}}\t{{.Status}}").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.Format = "{{.Names}}\t{{.Status}}"
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WithLast() {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--last", "1").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 1},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 1)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WithLatest() {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--latest").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 1)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WithNoTrunc() {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID                                                                    IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
sha256:abc123def456789012345678901234567890123456789012345678901234567890   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--no-trunc").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 1)
	suite.Equal("sha256:abc123def456789012345678901234567890123456789012345678901234567890", action.Wrapped.Containers[0].ContainerID)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WithQuiet() {
	logger := slog.Default()
	expectedOutput := "abc123def456\ndef456ghi789"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--quiet").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WithSize() {
	logger := slog.Default()
	expectedOutput := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES    SIZE
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1   133MB`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--size").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: true},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 1)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_CommandError() {
	logger := slog.Default()
	expectedError := "docker ps failed"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return("", errors.New(expectedError))

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), expectedError)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.Containers)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_ContextCancellation() {
	logger := slog.Default()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return("", context.Canceled)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.True(errors.Is(err, context.Canceled))
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.Containers)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_parseContainers() {
	logger := slog.Default()
	output := `CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES
abc123def456   nginx     "nginx -g 'daemon off"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1
def456ghi789   redis     "docker-entrypoint.s"    1 hour ago      Up 1 hour      6379/tcp   myapp_redis_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return(output, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(output, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 2)
	suite.Equal("abc123def456", action.Wrapped.Containers[0].ContainerID)
	suite.Equal("nginx", action.Wrapped.Containers[0].Image)
	suite.Equal("nginx -g 'daemon off", action.Wrapped.Containers[0].Command)
	suite.Equal("2 hours ago", action.Wrapped.Containers[0].Created)
	suite.Equal("Up 2 hours", action.Wrapped.Containers[0].Status)
	suite.Equal("0.0.0.0:8080->80/tcp", action.Wrapped.Containers[0].Ports)
	suite.Equal("myapp_web_1", action.Wrapped.Containers[0].Names)
	suite.Equal("def456ghi789", action.Wrapped.Containers[1].ContainerID)
	suite.Equal("redis", action.Wrapped.Containers[1].Image)
	suite.Equal("docker-entrypoint.s", action.Wrapped.Containers[1].Command)
	suite.Equal("1 hour ago", action.Wrapped.Containers[1].Created)
	suite.Equal("Up 1 hour", action.Wrapped.Containers[1].Status)
	suite.Equal("6379/tcp", action.Wrapped.Containers[1].Ports)
	suite.Equal("myapp_redis_1", action.Wrapped.Containers[1].Names)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_parseContainerLine() {
	action := &DockerPsAction{}
	line := "abc123def456   nginx     \"nginx -g 'daemon off\"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1"
	container := action.parseContainerLine(line)

	suite.Equal("abc123def456", container.ContainerID)
	suite.Equal("nginx", container.Image)
	suite.Equal("nginx -g 'daemon off", container.Command)
	suite.Equal("2 hours ago", container.Created)
	suite.Equal("Up 2 hours", container.Status)
	suite.Equal("0.0.0.0:8080->80/tcp", container.Ports)
	suite.Equal("myapp_web_1", container.Names)
	line = "def456ghi789   redis     \"docker-entrypoint.s\"    1 hour ago      Exited (0) 1 hour ago     6379/tcp   myapp_redis_1"
	container = action.parseContainerLine(line)

	suite.Equal("def456ghi789", container.ContainerID)
	suite.Equal("redis", container.Image)
	suite.Equal("docker-entrypoint.s", container.Command)
	suite.Equal("1 hour ago", container.Created)
	suite.Equal("Exited (0) 1 hour ago", container.Status)
	suite.Equal("6379/tcp", container.Ports)
	suite.Equal("myapp_redis_1", container.Names)
	line = "ghi789jkl012   postgres  \"postgres\"               3 hours ago     Up 3 hours                 myapp_db_1"
	container = action.parseContainerLine(line)

	suite.Equal("ghi789jkl012", container.ContainerID)
	suite.Equal("postgres", container.Image)
	suite.Equal("postgres", container.Command)
	suite.Equal("3 hours ago", container.Created)
	suite.Equal("Up 3 hours", container.Status)
	suite.Equal("", container.Ports)
	suite.Equal("myapp_db_1", container.Names)
	line = "jkl012mno345   alpine    \"sh\"                     4 hours ago     Up 4 hours                 myapp_alpine_1,alpine"
	container = action.parseContainerLine(line)

	suite.Equal("jkl012mno345", container.ContainerID)
	suite.Equal("alpine", container.Image)
	suite.Equal("sh", container.Command)
	suite.Equal("4 hours ago", container.Created)
	suite.Equal("Up 4 hours", container.Status)
	suite.Equal("", container.Ports)
	suite.Equal("myapp_alpine_1,alpine", container.Names)
	line = "mno345pqr678   ubuntu    \"bash -c 'echo hello'\"   5 hours ago     Up 5 hours                 myapp_ubuntu_1"
	container = action.parseContainerLine(line)

	suite.Equal("mno345pqr678", container.ContainerID)
	suite.Equal("ubuntu", container.Image)
	suite.Equal("bash -c 'echo hello'", container.Command)
	suite.Equal("5 hours ago", container.Created)
	suite.Equal("Up 5 hours", container.Status)
	suite.Equal("", container.Ports)
	suite.Equal("myapp_ubuntu_1", container.Names)
	line = "pqr678stu901   nginx     \"nginx -g 'daemon off\"   6 hours ago     Up 6 hours     0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp   myapp_nginx_1"
	container = action.parseContainerLine(line)

	suite.Equal("pqr678stu901", container.ContainerID)
	suite.Equal("nginx", container.Image)
	suite.Equal("nginx -g 'daemon off", container.Command)
	suite.Equal("6 hours ago", container.Created)
	suite.Equal("Up 6 hours", container.Status)
	suite.Equal("0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp", container.Ports)
	suite.Equal("myapp_nginx_1", container.Names)
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_EmptyOutput() {
	logger := slog.Default()
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return(expectedOutput, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Empty(action.Wrapped.Containers)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_WhitespaceOnlyOutput() {
	logger := slog.Default()
	output := "  \n  \n"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return(output, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(output, action.Wrapped.Output)
	suite.Empty(action.Wrapped.Containers)
	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_GetOutput() {
	action := &DockerPsAction{
		BaseAction:        task_engine.NewBaseAction(slog.Default()),
		ParameterResolver: *common.NewParameterResolver(slog.Default()),
		OutputBuilder:     *common.NewOutputBuilder(slog.Default()),
		Output:            "raw output",
		Containers:        []Container{{ContainerID: "abc123", Image: "nginx:latest"}},
	}

	out := action.GetOutput()
	suite.IsType(map[string]interface{}{}, out)
	m := out.(map[string]interface{})
	suite.Equal(1, m["count"])
	suite.Equal("raw output", m["rawOutput"])
	suite.Equal(true, m["success"])
	suite.Len(m["containers"], 1)
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_WithOptionMethods() {
	logger := slog.Default()
	expected := `CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS   PORTS   NAMES`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps", "--all", "--filter", "status=running", "--format", "{{.Names}}", "--last", "2", "--latest", "--no-trunc", "--quiet", "--size").Return(expected, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	suite.Require().NoError(err)

	// Apply option methods directly to cover them
	WithPsAll()(action.Wrapped)
	WithPsFilter("status=running")(action.Wrapped)
	WithPsFormat("{{.Names}}")(action.Wrapped)
	WithPsLast(2)(action.Wrapped)
	WithPsLatest()(action.Wrapped)
	WithPsNoTrunc()(action.Wrapped)
	WithPsQuiet()(action.Wrapped)
	WithPsSize()(action.Wrapped)

	action.Wrapped.SetCommandRunner(mockRunner)
	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expected, action.Wrapped.Output)
	suite.True(action.Wrapped.All)
	suite.Equal("status=running", action.Wrapped.Filter)
	suite.Equal("{{.Names}}", action.Wrapped.Format)
	suite.Equal(2, action.Wrapped.Last)
	suite.True(action.Wrapped.Latest)
	suite.True(action.Wrapped.NoTrunc)
	suite.True(action.Wrapped.Quiet)
	suite.True(action.Wrapped.Size)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerPsActionTestSuite) TestDockerPsAction_Execute_OutputWithTrailingWhitespace() {
	logger := slog.Default()
	output := "CONTAINER ID   IMAGE     COMMAND                  CREATED         STATUS         PORTS     NAMES\nabc123def456   nginx     \"nginx -g 'daemon off\"   2 hours ago     Up 2 hours     0.0.0.0:8080->80/tcp   myapp_web_1\n  \n"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "ps").Return(output, nil)

	action, err := NewDockerPsAction(logger).WithParameters(
		task_engine.StaticParameter{Value: ""},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: false},
		task_engine.StaticParameter{Value: 0},
	)
	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(output, action.Wrapped.Output)
	suite.Len(action.Wrapped.Containers, 1)
	suite.Equal("abc123def456", action.Wrapped.Containers[0].ContainerID)
	mockRunner.AssertExpectations(suite.T())
}
