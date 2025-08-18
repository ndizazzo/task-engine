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

// DockerComposePsActionTestSuite tests the DockerComposePsAction
type DockerComposePsActionTestSuite struct {
	suite.Suite
}

// TestDockerComposePsActionTestSuite runs the DockerComposePsAction test suite
func TestDockerComposePsActionTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposePsActionTestSuite))
}

// Tests for new constructor pattern with parameters
func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_WithParameters() {
	logger := slog.Default()

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}}, // services
		task_engine.StaticParameter{Value: false},      // all
		task_engine.StaticParameter{Value: ""},         // filter
		task_engine.StaticParameter{Value: ""},         // format
		task_engine.StaticParameter{Value: false},      // quiet
		task_engine.StaticParameter{Value: ""},         // workingDir
	)

	suite.Require().NoError(err)
	suite.NotNil(action)
	suite.Equal("docker-compose-ps-action", action.ID)
	suite.NotNil(action.Wrapped)
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_WithParameters() {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
myapp_db_1          postgres:13         "docker-entrypoint.s"    db                  2 hours ago         Up 2 hours         5432/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web", "db").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"web", "db"}}, // services
		task_engine.StaticParameter{Value: false},                 // all
		task_engine.StaticParameter{Value: ""},                    // filter
		task_engine.StaticParameter{Value: ""},                    // format
		task_engine.StaticParameter{Value: false},                 // quiet
		task_engine.StaticParameter{Value: ""},                    // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 2)
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	suite.Equal("nginx:latest", action.Wrapped.ServicesList[0].Image)
	suite.Equal("web", action.Wrapped.ServicesList[0].ServiceName)
	suite.Equal("Up 2 hours", action.Wrapped.ServicesList[0].Status)
	suite.Equal("0.0.0.0:8080->80/tcp", action.Wrapped.ServicesList[0].Ports)
	suite.Equal("myapp_db_1", action.Wrapped.ServicesList[1].Name)
	suite.Equal("postgres:13", action.Wrapped.ServicesList[1].Image)
	suite.Equal("db", action.Wrapped.ServicesList[1].ServiceName)
	suite.Equal("Up 2 hours", action.Wrapped.ServicesList[1].Status)
	suite.Equal("5432/tcp", action.Wrapped.ServicesList[1].Ports)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_WithAllParameter() {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
myapp_stopped_1     nginx:alpine        "nginx -g 'daemon off"   stopped             3 hours ago         Exited (0) 1 hour ago`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--all").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}}, // services (empty)
		task_engine.StaticParameter{Value: true},       // all = true
		task_engine.StaticParameter{Value: ""},         // filter
		task_engine.StaticParameter{Value: ""},         // format
		task_engine.StaticParameter{Value: false},      // quiet
		task_engine.StaticParameter{Value: ""},         // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Len(action.Wrapped.ServicesList, 2)
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	suite.Equal("myapp_stopped_1", action.Wrapped.ServicesList[1].Name)
	suite.Equal("Exited (0) 1 hour ago", action.Wrapped.ServicesList[1].Status)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_WithFilterParameter() {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--filter", "status=running").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}},       // services
		task_engine.StaticParameter{Value: false},            // all
		task_engine.StaticParameter{Value: "status=running"}, // filter
		task_engine.StaticParameter{Value: ""},               // format
		task_engine.StaticParameter{Value: false},            // quiet
		task_engine.StaticParameter{Value: ""},               // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal("status=running", action.Wrapped.Filter)
	suite.Len(action.Wrapped.ServicesList, 1)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_WithFormatParameter() {
	logger := slog.Default()
	expectedOutput := `myapp_web_1	Up 2 hours
myapp_db_1	Up 2 hours`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--format", "table {{.Name}}\t{{.Status}}").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}},                     // services
		task_engine.StaticParameter{Value: false},                          // all
		task_engine.StaticParameter{Value: ""},                             // filter
		task_engine.StaticParameter{Value: "table {{.Name}}\t{{.Status}}"}, // format
		task_engine.StaticParameter{Value: false},                          // quiet
		task_engine.StaticParameter{Value: ""},                             // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal("table {{.Name}}\t{{.Status}}", action.Wrapped.Format)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_WithQuietParameter() {
	logger := slog.Default()
	expectedOutput := `myapp_web_1
myapp_db_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--quiet").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}}, // services
		task_engine.StaticParameter{Value: false},      // all
		task_engine.StaticParameter{Value: ""},         // filter
		task_engine.StaticParameter{Value: ""},         // format
		task_engine.StaticParameter{Value: true},       // quiet = true
		task_engine.StaticParameter{Value: ""},         // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.True(action.Wrapped.Quiet)
	suite.Len(action.Wrapped.ServicesList, 2)
	suite.Equal("myapp_web_1", action.Wrapped.ServicesList[0].Name)
	suite.Equal("myapp_db_1", action.Wrapped.ServicesList[1].Name)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_WithWorkingDirParameter() {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}},         // services
		task_engine.StaticParameter{Value: false},              // all
		task_engine.StaticParameter{Value: ""},                 // filter
		task_engine.StaticParameter{Value: ""},                 // format
		task_engine.StaticParameter{Value: false},              // quiet
		task_engine.StaticParameter{Value: "/path/to/compose"}, // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.Equal("/path/to/compose", action.Wrapped.WorkingDir)
	suite.Len(action.Wrapped.ServicesList, 1)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_WithMultipleParameters() {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
myapp_stopped_1     nginx:alpine        "nginx -g 'daemon off"   stopped             3 hours ago         Exited (0) 1 hour ago`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--all", "--filter", "status=exited", "--format", "table {{.Name}}\t{{.Status}}", "web").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{"web"}},                // services
		task_engine.StaticParameter{Value: true},                           // all = true
		task_engine.StaticParameter{Value: "status=exited"},                // filter
		task_engine.StaticParameter{Value: "table {{.Name}}\t{{.Status}}"}, // format
		task_engine.StaticParameter{Value: false},                          // quiet
		task_engine.StaticParameter{Value: "/path/to/compose"},             // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal(expectedOutput, action.Wrapped.Output)
	suite.True(action.Wrapped.All)
	suite.Equal("status=exited", action.Wrapped.Filter)
	suite.Equal("table {{.Name}}\t{{.Status}}", action.Wrapped.Format)
	suite.False(action.Wrapped.Quiet)
	suite.Equal("/path/to/compose", action.Wrapped.WorkingDir)
	suite.Equal([]string{"web"}, action.Wrapped.Services)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_InvalidParameterTypes() {
	logger := slog.Default()

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}}, // services
		task_engine.StaticParameter{Value: "invalid"},  // all should be bool, not string
		task_engine.StaticParameter{Value: ""},         // filter
		task_engine.StaticParameter{Value: ""},         // format
		task_engine.StaticParameter{Value: false},      // quiet
		task_engine.StaticParameter{Value: ""},         // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "all parameter resolved to non-boolean value")
	action, err = constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}}, // services
		task_engine.StaticParameter{Value: false},      // all
		task_engine.StaticParameter{Value: ""},         // filter
		task_engine.StaticParameter{Value: ""},         // format
		task_engine.StaticParameter{Value: "invalid"},  // quiet should be bool, not string
		task_engine.StaticParameter{Value: ""},         // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(&mocks.MockCommandRunner{})

	err = action.Wrapped.Execute(context.Background())
	suite.Error(err)
	suite.Contains(err.Error(), "quiet parameter resolved to non-boolean value")
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_ServicesAsString() {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web", "db").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: "web,db"}, // services as comma-separated string
		task_engine.StaticParameter{Value: false},    // all
		task_engine.StaticParameter{Value: ""},       // filter
		task_engine.StaticParameter{Value: ""},       // format
		task_engine.StaticParameter{Value: false},    // quiet
		task_engine.StaticParameter{Value: ""},       // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal([]string{"web", "db"}, action.Wrapped.Services)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_ServicesAsSpaceSeparatedString() {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web", "db").Return(expectedOutput, nil)

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: "web db"}, // services as space-separated string
		task_engine.StaticParameter{Value: false},    // all
		task_engine.StaticParameter{Value: ""},       // filter
		task_engine.StaticParameter{Value: ""},       // format
		task_engine.StaticParameter{Value: false},    // quiet
		task_engine.StaticParameter{Value: ""},       // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.NoError(err)
	suite.Equal([]string{"web", "db"}, action.Wrapped.Services)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestNewDockerComposePsActionConstructor_Execute_CommandFailure() {
	logger := slog.Default()
	expectedError := "docker compose ps failed"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps").Return("", errors.New(expectedError))

	constructor := NewDockerComposePsAction(logger)
	action, err := constructor.WithParameters(
		task_engine.StaticParameter{Value: []string{}}, // services
		task_engine.StaticParameter{Value: false},      // all
		task_engine.StaticParameter{Value: ""},         // filter
		task_engine.StaticParameter{Value: ""},         // format
		task_engine.StaticParameter{Value: false},      // quiet
		task_engine.StaticParameter{Value: ""},         // workingDir
	)

	suite.Require().NoError(err)
	action.Wrapped.SetCommandRunner(mockRunner)

	err = action.Wrapped.Execute(context.Background())

	suite.Error(err)
	suite.Contains(err.Error(), expectedError)
	suite.Empty(action.Wrapped.Output)
	suite.Empty(action.Wrapped.ServicesList)

	mockRunner.AssertExpectations(suite.T())
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_parseServiceLine() {
	action := &DockerComposePsAction{}
	line := `myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`
	service := action.parseServiceLine(line)

	suite.Equal("myapp_web_1", service.Name)
	suite.Equal("nginx:latest", service.Image)
	suite.Equal("web", service.ServiceName)
	suite.Equal("Up 2 hours", service.Status)
	suite.Equal("0.0.0.0:8080->80/tcp", service.Ports)
	line = `myapp_db_1          postgres:13         "docker-entrypoint.s"    db                  2 hours ago         Exited (0) 2 hours ago         5432/tcp`
	service = action.parseServiceLine(line)

	suite.Equal("myapp_db_1", service.Name)
	suite.Equal("postgres:13", service.Image)
	suite.Equal("db", service.ServiceName)
	suite.Equal("Exited (0) 2 hours ago", service.Status)
	suite.Equal("5432/tcp", service.Ports)
	action.Quiet = true
	line = "myapp_web_1"
	service = action.parseServiceLine(line)

	suite.Equal("myapp_web_1", service.Name)
	suite.Equal("", service.Image)
	suite.Equal("", service.ServiceName)
	suite.Equal("", service.Status)
	suite.Equal("", service.Ports)
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_parseServices() {
	action := &DockerComposePsAction{}

	output := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
myapp_db_1          postgres:13         "docker-entrypoint.s"    db                  2 hours ago         Up 2 hours         5432/tcp`

	action.parseServices(output)

	suite.Len(action.ServicesList, 2)
	suite.Equal("myapp_web_1", action.ServicesList[0].Name)
	suite.Equal("nginx:latest", action.ServicesList[0].Image)
	suite.Equal("web", action.ServicesList[0].ServiceName)
	suite.Equal("Up 2 hours", action.ServicesList[0].Status)
	suite.Equal("0.0.0.0:8080->80/tcp", action.ServicesList[0].Ports)
	suite.Equal("myapp_db_1", action.ServicesList[1].Name)
	suite.Equal("postgres:13", action.ServicesList[1].Image)
	suite.Equal("db", action.ServicesList[1].ServiceName)
	suite.Equal("Up 2 hours", action.ServicesList[1].Status)
	suite.Equal("5432/tcp", action.ServicesList[1].Ports)
}

func (suite *DockerComposePsActionTestSuite) TestDockerComposePsAction_GetOutput() {
	action := &DockerComposePsAction{
		Output: "test output",
		ServicesList: []ComposeService{
			{Name: "web", ServiceName: "web", Status: "Up"},
			{Name: "db", ServiceName: "db", Status: "Up"},
		},
	}

	output := action.GetOutput()

	suite.IsType(map[string]interface{}{}, output)
	outputMap := output.(map[string]interface{})

	suite.Equal(2, outputMap["count"])
	suite.Equal("test output", outputMap["output"])
	suite.Equal(true, outputMap["success"])
	suite.Len(outputMap["services"], 2)
}
