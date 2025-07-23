package docker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/ndizazzo/task-engine/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerComposePsAction(t *testing.T) {
	logger := slog.Default()
	services := []string{"web", "db"}

	action := NewDockerComposePsAction(logger, services)

	assert.NotNil(t, action)
	assert.Equal(t, "docker-compose-ps-action", action.ID)
	assert.Equal(t, services, action.Wrapped.Services)
	assert.False(t, action.Wrapped.All)
	assert.Empty(t, action.Wrapped.Filter)
	assert.Empty(t, action.Wrapped.Format)
	assert.False(t, action.Wrapped.Quiet)
	assert.Empty(t, action.Wrapped.WorkingDir)
}

func TestNewDockerComposePsActionWithOptions(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}

	action := NewDockerComposePsAction(logger, services,
		WithComposePsAll(),
		WithComposePsFilter("status=running"),
		WithComposePsFormat("table {{.Name}}\t{{.Status}}"),
		WithComposePsQuiet(),
		WithComposePsWorkingDir("/path/to/compose"),
	)

	assert.NotNil(t, action)
	assert.Equal(t, services, action.Wrapped.Services)
	assert.True(t, action.Wrapped.All)
	assert.Equal(t, "status=running", action.Wrapped.Filter)
	assert.Equal(t, "table {{.Name}}\t{{.Status}}", action.Wrapped.Format)
	assert.True(t, action.Wrapped.Quiet)
	assert.Equal(t, "/path/to/compose", action.Wrapped.WorkingDir)
}

func TestDockerComposePsAction_Execute_Success(t *testing.T) {
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

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.ServicesList, 2)

	// Check first service
	assert.Equal(t, "myapp_web_1", action.Wrapped.ServicesList[0].Name)
	assert.Equal(t, "nginx:latest", action.Wrapped.ServicesList[0].Image)
	assert.Equal(t, "web", action.Wrapped.ServicesList[0].ServiceName)
	assert.Equal(t, "Up 2 hours", action.Wrapped.ServicesList[0].Status)
	assert.Equal(t, "0.0.0.0:8080->80/tcp", action.Wrapped.ServicesList[0].Ports)

	// Check second service
	assert.Equal(t, "myapp_db_1", action.Wrapped.ServicesList[1].Name)
	assert.Equal(t, "postgres:13", action.Wrapped.ServicesList[1].Image)
	assert.Equal(t, "db", action.Wrapped.ServicesList[1].ServiceName)
	assert.Equal(t, "Up 2 hours", action.Wrapped.ServicesList[1].Status)
	assert.Equal(t, "5432/tcp", action.Wrapped.ServicesList[1].Ports)

	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_Execute_NoServices(t *testing.T) {
	logger := slog.Default()
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, []string{})
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.ServicesList, 1)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_Execute_WithAll(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
myapp_web_2         nginx:latest        "nginx -g 'daemon off"   web                 1 hour ago          Exited (0) 1 hour ago`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--all", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services, WithComposePsAll())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.ServicesList, 2)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_Execute_WithFilter(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--filter", "status=running", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services, WithComposePsFilter("status=running"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.ServicesList, 1)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_Execute_WithFormat(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `myapp_web_1:Up 2 hours
myapp_db_1:Up 2 hours`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--format", "{{.Name}}:{{.Status}}", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services, WithComposePsFormat("{{.Name}}:{{.Status}}"))
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	// With custom format, we don't parse the output into structured data
	assert.Empty(t, action.Wrapped.ServicesList)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_Execute_WithQuiet(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := `myapp_web_1
myapp_db_1`

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "--quiet", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services, WithComposePsQuiet())
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	// With quiet mode, we don't parse the output into structured data
	assert.Empty(t, action.Wrapped.ServicesList)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_Execute_CommandError(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}
	expectedError := "permission denied"

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web").Return("", errors.New(expectedError))

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
	assert.Empty(t, action.Wrapped.ServicesList)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_Execute_ContextCancellation(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web").Return("", context.Canceled)

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_parseServices(t *testing.T) {
	tests := []struct {
		name             string
		output           string
		expectedServices []ComposeService
	}{
		{
			name:             "empty output",
			output:           "",
			expectedServices: []ComposeService(nil),
		},
		{
			name:             "only header",
			output:           "NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS",
			expectedServices: []ComposeService(nil),
		},
		{
			name: "single service",
			output: `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp`,
			expectedServices: []ComposeService{
				{
					Name:        "myapp_web_1",
					Image:       "nginx:latest",
					ServiceName: "web",
					Status:      "Up 2 hours",
					Ports:       "0.0.0.0:8080->80/tcp",
				},
			},
		},
		{
			name: "multiple services",
			output: `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_web_1         nginx:latest        "nginx -g 'daemon off"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp
myapp_db_1          postgres:13         "docker-entrypoint.s"    db                  2 hours ago         Up 2 hours         5432/tcp`,
			expectedServices: []ComposeService{
				{
					Name:        "myapp_web_1",
					Image:       "nginx:latest",
					ServiceName: "web",
					Status:      "Up 2 hours",
					Ports:       "0.0.0.0:8080->80/tcp",
				},
				{
					Name:        "myapp_db_1",
					Image:       "postgres:13",
					ServiceName: "db",
					Status:      "Up 2 hours",
					Ports:       "5432/tcp",
				},
			},
		},
		{
			name: "service without ports",
			output: `NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
myapp_worker_1      redis:alpine        "docker-entrypoint.s"    worker              1 hour ago          Up 1 hour`,
			expectedServices: []ComposeService{
				{
					Name:        "myapp_worker_1",
					Image:       "redis:alpine",
					ServiceName: "worker",
					Status:      "Up 1 hour",
					Ports:       "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			action := NewDockerComposePsAction(logger, []string{})

			action.Wrapped.parseServices(tt.output)

			assert.Equal(t, tt.expectedServices, action.Wrapped.ServicesList)
		})
	}
}

func TestDockerComposePsAction_parseServiceLine(t *testing.T) {
	tests := []struct {
		name            string
		line            string
		expectedService *ComposeService
	}{
		{
			name: "valid service line",
			line: "myapp_web_1         nginx:latest        \"nginx -g 'daemon off\"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp",
			expectedService: &ComposeService{
				Name:        "myapp_web_1",
				Image:       "nginx:latest",
				ServiceName: "web",
				Status:      "Up 2 hours",
				Ports:       "0.0.0.0:8080->80/tcp",
			},
		},
		{
			name: "service without ports",
			line: "myapp_worker_1      redis:alpine        \"docker-entrypoint.s\"    worker              1 hour ago          Up 1 hour",
			expectedService: &ComposeService{
				Name:        "myapp_worker_1",
				Image:       "redis:alpine",
				ServiceName: "worker",
				Status:      "Up 1 hour",
				Ports:       "",
			},
		},
		{
			name: "service with complex command",
			line: "myapp_app_1         node:16             \"node /app/server.js\"    app                 30 minutes ago      Up 30 minutes       3000/tcp",
			expectedService: &ComposeService{
				Name:        "myapp_app_1",
				Image:       "node:16",
				ServiceName: "app",
				Status:      "Up 30 minutes",
				Ports:       "3000/tcp",
			},
		},
		{
			name:            "insufficient parts",
			line:            "myapp_web_1 nginx:latest",
			expectedService: nil,
		},
		{
			name:            "empty line",
			line:            "",
			expectedService: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.Default()
			action := NewDockerComposePsAction(logger, []string{})

			result := action.Wrapped.parseServiceLine(tt.line)

			assert.Equal(t, tt.expectedService, result)
		})
	}
}

func TestDockerComposePsAction_Execute_EmptyOutput(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}
	expectedOutput := ""

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web").Return(expectedOutput, nil)

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, action.Wrapped.Output)
	assert.Empty(t, action.Wrapped.ServicesList)
	mockRunner.AssertExpectations(t)
}

func TestDockerComposePsAction_Execute_OutputWithTrailingWhitespace(t *testing.T) {
	logger := slog.Default()
	services := []string{"web"}
	rawOutput := "NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS\nmyapp_web_1         nginx:latest        \"nginx -g 'daemon off\"   web                 2 hours ago         Up 2 hours         0.0.0.0:8080->80/tcp\n  \n  "

	mockRunner := &mocks.MockCommandRunner{}
	mockRunner.On("RunCommand", "docker", "compose", "ps", "web").Return(rawOutput, nil)

	action := NewDockerComposePsAction(logger, services)
	action.Wrapped.SetCommandRunner(mockRunner)

	err := action.Wrapped.Execute(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, rawOutput, action.Wrapped.Output)
	assert.Len(t, action.Wrapped.ServicesList, 1)
	assert.Equal(t, "myapp_web_1", action.Wrapped.ServicesList[0].Name)
	mockRunner.AssertExpectations(t)
}
